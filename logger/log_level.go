package logger

type LogLevel int32

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
	TRACE
)

func (level LogLevel) String() string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	}
	return "INVALID"
}

func ParseLogLevel(s string) (lv LogLevel, ok bool) {
	switch s {
	case "trace", "TRACE", "T", "t":
		return TRACE, true
	case "debug", "DEBUG", "D", "d":
		return DEBUG, true
	case "info", "INFO", "I", "i":
		return INFO, true
	case "warn", "WARN", "W", "w":
		return WARN, true
	case "error", "ERROR", "E", "e":
		return ERROR, true
	}
	return
}

func MustParseLogLevel(s string) LogLevel {
	lv, ok := ParseLogLevel(s)
	if !ok {
		panic("ParseLogLevel " + s + " fail")
	}
	return lv
}
