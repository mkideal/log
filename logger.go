package log

// Logger represents basic logger interface
type Logger interface {
	// Trace outputs trace-level logs
	Trace(format string, args ...interface{})
	// Debug outputs debug-level logs
	Debug(format string, args ...interface{})
	// Info outputs info-level logs
	Info(format string, args ...interface{})
	// Warn outputs warn-level logs
	Warn(format string, args ...interface{})
	// Error outputs error-level logs
	Error(format string, args ...interface{})
	// Fatal outputs fatal-level logs
	Fatal(format string, args ...interface{})
}

// logger implements Logger to prints logging with extra prefix
type logger struct {
	prefix string
}

func (l logger) Trace(format string, args ...interface{}) {
	gprinter.Printf(1, LvTRACE, l.prefix, format, args...)
}

func (l logger) Debug(format string, args ...interface{}) {
	gprinter.Printf(1, LvDEBUG, l.prefix, format, args...)
}

func (l logger) Info(format string, args ...interface{}) {
	gprinter.Printf(1, LvINFO, l.prefix, format, args...)
}

func (l logger) Warn(format string, args ...interface{}) {
	gprinter.Printf(1, LvWARN, l.prefix, format, args...)
}

func (l logger) Error(format string, args ...interface{}) {
	gprinter.Printf(1, LvERROR, l.prefix, format, args...)
}

func (l logger) Fatal(format string, args ...interface{}) {
	gprinter.Printf(1, LvFATAL, l.prefix, format, args...)
}
