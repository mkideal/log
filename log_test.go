package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mkideal/log"
)

type testingLogWriter struct {
	discard bool
	buf     bytes.Buffer
}

func (w *testingLogWriter) Write(level log.Level, data []byte, headerLen int) error {
	if !w.discard {
		w.buf.WriteByte('[')
		w.buf.WriteString(level.String())
		w.buf.WriteByte(']')
		w.buf.WriteByte(' ')
		w.buf.Write(data[headerLen:])
	}
	return nil
}

func (w *testingLogWriter) Close() error { return nil }

func TestWriter(t *testing.T) {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LvTRACE), log.WithPrefix("testing"))
	log.Printf(log.LvTRACE, "hello %s", "log")
	logger := log.Prefix("prefix")
	logger.Debug().Print("hello world")
	log.Shutdown()
	got := writer.buf.String()
	want := "[TRACE] (testing) hello log\n[DEBUG] (testing/prefix) hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func ExampleFields() {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LvINFO), log.WithPrefix("testing"))
	log.Info().Int("int", 123456).Print("fields")
	log.Info().Int8("int8", -12).Print("fields")
	log.Info().Int16("int16", 1234).Print("fields")
	log.Info().Int32("int32", -12345678).Print("fields")
	log.Info().Int64("int64", 1234567890).Print("fields")
	log.Info().Uint("uint", 123456).Print("fields")
	log.Info().Uint8("uint8", 120).Print("fields")
	log.Info().Uint16("uint16", 12340).Print("fields")
	log.Info().Uint32("uint32", 123456780).Print("fields")
	log.Info().Uint64("uint64", 12345678900).Print("fields")
	log.Info().Float32("float32", 1234.5678).Print("fields")
	log.Info().Float64("float64", 0.123456789).Print("fields")
	log.Info().Complex64("complex64", 1+2i).Print("fields")
	log.Info().Complex128("complex128", 1).Print("fields")
	log.Info().Complex128("complex128", 2i).Print("fields")
	log.Info().Byte("byte", 'h').Print("fields")
	log.Info().Rune("rune", 'Å').Print("fields")
	log.Info().Bool("bool", true).Print("fields")
	log.Info().Bool("bool", false).Print("fields")
	log.Info().String("string", "hello").Print("fields")
	log.Info().Error("error", nil).Print("fields")
	log.Info().Error("error", errors.New("err")).Print("fields")
	log.Info().Any("any", nil).Print("fields")
	log.Info().Any("any", "nil").Print("fields")
	log.Info().Type("type", nil).Print("fields")
	log.Info().Type("type", "string").Print("fields")
	log.Info().Type("type", new(int)).Print("fields")
	log.Info().Duration("duration", time.Millisecond*1200).Print("fields")
	log.Info().String("$name", "hello").Print("fields")
	log.Info().String("name of", "hello").Print("fields")
	log.Prefix("prefix").Info().
		String("k1", "v1").
		Int("k2", 2).
		Print("prefix logging")
	log.Debug().String("key", "value").Print("not output")
	log.Shutdown()
	fmt.Print(writer.buf.String())
	// Output:
	// [INFO] (testing) {int:123456} fields
	// [INFO] (testing) {int8:-12} fields
	// [INFO] (testing) {int16:1234} fields
	// [INFO] (testing) {int32:-12345678} fields
	// [INFO] (testing) {int64:1234567890} fields
	// [INFO] (testing) {uint:123456} fields
	// [INFO] (testing) {uint8:120} fields
	// [INFO] (testing) {uint16:12340} fields
	// [INFO] (testing) {uint32:123456780} fields
	// [INFO] (testing) {uint64:12345678900} fields
	// [INFO] (testing) {float32:1234.5677} fields
	// [INFO] (testing) {float64:0.123456789} fields
	// [INFO] (testing) {complex64:1+2i} fields
	// [INFO] (testing) {complex128:1} fields
	// [INFO] (testing) {complex128:2i} fields
	// [INFO] (testing) {byte:'h'} fields
	// [INFO] (testing) {rune:'Å'} fields
	// [INFO] (testing) {bool:true} fields
	// [INFO] (testing) {bool:false} fields
	// [INFO] (testing) {string:"hello"} fields
	// [INFO] (testing) {error:nil} fields
	// [INFO] (testing) {error:"err"} fields
	// [INFO] (testing) {any:nil} fields
	// [INFO] (testing) {any:"nil"} fields
	// [INFO] (testing) {type:"nil"} fields
	// [INFO] (testing) {type:"string"} fields
	// [INFO] (testing) {type:"*int"} fields
	// [INFO] (testing) {duration:1.2s} fields
	// [INFO] (testing) {$name:"hello"} fields
	// [INFO] (testing) {"name of":"hello"} fields
	// [INFO] (testing/prefix) {k1:"v1" k2:2} prefix logging
}

func benchmarkSetup(b *testing.B) {
	b.StopTimer()
	writer := new(testingLogWriter)
	writer.discard = true
	log.Start(log.WithWriters(writer), log.WithSync(true), log.WithLevel(log.LvINFO))
	b.StartTimer()
}

func benchmarkTeardown(b *testing.B) {
	b.StopTimer()
	log.Shutdown()
	b.StartTimer()
}

func BenchmarkFormattingFields(b *testing.B) {
	benchmarkSetup(b)
	for i := 0; i < b.N; i++ {
		log.Info().
			Int("int", 123456).
			Uint("uint", 123456).
			Float64("float64", 0.123456789).
			String("string", "hello").
			Duration("duration", time.Microsecond*1234567890).
			Print("benchmark fields")
	}
	benchmarkTeardown(b)
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
