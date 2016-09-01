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

type With interface {
	LogWith(level Level, calldepth int, data []byte, format string, args ...interface{})
}

// Entry represents a logging entry
type Entry interface {
	Level() Level
	Timestamp() int64
	Body() []byte
	Desc() []byte
	Clone() Entry
}

// Handler handle the logging entry
type Handler interface {
	Handle(entry Entry) error
}

// HookableLogger is a logger which can hook handlers
type HookableLogger interface {
	Logger
	Hook(Handler)
}

// Stack gets the call stack
func Stack(calldepth int) []byte {
	var (
		e             = make([]byte, 1<<16) // 64k
		nbytes        = runtime.Stack(e, false)
		ignorelinenum = 2*calldepth + 1
		count         = 0
		startIndex    = 0
	)
	for i := range e {
		if e[i] == '\n' {
			count++
			if count == ignorelinenum {
				startIndex = i + 1
			}
		}
	}
	return e[startIndex:nbytes]
}

// logger implements interface HookableLogger and With
type logger struct {
	level    Level
	provider Provider
	noHeader int32

	entryListLocker sync.Mutex
	entryList       *entry

	running    int32
	writeQueue chan *entry
	quitNotify chan struct{}

	async       bool
	writeLocker sync.Mutex // used if async==false

	// hooked handlers
	handlers []Handler
}

// New creates async logger with provider
func New(provider Provider) HookableLogger {
	return newLogger(provider, true)
}

// NewSync creates a sync logger with provider
func NewSync(provider Provider) HookableLogger {
	return newLogger(provider, false)
}

// (NOTE): NewLoggerForTest creates a logger for testing
func NewLoggerForTest(provider Provider, async bool, with bool) HookableLogger {
	l := newLogger(provider, async)
	if with {
		return l
	}
	return l.logger
}

func newLogger(provider Provider, async bool) *withLogger {
	l := &logger{
		provider:   provider,
		entryList:  new(entry),
		writeQueue: make(chan *entry, 8192),
		quitNotify: make(chan struct{}),
		async:      async,
		handlers:   []Handler{},
	}
	return &withLogger{l}
}

type withLogger struct {
	*logger
}

// LogWith implements With interface
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
		for e := range l.writeQueue {
			if e.quit {
				break
			}
			l.writeBuffer(e)
		}
		atomic.StoreInt32(&l.running, 0)
		l.quitNotify <- struct{}{}
	}()
}

func (l *logger) writeBuffer(e *entry) {
	l.provider.Write(e.level, e.headerLength, e.Bytes())
	if len(l.handlers) > 0 {
		for _, h := range l.handlers {
			h.Handle(e)
		}
	}
	if e.level == FATAL {
		l.provider.Close()
		os.Exit(1)
	}
	l.putBuffer(e)
}

func (l *logger) Hook(h Handler) {
	l.handlers = append(l.handlers, h)
}

func (l *logger) Quit() {
	if !l.async || atomic.LoadInt32(&l.running) == 0 {
		return
	}
	l.writeQueue <- &entry{quit: true}
	<-l.quitNotify
	l.provider.Close()
}

func (l *logger) getBuffer() *entry {
	l.entryListLocker.Lock()
	if b := l.entryList; b != nil {
		l.entryList = b.next
		b.next = nil
		b.Reset()
		l.entryListLocker.Unlock()
		return b
	}
	l.entryListLocker.Unlock()
	return new(entry)
}

func (l *logger) putBuffer(e *entry) {
	if e.Len() > 256 { //FIXME
		return
	}
	l.entryListLocker.Lock()
	e.next = l.entryList
	l.entryList = e
	l.entryListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (l *logger) formatHeader(now time.Time, level Level, file string, line int) *entry {
	if line < 0 {
		line = 0
	}
	var (
		e                    = l.getBuffer()
		year, month, day     = now.Date()
		hour, minute, second = now.Clock()
		millisecond          = now.Nanosecond() / 1000000
	)
	e.timestamp = now.Unix()
	e.tmp[0] = '['
	e.tmp[1] = level.String()[0]
	e.tmp[2] = ' '
	fourDigits(e, 3, year)
	e.tmp[7] = '/'
	twoDigits(e, 8, int(month))
	e.tmp[10] = '/'
	twoDigits(e, 11, day)
	e.tmp[13] = ' '
	twoDigits(e, 14, hour)
	e.tmp[16] = ':'
	twoDigits(e, 17, minute)
	e.tmp[19] = ':'
	twoDigits(e, 20, second)
	e.tmp[22] = '.'
	threeDigits(e, 23, millisecond)
	e.tmp[26] = ' '
	e.Write(e.tmp[:27])
	e.WriteString(file)
	e.tmp[0] = ':'
	n := someDigits(e, 1, line)
	e.tmp[n+1] = ']'
	e.tmp[n+2] = ' '
	e.Write(e.tmp[:n+3])

	return e
}

func (l *logger) header(level Level, calldepth int) *entry {
	now := time.Now()
	if atomic.LoadInt32(&l.noHeader) != 0 {
		e := l.getBuffer()
		e.timestamp = now.Unix()
		return e
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
	e := l.header(level, calldepth+3)
	e.headerLength = e.Len()
	if len(data) > 0 {
		e.bodyBegin = e.Len()
		e.Write(data)
		e.bodyEnd = e.Len()
		if len(format) > 0 {
			e.WriteString(" | ")
		}
	}
	e.descBegin = e.Len()
	fmt.Fprintf(e, format, args...)
	e.descEnd = e.Len()
	if e.Len() > 0 && e.Bytes()[e.Len()-1] != '\n' {
		e.WriteByte('\n')
	}
	if level == FATAL {
		stackBuf := Stack(4)
		e.WriteString("========= BEGIN STACK TRACE =========\n")
		e.Write(stackBuf)
		e.WriteString("========== END STACK TRACE ==========\n")
	}
	if e.Len() == 0 {
		return
	}
	e.level = level
	if l.async {
		l.writeQueue <- e
	} else {
		l.writeLocker.Lock()
		l.writeBuffer(e)
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
