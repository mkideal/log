package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/mkideal/log/logger"
)

var (
	ErrWriterIsNil = errors.New("writer is nil")

	onceCreateLogDir sync.Once
	pid              = os.Getpid()
)

type fileConfig struct {
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
	MaxSize  int    `json:"max_size"`
}

func newFileConfig() fileConfig {
	return fileConfig{
		Dir:      ".",
		Filename: os.Args[0] + ".log",
		MaxSize:  1 << 26, // 64M
	}
}

type File struct {
	config      fileConfig
	currentSize int
	createdTime time.Time
	fileIndex   int

	mu     sync.Mutex
	writer *bufio.Writer
	file   *os.File
}

func NewFile(opts string) logger.Provider {
	config := newFileConfig()
	json.Unmarshal([]byte(opts), &config)
	p := &File{
		config:    config,
		fileIndex: -1,
	}
	p.rotate(time.Now())
	go func(f *File) {
		for range time.Tick(time.Second) {
			f.mu.Lock()
			f.writer.Flush()
			f.file.Sync()
			f.mu.Unlock()
		}
	}(p)
	return p
}

func (p *File) Write(level logger.LogLevel, headerLength int, data []byte) error {
	if p.writer == nil {
		return ErrWriterIsNil
	}
	now := time.Now()
	if !isSameDay(now, p.createdTime) {
		if err := p.rotate(now); err != nil {
			return err
		}
	}
	p.mu.Lock()
	n, err := p.writer.Write(data)
	p.mu.Unlock()
	p.currentSize += n
	if p.currentSize >= p.config.MaxSize {
		p.rotate(now)
	}
	return err
}

func (p *File) Close() {
	p.mu.Lock()
	if p.writer != nil {
		p.writer.Flush()
		p.file.Sync()
		p.file.Close()
	}
	p.mu.Unlock()
}

func (p *File) rotate(now time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer != nil {
		p.writer.Flush()
		p.file.Sync()
		p.file.Close()
	}
	p.currentSize = 0
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
	fmt.Fprintf(&buf, "File created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	n, err := p.file.Write(buf.Bytes())
	p.currentSize += n
	return err
}

func (p *File) create() (*os.File, error) {
	onceCreateLogDir.Do(p.createDir)
	year, month, day := p.createdTime.Date()
	hour, minute, _ := p.createdTime.Clock()
	name := fmt.Sprintf("%s.%04d%02d%02d-%02d%02d.%06d.%03d", p.config.Filename, year, int(month), day, hour, minute, pid, p.fileIndex)
	fullname := filepath.Join(p.config.Dir, name)
	f, err := os.Create(fullname)
	if err == nil {
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
