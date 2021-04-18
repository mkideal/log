package log

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Level represents log level
type Level int32

// Level constants
const (
	_       Level = iota // 0
	LvFATAL              // 1
	LvERROR              // 2
	LvWARN               // 3
	LvINFO               // 4
	LvDEBUG              // 5
	LvTRACE              // 6
	numLevel
)

var errUnrecognizedLogLevel = errors.New("unrecognized log level")

// Set implements flag.Value interface such that you can use level  as a command as following:
//
//	var level logger.Level
//	flag.Var(&level, "log_level", "log level: trace/debug/info/warn/error/fatal")
func (level *Level) Set(s string) error {
	lv, ok := ParseLevel(s)
	*level = lv
	if !ok {
		return errUnrecognizedLogLevel
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
	case LvFATAL:
		return "FATAL"
	case LvERROR:
		return "ERROR"
	case LvWARN:
		return "WARN"
	case LvINFO:
		return "INFO"
	case LvDEBUG:
		return "DEBUG"
	case LvTRACE:
		return "TRACE"
	}
	return "(" + strconv.Itoa(int(level)) + ")"
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
	return level.Set(s)
}

// MoreVerboseThan returns whether level more verbose than other
func (level Level) MoreVerboseThan(other Level) bool { return level > other }

// ParseLevel parses log level from string
func ParseLevel(s string) (lv Level, ok bool) {
	s = strings.ToUpper(s)
	switch s {
	case "FATAL", "F", LvFATAL.Literal():
		return LvFATAL, true
	case "ERROR", "E", LvERROR.Literal():
		return LvERROR, true
	case "WARN", "W", LvWARN.Literal():
		return LvWARN, true
	case "INFO", "I", LvINFO.Literal():
		return LvINFO, true
	case "DEBUG", "D", LvDEBUG.Literal():
		return LvDEBUG, true
	case "TRACE", "T", LvTRACE.Literal():
		return LvTRACE, true
	}
	return LvINFO, false
}

// MustParseLevel similars to ParseLevel, but panic if parse failed
func MustParseLevel(s string) Level {
	lv, ok := ParseLevel(s)
	if !ok {
		panic(errUnrecognizedLogLevel.Error() + ": " + s)
	}
	return lv
}

// HTTPHandlerGetLevel returns a http handler for getting log level
func HTTPHandlerGetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, GetLevel().String())
	})
}

// HTTPHandlerSetLevel sets new log level and returns old log level
// Returns status code `StatusBadRequest` if parse log level fail
func HTTPHandlerSetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		level := r.FormValue("level")
		lv, ok := ParseLevel(level)
		// invalid parameter
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid log level: "+level)
			return
		}
		// not modified
		oldLevel := GetLevel()
		if lv == oldLevel {
			w.WriteHeader(http.StatusNotModified)
			io.WriteString(w, oldLevel.String())
			return
		}
		// updated
		SetLevel(lv)
		io.WriteString(w, oldLevel.String())
	})
}

var (
	registerHTTPHandlersOnce sync.Once
)

func registerHTTPHandlers() {
	registerHTTPHandlersOnce.Do(func() {
		http.Handle("/log/level/get", HTTPHandlerGetLevel())
		http.Handle("/log/level/set", HTTPHandlerSetLevel())
	})
}
