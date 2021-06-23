package log

// IfLogger implements Context
type IfLogger byte

func (il IfLogger) ok() bool { return il&0x0F != 0 }

func (il IfLogger) If(ok bool) IfLogger {
	return If(ok)
}

func (il IfLogger) Else() IfLogger {
	p := il & 0xF0
	if p != 0 {
		return p
	}
	return 0xFF
}

func (il IfLogger) ElseIf(ok bool) IfLogger {
	p := il & 0xF0
	if p != 0 {
		return p
	}
	if ok {
		return 0xFF
	} else {
		return 0x00
	}
}

// With implements Context.With method
func (il IfLogger) With(values ...interface{}) ContextLogger {
	if len(values) == 1 {
		return &contextLogger{isTrue: il.ok(), data: values[0]}
	}
	return &contextLogger{isTrue: il.ok(), data: values}
}

// WithJSON implements Context.WithJSON method
func (il IfLogger) WithJSON(values ...interface{}) ContextLogger {
	return il.With(values...).SetFormatter(jsonFormatter)
}

// SetFormatter implements Context.SetFormatter method
func (il IfLogger) SetFormatter(f Formatter) ContextLogger {
	return &contextLogger{
		isTrue:    il.ok(),
		formatter: f,
	}
}

func (il IfLogger) Trace(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Trace(1, format, args...)
	}
	return il
}

func (il IfLogger) Debug(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Debug(1, format, args...)
	}
	return il
}

func (il IfLogger) Info(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Info(1, format, args...)
	}
	return il
}

func (il IfLogger) Warn(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Warn(1, format, args...)
	}
	return il
}

func (il IfLogger) Error(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Error(1, format, args...)
	}
	return il
}

func (il IfLogger) Fatal(format string, args ...interface{}) IfLogger {
	if il.ok() {
		glogger.Fatal(1, format, args...)
	}
	return il
}
