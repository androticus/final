package core

import (
	"strings"
	"testing"
)

func TestAmpOfInterpretPushesXT(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dropW := globalDictionaryContext.dict.GetByName("drop")
	if dropW == nil {
		t.Fatal("missing drop")
	}
	want := dropW.XT()
	runInterpretScript(t, task, rp, "&of drop\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || XT(got) != want {
		t.Fatalf("TOS = %v (%T), want XT %v", got, got, want)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}

func TestAmpOfInterpretExecSmoke(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Bool(false))
	runInterpretScript(t, task, rp, "&of not exec\n")
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || !bool(got) {
		t.Fatalf("after &of not exec = %v, want true", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}

func TestAmpOfCommaCompileLeavesXT(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dropW := globalDictionaryContext.dict.GetByName("drop")
	if dropW == nil {
		t.Fatal("missing drop")
	}
	want := dropW.XT()
	if err := compileBootstrapSource(task, rp, `: t &of, drop ;`); err != nil {
		t.Fatal(err)
	}
	runInterpretScript(t, task, rp, "t\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || XT(got) != want {
		t.Fatalf("t => %v (%T), want XT %v", got, got, want)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}

func TestAmpOfCompileLeavesXT(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dropW := globalDictionaryContext.dict.GetByName("drop")
	if dropW == nil {
		t.Fatal("missing drop")
	}
	want := dropW.XT()
	if err := compileBootstrapSource(task, rp, `: t &of drop ;`); err != nil {
		t.Fatal(err)
	}
	runInterpretScript(t, task, rp, "t\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || XT(got) != want {
		t.Fatalf("t => %v (%T), want XT %v", got, got, want)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}

func TestAmpOfIsImmediateNotCompileOnly(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("&of")
	if w == nil {
		t.Fatal("missing &of")
	}
	if !w.IsImmediate() {
		t.Fatal("&of should be Immediate")
	}
	if w.flags&WordFlagCompileOnly != 0 {
		t.Fatal("&of should not be CompileOnly")
	}
}

func TestAmpOfMissingNameErrors(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString("&of nosuchAmpOfWord\n")
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "word not found") {
		t.Fatalf("err = %v, want word not found", err)
	}
}

func TestAmpOfCommaMissingNameErrors(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad &of, nosuchAmpOfCommaWord ;`)
	if err == nil {
		t.Fatal("expected compile error for missing name")
	}
	if !strings.Contains(err.Error(), "word not found") {
		t.Fatalf("err = %v, want word not found", err)
	}
}

func TestAmpOfCommaIsCompileOnlyImmediate(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("&of,")
	if w == nil {
		t.Fatal("missing &of,")
	}
	if !w.IsImmediate() {
		t.Fatal("&of, should be Immediate")
	}
	if w.flags&WordFlagCompileOnly == 0 {
		t.Fatal("&of, should be CompileOnly")
	}
}
