package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestBootstrapREPLGetNextTokenFooLeavesIdentOnDS(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	var stderr StringBuilder
	rp.Stderr = &stderr
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("getNextToken foo\n")),
		label: "test",
	}}
	BootstrapREPL(tk, rp)
	if strings.Contains(stderr.ToString(), "unknown word") {
		t.Fatalf("stderr = %q, want no unknown word error", stderr.ToString())
	}
	wantDS(t, rp, Int(TokenIdent), String("foo"))
}

// After compileFromFile, getNextToken must read from the REPL stdin parser, not the
// compile unit parser left on rp.Parser (regression: inner getNextToken skipped foo,
// outer loop re-dispatched foo as unknown word).
func TestBootstrapREPLGetNextTokenAfterCompileFromFile(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	var stderr StringBuilder
	rp.Stderr = &stderr
	path := writeTempFinalSource(t, ": noop ;\n")
	script := `"` + path + `" compileFromFile getNextToken foo` + "\n"
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(script)),
		label: "test",
	}}
	BootstrapREPL(tk, rp)
	if strings.Contains(stderr.ToString(), `unknown word "foo"`) {
		t.Fatalf("stderr = %q, want getNextToken to consume foo from REPL line", stderr.ToString())
	}
	wantDS(t, rp, Int(TokenIdent), String("foo"))
}
