package log

import (
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type builder struct {
	buf []byte
}

// String returns the accumulated string.
func (b *builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *builder) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *builder) Cap() int { return cap(b.buf) }

func (b *builder) reset() {
	b.buf = b.buf[:0]
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *builder) grow(n int) {
	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
	copy(buf, b.buf)
	b.buf = buf
}

func (b *builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *builder) writeByte(c byte) {
	b.buf = append(b.buf, c)
}

func (b *builder) writeRune(r rune) {
	if r < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return
	}
	l := len(b.buf)
	if cap(b.buf)-l < utf8.UTFMax {
		b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[l:l+utf8.UTFMax], r)
	b.buf = b.buf[:l+n]
}

func (b *builder) writeString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *builder) writeInt(i int64) {
	b.buf = strconv.AppendInt(b.buf, i, 10)
}

func (b *builder) writeUint(i uint64) {
	b.buf = strconv.AppendUint(b.buf, i, 10)
}

func (b *builder) writeFloat32(f float32) {
	b.buf = strconv.AppendFloat(b.buf, float64(f), 'f', -1, 32)
}

func (b *builder) writeFloat64(f float64) {
	b.buf = strconv.AppendFloat(b.buf, f, 'f', -1, 64)
}

func (b *builder) writeBool(v bool) {
	b.buf = strconv.AppendBool(b.buf, v)
}
