package log

// IfLogger implements Context
type IfLogger bool

func (il IfLogger) If(ok bool) IfLogger     { return IfLogger(ok) }
func (il IfLogger) Else() IfLogger          { return !il }
func (il IfLogger) ElseIf(ok bool) IfLogger { return !il && IfLogger(ok) }

// With implements Context.With method
func (il IfLogger) With(v interface{}) ContextLogger {
	return &withLogger{
		isTrue: bool(il),
		data:   v,
	}
}

// WithN implements Context.WithN method
func (il IfLogger) WithN(objs ...interface{}) ContextLogger {
	return &withLogger{
		isTrue: bool(il),
		data:   objs,
	}
}

// WithJSON implements Context.WithJSON method
func (il IfLogger) WithJSON(v interface{}) ContextLogger {
	return &withLogger{
		isTrue:    bool(il),
		data:      v,
		formatter: jsonFormatter,
	}
}

// SetFormatter implements Context.SetFormatter method
func (il IfLogger) SetFormatter(f Formatter) ContextLogger {
	return &withLogger{
		isTrue:    bool(il),
		formatter: f,
	}
}

func (il IfLogger) Trace(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Trace(1, format, args...)
	}
	return il
}

func (il IfLogger) Debug(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Debug(1, format, args...)
	}
	return il
}

func (il IfLogger) Info(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Info(1, format, args...)
	}
	return il
}

func (il IfLogger) Warn(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Warn(1, format, args...)
	}
	return il
}

func (il IfLogger) Error(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Error(1, format, args...)
	}
	return il
}

func (il IfLogger) Fatal(format string, args ...interface{}) IfLogger {
	if il {
		glogger.Fatal(1, format, args...)
	}
	return il
}
