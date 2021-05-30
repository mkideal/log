package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	std "log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Level represents log level
type Level int32

// Level constants
const (
	_       Level = iota // 0
	LvFATAL              // 1
	LvERROR              // 2
	LvWARN               // 3
	LvINFO               // 4
	LvDEBUG              // 5
	LvTRACE              // 6

	numLevel = 6
)

var errUnrecognizedLogLevel = errors.New("unrecognized log level")

func (level Level) index() int { return int(level - 1) }

// Set implements flag.Value interface such that you can use level  as a command as following:
//
//	var level logger.Level
//	flag.Var(&level, "log_level", "log level: trace/debug/info/warn/error/fatal")
func (level *Level) Set(s string) error {
	lv, ok := ParseLevel(s)
	*level = lv
	if !ok {
		return errUnrecognizedLogLevel
	}
	return nil
}

// Literal returns literal value which is a number
func (level Level) Literal() string {
	return strconv.Itoa(int(level))
}

// String returns a serialized string of level
func (level Level) String() string {
	switch level {
	case LvFATAL:
		return "FATAL"
	case LvERROR:
		return "ERROR"
	case LvWARN:
		return "WARN"
	case LvINFO:
		return "INFO"
	case LvDEBUG:
		return "DEBUG"
	case LvTRACE:
		return "TRACE"
	}
	return "(" + strconv.Itoa(int(level)) + ")"
}

// MarshalJSON implements json.Marshaler
func (level Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + level.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (level *Level) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	if len(data) >= 2 {
		s, err = strconv.Unquote(string(data))
		if err != nil {
			return err
		}
	} else {
		s = string(data)
	}
	return level.Set(s)
}

// MoreVerboseThan returns whether level more verbose than other
func (level Level) MoreVerboseThan(other Level) bool { return level > other }

// ParseLevel parses log level from string
func ParseLevel(s string) (lv Level, ok bool) {
	s = strings.ToUpper(s)
	switch s {
	case "FATAL", "F", LvFATAL.Literal():
		return LvFATAL, true
	case "ERROR", "E", LvERROR.Literal():
		return LvERROR, true
	case "WARN", "W", LvWARN.Literal():
		return LvWARN, true
	case "INFO", "I", LvINFO.Literal():
		return LvINFO, true
	case "DEBUG", "D", LvDEBUG.Literal():
		return LvDEBUG, true
	case "TRACE", "T", LvTRACE.Literal():
		return LvTRACE, true
	}
	return LvINFO, false
}

// httpHandlerGetLevel returns a http handler for getting log level.
func httpHandlerGetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, GetLevel().String())
	})
}

// httpHandlerSetLevel sets new log level and returns old log level,
// Returns status code `StatusBadRequest` if parse log level fail.
func httpHandlerSetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		level := r.FormValue("level")
		lv, ok := ParseLevel(level)
		// invalid parameter
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid log level: "+level)
			return
		}
		// not modified
		oldLevel := GetLevel()
		if lv == oldLevel {
			w.WriteHeader(http.StatusNotModified)
			io.WriteString(w, oldLevel.String())
			return
		}
		// updated
		SetLevel(lv)
		io.WriteString(w, oldLevel.String())
	})
}

var (
	registerHTTPHandlersOnce sync.Once
)

func registerHTTPHandlers() {
	registerHTTPHandlersOnce.Do(func() {
		http.Handle("/log/level/get", httpHandlerGetLevel())
		http.Handle("/log/level/set", httpHandlerSetLevel())
	})
}

// Printer represents the printer for logging
type Printer interface {
	// Start starts the printer
	Start()
	// Quit quits the printer
	Shutdown()
	// Flush flushs all queued logs
	Flush()
	// SetCaller sets whether print caller information
	SetCaller(bool)
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
	caller int32

	entryListLocker sync.Mutex
	entryList       *entry

	async bool

	// used if async==false
	writeLocker sync.Mutex

	// used if async==true
	running int32
	queue   *queue
	queueMu sync.Mutex
	cond    *sync.Cond
	flush   chan chan struct{}
	quit    chan struct{}
	wait    chan struct{}
}

