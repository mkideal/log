package logger

import (
	"errors"
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
	LevelNum
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
	switch s {
	case "fatal", "FATAL", "F", "f", "0":
		return FATAL, true
	case "error", "ERROR", "E", "e", "1":
		return ERROR, true
	case "warn", "WARN", "W", "w", "2":
		return WARN, true
	case "info", "INFO", "I", "i", "3":
		return INFO, true
	case "debug", "DEBUG", "D", "d", "4":
		return DEBUG, true
	case "trace", "TRACE", "T", "t", "5":
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
