package log

// global printer
var gprinter = newStdPrinter()

type startOptions struct {
	httpHandler bool
	sync        bool
	level       Level
	prefix      string
	printer     Printer
	writers     []Writer
}

func (opt *startOptions) apply(options []Option) {
	for i := range options {
		if options[i] != nil {
			options[i](opt)
		}
	}
}

// Option is option for Start
type Option func(*startOptions)

// WithSync synchronize outputs log or not
func WithSync(yes bool) Option {
	return func(opt *startOptions) {
		opt.sync = yes
	}
}

// WithHTTPHandler enable or disable http handler for settting level
func WithHTTPHandler(yes bool) Option {
	return func(opt *startOptions) {
		opt.httpHandler = yes
	}
}

// WithLevel sets log level
func WithLevel(level Level) Option {
	return func(opt *startOptions) {
		opt.level = level
	}
}

// WithPrefix set log prefix
func WithPrefix(prefix string) Option {
	return func(opt *startOptions) {
		opt.prefix = prefix
	}
}

// WithPrinter specify custom printer
func WithPrinter(printer Printer) Option {
	return func(opt *startOptions) {
		opt.printer = printer
	}
}

// WithPrinter appends a custom writer
func WithWriter(writer Writer) Option {
	return func(opt *startOptions) {
		opt.writers = append(opt.writers, writer)
	}
}

// WithConsle appends a console writer
func WithConsle() Option {
	return WithWriter(newConsole())
}

// WithFile appends a file writer
func WithFile(fileOptions FileOptions) Option {
	return WithWriter(newFile(fileOptions))
}

// WithMultiFile appends a multifile writer
func WithMultiFile(multiFileOptions MultiFileOptions) Option {
	return WithWriter(newMultiFile(multiFileOptions))
}

// Start start global printer with options
func Start(options ...Option) error {
	var opt startOptions
	opt.apply(options)
	async := !opt.sync
	if opt.printer != nil && len(opt.writers) > 0 {
		println("log.Start: writers ignored because printer sepecfied")
	}
	if opt.printer == nil {
		switch len(opt.writers) {
		case 0:
			return nil
		case 1:
			opt.printer = newPrinter(opt.writers[0], async)
		default:
			opt.printer = newPrinter(mixWriter{opt.writers}, async)
		}
	}
	if opt.level != 0 {
		opt.printer.SetLevel(opt.level)
	}
	opt.printer.SetPrefix(opt.prefix)

	gprinter.Shutdown()
	gprinter = opt.printer
	gprinter.Start()
	if opt.httpHandler {
		registerHTTPHandlers()
	}
	return nil
}

// Shutdown shutdowns global printer
func Shutdown() {
	gprinter.Shutdown()
}

// GetLevel gets level of global printer
func GetLevel() Level {
	return gprinter.GetLevel()
}

// SetLevel sets level of gloabl printer
func SetLevel(level Level) {
	gprinter.SetLevel(level)
}

// Trace prints log with trace level
func Trace(format string, args ...interface{}) {
	gprinter.Printf(1, LvTRACE, "", format, args...)
}

// Debug prints log with debug level
func Debug(format string, args ...interface{}) {
	gprinter.Printf(1, LvDEBUG, "", format, args...)
}

// Info prints log with info level
func Info(format string, args ...interface{}) {
	gprinter.Printf(1, LvINFO, "", format, args...)
}

// Warn prints log with warning level
func Warn(format string, args ...interface{}) {
	gprinter.Printf(1, LvWARN, "", format, args...)
}

// Error prints log with error level
func Error(format string, args ...interface{}) {
	gprinter.Printf(1, LvERROR, "", format, args...)
}

// Trace prints log with fatal level
func Fatal(format string, args ...interface{}) {
	gprinter.Printf(1, LvFATAL, "", format, args...)
}

// Printf wraps global printer Printf method
func Printf(calldepth int, level Level, prefix, format string, args ...interface{}) {
	gprinter.Printf(calldepth, level, prefix, format, args...)
}

// PrefixLogger creates a logger with a extra prefix
func PrefixLogger(prefix string) Logger {
	return logger{printer: gprinter, prefix: prefix}
}
