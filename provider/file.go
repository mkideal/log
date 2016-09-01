package provider

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/mkideal/log/logger"
)

func init() {
	logger.Register("file", NewFile)
}

var (
	pid = os.Getpid()
)

// FileOpts represents options object of file provider
type FileOpts struct {
	Dir         string `json:"dir"`          // log directory(default: .)
	Filename    string `json:"filename"`     // log filename(default: <appName>.log)
	NoSymlink   bool   `json:"nosymlink"`    // doesn't create symlink to latest log file(default: false)
	MaxSize     int    `json:"maxsize"`      // max bytes number of every log file(default: 64M)
	DailyAppend bool   `json:"daily_append"` // append to existed file instead of creating a new file(default: true)
}

// NewFileOpts ...
func NewFileOpts() FileOpts {
	opts := FileOpts{}
	opts.setDefaults()
	opts.DailyAppend = true
	return opts
}

func (opts *FileOpts) setDefaults() {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Filename == "" {
		_, appName := filepath.Split(os.Args[0])
		opts.Filename = appName + ".log"
	}
	if opts.MaxSize == 0 {
		opts.MaxSize = 1 << 26 // 64M
	}
}

// File is provider that writes logs to file
type File struct {
	config           FileOpts
	currentSize      int
	createdTime      time.Time
	fileIndex        int
	onceCreateLogDir sync.Once

	mu      sync.Mutex
	writer  *bufio.Writer
	file    *os.File
	written bool
}

// NewFile creates file provider
func NewFile(opts string) logger.Provider {
	config := NewFileOpts()
	logger.UnmarshalOpts(opts, &config)
	config.setDefaults()
	return newFile(config)
}

func newFile(config FileOpts) *File {
	p := &File{
		config:    config,
		fileIndex: -1,
	}
	p.rotate(time.Now())
	go func(f *File) {
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
func (p *File) Write(level logger.Level, headerLength int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer == nil {
		return errWriterIsNil
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
	if p.currentSize >= p.config.MaxSize {
		p.rotate(now)
	}
	return err
}

func (p *File) closeCurrent() error {
	var err errorList
	if p.writer != nil {
		err.tryPush(p.writer.Flush())
		err.tryPush(p.file.Sync())
		err.tryPush(p.file.Close())
		p.written = false
	}
	p.currentSize = 0
	return err.err()
}

// Close closes current log file
func (p *File) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeCurrent()
}

func (p *File) rotate(now time.Time) error {
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
	fmt.Fprintf(&buf, "File opened at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	n, err := p.file.Write(buf.Bytes())
	p.currentSize += n
	p.writer.Flush()
	p.file.Sync()
	return err
}

func (p *File) create() (*os.File, error) {
	p.onceCreateLogDir.Do(p.createDir)

	// make filename
	y, m, d := p.createdTime.Date()
	var name string
	if p.config.DailyAppend {
		name = fmt.Sprintf("%s.%04d%02d%02d", p.config.Filename, y, m, d)
	} else {
		H, M, _ := p.createdTime.Clock()
		name = fmt.Sprintf("%s.%04d%02d%02d-%02d%02d.%06d", p.config.Filename, y, m, d, H, M, pid)
	}
	if p.fileIndex > 0 {
		name = fmt.Sprintf("%s.%03d", name, p.fileIndex)
	}

	// create file
	var (
		fullname = filepath.Join(p.config.Dir, name)
		f        *os.File
		err      error
	)
	if p.config.DailyAppend {
		f, err = os.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = os.OpenFile(fullname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	}
	if err == nil && !p.config.NoSymlink {
		symlink := filepath.Join(p.config.Dir, p.config.Filename)
		os.Remove(symlink)
		os.Symlink(name, symlink)
	}
	return f, err
}

func (p *File) createDir() {
	os.MkdirAll(p.config.Dir, 0755)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
