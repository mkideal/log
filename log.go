package log

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mkideal/log/logger"
	"github.com/mkideal/log/provider"
)

const (
	LvFATAL = logger.FATAL
	LvERROR = logger.ERROR
	LvWARN  = logger.WARN
	LvINFO  = logger.INFO
	LvDEBUG = logger.DEBUG
	LvTRACE = logger.TRACE
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// ParseLevel parses log level from string
func ParseLevel(s string) (logger.Level, bool) { return logger.ParseLevel(s) }

// MustParseLevel is similar to ParseLevel, but panics if parse failed
func MustParseLevel(s string) logger.Level { return logger.MustParseLevel(s) }

// global logger
var glogger = logger.NewStdLogger()

// Uninit uninits log package
func Uninit(err error) {
	glogger.Quit()
}

// InitWithLogger inits global logger with a specified logger
func InitWithLogger(l logger.Logger) error {
	glogger.Quit()
	glogger = l
	glogger.Run()
	return nil
}

// InitWithProvider inits global logger(sync) with a specified provider
func InitWithProvider(p logger.Provider) error {
	l := logger.New(p)
	l.SetLevel(LvINFO)
	return InitWithLogger(l)
}

// InitSyncWithProvider inits global logger(async) with a specified provider
func InitSyncWithProvider(p logger.Provider) error {
	l := logger.NewSync(p)
	l.SetLevel(LvINFO)
	return InitWithLogger(l)
}

// Init inits global logger with providerType and opts (opts is a json string or empty)
func Init(providerTypes string, opts interface{}) error {
	// splits providerTypes by '/'
	types := strings.Split(providerTypes, "/")
	if len(types) == 0 || len(providerTypes) == 0 {
		err := errors.New("empty providers")
		glogger.Error(1, "init log error: %v", err)
		return err
	}
	// gets opts string
	optsString := ""
	switch c := opts.(type) {
	case string:
		optsString = c
	case jsonStringer:
		optsString = c.JSON()
	case fmt.Stringer:
		optsString = c.String()
	default:
		optsString = fmt.Sprintf("%v", opts)
	}

	// clean repeated provider type
	usedTypes := map[string]bool{}
	for _, typ := range types {
		typ = strings.TrimSpace(typ)
		usedTypes[typ] = true
	}

	// creates providers
	var providers []logger.Provider
	for typ := range usedTypes {
		creator := logger.Lookup(typ)
		if creator == nil {
			err := errors.New("unregistered provider type: " + typ)
			glogger.Error(1, "init log error: %v", err)
			return err
		}
		p := creator(optsString)
		if len(usedTypes) == 1 {
			return InitWithProvider(p)
		}
		providers = append(providers, p)
	}
	return InitWithProvider(provider.NewMixProvider(providers[0], providers[1:]...))
}

// InitFile inits with file provider by log file fullpath
func InitFile(fullpath string) error {
	return Init("file", makeFileOpts(fullpath))
}

func makeFileOpts(fullpath string) string {
	dir, filename := filepath.Split(fullpath)
	if dir == "" {
		dir = "."
	}
	return fmt.Sprintf(`{"dir":%s,"filename":%s}`, strconv.Quote(dir), strconv.Quote(filename))
}

// InitConsole inits with console provider by toStderrLevel
func InitConsole(toStderrLevel logger.Level) error {
	return Init("console", makeConsoleOpts(toStderrLevel))
}

func InitColoredConsole(toStderrLevel logger.Level) error {
	return Init("colored_console", makeConsoleOpts(toStderrLevel))
}

func makeConsoleOpts(toStderrLevel logger.Level) string {
	return fmt.Sprintf(`{"tostderrlevel":%d}`, toStderrLevel)
}

// InitFileAndConsole inits with console and file providers
func InitFileAndConsole(fullpath string, toStderrLevel logger.Level) error {
	fileOpts := makeFileOpts(fullpath)
	consoleOpts := makeConsoleOpts(toStderrLevel)
	p := provider.NewMixProvider(provider.NewFile(fileOpts), provider.NewConsole(consoleOpts))
	return InitWithProvider(p)
}

// InitMultiFile inits with multifile provider
func InitMultiFile(rootdir, filename string) error {
	return Init("multifile", makeMultiFileOpts(rootdir, filename))
}

func makeMultiFileOpts(rootdir, filename string) string {
	return fmt.Sprintf(`{"rootdir":"%s","filename":"%s"}`, rootdir, filename)
}

// InitMultiFileAndConsole inits with console and multifile providers
func InitMultiFileAndConsole(rootdir, filename string, toStderrLevel logger.Level) error {
	multifileOpts := makeMultiFileOpts(rootdir, filename)
	consoleOpts := makeConsoleOpts(toStderrLevel)
	p := provider.NewMixProvider(provider.NewMultiFile(multifileOpts), provider.NewConsole(consoleOpts))
	return InitWithProvider(p)
}

func NoHeader()                                { glogger.NoHeader() }
func GetLevel() logger.Level                   { return glogger.GetLevel() }
func SetLevel(level logger.Level)              { glogger.SetLevel(level) }
func Trace(format string, args ...interface{}) { glogger.Trace(1, format, args...) }
func Debug(format string, args ...interface{}) { glogger.Debug(1, format, args...) }
func Info(format string, args ...interface{})  { glogger.Info(1, format, args...) }
func Warn(format string, args ...interface{})  { glogger.Warn(1, format, args...) }
func Error(format string, args ...interface{}) { glogger.Error(1, format, args...) }
func Fatal(format string, args ...interface{}) { glogger.Fatal(1, format, args...) }

func Printf(calldepth int, level logger.Level, format string, args ...interface{}) {
	switch level {
	case LvTRACE:
		glogger.Trace(calldepth, format, args...)
	case LvDEBUG:
		glogger.Debug(calldepth, format, args...)
	case LvINFO:
		glogger.Info(calldepth, format, args...)
	case LvWARN:
		glogger.Warn(calldepth, format, args...)
	case LvERROR:
		glogger.Error(calldepth, format, args...)
	case LvFATAL:
		glogger.Fatal(calldepth, format, args...)
	}
}

func Print(calldepth int, level logger.Level, args ...interface{}) {
	if level <= glogger.GetLevel() {
		msg := fmt.Sprint(args...)
		switch level {
		case LvTRACE:
			glogger.Trace(calldepth, msg)
		case LvDEBUG:
			glogger.Debug(calldepth, msg)
		case LvINFO:
			glogger.Info(calldepth, msg)
		case LvWARN:
			glogger.Warn(calldepth, msg)
		case LvERROR:
			glogger.Error(calldepth, msg)
		case LvFATAL:
			glogger.Fatal(calldepth, msg)
		}
	}
}

// HTTPHandlerGetLevel returns a http handler for getting log level
func HTTPHandlerGetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, GetLevel().String())
	})
}

