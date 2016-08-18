package logger

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type mockLogger struct {
	level    Level
	noHeader int32
	locker   sync.Mutex
	writer   io.Writer
}

func NewMockLogger(writer io.Writer) Logger {
	return &mockLogger{writer: writer}
}

func (l *mockLogger) Run()              {}
func (l *mockLogger) Quit()             {}
func (l *mockLogger) NoHeader()         { atomic.StoreInt32(&l.noHeader, 1) }
func (l *mockLogger) GetLevel() Level   { return Level(atomic.LoadInt32((*int32)(&l.level))) }
func (l *mockLogger) SetLevel(lv Level) { atomic.StoreInt32((*int32)(&l.level), int32(lv)) }

func (l *mockLogger) output(level Level, calldepth int, b []byte, format string, args ...interface{}) {
	l.locker.Lock()
	defer l.locker.Unlock()
	if len(b) > 0 {
		l.writer.Write(b)
		if len(format) > 0 {
			io.WriteString(l.writer, " | ")
		}
	}
	fmt.Fprintf(l.writer, format, args...)
	fmt.Fprintf(l.writer, "\n")
}

func (l *mockLogger) Trace(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= TRACE {
		l.output(TRACE, calldepth, nil, format, args...)
	}
}

func (l *mockLogger) Debug(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= DEBUG {
		l.output(DEBUG, calldepth, nil, format, args...)
	}
}

func (l *mockLogger) Info(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= INFO {
		l.output(INFO, calldepth, nil, format, args...)
	}
}

func (l *mockLogger) Warn(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= WARN {
		l.output(WARN, calldepth, nil, format, args...)
	}
}

func (l *mockLogger) Error(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= ERROR {
		l.output(ERROR, calldepth, nil, format, args...)
	}
}

func (l *mockLogger) Fatal(calldepth int, format string, args ...interface{}) {
	l.output(FATAL, calldepth, nil, format, args...)
	select {}
}

// LogWith implements WithLogger
func (l *mockLogger) LogWith(level Level, calldepth int, b []byte, format string, args ...interface{}) {
	if l.GetLevel() >= level {
		l.output(level, calldepth, b, format, args...)
	}
}
