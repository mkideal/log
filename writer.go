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

type multiWriter struct {
	writers []Writer
}

// Write writes log to all inner writers
func (w multiWriter) Write(level Level, data []byte, headerLen int) error {
	var lastErr error
	for i := range w.writers {
		if err := w.writers[i].Write(level, data, headerLen); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close closes all inner writers
func (w multiWriter) Close() error {
	var lastErr error
	for i := range w.writers {
		if err := w.writers[i].Close(); err != nil {
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
func (w *console) Write(level Level, data []byte, _ int) error {
	if level <= w.toStderrLevel {
		_, err := w.stderr.Write(data)
		return err
	}
	_, err := w.stdout.Write(data)
	return err
}

// Close implements Writer Close method
func (w *console) Close() error { return nil }

// File contains the basic writable file operations for logging
type File interface {
	io.WriteCloser
	// Sync commits the current contents of the file to stable storage.
	// Typically, this means flushing the file system's in-memory copy
	// of recently written data to disk.
	Sync() error
}

// FS wraps the basic fs operations for logging
type FS interface {
	OpenFile(name string, flag int, perm os.FileMode) (File, error) // OpenFile opens the file
	Remove(name string) error                                       // Remove removes the file
	Symlink(oldname, newname string) error                          // Symlink creates file symlink
	MkdirAll(path string, perm os.FileMode) error                   // MkdirAll creates a directory
}

// stdFS wraps the standard filesystem
type stdFS struct{}

var defaultFS stdFS

// OpenFile implements FS OpenFile method
func (fs stdFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, err := os.OpenFile(name, flag, perm)
	return f, err
}

// Remove implements FS Remove method
func (fs stdFS) Remove(name string) error { return os.Remove(name) }

// Symlink implements FS Symlink method
func (fs stdFS) Symlink(oldname, newname string) error { return os.Symlink(oldname, newname) }

// MkdirAll implements FS MkdirAll method
func (fs stdFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

// FileHeader represents header type of file
type FileHeader int

// FileHeader constants
const (
	NoHeader   FileHeader = 0 // no header in file
	HTMLHeader FileHeader = 1 // append html header in file
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
	DateFormat   string     `json:"dateformat"`   // date format string for filename (default: %04d%02d%02d)
	Header       FileHeader `json:"header"`       // header type of file (default: NoHeader)

	FS FS `json:"-"` // custom filesystem (default: stdFS)
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
	if opts.FS == nil {
		opts.FS = defaultFS
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
	file    File
	written bool
}

func newFile(options FileOptions) *file {
	options.setDefaults()
	w := &file{
		options:   options,
		fileIndex: -1,
	}
	w.rotate(time.Now())
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
	}(w)
	return w
}

// Write writes log to file
func (w *file) Write(level Level, data []byte, _ int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return errNilWriter
	}
	now := time.Now()
	if !isSameDay(now, w.createdTime) {
		if err := w.rotate(now); err != nil {
			return err
		}
	}
	n, err := w.writer.Write(data)
	w.written = true
	w.currentSize += n
	if w.currentSize >= w.options.MaxSize {
		w.rotate(now)
	}
	return err
}

func (w *file) closeCurrent() error {
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return err
		}
		if err := w.file.Sync(); err != nil {
			return err
		}
		if err := w.file.Close(); err != nil {
			return err
		}
		w.written = false
	}
	w.currentSize = 0
	return nil
}

// Close closes current log file
func (w *file) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeCurrent()
}

func (w *file) rotate(now time.Time) error {
	w.closeCurrent()
	if isSameDay(now, w.createdTime) {
		w.fileIndex = (w.fileIndex + 1) % 1000
	} else {
		w.fileIndex = 0
	}
	w.createdTime = now

	var err error
	w.file, err = w.create()
	if err != nil {
		return err
	}

	w.writer = bufio.NewWriterSize(w.file, 1<<14) // 16k
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "File opened at: %s.\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s.\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if header, ok := fileHeaders[w.options.Header]; ok {
		fmt.Fprintln(&buf, header)
	}
	n, err := w.file.Write(buf.Bytes())
	w.currentSize += n
	w.writer.Flush()
	w.file.Sync()
	return err
}

func (w *file) create() (File, error) {
	w.onceCreateLogDir.Do(w.createDir)

	// make filename
	var (
		y, m, d = w.createdTime.Date()
		name    string
		prefix  = w.options.Filename
		date    = fmt.Sprintf(w.options.DateFormat, y, m, d)
	)
	if w.options.Filename != "" {
		prefix += "."
	}
	if w.options.Rotate {
		name = fmt.Sprintf("%s%s", prefix, date)
	} else {
		H, M, _ := w.createdTime.Clock()
		name = fmt.Sprintf("%s%s-%02d%02d.%06d", prefix, date, H, M, pid)
	}
	if w.fileIndex > 0 {
		name = fmt.Sprintf("%s.%03d", name, w.fileIndex)
	}
	if !strings.HasSuffix(name, w.options.Suffix) {
		name += w.options.Suffix
	}

	// create file
	var (
		fullname = filepath.Join(w.options.Dir, name)
		f        File
		err      error
	)
	if !w.options.NoSymlink {
		fullname = filepath.Join(w.options.Dir, w.options.SymlinkedDir, name)
	}
	if w.options.Rotate {
		f, err = w.options.FS.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = w.options.FS.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	}
	if err == nil && !w.options.NoSymlink {
		tmp := w.options.Filename
		if tmp == "" {
			tmp = filepath.Base(os.Args[0])
		}
		symlink := filepath.Join(w.options.Dir, tmp+w.options.Suffix)
		w.options.FS.Remove(symlink)
		w.options.FS.Symlink(filepath.Join(w.options.SymlinkedDir, name), symlink)
	}
	return f, err
}

func (w *file) createDir() {
	w.options.FS.MkdirAll(w.options.Dir, 0755)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// MultiFileOptions represents options for multi file writer
type MultiFileOptions struct {
	FileOptions
	FatalDir string `json:"fataldir"` // fatal subdirectory (default: fatal)
	ErrorDir string `json:"errordir"` // error subdirectory (default: error)
	WarnDir  string `json:"warndir"`  // warn subdirectory (default: warn)
	InfoDir  string `json:"infodir"`  // info subdirectory (default: info)
	DebugDir string `json:"debugdir"` // debug subdirectory (default: debug)
	TraceDir string `json:"tracedir"` // trace subdirectory (default: trace)
}

func (opts *MultiFileOptions) setDefaults() {
	opts.FileOptions.setDefaults()
	if opts.FatalDir == "" {
		opts.FatalDir = "fatal"
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
}

type multiFile struct {
	options MultiFileOptions
	files   [numLevel]*file
	group   map[string][]Level
}

func absPath(path string) string {
	s, _ := filepath.Abs(path)
	return s
}

func newMultiFile(options MultiFileOptions) *multiFile {
	options.setDefaults()
	w := new(multiFile)
	w.options = options
	w.group = map[string][]Level{}
	for level := LvFATAL; level <= LvTRACE; level++ {
		dir := w.levelDir(level)
		if levels, ok := w.group[dir]; ok {
			w.group[dir] = append(levels, level)
		} else {
			w.group[dir] = []Level{level}
		}
	}
	return w
}

func (w *multiFile) Write(level Level, data []byte, headerLen int) error {
	if w.files[level.index()] == nil {
		if err := w.initForLevel(level); err != nil {
			return err
		}
	}
	return w.files[level.index()].Write(level, data, headerLen)
}

func (w *multiFile) Close() error {
	var lastErr error
	for i := range w.files {
		if w.files[i] != nil {
			if err := w.files[i].Close(); err != nil {
				lastErr = err
			}
			w.files[i] = nil
		}
	}
	return lastErr
}

func (w *multiFile) initForLevel(level Level) error {
	index := level.index()
	if index < 0 || index >= len(w.files) {
		return errOutOfRange
	}
	f := newFile(w.optionsOfLevel(level))
	w.files[index] = f
	if levels, ok := w.group[absPath(f.options.Dir)]; ok {
		for _, lv := range levels {
			if w.files[lv.index()] == nil {
				w.files[lv.index()] = f
			}
		}
	}
	return nil
}

func (w *multiFile) optionsOfLevel(level Level) FileOptions {
	options := w.options.FileOptions
	options.Dir = filepath.Join(options.Dir, w.levelDir(level))
	return options
}

func (w *multiFile) levelDir(level Level) string {
	switch level {
	case LvFATAL:
		return w.options.FatalDir
	case LvERROR:
		return w.options.ErrorDir
	case LvWARN:
		return w.options.WarnDir
	case LvINFO:
		return w.options.InfoDir
	case LvDEBUG:
		return w.options.DebugDir
	default:
		return w.options.TraceDir
	}
}
