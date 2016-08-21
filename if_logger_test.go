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
