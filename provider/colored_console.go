package provider

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mkideal/log/logger"
)

func init() {
	logger.Register("colored_console", NewColoredConsole)
}

// Console is a provider that writes logs to console
type ColoredConsole struct {
	config          ConsoleOpts
	stdout          io.Writer
	stderr          io.Writer
	stdoutColorable bool
	stderrColorable bool
}

// NewConsole creates a console provider
func NewColoredConsole(opts string) logger.Provider {
	config := NewConsoleOpts()
	logger.UnmarshalOpts(opts, &config)
	p := &ColoredConsole{
		config: config,
		stdout: colorable.NewColorableStdout(),
		stderr: colorable.NewColorableStderr(),
	}
	if f, ok := p.stdout.(*os.File); ok {
		p.stdoutColorable = isatty.IsTerminal(f.Fd())
	}
	if f, ok := p.stderr.(*os.File); ok {
		p.stderrColorable = isatty.IsTerminal(f.Fd())
	}
	return p
}

// Write implements Provider.Write method
func (p *ColoredConsole) Write(level logger.Level, headerLength int, data []byte) error {
	writer := p.stdout
	colorable := p.stdoutColorable
	if level <= p.config.ToStderrLevel {
		writer = p.stderr
		colorable = p.stderrColorable
	}
	var cpre []byte
	if colorable {
		cpre = p.color(level)
		if cpre != nil {
			writer.Write(cpre)
		}
	}
	_, err := writer.Write(data)
	if colorable && err == nil && cpre != nil {
		writer.Write(colorEnd)
	}
	return err
}

// Close implements Provider.Close method
func (p *ColoredConsole) Close() error { return nil }

func (p *ColoredConsole) color(level logger.Level) []byte {
	switch level {
	case logger.FATAL:
		return mag
	case logger.ERROR:
		return red
	case logger.WARN:
		return yel
	case logger.INFO:
		return cyn
	case logger.DEBUG:
		return nil
	case logger.TRACE:
		return gry
	}
	return nil
}

var (
	mag = []byte("\x1b[35m")
	red = []byte("\x1b[31m")
	yel = []byte("\x1b[33m")
	cyn = []byte("\x1b[36m")
	gry = []byte("\x1b[90m")

	colorEnd = []byte("\x1b[0m")
)
