package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mkideal/log/logger"
)

// Formatter format the data of context
type Formatter interface {
	Format(v interface{}) []byte
}

type JSONFormatter struct{}

var jsonFormatter = JSONFormatter{}

func (f JSONFormatter) Format(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// M aliases map
type M map[string]interface{}

// S aliases slice
type S []interface{}

// Context represents a context of logger
type Context interface {
	With(values ...interface{}) ContextLogger
	WithJSON(values ...interface{}) ContextLogger
	SetFormatter(f Formatter) ContextLogger
}

type ContextLogger interface {
	Context
	Trace(format string, args ...interface{}) ContextLogger
	Debug(format string, args ...interface{}) ContextLogger
	Info(format string, args ...interface{}) ContextLogger
	Warn(format string, args ...interface{}) ContextLogger
	Error(format string, args ...interface{}) ContextLogger
	Fatal(format string, args ...interface{}) ContextLogger
}

// contextLogger implements ContextLogger
type contextLogger struct {
	isTrue    bool
	data      interface{}
	formatter Formatter
	b         []byte
}

var bytesTrue = []byte("true")
var bytesFalse = []byte("false")

func toBytes(v interface{}) []byte {
	switch data := v.(type) {
	case string:
		return []byte(data)
	case int:
		return strconv.AppendInt(nil, int64(data), 10)
	case int8:
		return strconv.AppendInt(nil, int64(data), 10)
	case int16:
		return strconv.AppendInt(nil, int64(data), 10)
	case int32:
		return strconv.AppendInt(nil, int64(data), 10)
	case int64:
		return strconv.AppendInt(nil, data, 10)
	case uint:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uintptr:
		return strconv.AppendUint([]byte{'0', 'x'}, uint64(data), 16)
	case uint8:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint16:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint32:
		return strconv.AppendUint(nil, uint64(data), 10)
	case uint64:
		return strconv.AppendUint(nil, data, 10)
	case bool:
		if data {
			return bytesTrue
		}
		return bytesFalse
	default:
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "%v", data)
		return buf.Bytes()
	}
}

func (l *contextLogger) bytes() []byte {
	if l.b != nil {
		return l.b
	}
	if l.formatter != nil {
		l.b = l.formatter.Format(l.data)
		return l.b
	}
	l.b = toBytes(l.data)
	return l.b
}

func (l *contextLogger) With(values ...interface{}) ContextLogger {
	l.b = nil
	if l.data == nil {
		if len(values) == 0 {
			l.data = values[0]
		} else {
			l.data = values
		}
		return l
	}
	for _, obj := range values {
		l.migrateValue(obj)
	}
	return l
}

func (l *contextLogger) WithJSON(values ...interface{}) ContextLogger {
	return l.With(values...).SetFormatter(jsonFormatter)
}

func (l *contextLogger) migrateValue(v interface{}) {
	m1, ok1 := l.data.(M)
	m2, ok2 := v.(M)
	if ok1 && ok2 {
		for key, val := range m2 {
			m1[key] = val
		}
		l.data = m1
	} else {
		s1, ok1 := l.data.(S)
		s2, ok2 := v.(S)
		if ok1 {
			if ok2 {
				for _, val := range s2 {
					s1 = append(s1, val)
				}
			} else {
				s1 = append(s1, v)
			}
			l.data = s1
		} else {
			if ok2 {
				l.data = append(S{l.data}, s2...)
			} else {
				l.data = S{l.data, v}
			}
		}
	}
}

func (l *contextLogger) SetFormatter(f Formatter) ContextLogger {
	l.formatter = f
	l.b = nil
	return l
}

func (l *contextLogger) formatMessage(format string, args ...interface{}) string {
	buf := bytes.NewBuffer(l.bytes())
	if buf.Len() > 0 && len(format) > 0 {
		buf.WriteString(" | ")
	}
	fmt.Fprintf(buf, format, args...)
	return buf.String()
}

func (l *contextLogger) Trace(format string, args ...interface{}) ContextLogger {
	if l.isTrue && glogger.GetLevel() >= LvTRACE {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvTRACE, 1, l.bytes(), format, args...)
		} else {
			glogger.Trace(1, l.formatMessage(format, args...))
		}
	}
	return l
}

func (l *contextLogger) Debug(format string, args ...interface{}) ContextLogger {
	if l.isTrue && glogger.GetLevel() >= LvDEBUG {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvDEBUG, 1, l.bytes(), format, args...)
		} else {
			glogger.Debug(1, l.formatMessage(format, args...))
		}
	}
	return l
}

func (l *contextLogger) Info(format string, args ...interface{}) ContextLogger {
	if l.isTrue && glogger.GetLevel() >= LvINFO {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvINFO, 1, l.bytes(), format, args...)
		} else {
			glogger.Info(1, l.formatMessage(format, args...))
		}
	}
	return l
}

func (l *contextLogger) Warn(format string, args ...interface{}) ContextLogger {
	if l.isTrue && glogger.GetLevel() >= LvWARN {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvWARN, 1, l.bytes(), format, args...)
		} else {
			glogger.Warn(1, l.formatMessage(format, args...))
		}
	}
	return l
}

func (l *contextLogger) Error(format string, args ...interface{}) ContextLogger {
	if l.isTrue && glogger.GetLevel() >= LvERROR {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvERROR, 1, l.bytes(), format, args...)
		} else {
			glogger.Error(1, l.formatMessage(format, args...))
		}
	}
	return l
}

func (l *contextLogger) Fatal(format string, args ...interface{}) ContextLogger {
	if l.isTrue {
		if wl, ok := glogger.(logger.With); ok {
			wl.LogWith(LvFATAL, 1, l.bytes(), format, args...)
		} else {
			glogger.Fatal(1, l.formatMessage(format, args...))
		}
	}
	return l
}
