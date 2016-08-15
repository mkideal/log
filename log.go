package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

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

// InitWithProvider inits global logger with a specified provider
func InitWithProvider(p logger.Provider) error {
	glogger.Quit()
	glogger = logger.New(p)
	glogger.SetLevel(LvINFO)
	glogger.Run()
	return nil
}

// Init inits global logger with providerType and opts
// * providerType: providerType should be one of {file, console}
// * opts        : opts is a json string or empty
func Init(providerType, opts string) error {
	pcreator := logger.Lookup(providerType)
	if pcreator == nil {
		return errors.New("unsupported provider type: " + providerType)
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
func If(ok bool) IfLogger { return IfLogger(ok) }

type IfLogger bool

func (il IfLogger) Else() IfLogger          { return !il }
func (il IfLogger) ElseIf(ok bool) IfLogger { return !il && IfLogger(ok) }

func (il IfLogger) Trace(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Trace(1, format, args...)
	}
	return il
}

func (il IfLogger) Debug(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Debug(1, format, args...)
	}
	return il
}

func (il IfLogger) Info(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Info(1, format, args...)
	}
	return il
}

func (il IfLogger) Warn(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Warn(1, format, args...)
	}
	return il
}

func (il IfLogger) Error(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Error(1, format, args...)
	}
	return il
}

func (il IfLogger) Fatal(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Fatal(1, format, args...)
	}
	return il
}

// With returns a WithLogger
func With(v interface{}) *WithLogger { return &WithLogger{data: v} }

// WithN returns a WithLogger which data is a slice of interface{}
func WithN(objs ...interface{}) *WithLogger { return &WithLogger{data: objs} }

// WithJSON returns a WithLogger which use JSONFormatter
func WithJSON(v interface{}) *WithLogger { return &WithLogger{data: v, formatter: jsonFormatter} }

type WithLogger struct {
	data      interface{}
	formatter Formatter
	b         []byte
}

var bytesTrue = []byte("true")
var bytesFalse = []byte("false")

func toBytes(v interface{}) []byte {
	switch data := v.(type) {
	case string:
		return []byte(data)
	case int:
		return strconv.AppendInt(nil, int64(data), 10)
	case int8:
		return strconv.AppendInt(nil, int64(data), 10)
	case int16:
		return strconv.AppendInt(nil, int64(data), 10)
	case int32:
		return strconv.AppendInt(nil, int64(data), 10)
	case int64:
		return strconv.AppendInt(nil, data, 10)
	case uint:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uintptr:
		return strconv.AppendUint([]byte{'0', 'x'}, uint64(data), 16)
	case uint8:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint16:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint32:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint64:
		return strconv.AppendUint(nil, data, 10)
	case float32:
		return strconv.AppendFloat(nil, float64(data), 'E', -1, 64)
	case float64:
		return strconv.AppendFloat(nil, data, 'E', -1, 64)
	case bool:
		if data {
			return bytesTrue
		}
		return bytesFalse
	default:
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "%v", data)
		return buf.Bytes()
	}
}

func (wl *WithLogger) Bytes() []byte {
	if wl.b != nil {
		return wl.b
	}
	if wl.formatter != nil {
		wl.b = wl.formatter.Format(wl.data)
		return wl.b
	}
	wl.b = toBytes(wl.data)
	return wl.b
}

func (wl *WithLogger) SetFormatter(f Formatter) *WithLogger {
	wl.formatter = f
	wl.b = nil
	return wl
}

func (wl *WithLogger) Trace(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvTRACE {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvTRACE, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Trace(1, buf.String())
		}
	}
}

func (wl *WithLogger) Debug(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvDEBUG {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvDEBUG, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Debug(1, buf.String())
		}
	}
}

func (wl *WithLogger) Info(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvINFO {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvINFO, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Info(1, buf.String())
		}
	}
}

func (wl *WithLogger) Warn(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvWARN {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvWARN, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Warn(1, buf.String())
		}
	}
}

func (wl *WithLogger) Error(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvERROR {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvERROR, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Error(1, buf.String())
		}
	}
}

func (wl *WithLogger) Fatal(format string, args ...interface{}) {
	if glogger.GetLevel() >= LvFATAL {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvFATAL, 1, wl.Bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.Bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Fatal(1, buf.String())
		}
	}
}

type Formatter interface {
	Format(v interface{}) []byte
}

type JSONFormatter struct{}

var jsonFormatter = JSONFormatter{}

func (f JSONFormatter) Format(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

type M map[string]interface{}
type S []interface{}
