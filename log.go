package log

import (
	"fmt"
	"path/filepath"

	"github.com/mkideal/log/logger"
	"github.com/mkideal/log/provider"
)

// global logger
var glogger = logger.NewStdLogger()

func InitWithCustomLogger(l logger.Logger) {
	glogger = l
	glogger.Run()
}

func InitWithCustomProvider(p logger.Provider) {
	glogger := logger.NewLogger(p)
	glogger.Run()
}

func InitWithFile(level logger.LogLevel, logfile string) {
	dir, file := filepath.Split(logfile)
	if dir == "" {
		dir = "."
	}
	glogger = logger.NewLogger(provider.NewFile(fmt.Sprintf(`{"dir":"%s","filename":"%s"}`, dir, file)))
	glogger.SetLevel(level)
	glogger.Run()
}

func GetLevel() logger.LogLevel                { return glogger.GetLevel() }
func SetLevel(level logger.LogLevel)           { glogger.SetLevel(level) }
func Trace(format string, args ...interface{}) { glogger.Trace(1, format, args...) }
func Debug(format string, args ...interface{}) { glogger.Debug(1, format, args...) }
func Info(format string, args ...interface{})  { glogger.Info(1, format, args...) }
func Warn(format string, args ...interface{})  { glogger.Warn(1, format, args...) }
func Error(format string, args ...interface{}) { glogger.Error(1, format, args...) }
