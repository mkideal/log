package log

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mkideal/log/provider"
)

type Logger interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Run()
}

type buffer struct {
	bytes.Buffer
	tmp  [64]byte
	next *buffer
}

type logger struct {
	level    LogLevel
	provider provider.LoggerProvider

	bufferListLocker sync.Mutex
	bufferList       *buffer
}

func NewLogger(provider provider.LoggerProvider) Logger {
	return newLogger(provider)
}

func newLogger(provider provider.LoggerProvider) *logger {
	return &logger{
		provider:   provider,
		bufferList: new(buffer),
	}
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

const digits = "0123456789"

func twoDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func threeDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+2] = digits[v%10]
	v /= 10
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func fourDigits(buf *buffer, begin int, v int) {
	buf.tmp[begin+3] = digits[v%10]
	v /= 10
	buf.tmp[begin+2] = digits[v%10]
	v /= 10
	buf.tmp[begin+1] = digits[v%10]
	v /= 10
	buf.tmp[begin] = digits[v%10]
}

func someDigits(buf *buffer, begin int, v int) int {
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[v%10]
		v /= 10
		if v == 0 {
			break
		}
	}
	return copy(buf.tmp[begin:], buf.tmp[j:])
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (l *logger) formatHeader(level LogLevel, file string, line int) *buffer {
	now := time.Now()
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
	buf.tmp[22] = ' '
	buf.tmp[23] = '.'
	buf.tmp[24] = ' '
	threeDigits(buf, 25, millisecond)
	buf.Write(buf.tmp[:26])
	buf.WriteString(file)
	buf.tmp[0] = ':'
	n := someDigits(buf, 1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])

	return buf
}

func (l *logger) header(level LogLevel, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 0
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return l.formatHeader(level, file, line), file, line
}

func (l *logger) printDepth(level LogLevel, depth int, format string, args ...interface{}) {
	buf, _, _ := l.header(level, depth)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.provider.Write(level, buf.Bytes())
}

func (l *logger) GetLevel() LogLevel   { return LogLevel(atomic.LoadInt32((*int32)(&l.level))) }
func (l *logger) SetLevel(lv LogLevel) { atomic.StoreInt32((*int32)(&l.level), int32(lv)) }

func (l *logger) Trace(format string, args ...interface{}) {
	if l.level >= TRACE {
		l.printDepth(TRACE, 3, format, args...)
	}
}

func (l *logger) Debug(format string, args ...interface{}) {
	if l.level >= DEBUG {
		l.printDepth(DEBUG, 3, format, args...)
	}
}

func (l *logger) Info(format string, args ...interface{}) {
	if l.level >= INFO {
		l.printDepth(INFO, 3, format, args...)
	}
}

func (l *logger) Warn(format string, args ...interface{}) {
	if l.level >= WARN {
		l.printDepth(WARN, 3, format, args...)
	}
}

func (l *logger) Error(format string, args ...interface{}) {
	l.printDepth(ERROR, 3, format, args...)
}
