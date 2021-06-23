package provider

import (
	"io"
	"os"

	"github.com/mkideal/log/logger"
)

func init() {
	logger.Register("console", NewConsole)
}

// ConsoleOpts represents options object of console provider
type ConsoleOpts struct {
	ToStderrLevel logger.Level `json:"tostderrlevel"` // level which write to stderr from
}

// NewConsoleOpts ...
func NewConsoleOpts() ConsoleOpts {
	return ConsoleOpts{ToStderrLevel: logger.ERROR}
}

// Console is a provider that writes logs to console
type Console struct {
	config ConsoleOpts
	stdout io.Writer
	stderr io.Writer
}

// NewConsole creates a console provider
func NewConsole(opts string) logger.Provider {
	return NewConsoleWithWriter(opts, os.Stdout, os.Stderr)
}

// NewConsoleWithWriter creates a  console provider by specified writers
func NewConsoleWithWriter(opts string, stdout, stderr io.Writer) logger.Provider {
	config := NewConsoleOpts()
	logger.UnmarshalOpts(opts, &config)
	return &Console{
		config: config,
		stdout: stdout,
		stderr: stderr,
	}
}

// Write implements Provider.Write method
func (p *Console) Write(level logger.Level, headerLength int, data []byte) error {
	if level <= p.config.ToStderrLevel {
		_, err := p.stderr.Write(data)
		return err
	}
	_, err := p.stdout.Write(data)
	return err
}

// Close implements Provider.Close method
func (p *Console) Close() error { return nil }
