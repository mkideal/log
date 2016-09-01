package logger

import (
	"bytes"
	"fmt"
	"os"
	"sync/atomic"

	"testing"
)

type testingLogger struct {
	t1    testingLoggerWriter
	t2    testingLoggerWriterWithCalldepth
	level Level
}

type testingLoggerWriter interface {
	Log(args ...interface{})
}

type testingLoggerWriterWithCalldepth interface {
	LogCalldepth(calldepth int, args ...interface{})
}

// NewTestingLogger creates testing logger
func NewTestingLogger(t *testing.T) Logger {
	l := &testingLogger{
		t1:    t,
		level: TRACE,
	}
	if t2, ok := l.t1.(testingLoggerWriterWithCalldepth); ok {
		l.t2 = t2
	}
	return l
}

func (l *testingLogger) Run()                 {}
func (l *testingLogger) Quit()                {}
func (l *testingLogger) NoHeader()            {}
func (l *testingLogger) GetLevel() Level      { return Level(atomic.LoadInt32((*int32)(&l.level))) }
func (l *testingLogger) SetLevel(level Level) { atomic.StoreInt32((*int32)(&l.level), int32(level)) }

func (l *testingLogger) output(calldepth int, level Level, format string, args ...interface{}) {
	if level != FATAL {
		calldepth += 5
		msg := fmt.Sprintf(format, args...)
		if l.t2 != nil {
			l.t2.LogCalldepth(calldepth, msg)
		} else {
			l.t1.Log(calldepth, msg)
		}
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

func (l *testingLogger) Trace(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= TRACE {
		l.output(calldepth, TRACE, format, args...)
	}
}

func (l *testingLogger) Debug(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= DEBUG {
		l.output(calldepth, DEBUG, format, args...)
	}
}

func (l *testingLogger) Info(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= INFO {
		l.output(calldepth, INFO, format, args...)
	}
}

func (l *testingLogger) Warn(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= WARN {
		l.output(calldepth, WARN, format, args...)
	}
}

func (l *testingLogger) Error(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= ERROR {
		l.output(calldepth, ERROR, format, args...)
	}
}

func (l *testingLogger) Fatal(calldepth int, format string, args ...interface{}) {
	l.output(calldepth, FATAL, format, args...)
}