// HTTPHandlerSetLevel sets new log level and returns old log level
// Returns status code `StatusBadRequest` if parse log level fail
func HTTPHandlerSetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		level := r.FormValue("level")
		lv, ok := ParseLevel(level)
		// invalid parameter
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid log level: "+level)
			return
		}
		// not modified
		oldLevel := GetLevel()
		if lv == oldLevel {
			w.WriteHeader(http.StatusNotModified)
			io.WriteString(w, oldLevel.String())
			return
		}
		// updated
		SetLevel(lv)
		io.WriteString(w, oldLevel.String())
	})
}

// SetLevelFromString parses level from string and set parsed level
// (NOTE): set level to INFO if parse failed
func SetLevelFromString(s string) logger.Level {
	level, _ := ParseLevel(s)
	glogger.SetLevel(level)
	return level
}

// If returns an `IfLogger`
func If(ok bool) IfLogger {
	if ok {
		return IfLogger(0xFF)
	} else {
		return IfLogger(0x00)
	}
}

// With returns a ContextLogger
func With(values ...interface{}) ContextLogger {
	if len(values) == 1 {
		return &contextLogger{isTrue: true, data: values[0]}
	}
	return &contextLogger{isTrue: true, data: values}
}

// WithJSON returns a ContextLogger using JSONFormatter
func WithJSON(values ...interface{}) ContextLogger {
	if len(values) == 1 {
		return &contextLogger{isTrue: true, data: values[0], formatter: jsonFormatter}
	}
	return &contextLogger{isTrue: true, data: values, formatter: jsonFormatter}
}
