package log_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/mkideal/log"
)

type testingLogWriter struct {
	buf bytes.Buffer
}

func (w *testingLogWriter) Write(level log.Level, data []byte, headerLen int) error {
	w.buf.WriteByte('[')
	w.buf.WriteString(level.String())
	w.buf.WriteByte(']')
	w.buf.WriteByte(' ')
	w.buf.Write(data[headerLen:])
	return nil
}

func (w *testingLogWriter) Close() error { return nil }

func TestWriter(t *testing.T) {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LvTRACE), log.WithPrefix("testing"))
	log.Printf(1, log.LvTRACE, "prefix", "hello %s", "log")
	logger := log.NewLogger("with-prefix")
	logger.Debug("hello world")
	log.Shutdown()
	got := writer.buf.String()
	want := "[TRACE] (testing/prefix) hello log\n[DEBUG] (testing/with-prefix) hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func ExampleFields() {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LvTRACE), log.WithPrefix("testing"))
	log.For(log.LvINFO).
		Int("int", 123456).
		Int8("int8", -12).
		Int16("int16", 1234).
		Int32("int32", -12345678).
		Int64("int64", 1234567890).
		Uint("uint", 123456).
		Uint8("uint8", 120).
		Uint16("uint16", 12340).
		Uint32("uint32", 123456780).
		Uint64("uint64", 12345678900).
		Float32("float32", 1234.5678).
		Float64("float64", 0.123456789).
		Byte("byte", 'h').
		Rune("rune", 'Å').
		Bool("bool1", true).
		Bool("bool2", false).
		String("string", "hello").
		Printf("fields")
	log.Shutdown()
	fmt.Print(writer.buf.String())
	// Output:
	// [INFO] (testing) {int:123456 int8:-12 int16:1234 int32:-12345678 int64:1234567890 uint:123456 uint8:120 uint16:12340 uint32:123456780 uint64:12345678900 float32:1234.5677 float64:0.123456789 byte:h rune:Å bool1:true bool2:false string:hello} fields
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
func (fs testFS) OpenFile(name string, flag int, perm os.FileMode) (log.File, error) {
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
