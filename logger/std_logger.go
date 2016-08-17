package logger

import (
	"bytes"
	"fmt"
	stdlog "log"
	"os"
	"sync/atomic"
)

type stdLogger Level

// NewStdLogger creates std logger
func NewStdLogger() Logger {
	l := new(stdLogger)
	*l = stdLogger(INFO)
	return l
}

func (l *stdLogger) Run()                 {}
func (l *stdLogger) Quit()                {}
func (l *stdLogger) NoHeader()            { stdlog.SetPrefix("") }
func (l *stdLogger) GetLevel() Level      { return Level(atomic.LoadInt32((*int32)(l))) }
func (l *stdLogger) SetLevel(level Level) { atomic.StoreInt32((*int32)(l), int32(level)) }

func (l *stdLogger) output(calldepth int, level Level, format string, args ...interface{}) {
	if level != FATAL {
		stdlog.Output(calldepth+3, fmt.Sprintf(format, args...))
	} else {
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, format, args...)
		if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteByte('\n')
		}
		stackBuf := Stack(4)
		buf.WriteString("========= BEGIN STACK TRACE =========\n")
		buf.Write(stackBuf)
		buf.WriteString("========== END STACK TRACE ==========\n")
		l.output(calldepth, TRACE, buf.String())
		os.Exit(1)
	}
}

func (l *stdLogger) Trace(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= TRACE {
		l.output(calldepth, TRACE, format, args...)
	}
}

func (l *stdLogger) Debug(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= DEBUG {
		l.output(calldepth, DEBUG, format, args...)
	}
}

func (l *stdLogger) Info(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= INFO {
		l.output(calldepth, INFO, format, args...)
	}
}

func (l *stdLogger) Warn(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= WARN {
		l.output(calldepth, WARN, format, args...)
	}
}

func (l *stdLogger) Error(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= ERROR {
		l.output(calldepth, ERROR, format, args...)
	}
}

func (l *stdLogger) Fatal(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, FATAL, format, args...)
}
