package core

import (
	"fmt"
	"strings"
	"testing"
)

const jsonParseLimitProbeSrc = `
: jsonParseLimitProbe
  try
    json>
    true
  catch
    false
  /try
;
`

func jsonParseLimitsList(t *testing.T, task *Task, rp *RunProvider) *IntList {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName("jsonParse_limits")
	if w == nil {
		t.Fatal("missing jsonParse_limits")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("jsonParse_limits: %v", err)
	}
	l, ok := rp.DataStack.Pop().(*IntList)
	if !ok {
		t.Fatalf("jsonParse_limits TOS = %T, want *IntList", rp.DataStack.Peek())
	}
	return l
}

func pokeJSONParseLimit(t *testing.T, limits *IntList, idx, cap int) func() {
	t.Helper()
	old := limits.At(idx)
	limits.Put(idx, cap)
	return func() { limits.Put(idx, old) }
}

func compileJSONParseLimitProbe(t *testing.T, task *Task, rp *RunProvider) {
	t.Helper()
	if err := compileBootstrapSource(task, rp, jsonParseLimitProbeSrc); err != nil {
		t.Fatal(err)
	}
}

func runJSONParseLimitProbe(t *testing.T, task *Task, rp *RunProvider, json string) (ok bool, msg string, root any) {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName("jsonParseLimitProbe")
	if w == nil {
		t.Fatal("missing jsonParseLimitProbe")
	}
	rp.DataStack.Push(String(json))
	if err := runWordInlineExec(task, rp, w); err != nil {
		if fault, ok := err.(*RunFault); ok {
			t.Fatalf("probe: %s\n%s", fault.Message, FormatRunFaultReport(fault))
		}
		t.Fatalf("probe: %v", err)
	}
	flag, isBool := rp.DataStack.Pop().(Bool)
	if !isBool {
		t.Fatalf("probe flag = %T, want Bool", flag)
	}
	if flag {
		return true, "", rp.DataStack.Pop()
	}
	m, _ := rp.DataStack.Pop().(String)
	return false, string(m), nil
}

func assertJSONParseLimitThrow(t *testing.T, task *Task, rp *RunProvider, json, label string, cap int) {
	t.Helper()
	ok, msg, _ := runJSONParseLimitProbe(t, task, rp, json)
	if ok {
		t.Fatalf("parse %q succeeded, want throw (%s limit %d)", json, label, cap)
	}
	if !strings.Contains(msg, label) {
		t.Fatalf("throw %q: want label %q", msg, label)
	}
	if !strings.Contains(msg, fmt.Sprintf("%d", cap)) {
		t.Fatalf("throw %q: want cap %d", msg, cap)
	}
}

func assertJSONParseLimitOK(t *testing.T, task *Task, rp *RunProvider, json string) {
	t.Helper()
	ok, msg, _ := runJSONParseLimitProbe(t, task, rp, json)
	if !ok {
		t.Fatalf("parse %q threw %q, want success", json, msg)
	}
}

func nestedJSONArrays(depth int) string {
	return strings.Repeat("[", depth) + "null" + strings.Repeat("]", depth)
}

func identKeyJSON(n int) string {
	return `{"` + strings.Repeat("k", n) + `":1}`
}

func TestJsonParseLimitsInputLength(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	limits := jsonParseLimitsList(t, task, rp)
	defer pokeJSONParseLimit(t, limits, 0, 4)()
	assertJSONParseLimitOK(t, task, rp, "true")
	assertJSONParseLimitThrow(t, task, rp, "false", "input length", 4)
}

func TestJsonParseLimitsNesting(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	assertJSONParseLimitOK(t, task, rp, nestedJSONArrays(64))
	assertJSONParseLimitThrow(t, task, rp, nestedJSONArrays(65), "nesting depth", 64)
}

func TestJsonParseLimitsNodes(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	limits := jsonParseLimitsList(t, task, rp)
	defer pokeJSONParseLimit(t, limits, 2, 2)()
	assertJSONParseLimitOK(t, task, rp, "[0]")
	assertJSONParseLimitThrow(t, task, rp, "[0,0]", "node count", 2)
}

func TestJsonParseLimitsIdentVsString(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	limits := jsonParseLimitsList(t, task, rp)

	restoreIdent := pokeJSONParseLimit(t, limits, 3, 2)
	assertJSONParseLimitOK(t, task, rp, `{"ab":1}`)
	assertJSONParseLimitThrow(t, task, rp, `{"abc":1}`, "identifier length", 2)
	restoreIdent()

	assertJSONParseLimitOK(t, task, rp, identKeyJSON(2048))
	assertJSONParseLimitThrow(t, task, rp, identKeyJSON(2049), "identifier length", 2048)

	restoreString := pokeJSONParseLimit(t, limits, 4, 2)
	assertJSONParseLimitOK(t, task, rp, `"ab"`)
	assertJSONParseLimitThrow(t, task, rp, `"abc"`, "string length", 2)
	assertJSONParseLimitThrow(t, task, rp, `{"ok":"abc"}`, "string length", 2)
	restoreString()
}

func TestJsonParseLimitsNumber(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	limits := jsonParseLimitsList(t, task, rp)
	defer pokeJSONParseLimit(t, limits, 5, 2)()
	assertJSONParseLimitOK(t, task, rp, "12")
	assertJSONParseLimitThrow(t, task, rp, "123", "number token length", 2)
}

func TestJsonParseLimitsUnicodeStringCountsOneRune(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	compileJSONParseLimitProbe(t, task, rp)
	limits := jsonParseLimitsList(t, task, rp)
	defer pokeJSONParseLimit(t, limits, 4, 1)()
	assertJSONParseLimitOK(t, task, rp, `"\u0041"`)
	assertJSONParseLimitThrow(t, task, rp, `"\u0041\u0042"`, "string length", 1)
}
