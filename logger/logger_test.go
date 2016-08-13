package logger

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogLevel(t *testing.T) {
	assert.Equal(t, "TRACE", TRACE.String())
	assert.Equal(t, "DEBUG", DEBUG.String())
	assert.Equal(t, "INFO", INFO.String())
	assert.Equal(t, "WARN", WARN.String())
	assert.Equal(t, "ERROR", ERROR.String())

	assert.Equal(t, TRACE, MustParseLogLevel("TRACE"))
	assert.Equal(t, DEBUG, MustParseLogLevel("DEBUG"))
	assert.Equal(t, INFO, MustParseLogLevel("INFO"))
	assert.Equal(t, WARN, MustParseLogLevel("WARN"))
	assert.Equal(t, ERROR, MustParseLogLevel("ERROR"))

	assert.Equal(t, TRACE, MustParseLogLevel("trace"))
	assert.Equal(t, DEBUG, MustParseLogLevel("debug"))
	assert.Equal(t, INFO, MustParseLogLevel("info"))
	assert.Equal(t, WARN, MustParseLogLevel("warn"))
	assert.Equal(t, ERROR, MustParseLogLevel("error"))

	assert.Equal(t, TRACE, MustParseLogLevel("T"))
	assert.Equal(t, DEBUG, MustParseLogLevel("D"))
	assert.Equal(t, INFO, MustParseLogLevel("I"))
	assert.Equal(t, WARN, MustParseLogLevel("W"))
	assert.Equal(t, ERROR, MustParseLogLevel("E"))

	assert.Equal(t, TRACE, MustParseLogLevel("t"))
	assert.Equal(t, DEBUG, MustParseLogLevel("d"))
	assert.Equal(t, INFO, MustParseLogLevel("i"))
	assert.Equal(t, WARN, MustParseLogLevel("w"))
	assert.Equal(t, ERROR, MustParseLogLevel("e"))

	assert.Panics(t, func() { MustParseLogLevel("xTrace") })
	assert.Panics(t, func() { MustParseLogLevel("xDebug") })
	assert.Panics(t, func() { MustParseLogLevel("xInfo") })
	assert.Panics(t, func() { MustParseLogLevel("xWarn") })
	assert.Panics(t, func() { MustParseLogLevel("xError") })
}

func TestFormatHeader(t *testing.T) {
	l := newLogger(nil)
	for expected, now := range map[string]time.Time{
		"[T 2000/01/02 03:04:05.006 ???:0] ": time.Date(2000, 1, 2, 3, 4, 5, 6E6, time.Local),
		"[T 2000/11/22 11:22:11.022 ???:0] ": time.Date(2000, 11, 22, 11, 22, 11, 22E6, time.Local),
	} {
		buf := l.formatHeader(now, TRACE, "???", 0)
		if got := buf.String(); got != expected {
			t.Errorf("unexpected formatHeader: `%s' vs `%s'", got, expected)
		}
	}
}

type mockProvider struct {
	sync.WaitGroup
	data *bytes.Buffer
}

func newMockProvider() *mockProvider {
	return &mockProvider{data: bytes.NewBufferString("")}
}

func (p *mockProvider) Write(level LogLevel, data []byte) error {
	index := bytes.LastIndex(data, []byte("] "))
	if index >= 0 {
		data = data[index+2:]
	}
	_, err := p.data.Write(data)
	p.Done()
	return err
}

func TestLogPrint(t *testing.T) {
	p := newMockProvider()
	l := newLogger(p)
	l.Run()

	print := func() {
		l.Trace("hello %v", TRACE)
		l.Debug("hello %v", DEBUG)
		l.Info("hello %v", INFO)
		l.Warn("hello %v", WARN)
		l.Error("hello %v", ERROR)
	}

	p.Add(5 + 4 + 3 + 2 + 1)
	for _, level := range []LogLevel{TRACE, DEBUG, INFO, WARN, ERROR} {
		l.SetLevel(level)
		print()
	}

	expected := `hello TRACE
hello DEBUG
hello INFO
hello WARN
hello ERROR
hello DEBUG
hello INFO
hello WARN
hello ERROR
hello INFO
hello WARN
hello ERROR
hello WARN
hello ERROR
hello ERROR
`

	p.Wait()
	if got := p.data.String(); got != expected {
		t.Errorf("unexpected print: `%s' vs `%s'", got, expected)
	}
}
