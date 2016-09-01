package log

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

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

// ParseLevel parses log level from string
func ParseLevel(s string) (logger.Level, bool) { return logger.ParseLevel(s) }

// MustParseLevel similars to ParseLevel, but panic if parse fail
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
func Init(providerType, opts string) error {
	pcreator := logger.Lookup(providerType)
	if pcreator == nil {
		return errors.New("unregistered provider type: " + providerType)
	}
	return InitWithProvider(pcreator(opts))
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
	return fmt.Sprintf(`{"dir":"%s","filename":"%s"}`, dir, filename)
}

// InitConsole inits with console provider by toStderrLevel
func InitConsole(toStderrLevel logger.Level) error {
	return Init("console", makeConsoleOpts(toStderrLevel))
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

// InitTesting inits logger for testing
func InitTesting(t *testing.T) {
	InitWithLogger(logger.NewTestingLogger(t))
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
