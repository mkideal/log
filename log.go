package log

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mkideal/log/logger"
	"github.com/mkideal/log/provider"
)

const (
	ERROR = logger.ERROR
	WARN  = logger.WARN
	INFO  = logger.INFO
	DEBUG = logger.DEBUG
	TRACE = logger.TRACE
	FATAL = logger.FATAL
)

func ParseLevel(s string) (logger.Level, bool) { return logger.ParseLevel(s) }
func MustParseLevel(s string) logger.Level     { return logger.MustParseLevel(s) }

// global logger
var glogger = logger.NewStdLogger()

func Uninit(err error) {
	glogger.Quit()
}

func InitWithLogger(l logger.Logger) error {
	glogger = l
	glogger.Run()
	return nil
}

func InitWithProvider(p logger.Provider) error {
	glogger = logger.NewLogger(p)
	glogger.SetLevel(INFO)
	glogger.Run()
	return nil
}

// Init inits global logger with providerType and opts
// * providerType: providerType should be one of {file, console}
// * opts        : opts is a json string or empty
func Init(providerType, opts string) error {
	var p logger.Provider
	switch providerType {
	case "file":
		// NOTE: opts for file:
		// dir     : log directory
		// filename: log filename
		// maxsize: max bytes number of one log file(default=1<<26, i.e. 64M)
		//
		// EXAMPLE:
		// `{"dir":"log","filename":"app.log"}`
		// `{"dir":"log","filename":"app.log","maxsize":819200}`
		p = provider.NewFile(opts)
	case "console":
		p = provider.NewConsole()
	default:
		return errors.New("unsupported provider type: " + providerType)
	}
	return InitWithProvider(p)
}

func InitFile(fullpath string) error {
	dir, filename := filepath.Split(fullpath)
	if dir == "" {
		dir = "."
	}
	return Init("file", fmt.Sprintf(`{"dir":"%s","filename":"%s"}`, dir, filename))
}

func InitConsole() error {
	return Init("console", "")
}

func GetLevel() logger.Level                   { return glogger.GetLevel() }
func SetLevel(level logger.Level)              { glogger.SetLevel(level) }
func Trace(format string, args ...interface{}) { glogger.Trace(1, format, args...) }
func Debug(format string, args ...interface{}) { glogger.Debug(1, format, args...) }
func Info(format string, args ...interface{})  { glogger.Info(1, format, args...) }
func Warn(format string, args ...interface{})  { glogger.Warn(1, format, args...) }
func Error(format string, args ...interface{}) { glogger.Error(1, format, args...) }
