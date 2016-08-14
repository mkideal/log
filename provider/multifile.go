package provider

import (
	"os"
	"path/filepath"

	"github.com/mkideal/log/logger"
)

func init() {
	logger.Register("multifile", NewMultiFile)
}

type MultiFileOpts struct {
	RootDir   string `json:"rootdir"`   // log directory(default: .)
	ErrorDir  string `json:"errordir"`  // error subdirectory(default: error)
	WarnDir   string `json:"warndir"`   // warn subdirectory(default: warn)
	InfoDir   string `json:"infodir"`   // info subdirectory(default: info)
	DebugDir  string `json:"debugdir"`  // debug subdirectory(default: debug)
	TraceDir  string `json:"tracedir"`  // trace subdirectory(default: trace)
	Filename  string `json:"filename"`  // log filename(default: <appName>.log)
	NoSymlink bool   `json:"nosymlink"` // doesn't create symlink to latest log file(default: false)
	MaxSize   int    `json:"maxsize"`   // max bytes number of every log file(default: 64M)
}

func NewMultiFileOpts() MultiFileOpts {
	return MultiFileOpts{
		RootDir: ".",
	}
}

func (opts *MultiFileOpts) setDefaults() {
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
	if opts.Filename == "" {
		_, appName := filepath.Split(os.Args[0])
		opts.Filename = appName + ".log"
	}
}

type MultiFile struct {
	config MultiFileOpts
	files  [logger.LevelNum]*File
}

func NewMultiFile(opts string) logger.Provider {
	p := new(MultiFile)
	p.config = NewMultiFileOpts()
	logger.UnmarshalOpts(opts, &p.config)
	p.config.setDefaults()
	return p
}

func (p *MultiFile) Write(level logger.Level, headerLength int, data []byte) error {
	if p.files[level] == nil {
		if err := p.initForLevel(level); err != nil {
			return err
		}
	}
	return p.files[level].Write(level, headerLength, data)
}

func (p *MultiFile) Close() error {
	var errs errorList
	for i := range p.files {
		if i == 0 {
			continue
		}
		if p.files[i] != nil {
			errs.tryPush(p.files[i].Close())
			p.files[i] = nil
		}
	}
	return errs.err()
}

func (p *MultiFile) initForLevel(level logger.Level) error {
	if level < 0 || int(level) >= len(p.files) {
		return errOutOfRange
	}
	p.files[level] = newFile(p.configForLevel(level))
	if level == logger.FATAL {
		p.files[logger.ERROR] = p.files[level]
	} else if level == logger.ERROR {
		p.files[logger.FATAL] = p.files[level]
	}
	return nil
}

func (p *MultiFile) configForLevel(level logger.Level) FileOpts {
	config := FileOpts{
		MaxSize:   p.config.MaxSize,
		NoSymlink: p.config.NoSymlink,
		Filename:  p.config.Filename,
	}
	switch level {
	case logger.FATAL, logger.ERROR:
		config.Dir = filepath.Join(p.config.RootDir, p.config.ErrorDir)
	case logger.WARN:
		config.Dir = filepath.Join(p.config.RootDir, p.config.WarnDir)
	case logger.INFO:
		config.Dir = filepath.Join(p.config.RootDir, p.config.InfoDir)
	case logger.DEBUG:
		config.Dir = filepath.Join(p.config.RootDir, p.config.DebugDir)
	default:
		config.Dir = filepath.Join(p.config.RootDir, p.config.TraceDir)
	}
	return config
}
