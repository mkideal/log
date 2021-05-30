package log

import (
	"bytes"
	"os"
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
	Start(WithWriters(writer), WithLevel(LvTRACE), WithPrefix("testing"))
	Printf(1, LvTRACE, "prefix", "hello %s", "log")
	logger := NewLogger("with-prefix")
	logger.Debug("hello world")
	Shutdown()
	got := writer.buf.String()
	want := "[TRACE](testing/prefix) hello log\n[DEBUG](testing/with-prefix) hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}

// testFS implements File interface
type testFile struct {
	content bytes.Buffer
}

func (t *testFile) Write(p []byte) (int, error) { return t.content.Write(p) }
func (t *testFile) Close() error                { return nil }
func (t *testFile) Sync() error                 { return nil }

// testFS implements FS interface
type testFS struct {
	files map[string]*testFile
}

func newTestFS() *testFS {
	return &testFS{
		files: make(map[string]*testFile),
	}
}

// OpenFile implements FS OpenFile method
func (fs testFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, ok := fs.files[name]
	if ok {
		if flag&os.O_CREATE != 0 && flag&os.O_EXCL == 0 {
			return nil, os.ErrExist
		}
		if flag&os.O_TRUNC != 0 {
			f.content.Reset()
		}
	} else if flag&os.O_CREATE != 0 {
		f = &testFile{}
		fs.files[name] = f
	} else {
		return nil, os.ErrNotExist
	}
	return f, nil
}

// Remove implements FS Remove method
func (fs *testFS) Remove(name string) error {
	if _, ok := fs.files[name]; !ok {
		return os.ErrNotExist
	}
	delete(fs.files, name)
	return nil
}

// Symlink implements FS Symlink method
func (fs testFS) Symlink(oldname, newname string) error { return nil }

// MkdirAll implements FS MkdirAll method
func (fs testFS) MkdirAll(path string, perm os.FileMode) error { return nil }

func TestFile(t *testing.T) {
	// (TODO): test writer `file`
}
