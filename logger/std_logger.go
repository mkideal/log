package logger

import (
	"fmt"
	stdlog "log"
	"sync/atomic"
)

type stdLogger LogLevel

func NewStdLogger() Logger {
	l := new(stdLogger)
	*l = stdLogger(DEBUG)
	return l
}

func (l *stdLogger) Run()                    {}
func (l *stdLogger) GetLevel() LogLevel      { return LogLevel(atomic.LoadInt32((*int32)(l))) }
func (l *stdLogger) SetLevel(level LogLevel) { atomic.StoreInt32((*int32)(l), int32(level)) }

func (l *stdLogger) output(calldepth int, level LogLevel, format string, args ...interface{}) {
	if l.GetLevel() >= level {
		stdlog.Output(calldepth+3, fmt.Sprintf(format, args...))
	}
}

func (l *stdLogger) Trace(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, TRACE, format, args...)
}

func (l *stdLogger) Debug(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, DEBUG, format, args...)
}

func (l *stdLogger) Info(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, INFO, format, args...)
}

func (l *stdLogger) Warn(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, WARN, format, args...)
}

func (l *stdLogger) Error(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, ERROR, format, args...)
}
