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
	opts := MultiFileOpts{
		RootDir: ".",
	}
	opts.setDefaults()
	return opts
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
	files  [logger.NumLevel]*File
	group  map[string][]logger.Level
}

func abs(path string) string {
	s, _ := filepath.Abs(path)
	return s
}

func NewMultiFile(opts string) logger.Provider {
	p := new(MultiFile)
	p.config = NewMultiFileOpts()
	logger.UnmarshalOpts(opts, &p.config)
	p.config.setDefaults()
	dirs := map[logger.Level]string{
		logger.TRACE: abs(filepath.Join(p.config.RootDir, p.config.TraceDir)),
		logger.DEBUG: abs(filepath.Join(p.config.RootDir, p.config.DebugDir)),
		logger.INFO:  abs(filepath.Join(p.config.RootDir, p.config.InfoDir)),
		logger.WARN:  abs(filepath.Join(p.config.RootDir, p.config.WarnDir)),
		logger.ERROR: abs(filepath.Join(p.config.RootDir, p.config.ErrorDir)),
		logger.FATAL: abs(filepath.Join(p.config.RootDir, p.config.ErrorDir)),
	}
	p.group = map[string][]logger.Level{}
	for lv, dir := range dirs {
		if levels, ok := p.group[dir]; ok {
			p.group[dir] = append(levels, lv)
		} else {
			p.group[dir] = []logger.Level{lv}
		}
	}
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
	f := newFile(p.configForLevel(level))
	p.files[level] = f
	if levels, ok := p.group[abs(f.config.Dir)]; ok {
		for _, lv := range levels {
			if p.files[lv] == nil {
				p.files[lv] = f
			}
		}
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
