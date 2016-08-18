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
	With(objs ...interface{}) ContextLogger
	WithJSON(objs ...interface{}) ContextLogger
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

// withLogger implements ContextLogger
type withLogger struct {
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
	case float32:
		return strconv.AppendFloat(nil, float64(data), 'E', -1, 64)
	case float64:
		return strconv.AppendFloat(nil, data, 'E', -1, 64)
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

func (wl *withLogger) bytes() []byte {
	if wl.b != nil {
		return wl.b
	}
	if wl.formatter != nil {
		wl.b = wl.formatter.Format(wl.data)
		return wl.b
	}
	wl.b = toBytes(wl.data)
	return wl.b
}

func (wl *withLogger) With(objs ...interface{}) ContextLogger {
	wl.b = nil
	if wl.data == nil {
		wl.data = objs
		return wl
	}
	for _, obj := range objs {
		wl.migrateValue(obj)
	}
	return wl
}

func (wl *withLogger) WithJSON(objs ...interface{}) ContextLogger {
	return wl.With(objs...).SetFormatter(jsonFormatter)
}

func (wl *withLogger) migrateValue(v interface{}) {
	m1, ok1 := wl.data.(M)
	m2, ok2 := v.(M)
	if ok1 && ok2 {
		for key, val := range m2 {
			m1[key] = val
		}
		wl.data = m1
	} else {
		s1, ok1 := wl.data.(S)
		s2, ok2 := v.(S)
		if ok1 {
			if ok2 {
				for _, val := range s2 {
					s1 = append(s1, val)
				}
			} else {
				s1 = append(s1, v)
			}
			wl.data = s1
		} else {
			if ok2 {
				wl.data = append(S{wl.data}, s2...)
			} else {
				wl.data = S{wl.data, v}
			}
		}
	}
}

func (wl *withLogger) SetFormatter(f Formatter) ContextLogger {
	wl.formatter = f
	wl.b = nil
	return wl
}

func (wl *withLogger) Trace(format string, args ...interface{}) ContextLogger {
	if wl.isTrue && glogger.GetLevel() >= LvTRACE {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvTRACE, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Trace(1, buf.String())
		}
	}
	return wl
}

func (wl *withLogger) Debug(format string, args ...interface{}) ContextLogger {
	if wl.isTrue && glogger.GetLevel() >= LvDEBUG {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvDEBUG, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Debug(1, buf.String())
		}
	}
	return wl
}

func (wl *withLogger) Info(format string, args ...interface{}) ContextLogger {
	if wl.isTrue && glogger.GetLevel() >= LvINFO {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvINFO, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Info(1, buf.String())
		}
	}
	return wl
}

func (wl *withLogger) Warn(format string, args ...interface{}) ContextLogger {
	if wl.isTrue && glogger.GetLevel() >= LvWARN {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvWARN, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Warn(1, buf.String())
		}
	}
	return wl
}

func (wl *withLogger) Error(format string, args ...interface{}) ContextLogger {
	if wl.isTrue && glogger.GetLevel() >= LvERROR {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvERROR, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Error(1, buf.String())
		}
	}
	return wl
}

func (wl *withLogger) Fatal(format string, args ...interface{}) ContextLogger {
	if wl.isTrue {
		if l, ok := glogger.(logger.WithLogger); ok {
			l.LogWith(LvFATAL, 1, wl.bytes(), format, args...)
		} else {
			buf := bytes.NewBuffer(wl.bytes())
			fmt.Fprintf(buf, format, args...)
			glogger.Fatal(1, buf.String())
		}
	}
	return wl
}
