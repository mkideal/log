package log

import (
	"bytes"
	"fmt"
	"log"
	std "log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxWaitTimeForVerboseLevel   = 100 * time.Millisecond // trace, debug
	maxWaitTimeForImportantLevel = 3 * time.Second        // info, warn, error, fatal
)

// Printer represents the printer for logging
type Printer interface {
	// Start starts the printer
	Start()
	// Quit quits the printer
	Shutdown()
	// GetLevel gets current log level
	GetLevel() Level
	// SetLevel sets log level
	SetLevel(Level)
	// SetPrefix sets log prefix
	SetPrefix(string)
	// Printf outputs leveled logs with specified calldepth and extra prefix
	Printf(calldepth int, level Level, prefix, format string, args ...interface{})
}

// stack gets the call stack
func stack(calldepth int) []byte {
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

// printer implements Printer
type printer struct {
	level  Level
	prefix string
	writer Writer

	entryListLocker sync.Mutex
	entryList       *entry

	running int32
	queue   chan *entry
	wait    chan struct{}

	async       bool
	writeLocker sync.Mutex // used if async==false
}

// newPrinter creates built in printer
func newPrinter(writer Writer, async bool) Printer {
	return &printer{
		writer:    writer,
		entryList: new(entry),
		queue:     make(chan *entry, 8192),
		wait:      make(chan struct{}),
		async:     async,
	}
}

// Start implements Printer Start method
func (p *printer) Start() {
	if !p.async || atomic.AddInt32(&p.running, 1) > 1 {
		return
	}
	go func() {
		for e := range p.queue {
			if e.quit {
				break
			}
			p.writeBuffer(e)
		}
		atomic.StoreInt32(&p.running, 0)
		close(p.wait)
	}()
}

// Shutdown implements Printer Shutdown method
func (p *printer) Shutdown() {
	if !p.async || atomic.LoadInt32(&p.running) == 0 {
		return
	}
	p.queue <- &entry{quit: true}
	<-p.wait
	p.writer.Close()
}

func (p *printer) writeBuffer(e *entry) {
	p.writer.Write(e.level, e.Bytes(), e.headerLen)
	if e.level == LvFATAL {
		p.writer.Close()
		os.Exit(1)
	}
	p.putBuffer(e)
}

func (p *printer) getBuffer() *entry {
	p.entryListLocker.Lock()
	if b := p.entryList; b != nil {
		p.entryList = b.next
		b.next = nil
		b.Reset()
		p.entryListLocker.Unlock()
		return b
	}
	p.entryListLocker.Unlock()
	return new(entry)
}

func (p *printer) putBuffer(e *entry) {
	if e.Len() > 256 {
		return
	}
	p.entryListLocker.Lock()
	e.next = p.entryList
	p.entryList = e
	p.entryListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (p *printer) formatHeader(now time.Time, level Level, file string, line int) *entry {
	if line < 0 {
		line = 0
	}
	var (
		e                    = p.getBuffer()
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
	e.tmp[26] = ']'
	e.tmp[27] = ' '
	e.tmp[28] = '['
	e.Write(e.tmp[:29])
	e.WriteString(file)
	e.tmp[0] = ':'
	n := someDigits(e, 1, line)
	e.tmp[n+1] = ']'
	e.tmp[n+2] = ' '
	e.Write(e.tmp[:n+3])

	return e
}

func (p *printer) header(level Level, calldepth int) *entry {
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
	return p.formatHeader(time.Now(), level, file, line)
}

func (p *printer) output(level Level, calldepth int, prefix, format string, args ...interface{}) {
	e := p.header(level, calldepth+3)
	e.headerLen = e.Len()
	if len(p.prefix) > 0 {
		e.WriteString("(")
		e.WriteString(p.prefix)
		if len(prefix) > 0 {
			e.WriteString("/")
			e.WriteString(prefix)
		}
		e.WriteString(") ")
	} else if len(prefix) > 0 {
		e.WriteString("(")
		e.WriteString(prefix)
		e.WriteString(") ")
	}
	e.descBegin = e.Len()
	if len(args) == 0 {
		fmt.Fprint(e, format)
	} else {
		fmt.Fprintf(e, format, args...)
	}
	e.descEnd = e.Len()
	if e.Len() > 0 && e.Bytes()[e.Len()-1] != '\n' {
		e.WriteByte('\n')
	}
	if level == LvFATAL {
		stackBuf := stack(4)
		e.WriteString("========= BEGIN STACK TRACE =========\n")
		e.Write(stackBuf)
		e.WriteString("========== END STACK TRACE ==========\n")
	}
	if e.Len() == 0 {
		return
	}
	e.level = level
	maxWaitTime := maxWaitTimeForImportantLevel
	if e.level.MoreVerboseThan(LvINFO) {
		maxWaitTime = maxWaitTimeForVerboseLevel
	}
	if p.async && atomic.LoadInt32(&p.running) != 0 {
		select {
		case p.queue <- e:
		case <-time.After(maxWaitTime):
		}
	} else {
		p.writeLocker.Lock()
		p.writeBuffer(e)
		p.writeLocker.Unlock()
	}
}

// GetLevel implements Printer GetLevel method
func (p *printer) GetLevel() Level {
	return Level(atomic.LoadInt32((*int32)(&p.level)))
}

// SetLevel implements Printer SetLevel method
func (p *printer) SetLevel(level Level) {
	atomic.StoreInt32((*int32)(&p.level), int32(level))
}

// SetPrefix implements Printer SetPrefix method, SetPrefix is not concurrent-safe
func (p *printer) SetPrefix(prefix string) {
	p.prefix = prefix
}

// Printf implements Printer Printf method
func (p *printer) Printf(calldepth int, level Level, prefix, format string, args ...interface{}) {
	if p.GetLevel() >= level {
		p.output(level, calldepth, prefix, format, args...)
	}
	if level == LvFATAL {
		// blocked
		select {}
	}
}

// stdPrinter wraps golang standard log package
type stdPrinter struct {
	level Level
}

// newStdPrinter creates std logger
func newStdPrinter() Printer {
	p := new(stdPrinter)
	*p = stdPrinter{level: LvINFO}
	return p
}

// Start implements Printer Start method
func (stdPrinter) Start() {}

// Shutdown implements Printer Shutdown method
func (stdPrinter) Shutdown() {}

// NoHeader implements Printer NoHeader method
func (stdPrinter) SetHeader() { std.SetPrefix("") }

// GetLevel implements Printer GetLevel method
func (p stdPrinter) GetLevel() Level { return Level(atomic.LoadInt32((*int32)(&p.level))) }

// SetLevel implements Printer SetLevel method
func (p *stdPrinter) SetLevel(level Level) { atomic.StoreInt32((*int32)(&p.level), int32(level)) }

// SetPrefix implements Printer SetPrefix method
func (p *stdPrinter) SetPrefix(prefix string) { log.SetPrefix(prefix) }

// Printf implements Printer Printf method
func (p stdPrinter) Printf(calldepth int, level Level, prefix, format string, args ...interface{}) {
	if p.GetLevel() >= level {
		p.output(calldepth, level, prefix+format, args...)
	}
}

func (p stdPrinter) output(calldepth int, level Level, format string, args ...interface{}) {
	if level != LvFATAL {
		if len(args) == 0 {
			std.Output(calldepth+3, fmt.Sprint(format))
		} else {
			std.Output(calldepth+3, fmt.Sprintf(format, args...))
		}
	} else {
		buf := new(bytes.Buffer)
		if len(args) == 0 {
			fmt.Fprint(buf, format)
		} else {
			fmt.Fprintf(buf, format, args...)
		}
		if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteByte('\n')
		}
		stackBuf := stack(4)
		buf.WriteString("========= BEGIN STACK TRACE =========\n")
		buf.Write(stackBuf)
		buf.WriteString("========== END STACK TRACE ==========\n")
		std.Output(calldepth+3, buf.String())
		os.Exit(1)
	}
}
