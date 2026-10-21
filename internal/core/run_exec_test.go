package core

import "testing"

func TestOpExecDeferColonViaInlineExec(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString("foo ")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	colonW := globalDictionaryContext.dict.GetByName(":")
	if colonW == nil {
		t.Fatal("missing :")
	}
	if err := runExecWord(task, rp, colonW); err != nil {
		t.Fatalf("runExecWord(:): %v", err)
	}
	if !GetCompileState(task.compilerForTarget()) {
		t.Fatal("expected compile state after :")
	}
	if globalDictionaryContext.dict.GetByName("foo") != nil {
		t.Fatal("expected : foo to leave word on building stack, not in dictionary")
	}
	if w := globalDictionaryContext.dict.PeekBuilding(); w == nil || w.name != "foo" {
		t.Fatal("expected building word foo after :")
	}
}
