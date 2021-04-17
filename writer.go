package log

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	pid           = os.Getpid()
	errNilWriter  = errors.New("nil writer")
	errOutOfRange = errors.New("out of range")
)

// Writer represents a writer for logging
type Writer interface {
	Write(level Level, data []byte, headerLen int) error
	Close() error
}

type mixWriter struct {
	writers []Writer
}

// Write writes log to all inner writers
func (p mixWriter) Write(level Level, data []byte, headerLen int) error {
	var lastErr error
	for _, op := range p.writers {
		if err := op.Write(level, data, headerLen); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close closes all inner writers
func (p mixWriter) Close() error {
	var lastErr error
	for _, op := range p.writers {
		if err := op.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// console is a writer that writes logs to console
type console struct {
	toStderrLevel Level     // level which write to stderr from
	stdout        io.Writer // os.Stdout used if nil
	stderr        io.Writer // os.Stderr used if nil
}

// newConsole creates a console writer
func newConsole() *console {
	return &console{
		toStderrLevel: LvWARN,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
	}
}

// Write implements Writer Write method
func (p *console) Write(level Level, data []byte, _ int) error {
	if level <= p.toStderrLevel {
		_, err := p.stderr.Write(data)
		return err
	}
	_, err := p.stdout.Write(data)
	return err
}

// Close implements Writer Close method
func (p *console) Close() error { return nil }

// FileHeader represents header type of file
type FileHeader int

const (
	NoHeader   FileHeader = 0
	HTMLHeader FileHeader = 1
)

var fileHeaders = map[FileHeader]string{
	HTMLHeader: `<br/><head>
	<meta charset="UTF-8">
	<style>
		@media screen and (min-width: 1000px) {
			.item { width: 950px; padding-top: 6px; padding-bottom: 12px; padding-left: 24px; padding-right: 16px; }
			.metadata { font-size: 18px; }
			.content { font-size: 16px; }
			.datetime { font-size: 14px; }
		}
		@media screen and (max-width: 1000px) {
			.item { width: 95%; padding-top: 4px; padding-bottom: 8px; padding-left: 16px; padding-right: 12px; }
			.metadata { font-size: 14px; }
			.content { font-size: 13px; }
			.datetime { font-size: 12px; }
		}
		.item { max-width: 95%; box-shadow: rgba(60,64,67,.3) 0 1px 2px 0, rgba(60, 64, 67, .15) 0 1px 3px 1px; background: white; border-radius: 4px; margin: 20px auto; }
		.datetime { color: #00000080; display: block; }
		.metadata { color: #df005f; }
		pre {
			white-space: pre-wrap;       /* Since CSS 2.1 */
			white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
			white-space: -pre-wrap;      /* Opera 4-6 */
			white-space: -o-pre-wrap;    /* Opera 7 */
			word-wrap: break-word;       /* Internet Explorer 5.5+ */
		}
	</style>
</head>`,
}

// FileOptions represents options of file writer
type FileOptions struct {
	Dir          string     `json:"dir"`          // log directory (default: .)
	Filename     string     `json:"filename"`     // log filename (default: <appName>.log)
	SymlinkedDir string     `json:"symlinkeddir"` // symlinked directory is symlink enabled (default: symlinked)
	NoSymlink    bool       `json:"nosymlink"`    // doesn't create symlink to latest log file (default: false)
	MaxSize      int        `json:"maxsize"`      // max bytes number of every log file(default: 64M)
	Rotate       bool       `json:"rotate"`       // enable log rotate (default: no)
	Suffix       string     `json:"suffix"`       // filename suffixa(default: .log)
	DateFormat   string     `json:"dateformat"`   // date format string (default: %04d%02d%02d)
	Header       FileHeader `json:"header"`       // header type of file (default: NoHeader)
}

func (opts *FileOptions) setDefaults() {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.MaxSize == 0 {
		opts.MaxSize = 1 << 26 // 64M
	}
	if opts.DateFormat == "" {
		opts.DateFormat = "%04d%02d%02d"
	}
	if opts.Suffix == "" {
		opts.Suffix = ".log"
	}
	if opts.SymlinkedDir == "" {
		opts.SymlinkedDir = "symlinked"
	}
}

// file is a writer which writes logs to file
type file struct {
	options          FileOptions
	currentSize      int
	createdTime      time.Time
	fileIndex        int
	onceCreateLogDir sync.Once

	mu      sync.Mutex
	writer  *bufio.Writer
	file    *os.File
	written bool
}

func newFile(options FileOptions) *file {
	options.setDefaults()
	p := &file{
		options:   options,
		fileIndex: -1,
	}
	p.rotate(time.Now())
	go func(f *file) {
		for range time.Tick(time.Second) {
			f.mu.Lock()
			if f.written {
				f.writer.Flush()
				f.file.Sync()
				f.written = false
			}
			f.mu.Unlock()
		}
	}(p)
	return p
}

// Write writes log to file
func (p *file) Write(level Level, data []byte, _ int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer == nil {
		return errNilWriter
	}
	now := time.Now()
	if !isSameDay(now, p.createdTime) {
		if err := p.rotate(now); err != nil {
			return err
		}
	}
	n, err := p.writer.Write(data)
	p.written = true
	p.currentSize += n
	if p.currentSize >= p.options.MaxSize {
		p.rotate(now)
	}
	return err
}

func (p *file) closeCurrent() error {
	if p.writer != nil {
		if err := p.writer.Flush(); err != nil {
			return err
		}
		if err := p.file.Sync(); err != nil {
			return err
		}
		if err := p.file.Close(); err != nil {
			return err
		}
		p.written = false
	}
	p.currentSize = 0
	return nil
}

// Close closes current log file
func (p *file) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeCurrent()
}

func (p *file) rotate(now time.Time) error {
	p.closeCurrent()
	if isSameDay(now, p.createdTime) {
		p.fileIndex = (p.fileIndex + 1) % 1000
	} else {
		p.fileIndex = 0
	}
	p.createdTime = now

	var err error
	p.file, err = p.create()
	if err != nil {
		return err
	}

	p.writer = bufio.NewWriterSize(p.file, 1<<14) // 16k
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "File opened at: %s.\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s.\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if header, ok := fileHeaders[p.options.Header]; ok {
		fmt.Fprintln(&buf, header)
	}
	n, err := p.file.Write(buf.Bytes())
	p.currentSize += n
	p.writer.Flush()
	p.file.Sync()
	return err
}

func (p *file) create() (*os.File, error) {
	p.onceCreateLogDir.Do(p.createDir)

	// make filename
	var (
		y, m, d = p.createdTime.Date()
		name    string
		prefix  = p.options.Filename
		date    = fmt.Sprintf(p.options.DateFormat, y, m, d)
	)
	if p.options.Filename != "" {
		prefix += "."
	}
	if p.options.Rotate {
		name = fmt.Sprintf("%s%s", prefix, date)
	} else {
		H, M, _ := p.createdTime.Clock()
		name = fmt.Sprintf("%s%s-%02d%02d.%06d", prefix, date, H, M, pid)
	}
	if p.fileIndex > 0 {
		name = fmt.Sprintf("%s.%03d", name, p.fileIndex)
	}
	if !strings.HasSuffix(name, p.options.Suffix) {
		name += p.options.Suffix
	}

	// create file
	var (
		fullname = filepath.Join(p.options.Dir, name)
		f        *os.File
		err      error
	)
	if !p.options.NoSymlink {
		fullname = filepath.Join(p.options.Dir, p.options.SymlinkedDir, name)
	}
	if p.options.Rotate {
		f, err = os.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = os.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	}
	if err == nil && !p.options.NoSymlink {
		tmp := p.options.Filename
		if tmp == "" {
			tmp = filepath.Base(os.Args[0])
		}
		symlink := filepath.Join(p.options.Dir, tmp+p.options.Suffix)
		os.Remove(symlink)
		os.Symlink(filepath.Join(p.options.SymlinkedDir, name), symlink)
	}
	return f, err
}

func (p *file) createDir() {
	os.MkdirAll(p.options.Dir, 0755)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

type MultiFileOptions struct {
	RootDir      string `json:"rootdir"`      // log directory (default: .)
	ErrorDir     string `json:"errordir"`     // error subdirectory (default: error)
	WarnDir      string `json:"warndir"`      // warn subdirectory (default: warn)
	InfoDir      string `json:"infodir"`      // info subdirectory (default: info)
	DebugDir     string `json:"debugdir"`     // debug subdirectory (default: debug)
	TraceDir     string `json:"tracedir"`     // trace subdirectory (default: trace)
	Filename     string `json:"filename"`     // log filename (default: <appName>.log)
	SymlinkedDir string `json:"symlinkeddir"` // symlinked directory is symlink enabled (default: symlinked)
	NoSymlink    bool   `json:"nosymlink"`    // doesn't create symlink to latest log file (default: false)
	MaxSize      int    `json:"maxsize"`      // max bytes number of every log file (default: 64M)
	Rotate       bool   `json:"rotate"`       // enable log rotate (default: true)
	Suffix       string `json:"suffix"`       // filename suffix (default: log)
	DateFormat   string `json:"dateformat"`   // date format string (default: %04d%02d%02d)
}

func (opts *MultiFileOptions) setDefaults() {
	if opts.RootDir == "" {
		opts.RootDir = "."
	}
	if opts.ErrorDir == "" {
		opts.ErrorDir = "error"
	}
	if opts.WarnDir == "" {
		opts.WarnDir = "warn"
	}
	if opts.InfoDir == "" {
		opts.InfoDir = "info"
	}
	if opts.DebugDir == "" {
		opts.DebugDir = "debug"
	}
	if opts.TraceDir == "" {
		opts.TraceDir = "trace"
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = 1 << 26
	}
	if opts.DateFormat == "" {
		opts.DateFormat = "%04d%02d%02d"
	}
}

type multiFile struct {
	options MultiFileOptions
	files   [numLevel]*file
	group   map[string][]Level
}

func abs(path string) string {
	s, _ := filepath.Abs(path)
	return s
}

func newMultiFile(options MultiFileOptions) *multiFile {
	options.setDefaults()
	p := new(multiFile)
	p.options = options
	dirs := map[Level]string{
		LvTRACE: abs(filepath.Join(p.options.RootDir, p.options.TraceDir)),
		LvDEBUG: abs(filepath.Join(p.options.RootDir, p.options.DebugDir)),
		LvINFO:  abs(filepath.Join(p.options.RootDir, p.options.InfoDir)),
		LvWARN:  abs(filepath.Join(p.options.RootDir, p.options.WarnDir)),
		LvERROR: abs(filepath.Join(p.options.RootDir, p.options.ErrorDir)),
		LvFATAL: abs(filepath.Join(p.options.RootDir, p.options.ErrorDir)),
	}
	p.group = map[string][]Level{}
	for lv, dir := range dirs {
		if levels, ok := p.group[dir]; ok {
			p.group[dir] = append(levels, lv)
		} else {
			p.group[dir] = []Level{lv}
		}
	}
	return p
}

func (p *multiFile) Write(level Level, data []byte, headerLen int) error {
	if p.files[level] == nil {
		if err := p.initForLevel(level); err != nil {
			return err
		}
	}
	return p.files[level].Write(level, data, headerLen)
}

func (p *multiFile) Close() error {
	var lastErr error
	for i := range p.files {
		if i == 0 {
			continue
		}
		if p.files[i] != nil {
			if err := p.files[i].Close(); err != nil {
				lastErr = err
			}
			p.files[i] = nil
		}
	}
	return lastErr
}

func (p *multiFile) initForLevel(level Level) error {
	if level < 0 || int(level) >= len(p.files) {
		return errOutOfRange
	}
	f := newFile(p.optionsOfLevel(level))
	p.files[level] = f
	if levels, ok := p.group[abs(f.options.Dir)]; ok {
		for _, lv := range levels {
			if p.files[lv] == nil {
				p.files[lv] = f
			}
		}
	}
	return nil
}

func (p *multiFile) optionsOfLevel(level Level) FileOptions {
	options := FileOptions{
		MaxSize:    p.options.MaxSize,
		NoSymlink:  p.options.NoSymlink,
		Filename:   p.options.Filename,
		Rotate:     p.options.Rotate,
		Suffix:     p.options.Suffix,
		DateFormat: p.options.DateFormat,
	}
	switch level {
	case LvFATAL, LvERROR:
		options.Dir = filepath.Join(p.options.RootDir, p.options.ErrorDir)
	case LvWARN:
		options.Dir = filepath.Join(p.options.RootDir, p.options.WarnDir)
	case LvINFO:
		options.Dir = filepath.Join(p.options.RootDir, p.options.InfoDir)
	case LvDEBUG:
		options.Dir = filepath.Join(p.options.RootDir, p.options.DebugDir)
	default:
		options.Dir = filepath.Join(p.options.RootDir, p.options.TraceDir)
	}
	return options
}
