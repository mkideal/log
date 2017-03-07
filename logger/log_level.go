package logger

import (
	"errors"
	"strconv"
	"strings"
)

// Level represents log level
type Level int32

const (
	FATAL    Level = iota // 0
	ERROR                 // 1
	WARN                  // 2
	INFO                  // 3
	DEBUG                 // 4
	TRACE                 // 5
	NumLevel              // 6 log levels
)

var ErrUnrecognizedLogLevel = errors.New("unrecognized log level")

// Set implements flag.Value interface such that you can use level  as a command as following:
//
//	var level logger.Level
//	flag.Var(&level, "log_level", "log level: trace/debug/info/warn/error/fatal")
func (level *Level) Set(s string) error {
	return level.Decode(s)
}

// Decode implements github.com/mkideal/cli.Decoder interface
func (level *Level) Decode(s string) error {
	lv, ok := ParseLevel(s)
	*level = lv
	if !ok {
		return ErrUnrecognizedLogLevel
	}
	return nil
}

// Literal returns literal value which is a number
func (level Level) Literal() string {
	return strconv.Itoa(int(level))
}

// String returns a serialized string of level
func (level Level) String() string {
	switch level {
	case FATAL:
		return "FATAL"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	}
	return "INVALID"
}

// MarshalJSON implements json.Marshaler
func (level Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + level.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (level *Level) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	if len(data) >= 2 {
		s, err = strconv.Unquote(string(data))
		if err != nil {
			return err
		}
	} else {
		s = string(data)
	}
	return level.Decode(s)
}

func (level Level) MoreVerboseThan(other Level) bool { return level > other }

// ParseLevel parses log level from string
func ParseLevel(s string) (lv Level, ok bool) {
	s = strings.ToUpper(s)
	switch s {
	case "FATAL", "F", FATAL.Literal():
		return FATAL, true
	case "ERROR", "E", ERROR.Literal():
		return ERROR, true
	case "WARN", "W", WARN.Literal():
		return WARN, true
	case "INFO", "I", INFO.Literal():
		return INFO, true
	case "DEBUG", "D", DEBUG.Literal():
		return DEBUG, true
	case "TRACE", "T", TRACE.Literal():
		return TRACE, true
	}
	return INFO, false
}

// MustParseLevel similars to ParseLevel, but panic if parse failed
func MustParseLevel(s string) Level {
	lv, ok := ParseLevel(s)
	if !ok {
		panic(ErrUnrecognizedLogLevel.Error() + ": " + s)
	}
	return lv
}
