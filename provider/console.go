package provider

import (
	"io"
	"os"

	"github.com/mkideal/log/logger"
)

func init() {
	logger.Register("console", NewConsole)
}

type ConsoleOpts struct {
	ToStderrLevel logger.Level `json:"tostderrlevel"`
}

func NewConsoleOpts() ConsoleOpts {
	return ConsoleOpts{ToStderrLevel: logger.ERROR}
}

type Console struct {
	config ConsoleOpts
	stdout io.Writer
	stderr io.Writer
}

func NewConsole(opts string) logger.Provider {
	return NewConsoleWithWriter(opts, os.Stdout, os.Stderr)
}

func NewConsoleWithWriter(opts string, stdout, stderr io.Writer) logger.Provider {
	config := NewConsoleOpts()
	logger.UnmarshalOpts(opts, &config)
	return &Console{
		config: config,
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *Console) Write(level logger.Level, headerLength int, data []byte) error {
	if level <= p.config.ToStderrLevel {
		_, err := p.stderr.Write(data)
		return err
	}
	_, err := p.stdout.Write(data)
	return err
}

func (p *Console) Close() error { return nil }
