package log

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode"
)

// Fields holds context fields
type Fields struct {
	level   Level
	prefix  string
	builder builder
}

var fieldsPool = sync.Pool{
	New: func() interface{} {
		return new(Fields)
	},
}

func getFields(level Level, prefix Prefix) *Fields {
	if gprinter.GetLevel() < level {
		return nil
	}
	fields := fieldsPool.Get().(*Fields)
	fields.reset(level, string(prefix))
	return fields
}

func putFields(fields *Fields) {
	if fields.builder.Cap() < 1024 {
		fieldsPool.Put(fields)
	}
}

func (fields *Fields) reset(level Level, prefix string) {
	fields.level = level
	fields.prefix = prefix
	fields.builder.reset()
}

func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if !isIdentRune(c, i) {
			return false
		}
	}
	return true
}

func isIdentRune(ch rune, i int) bool {
	return ch == '_' || ch == '-' || ch == '.' || ch == '#' || ch == '$' || ch == '/' ||
		unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func (fields *Fields) writeKey(key string) {
	if fields.builder.Len() == 0 {
		fields.builder.writeByte('{')
	} else {
		fields.builder.writeByte(' ')
	}
	if isIdent(key) {
		fields.builder.writeString(key)
	} else {
		fields.builder.writeQuotedString(key)
	}
	fields.builder.writeByte(':')
}

// Print prints logging with context fields. After this call,
// the fields not available.
func (fields *Fields) Print(s string) {
	if fields == nil {
		return
	}
	if fields.builder.Len() > 0 {
		fields.builder.writeString("} ")
	}
	fields.builder.writeString(s)
	gprinter.Printf(1, fields.level, fields.prefix, fields.builder.String())
	putFields(fields)
}

func (fields *Fields) Int(key string, value int) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int8(key string, value int8) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int16(key string, value int16) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int32(key string, value int32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int64(key string, value int64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(value)
	}
	return fields
}

func (fields *Fields) Uint(key string, value uint) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint8(key string, value uint8) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint16(key string, value uint16) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint32(key string, value uint32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint64(key string, value uint64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(value)
	}
	return fields
}

func (fields *Fields) Float32(key string, value float32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeFloat32(value)
	}
	return fields
}

func (fields *Fields) Float64(key string, value float64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeFloat64(value)
	}
	return fields
}

func (fields *Fields) Byte(key string, value byte) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeByte('\'')
		fields.builder.writeByte(value)
		fields.builder.writeByte('\'')
	}
	return fields
}

func (fields *Fields) Rune(key string, value rune) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.buf = strconv.AppendQuoteRune(fields.builder.buf, value)
	}
	return fields
}

func (fields *Fields) Bool(key string, value bool) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeBool(value)
	}
	return fields
}

func (fields *Fields) String(key string, value string) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeQuotedString(value)
	}
	return fields
}

func (fields *Fields) Error(key string, value error) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("nil")
		} else {
			fields.builder.buf = strconv.AppendQuote(fields.builder.buf, value.Error())
		}
	}
	return fields
}

func (fields *Fields) Any(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("nil")
		} else {
			switch x := value.(type) {
			case error:
				fields.builder.writeQuotedString(x.Error())
			case fmt.Stringer:
				fields.builder.writeQuotedString(x.String())
			case string:
				fields.builder.writeQuotedString(x)
			default:
				fmt.Fprintf(&fields.builder, "%q", value)
			}
		}
	}
	return fields
}

func (fields *Fields) Type(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString(`"nil"`)
		} else {
			fields.builder.writeQuotedString(reflect.TypeOf(value).String())
		}
	}
	return fields
}

func (fields *Fields) Exec(key string, stringer func() string) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeQuotedString(stringer())
	}
	return fields
}

func (fields *Fields) Time(key string, value time.Time) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.buf = value.AppendFormat(fields.builder.buf, time.RFC3339Nano)
	}
	return fields
}

func (fields *Fields) Duration(key string, value time.Duration) *Fields {
	if fields != nil {
		fields.writeKey(key)
		const reserved = 32
		l := len(fields.builder.buf)
		if cap(fields.builder.buf)-l < reserved {
			fields.builder.grow(reserved)
		}
		n := formatDuration(fields.builder.buf[l:l+reserved], value)
		fields.builder.buf = fields.builder.buf[:l+n]
	}
	return fields
}

// String returns a string representing the duration in the form "72h3m0.5s".
// Leading zero units are omitted. As a special case, durations less than one
// second format use a smaller unit (milli-, micro-, or nanoseconds) to ensure
// that the leading digit is non-zero. The zero duration formats as 0s.
func formatDuration(buf []byte, d time.Duration) int {
	// Largest time is 2540400h10m10.000000000s
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	if u < uint64(time.Second) {
		// Special case: if duration is smaller than a second,
		// use smaller units, like 1.2ms
		var prec int
		w--
		buf[w] = 's'
		w--
		switch {
		case u == 0:
			buf[0] = '0'
			buf[1] = 's'
			return 2
		case u < uint64(time.Microsecond):
			// print nanoseconds
			prec = 0
			buf[w] = 'n'
		case u < uint64(time.Millisecond):
			// print microseconds
			prec = 3
			// U+00B5 'µ' micro sign == 0xC2 0xB5
			w-- // Need room for two bytes.
			copy(buf[w:], "µ")
		default:
			// print milliseconds
			prec = 6
			buf[w] = 'm'
		}
		w, u = fmtFrac(buf[:w], u, prec)
		w = fmtInt(buf[:w], u)
	} else {
		w--
		buf[w] = 's'

		w, u = fmtFrac(buf[:w], u, 9)

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		// u is now integer minutes
		if u > 0 {
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], u%60)
			u /= 60

			// u is now integer hours
			// Stop at hours because days can be different lengths.
			if u > 0 {
				w--
				buf[w] = 'h'
				w = fmtInt(buf[:w], u)
			}
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	copy(buf, buf[w:])
	return len(buf) - w
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	print := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		print = print || digit != 0
		if print {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if print {
		w--
		buf[w] = '.'
	}
	return w, v
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}
