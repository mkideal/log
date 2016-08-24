package logger

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLevel(t *testing.T) {
	assert.Equal(t, "TRACE", TRACE.String())
	assert.Equal(t, "DEBUG", DEBUG.String())
	assert.Equal(t, "INFO", INFO.String())
	assert.Equal(t, "WARN", WARN.String())
	assert.Equal(t, "ERROR", ERROR.String())
	assert.Equal(t, "FATAL", FATAL.String())

	assert.Equal(t, TRACE, MustParseLevel("TRACE"))
	assert.Equal(t, DEBUG, MustParseLevel("DEBUG"))
	assert.Equal(t, INFO, MustParseLevel("INFO"))
	assert.Equal(t, WARN, MustParseLevel("WARN"))
	assert.Equal(t, ERROR, MustParseLevel("ERROR"))
	assert.Equal(t, FATAL, MustParseLevel("FATAL"))

	assert.Equal(t, TRACE, MustParseLevel("trace"))
	assert.Equal(t, DEBUG, MustParseLevel("debug"))
	assert.Equal(t, INFO, MustParseLevel("info"))
	assert.Equal(t, WARN, MustParseLevel("warn"))
	assert.Equal(t, ERROR, MustParseLevel("error"))
	assert.Equal(t, FATAL, MustParseLevel("fatal"))

	assert.Equal(t, TRACE, MustParseLevel("Trace"))
	assert.Equal(t, DEBUG, MustParseLevel("Debug"))
	assert.Equal(t, INFO, MustParseLevel("Info"))
	assert.Equal(t, WARN, MustParseLevel("Warn"))
	assert.Equal(t, ERROR, MustParseLevel("Error"))
	assert.Equal(t, FATAL, MustParseLevel("Fatal"))

	assert.Equal(t, TRACE, MustParseLevel("T"))
	assert.Equal(t, DEBUG, MustParseLevel("D"))
	assert.Equal(t, INFO, MustParseLevel("I"))
	assert.Equal(t, WARN, MustParseLevel("W"))
	assert.Equal(t, ERROR, MustParseLevel("E"))
	assert.Equal(t, FATAL, MustParseLevel("F"))

	assert.Equal(t, TRACE, MustParseLevel("t"))
	assert.Equal(t, DEBUG, MustParseLevel("d"))
	assert.Equal(t, INFO, MustParseLevel("i"))
	assert.Equal(t, WARN, MustParseLevel("w"))
	assert.Equal(t, ERROR, MustParseLevel("e"))
	assert.Equal(t, FATAL, MustParseLevel("f"))

	assert.Equal(t, TRACE, MustParseLevel("5"))
	assert.Equal(t, DEBUG, MustParseLevel("4"))
	assert.Equal(t, INFO, MustParseLevel("3"))
	assert.Equal(t, WARN, MustParseLevel("2"))
	assert.Equal(t, ERROR, MustParseLevel("1"))
	assert.Equal(t, FATAL, MustParseLevel("0"))

	assert.Panics(t, func() { MustParseLevel("xTrace") })
	assert.Panics(t, func() { MustParseLevel("xDebug") })
	assert.Panics(t, func() { MustParseLevel("xInfo") })
	assert.Panics(t, func() { MustParseLevel("xWarn") })
	assert.Panics(t, func() { MustParseLevel("xError") })
	assert.Panics(t, func() { MustParseLevel("xFatal") })
}

func TestFormatHeader(t *testing.T) {
	l := newLogger(nil, true)
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
	data *bytes.Buffer
}

func newMockProvider() *mockProvider {
	return &mockProvider{data: bytes.NewBufferString("")}
}

func (p *mockProvider) Write(level Level, headerLength int, data []byte) error {
	_, err := p.data.Write(data[headerLength:])
	return err
}

func (p *mockProvider) Close() error { return nil }

func TestLogPrint(t *testing.T) {
	p := newMockProvider()
	l := newLogger(p, false)

	print := func() {
		l.Trace(0, "hello %v", TRACE)
		l.Debug(0, "hello %v", DEBUG)
		l.Info(0, "hello %v", INFO)
		l.Warn(0, "hello %v", WARN)
		l.Error(0, "hello %v", ERROR)
	}

	for _, level := range []Level{TRACE, DEBUG, INFO, WARN, ERROR} {
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

	if got := p.data.String(); got != expected {
		t.Errorf("unexpected print: `%s' vs `%s'", got, expected)
	}
}

type mockHandler struct {
	level Level
	body  []byte
	desc  []byte
}

func (h *mockHandler) Handle(e Entry) error {
	e2 := e.Clone()
	h.level = e2.Level()
	h.body = e2.Body()
	h.desc = e2.Desc()
	return nil
}

func TestHook(t *testing.T) {
	l := newLogger(newMockProvider(), false)
	l.SetLevel(TRACE)
	hanlder := new(mockHandler)
	l.Hook(hanlder)

	data := []byte("data")
	desc := "description"
	l.LogWith(INFO, 1, data, desc)

	assert.Equal(t, string(data), string(hanlder.body))
	assert.Equal(t, desc, string(hanlder.desc))
	assert.Equal(t, INFO, hanlder.level)
}
