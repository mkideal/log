package logger

import (
	"errors"
	"strings"
)

// Level represents log level
type Level int32

const (
	FATAL Level = iota
	ERROR
	WARN
	INFO
	DEBUG
	TRACE
	NumLevel
)

var ErrUnrecognizedLogLevel = errors.New("unrecognized log level")

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

func (level *Level) Decode(s string) error {
	lv, ok := ParseLevel(s)
	*level = lv
	if !ok {
		return ErrUnrecognizedLogLevel
	}
	return nil
}

// ParseLevel parses log level from string
func ParseLevel(s string) (lv Level, ok bool) {
	s = strings.ToUpper(s)
	switch s {
	case "FATAL", "F", "0":
		return FATAL, true
	case "ERROR", "E", "1":
		return ERROR, true
	case "WARN", "W", "2":
		return WARN, true
	case "INFO", "I", "3":
		return INFO, true
	case "DEBUG", "D", "4":
		return DEBUG, true
	case "TRACE", "T", "5":
		return TRACE, true
	}
	return INFO, false
}

// MustParseLevel similars to ParseLevel, but panic if parse fail
func MustParseLevel(s string) Level {
	lv, ok := ParseLevel(s)
	if !ok {
		panic("ParseLevel " + s + " fail")
	}
	return lv
}
