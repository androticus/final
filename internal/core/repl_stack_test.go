package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestBootstrapInitLeavesCleanStack(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	assertEmptyDataStack(t, rp, "after reset")
	if cs := NewCompilerState(globalDictionaryContext); GetCompileState(cs) {
		t.Fatal("compiler still in compile state after reset")
	}
}

func TestIfImmediateLeavesOrig(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	cs := tk.compilerForTarget()
	bootstrapCreate(cs, "testif")
	bootstrapEmitLiteral(cs, Int(1))
	ifW := globalDictionaryContext.dict.GetByName("if")
	if ifW == nil {
		t.Fatal("missing if")
	}
	if err := runWordInlineExec(tk, rp, ifW); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("after if ds len %d, want 1 (orig)", rp.DataStack.Len())
	}
}

func TestNifImmediateLeavesOrig(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	cs := tk.compilerForTarget()
	bootstrapCreate(cs, "testnif")
	bootstrapEmitLiteral(cs, Int(1))
	nifW := globalDictionaryContext.dict.GetByName("nif")
	if nifW == nil {
		t.Fatal("missing nif")
	}
	if err := runWordInlineExec(tk, rp, nifW); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("after nif ds len %d, want 1 (orig)", rp.DataStack.Len())
	}
}

func TestBootstrapREPLColonDiv0(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	var stdout, stderr StringBuilder
	rp.Stdout = &stdout
	rp.Stderr = &stderr
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": div0 1 0 / ;\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("div0") == nil {
		t.Fatalf("div0 not defined; stdout=%q stderr=%q", stdout.ToString(), stderr.ToString())
	}
}

func TestBootstrapREPLColonMinimal(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	var stdout, stderr StringBuilder
	rp.Stdout = &stdout
	rp.Stderr = &stderr
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": foo ;\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("foo") == nil {
		t.Fatalf("foo not defined; stdout=%q stderr=%q compiling=%v", stdout.ToString(), stderr.ToString(), GetCompileState(tk.compilerForTarget()))
	}
}

func TestBootstrapREPLColonTestIfElse(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	var stderr StringBuilder
	rp.Stderr = &stderr
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": test-if if 99 else 0 then ;\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatalf("BootstrapREPL: %v stderr=%q", err, stderr.ToString())
	}
	if globalDictionaryContext.dict.GetByName("test-if") == nil {
		t.Fatalf("test-if not defined after REPL line; stderr=%q", stderr.ToString())
	}
}

func TestCompileTestIfElseAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(tk, rp, `: test-if if 99 else 0 then ;`); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("test-if") == nil {
		t.Fatal("missing test-if")
	}
}

func TestCompileTestIfAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(tk, rp, `: testif 1 if 99 then ;`); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("testif") == nil {
		t.Fatal("missing testif")
	}
}

func TestCompileForeachRangeAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(tk, rp, `: sum10 0 0 4 range foreach i + /foreach ;`); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("sum10") == nil {
		t.Fatal("missing sum10")
	}
}

func TestRunForeachRangeSumAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": sum10 0 0 4 range foreach i + /foreach ;\nsum10\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 10 {
		t.Fatalf("sum10 = %v, want 10", got)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("return stack not clean: len %d", rp.ReturnStack.Len())
	}
}

func TestRunForeachEmptyRangeAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": empty 0 0 -1 range foreach 1 + /foreach ;\nempty\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("return stack not clean: len %d", rp.ReturnStack.Len())
	}
}

func TestRunBeginAgainAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": bump 0 begin 1 + dup 3 == until ;\nbump\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("bump = %v, want 3", got)
	}
}

func TestRunWhileRepeatAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": down 3 begin dup 0 <> while 1 - repeat ;\ndown\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 0 {
		t.Fatalf("down = %v, want 0", got)
	}
}

func assertEmptyDataStack(t *testing.T, rp *RunProvider, label string) {
	t.Helper()
	if rp.DataStack.Len() != 0 {
		for i := rp.DataStack.Len() - 1; i >= 0; i-- {
			t.Logf("%s ds[%d] = %T %v", label, i, *rp.DataStack.AddrAt(i), *rp.DataStack.AddrAt(i))
		}
		t.Fatalf("%s: DataStack len = %d, want 0", label, rp.DataStack.Len())
	}
}
