package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestParserNextDsDollar(t *testing.T) {
	p := NewParser(nil)
	p.bindString("1 2 ds$")
	val, typ, err := p.Next()
	if err != nil || typ != LiteralTokenType(TokenInt) || val != Int(1) {
		t.Fatalf("first token = %v %v err=%v", val, typ, err)
	}
	val, typ, err = p.Next()
	if err != nil || typ != LiteralTokenType(TokenInt) || val != Int(2) {
		t.Fatalf("second token = %v %v err=%v", val, typ, err)
	}
	val, typ, err = p.Next()
	if err != nil || !IsPlainIdentToken(typ) || val != String("ds$") {
		t.Fatalf("third token = %v %v err=%v", val, typ, err)
	}
	_, typ, err = p.Next()
	if err != nil || typ != TokenEOF {
		t.Fatalf("eof token typ=%v err=%v", typ, err)
	}
}

func TestPushNextStackLayout(t *testing.T) {
	p := NewParser(nil)
	p.bindString("42")
	ds := NewStack[any]("test", 8)
	if err := p.PushNext(ds); err != nil {
		t.Fatal(err)
	}
	if ds.Len() != 2 {
		t.Fatalf("len = %d", ds.Len())
	}
	if ds.Pop().(Int) != Int(LiteralTokenType(TokenInt)) {
		t.Fatal("TOS should be TokenType")
	}
	if ds.Pop().(Int) != Int(42) {
		t.Fatal("value should be 42")
	}
}

func TestBootstrapREPLByeExits(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("bye\n1 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq should be cleared after exit")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (1 + should not run after bye)", rp.DataStack.Len())
	}
}

func TestBootstrapREPLByeExitsLocalTask(t *testing.T) {
	ResetTestRuntime()
	tk := CreateLocalTask()
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("bye\n1 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("core byeReq should be cleared after exit")
	}
	if tk.dc.dt.Len() != 0 {
		t.Fatal("local DC must not reserve bootstrap DT slots")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (1 + should not run after bye)", rp.DataStack.Len())
	}
}

func TestBootstrapREPLFromReaderAdd(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("1 2 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("got %v, want 3", got)
	}
}

func TestBootstrapREPLUnknownWordContinues(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("nosuchword\n1 2 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("got %v, want 3 after error line", got)
	}
}

func TestBootstrapREPLPipedMultilineRawString(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("`line1\nline2` drop\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after drop", rp.DataStack.Len())
	}
}

func TestBootstrapColonStartWordDoc(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString("// Adds two integers\n: add2 1 + ;\n")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("add2")
	if w == nil {
		t.Fatal("missing add2")
	}
	if w.doc != "Adds two integers" {
		t.Fatalf("add2.doc = %q, want Adds two integers", w.doc)
	}
}

func TestBootstrapColonSemicolonDiscardsTrailingComment(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString(": first // trailing\n;\n// prelude\n: second 1 ;\n")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	first := globalDictionaryContext.dict.GetByName("first")
	second := globalDictionaryContext.dict.GetByName("second")
	if first == nil || second == nil {
		t.Fatal("missing first or second")
	}
	if first.doc != "" {
		t.Fatalf("first.doc = %q, want empty", first.doc)
	}
	if second.doc != "prelude" {
		t.Fatalf("second.doc = %q, want prelude", second.doc)
	}
}

func TestBootstrapREPLColonInc(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": inc 1 + ;\n2 inc\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("got %v, want 3", got)
	}
}

func TestFindNameDsDollar(t *testing.T) {
	ResetTestRuntime()
	w, xt, ok := FindName("ds$")
	if !ok || w == nil || w.opCode != OpDsS {
		t.Fatalf("FindName(ds$) = %v %v %v", w, xt, ok)
	}
}

func TestSemicolonOutsideCompileErrors(t *testing.T) {
	ResetTestRuntime()
	tk, _ := NewTask(false)
	rp := NewTestRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(";\n1 2 +\n")),
		label: "test",
	}}
	if err := BootstrapREPL(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("got %v, want 3 after ; error", got)
	}
}
