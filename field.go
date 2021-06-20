package log

import (
	"fmt"
	"sync"
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

func (fields *Fields) writeKey(key string) {
	if fields.builder.Len() == 0 {
		fields.builder.writeByte('{')
	} else {
		fields.builder.writeByte(' ')
	}
	fields.builder.writeString(key)
	fields.builder.writeByte(':')
}

// Printf prints logging with context fields. After this call,
// the fields not available.
func (fields *Fields) Printf(format string, args ...interface{}) {
	if fields == nil {
		return
	}
	if fields.builder.Len() > 0 {
		fields.builder.writeString("} ")
	}
	if len(args) == 0 {
		fields.builder.writeString(format)
	} else {
		fmt.Fprintf(&fields.builder, format, args...)
	}
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
		fields.builder.writeByte(value)
	}
	return fields
}

func (fields *Fields) Rune(key string, value rune) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeRune(value)
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
		fields.builder.writeString(value)
	}
	return fields
}

func (fields *Fields) Error(key string, value error) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("<nil>")
		} else {
			fields.builder.writeString(value.Error())
		}
	}
	return fields
}

func (fields *Fields) Any(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("<nil>")
		} else {
			switch x := value.(type) {
			case string:
				fields.builder.writeString(x)
			case error:
				fields.builder.writeString(x.Error())
			case fmt.Stringer:
				fields.builder.writeString(x.String())
			default:
				fmt.Fprintf(&fields.builder, "%v", value)
			}
		}
	}
	return fields
}

func (fields *Fields) Type(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("nil")
		} else {
			fmt.Fprintf(&fields.builder, "%T", value)
		}
	}
	return fields
}
