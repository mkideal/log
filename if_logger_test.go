package log

import (
	"bytes"
	"testing"
)

func TestIfLogger_IfElse(t *testing.T) {
	w := new(bytes.Buffer)
	initMockLogger(w, true)

	If(true).Info("if-true")
	checkTestResult(t, w, "if-true", "if-true")

	If(true).Info("if-true").
		Else().Info("else")
	checkTestResult(t, w, "if-true", "if-true-else")

	If(true).Info("if-true").
		ElseIf(true).Info("elseif-true")
	checkTestResult(t, w, "if-true", "if-true-elseif-true")

	If(true).Info("if-true").
		ElseIf(false).Info("elseif-false").
		Else().Info("else")
	checkTestResult(t, w, "if-true", "if-true-elseif-false-else")

	If(false).Info("if-false")
	checkTestResult(t, w, "", "if-false")

	If(false).Info("if-false").
		Else().Info("else")
	checkTestResult(t, w, "else", "if-false-else")

	If(false).Info("if-false").
		ElseIf(true).Info("elseif-true")
	checkTestResult(t, w, "elseif-true", "if-false-elseif-true")

	If(false).Info("if-false").
		ElseIf(false).Info("elseif-false").
		Else().Info("else")
	checkTestResult(t, w, "else", "if-false-elseif-false-else")
}

func TestIfLogger_Print(t *testing.T) {
	w := new(bytes.Buffer)
	initMockLogger(w, true)

	If(true).Trace("trace").Debug("debug").Info("info").Warn("warn").Error("error")

	checkTestResult(t, w, "trace\ndebug\ninfo\nwarn\nerror\n", "iflogger-print")
}

func TestIfLogger_With(t *testing.T) {
	w := new(bytes.Buffer)
	initMockLogger(w, true)

	If(true).With(1).Info("")
	checkTestResult(t, w, "1", "true-with-1")

	If(false).With(1).Info("")
	checkTestResult(t, w, "", "false-with-1")

	If(true).WithJSON(struct{ A int }{1}).Info("")
	checkTestResult(t, w, `{"A":1}`, "true-withjson-1")

	If(false).WithJSON(struct{ A int }{1}).Info("")
	checkTestResult(t, w, "", "false-withjson-1")
}
