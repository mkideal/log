package log

import (
	"bytes"
	"testing"
)

type testingLogWriter struct {
	buf bytes.Buffer
}

func (w *testingLogWriter) Write(level Level, data []byte, headerLen int) error {
	w.buf.WriteByte('[')
	w.buf.WriteString(level.String())
	w.buf.WriteByte(']')
	w.buf.Write(data[headerLen:])
	return nil
}

func (w *testingLogWriter) Close() error { return nil }

func TestWriter(t *testing.T) {
	writer := new(testingLogWriter)
	Start(WithWriter(writer), WithLevel(LvTRACE), WithPrefix("testing"))
	Printf(1, LvTRACE, "prefix", "hello %s", "log")
	logger := PrefixLogger("with-prefix")
	logger.Debug("hello world")
	Shutdown()
	got := writer.buf.String()
	want := "[TRACE](testing/prefix) hello log\n[DEBUG](testing/with-prefix) hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}