// newPrinter creates built in printer
func newPrinter(writer Writer, async bool) Printer {
	p := &printer{
		writer:    writer,
		entryList: new(entry),
		async:     async,
	}
	if async {
		p.queue = newQueue()
		p.cond = sync.NewCond(&p.queueMu)
		p.flush = make(chan chan struct{}, 1)
		p.quit = make(chan struct{})
		p.wait = make(chan struct{})
	}
	return p
}

// Start implements Printer Start method
func (p *printer) Start() {
	if p.queue == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return
	}
	go p.run()
}

func (p *printer) run() {
	for {
		p.cond.L.Lock()
		if p.queue.size() == 0 {
			p.cond.Wait()
		}
		entries := p.queue.popAll()
		p.cond.L.Unlock()
		p.writeEntries(entries)
		if p.consumeSignals() {
			break
		}
	}
}

func (p *printer) consumeSignals() bool {
	for {
		select {
		case resp := <-p.flush:
			p.flushAll()
			close(resp)
			continue
		case <-p.quit:
			p.flushAll()
			close(p.wait)
			return true
		default:
			return false
		}
	}
}

func (p *printer) flushAll() {
	p.cond.L.Lock()
	entries := p.queue.popAll()
	p.cond.L.Unlock()
	p.writeEntries(entries)
}

func (p *printer) writeEntries(entries []*entry) {
	for _, e := range entries {
		p.writeEntry(e)
	}
}

// Shutdown implements Printer Shutdown method
func (p *printer) Shutdown() {
	if p.queue == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return
	}
	close(p.quit)
	p.cond.Signal()
	<-p.wait
	p.writer.Close()
}

// Flush implements Printer Flush method
func (p *printer) Flush() {
	wait := make(chan struct{})
	p.flush <- wait
	p.cond.Signal()
	<-wait
}

// SetCaller implements Printer SetCaller method
func (p *printer) SetCaller(yes bool) {
	if yes {
		atomic.StoreInt32(&p.caller, 1)
	} else {
		atomic.StoreInt32(&p.caller, 0)
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
		p.Shutdown()
		os.Exit(1)
	}
}

func (p *printer) writeEntry(e *entry) {
	p.writer.Write(e.level, e.Bytes(), e.headerLen)
	p.putEntry(e)
}

