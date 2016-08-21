package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Logger is the top-level object of log package
type Logger interface {
	// Run startup logger
	Run()
	// Quit quits logger
	Quit()
	// NoHeader ignores header while output logs
	NoHeader()
	// GetLevel gets current log level
	GetLevel() Level
	// SetLevel sets log level
	SetLevel(level Level)
	// Trace outputs trace-level logs
	Trace(calldepth int, format string, args ...interface{})
	// Debug outputs debug-level logs
	Debug(calldepth int, format string, args ...interface{})
	// Info outputs info-level logs
	Info(calldepth int, format string, args ...interface{})
	// Warn outputs warn-level logs
	Warn(calldepth int, format string, args ...interface{})
	// Error outputs error-level logs
	Error(calldepth int, format string, args ...interface{})
	// Fatal outputs fatal-level logs
	Fatal(calldepth int, format string, args ...interface{})
}

type WithLogger interface {
	LogWith(level Level, calldepth int, data []byte, format string, args ...interface{})
}

func Stack(calldepth int) []byte {
	var (
		buf           = make([]byte, 1<<16) // 64k
		nbytes        = runtime.Stack(buf, false)
		ignorelinenum = 2*calldepth + 1
		count         = 0
		startIndex    = 0
	)
	for i := range buf {
		if buf[i] == '\n' {
			count++
			if count == ignorelinenum {
				startIndex = i + 1
			}
		}
	}
	return buf[startIndex:nbytes]
}

// logger implements Logger interface
type logger struct {
	level    Level
	provider Provider
	noHeader int32

	bufferListLocker sync.Mutex
	bufferList       *buffer

	running    int32
	writeQueue chan *buffer
	quitNotify chan struct{}

	async       bool
	writeLocker sync.Mutex // used while async==false
}

// New creates async logger with provider
func New(provider Provider) Logger {
	return newLogger(provider, true)
}

// NewSync creates a sync logger with provider
func NewSync(provider Provider) Logger {
	return newLogger(provider, false)
}

// NewLoggerForTest creates a logger for test
func NewLoggerForTest(provider Provider, async bool, with bool) Logger {
	l := newLogger(provider, async)
	if with {
		return l
	}
	return l.logger
}

func newLogger(provider Provider, async bool) *withLogger {
	l := &logger{
		provider:   provider,
		bufferList: new(buffer),
		writeQueue: make(chan *buffer, 8192),
		quitNotify: make(chan struct{}),
		async:      async,
	}
	return &withLogger{l}
}

type withLogger struct {
	*logger
}

// LogWith implements WithLogger
func (l *withLogger) LogWith(level Level, calldepth int, data []byte, format string, args ...interface{}) {
	if l.GetLevel() >= level {
		l.output(level, calldepth, data, format, args...)
	}
}

func (l *logger) Run() {
	if !l.async || atomic.AddInt32(&l.running, 1) > 1 {
		return
	}
	go func() {
		for buf := range l.writeQueue {
			if buf.quit {
				break
			}
			l.writeBuffer(buf)
		}
		atomic.StoreInt32(&l.running, 0)
		l.quitNotify <- struct{}{}
	}()
}

func (l *logger) writeBuffer(buf *buffer) {
	l.provider.Write(buf.level, buf.headerLength, buf.Bytes())
	if buf.level == FATAL {
		l.provider.Close()
		os.Exit(1)
	}
	l.putBuffer(buf)
}

func (l *logger) Quit() {
	if !l.async || atomic.LoadInt32(&l.running) == 0 {
		return
	}
	l.writeQueue <- &buffer{quit: true}
	<-l.quitNotify
	l.provider.Close()
}

func (l *logger) getBuffer() *buffer {
	l.bufferListLocker.Lock()
	if b := l.bufferList; b != nil {
		l.bufferList = b.next
		b.next = nil
		b.Reset()
		l.bufferListLocker.Unlock()
		return b
	}
	l.bufferListLocker.Unlock()
	return new(buffer)
}

func (l *logger) putBuffer(buf *buffer) {
	if buf.Len() > 256 { //FIXME
		return
	}
	l.bufferListLocker.Lock()
	buf.next = l.bufferList
	l.bufferList = buf
	l.bufferListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (l *logger) formatHeader(now time.Time, level Level, file string, line int) *buffer {
	if line < 0 {
		line = 0
	}
	var (
		buf                  = l.getBuffer()
		year, month, day     = now.Date()
		hour, minute, second = now.Clock()
		millisecond          = now.Nanosecond() / 1000000
	)
	buf.tmp[0] = '['
	buf.tmp[1] = level.String()[0]
	buf.tmp[2] = ' '
	fourDigits(buf, 3, year)
	buf.tmp[7] = '/'
	twoDigits(buf, 8, int(month))
	buf.tmp[10] = '/'
	twoDigits(buf, 11, day)
	buf.tmp[13] = ' '
	twoDigits(buf, 14, hour)
	buf.tmp[16] = ':'
	twoDigits(buf, 17, minute)
	buf.tmp[19] = ':'
	twoDigits(buf, 20, second)
	buf.tmp[22] = '.'
	threeDigits(buf, 23, millisecond)
	buf.tmp[26] = ' '
	buf.Write(buf.tmp[:27])
	buf.WriteString(file)
	buf.tmp[0] = ':'
	n := someDigits(buf, 1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])

	return buf
}

func (l *logger) header(level Level, calldepth int) *buffer {
	if atomic.LoadInt32(&l.noHeader) != 0 {
		return l.getBuffer()
	}
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return l.formatHeader(time.Now(), level, file, line)
}

func (l *logger) output(level Level, calldepth int, data []byte, format string, args ...interface{}) {
	buf := l.header(level, calldepth+3)
	buf.headerLength = buf.Len()
	if len(data) > 0 {
		buf.Write(data)
		if len(format) > 0 {
			buf.WriteString(" | ")
		}
	}
	fmt.Fprintf(buf, format, args...)
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	if level == FATAL {
		stackBuf := Stack(4)
		buf.WriteString("========= BEGIN STACK TRACE =========\n")
		buf.Write(stackBuf)
		buf.WriteString("========== END STACK TRACE ==========\n")
	}
	if buf.Len() == 0 {
		return
	}
	buf.level = level
	if l.async {
		l.writeQueue <- buf
	} else {
		l.writeLocker.Lock()
		l.writeBuffer(buf)
		l.writeLocker.Unlock()
	}
}

func (l *logger) NoHeader()         { atomic.StoreInt32(&l.noHeader, 1) }
func (l *logger) GetLevel() Level   { return Level(atomic.LoadInt32((*int32)(&l.level))) }
func (l *logger) SetLevel(lv Level) { atomic.StoreInt32((*int32)(&l.level), int32(lv)) }

func (l *logger) Trace(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= TRACE {
		l.output(TRACE, calldepth, nil, format, args...)
	}
}

func (l *logger) Debug(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= DEBUG {
		l.output(DEBUG, calldepth, nil, format, args...)
	}
}

func (l *logger) Info(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= INFO {
		l.output(INFO, calldepth, nil, format, args...)
	}
}

func (l *logger) Warn(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= WARN {
		l.output(WARN, calldepth, nil, format, args...)
	}
}

func (l *logger) Error(calldepth int, format string, args ...interface{}) {
	if l.GetLevel() >= ERROR {
		l.output(ERROR, calldepth, nil, format, args...)
	}
}

func (l *logger) Fatal(calldepth int, format string, args ...interface{}) {
	l.output(FATAL, calldepth, nil, format, args...)
	select {}
}
