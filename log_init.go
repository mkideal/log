package log

import (
	"github.com/mkideal/log/provider"
)

// global logger
var glogger *logger

func InitWithFile(level LogLevel, logfile string) {
	glogger = newLogger(provider.NewFile(logfile))
	glogger.level = level
	glogger.Run()
}
