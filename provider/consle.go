package provider

import (
	"io"
	"os"

	"github.com/mkideal/log/logger"
)

type Console struct {
	stdout io.Writer
	stderr io.Writer
}

func NewConsole() logger.Provider {
	return NewConsoleWithWriter(os.Stdout, os.Stderr)
}

func NewConsoleWithWriter(stdout, stderr io.Writer) logger.Provider {
	return &Console{
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *Console) Write(level logger.Level, headerLength int, data []byte) error {
	if level == logger.ERROR {
		_, err := p.stderr.Write(data)
		return err
	}
	_, err := p.stdout.Write(data)
	return err
}

func (p *Console) Close() {}
