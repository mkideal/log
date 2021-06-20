package log

import (
	"fmt"
	"sync"
)

type Fields struct {
	level   Level
	builder builder
}

var fieldsPool = sync.Pool{
	New: func() interface{} {
		return new(Fields)
	},
}

func getFields(level Level) *Fields {
	fields := fieldsPool.Get().(*Fields)
	fields.reset(level)
	return fields
}

func putFields(fields *Fields) {
	if fields.builder.Cap() < 1024 {
		fieldsPool.Put(fields)
	}
}

func For(level Level) *Fields {
	if gprinter.GetLevel() >= level {
		return getFields(level)
	}
	return nil
}

func (fields *Fields) reset(level Level) {
	fields.level = level
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
	gprinter.Printf(1, fields.level, "", fields.builder.String())
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
