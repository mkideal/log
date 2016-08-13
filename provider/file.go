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

	pid = os.Getpid()
)

type FileConfig struct {
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
	MaxSize  int    `json:"maxsize"`
}

func NewFileConfig() FileConfig {
	_, appName := filepath.Split(os.Args[0])
	return FileConfig{
		Dir:      ".",
		Filename: appName + ".log",
		MaxSize:  1 << 26, // 64M
	}
}

type File struct {
	config           FileConfig
	currentSize      int
	createdTime      time.Time
	fileIndex        int
	onceCreateLogDir sync.Once

	mu      sync.Mutex
	writer  *bufio.Writer
	file    *os.File
	written bool
}

func NewFile(opts string) logger.Provider {
	config := NewFileConfig()
	json.Unmarshal([]byte(opts), &config)
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

func (p *File) Write(level logger.Level, headerLength int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer == nil {
		return ErrWriterIsNil
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

func (p *File) closeCurrent() {
	if p.writer != nil {
		p.writer.Flush()
		p.file.Sync()
		p.file.Close()
		p.written = false
	}
	p.currentSize = 0
}

func (p *File) Close() {
	p.mu.Lock()
	p.closeCurrent()
	p.mu.Unlock()
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
	fmt.Fprintf(&buf, "File created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	n, err := p.file.Write(buf.Bytes())
	p.currentSize += n
	p.writer.Flush()
	p.file.Sync()
	return err
}

func (p *File) create() (*os.File, error) {
	p.onceCreateLogDir.Do(p.createDir)
	y, m, d := p.createdTime.Date()
	H, M, _ := p.createdTime.Clock()
	name := fmt.Sprintf("%s.%04d%02d%02d-%02d%02d.%06d.%03d", p.config.Filename, y, m, d, H, M, pid, p.fileIndex)
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
