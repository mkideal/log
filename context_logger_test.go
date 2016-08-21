package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/mkideal/log/logger"
	"github.com/mkideal/log/provider"
)

func initMockLogger(w io.Writer, with bool) {
	l := logger.NewLoggerForTest(provider.NewConsoleWithWriter("", w, w), false, with)
	InitWithLogger(l)
	l.SetLevel(LvTRACE)
	NoHeader()
}

type test struct {
	data    interface{}
	data2   interface{}
	output  string
	output2 string
}

var tests = func() []test {
	return []test{
		{"abc", 1, "abc", "[abc 1]"},
		{int(1), "abc", "1", "[1 abc]"},
		{int8(1), 2, "1", "[1 2]"},
		{int16(1), S{2}, "1", "[1 2]"},
		{int32(1), []int{2}, "1", "[1 [2]]"},
		{int64(1), true, "1", "[1 true]"},
		{uint(1), false, "1", "[1 false]"},
		{uint8(1), 2, "1", "[1 2]"},
		{uint16(1), 2, "1", "[1 2]"},
		{uint32(1), 2, "1", "[1 2]"},
		{uint64(1), 2, "1", "[1 2]"},
		{uintptr(1), 2, "0x1", "[1 2]"},
		{true, false, "true", "[true false]"},
		{false, true, "false", "[false true]"},
		{float32(1), 2, "1", "[1 2]"},
		{float64(1), 2, "1", "[1 2]"},
		{S{1}, 2, "[1]", "[1 2]"},
		{S{1}, S{2}, "[1]", "[1 2]"},
		{S{1, 2}, S{3}, "[1 2]", "[1 2 3]"},
		{S{1, 2}, S{3, 4}, "[1 2]", "[1 2 3 4]"},
		{M{"a": 1}, 2, "map[a:1]", "[map[a:1] 2]"},
		{M{"a": 1}, M{"a": 2}, "map[a:1]", "map[a:2]"},
		{struct{ A int }{1}, 2, "{1}", "[{1} 2]"},
		{struct{ A int }{1}, struct{ A int }{2}, "{1}", "[{1} {2}]"},
		{nil, nil, "<nil>", "[<nil>]"},
	}
}

func checkTestResult(t *testing.T, w *bytes.Buffer, expected string, prefix string) {
	got := w.String()
	w.Reset()
	if expected != "" && expected[len(expected)-1] != '\n' {
		expected += "\n"
	}
	if got != expected {
		t.Errorf("%s: got `%s', expected `%s'", prefix, got, expected)
	}
}

func TestContextLogger_With(t *testing.T) {
	w := new(bytes.Buffer)
	for _, with := range [2]bool{true, false} {
		initMockLogger(w, with)
		for i, tc := range tests() {
			lv := logger.Level(i) % logger.NumLevel
			switch lv {
			case logger.TRACE:
				With(tc.data).Trace("")
			case logger.DEBUG:
				With(tc.data).Debug("")
			case logger.INFO:
				With(tc.data).Info("")
			case logger.WARN:
				With(tc.data).Warn("")
			default:
				With(tc.data).Error("")
			}
			checkTestResult(t, w, tc.output, fmt.Sprintf("%3dth with case, source=%v,level=%v", i, tc.data, lv))
			With(tc.data).With(tc.data2).Debug("")
			checkTestResult(t, w, tc.output2, fmt.Sprintf("%3dth with2 case, source=%v,%v", i, tc.data, tc.data2))
		}
	}
}

func TestContextLogger_WithJSON(t *testing.T) {
	w := new(bytes.Buffer)
	for _, with := range [2]bool{true, false} {
		initMockLogger(w, with)
		for i, tc := range tests() {
			WithJSON(tc.data).Info("")
			b, _ := json.Marshal(tc.data)
			checkTestResult(t, w, string(b), fmt.Sprintf("%3dth withjson case, source=%v", i, tc.data))
		}
	}
}