func (p *printer) getEntry() *entry {
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

func (p *printer) putEntry(e *entry) {
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
		e                    = p.getEntry()
		year, month, day     = now.Date()
		hour, minute, second = now.Clock()
		millisecond          = now.Nanosecond() / 1000000
		noCaller             = len(file) == 0
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
	if !noCaller {
		e.tmp[28] = '['
		e.Write(e.tmp[:29])
		e.WriteString(file)
		e.tmp[0] = ':'
		n := someDigits(e, 1, line)
		e.tmp[n+1] = ']'
		e.tmp[n+2] = ' '
		e.Write(e.tmp[:n+3])
	} else {
		e.Write(e.tmp[:28])
	}

	return e
}

func (p *printer) header(level Level, calldepth int) *entry {
	var (
		file     string
		line     int
		ok       bool
		noCaller = atomic.LoadInt32(&p.caller) == 0
	)
	if !noCaller {
		_, file, line, ok = runtime.Caller(calldepth)
	}
	if !ok {
		if !noCaller {
			file = "???"
			line = 0
		}
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
	if len(args) == 0 {
		fmt.Fprint(e, format)
	} else {
		fmt.Fprintf(e, format, args...)
	}
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
	if p.queue != nil && atomic.LoadInt32(&p.running) != 0 {
		p.cond.L.Lock()
		if p.queue.push(e) == 1 {
			p.cond.Signal()
		}
		p.cond.L.Unlock()
	} else {
		p.writeLocker.Lock()
		p.writeEntry(e)
		p.writeLocker.Unlock()
	}
}

// stdPrinter wraps golang standard log package
type stdPrinter struct {
	level Level
}

// newStdPrinter creates std logger
func newStdPrinter() Printer {
	return &stdPrinter{level: LvINFO}
}

// Start implements Printer Start method
func (stdPrinter) Start() {}

// Shutdown implements Printer Shutdown method
func (stdPrinter) Shutdown() {}

// Flush implements Printer Flush method
func (stdPrinter) Flush() {}

// SetPrefix implements Printer SetPrefix method
func (stdPrinter) SetPrefix(prefix string) { std.SetPrefix(prefix) }

// SetCaller implements Printer SetCaller method
func (stdPrinter) SetCaller(yes bool) {
	std.SetFlags(std.Flags() & (^std.Llongfile))
	if yes {
		std.SetFlags(std.Flags() | std.Lshortfile)
	} else {
		std.SetFlags(std.Flags() & (^std.Lshortfile))
	}
}

// GetLevel implements Printer GetLevel method
func (p stdPrinter) GetLevel() Level { return Level(atomic.LoadInt32((*int32)(&p.level))) }

// SetLevel implements Printer SetLevel method
func (p *stdPrinter) SetLevel(level Level) { atomic.StoreInt32((*int32)(&p.level), int32(level)) }

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

// global printer
var gprinter = newStdPrinter()

type startOptions struct {
	httpHandler bool
	caller      bool
	sync        bool
	level       Level
	prefix      string
	printer     Printer
	writers     []Writer
}

func (opt *startOptions) apply(options []Option) {
	for i := range options {
		if options[i] != nil {
			options[i](opt)
		}
	}
}

// Option is option for Start
type Option func(*startOptions)

// WithSync synchronize outputs log or not
func WithSync(yes bool) Option {
	return func(opt *startOptions) {
		opt.sync = yes
	}
}

// WithHTTPHandler enable or disable http handler for settting level
func WithHTTPHandler(yes bool) Option {
	return func(opt *startOptions) {
		opt.httpHandler = yes
	}
}

// WithCaller enable or disable caller information
func WithCaller(yes bool) Option {
	return func(opt *startOptions) {
		opt.caller = yes
	}
}

// WithLevel sets log level
func WithLevel(level Level) Option {
	return func(opt *startOptions) {
		opt.level = level
	}
}

// WithPrefix set log prefix
func WithPrefix(prefix string) Option {
	return func(opt *startOptions) {
		opt.prefix = prefix
	}
}

// WithPrinter specify custom printer
func WithPrinter(printer Printer) Option {
	if printer == nil {
		panic("log: with a nil printer")
	}
	return func(opt *startOptions) {
		if opt.printer != nil {
			panic("log: printer already specified")
		}
		if len(opt.writers) > 0 {
			panic("log: couldn't specify printer if any writer specified")
		}
		opt.printer = printer
	}
}

// WithWriters appends the writers
func WithWriters(writers ...Writer) Option {
	if len(writers) == 0 {
		return nil
	}
	for i, writer := range writers {
		if writer == nil {
			panic("log: with a nil(" + strconv.Itoa(i+1) + "th) writer")
		}
	}
	var copied = make([]Writer, len(writers))
	copy(copied, writers)
	return func(opt *startOptions) {
		if opt.printer != nil {
			panic("log: couldn't specify writer if a printer specified")
		}
		opt.writers = append(opt.writers, copied...)
	}
}

// WithConsle appends a console writer
func WithConsle() Option {
	return WithWriters(newConsole())
}

// WithFile appends a file writer
func WithFile(fileOptions FileOptions) Option {
	return WithWriters(newFile(fileOptions))
}

// WithMultiFile appends a multifile writer
func WithMultiFile(multiFileOptions MultiFileOptions) Option {
	return WithWriters(newMultiFile(multiFileOptions))
}

// Start starts logging with options
func Start(options ...Option) error {
	var opt startOptions
	opt.apply(options)
	async := !opt.sync
	changed := true
	if opt.printer == nil {
		switch len(opt.writers) {
		case 0:
			opt.printer = gprinter
			changed = false
		case 1:
			opt.printer = newPrinter(opt.writers[0], async)
		default:
			opt.printer = newPrinter(multiWriter{opt.writers}, async)
		}
	}
	if opt.level != 0 {
		opt.printer.SetLevel(opt.level)
	}
	opt.printer.SetPrefix(opt.prefix)
	opt.printer.SetCaller(opt.caller)

	if changed {
		gprinter.Shutdown()
		gprinter = opt.printer
		gprinter.Start()
	}
	if opt.httpHandler {
		registerHTTPHandlers()
	}
	return nil
}

// Shutdown shutdowns global printer
func Shutdown() {
	gprinter.Shutdown()
}

// SetCaller sets whether print caller information
func SetCaller(yes bool) {
	gprinter.SetCaller(yes)
}

// GetLevel gets current log level
func GetLevel() Level {
	return gprinter.GetLevel()
}

// SetLevel sets current log level
func SetLevel(level Level) {
	gprinter.SetLevel(level)
}

// Trace prints log with trace level
func Trace(format string, args ...interface{}) {
	gprinter.Printf(1, LvTRACE, "", format, args...)
}

// Debug prints log with level debug
func Debug(format string, args ...interface{}) {
	gprinter.Printf(1, LvDEBUG, "", format, args...)
}

// Info prints log with level info
func Info(format string, args ...interface{}) {
	gprinter.Printf(1, LvINFO, "", format, args...)
}

// Warn prints log with level warning
func Warn(format string, args ...interface{}) {
	gprinter.Printf(1, LvWARN, "", format, args...)
}

// Error prints log with level error
func Error(format string, args ...interface{}) {
	gprinter.Printf(1, LvERROR, "", format, args...)
}

// Fatal prints log with level fatal
func Fatal(format string, args ...interface{}) {
	gprinter.Printf(1, LvFATAL, "", format, args...)
}

// Printf wraps the global printer Printf method
func Printf(calldepth int, level Level, prefix, format string, args ...interface{}) {
	gprinter.Printf(calldepth, level, prefix, format, args...)
}

// NewLogger creates a logger with a extra prefix print logging to global printer
func NewLogger(prefix string) Logger {
	return logger{prefix: prefix}
}

// Logger represents basic logger interface
type Logger interface {
	// Trace outputs trace-level logs
	Trace(format string, args ...interface{})
	// Debug outputs debug-level logs
	Debug(format string, args ...interface{})
	// Info outputs info-level logs
	Info(format string, args ...interface{})
	// Warn outputs warn-level logs
	Warn(format string, args ...interface{})
	// Error outputs error-level logs
	Error(format string, args ...interface{})
	// Fatal outputs fatal-level logs
	Fatal(format string, args ...interface{})
}

// logger implements Logger to prints logging with extra prefix to global printer
type logger struct {
	prefix string
}

// Trace implements Logger Trace method
func (l logger) Trace(format string, args ...interface{}) {
	gprinter.Printf(1, LvTRACE, l.prefix, format, args...)
}

// Debug implements Logger Debug method
func (l logger) Debug(format string, args ...interface{}) {
	gprinter.Printf(1, LvDEBUG, l.prefix, format, args...)
}

// Info implements Logger Info method
func (l logger) Info(format string, args ...interface{}) {
	gprinter.Printf(1, LvINFO, l.prefix, format, args...)
}

// Warn implements Logger Warn method
func (l logger) Warn(format string, args ...interface{}) {
	gprinter.Printf(1, LvWARN, l.prefix, format, args...)
}

// Error implements Logger Error method
func (l logger) Error(format string, args ...interface{}) {
	gprinter.Printf(1, LvERROR, l.prefix, format, args...)
}

// Fatal implements Logger Fatal method
func (l logger) Fatal(format string, args ...interface{}) {
	gprinter.Printf(1, LvFATAL, l.prefix, format, args...)
}
