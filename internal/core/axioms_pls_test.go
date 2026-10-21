package core

import "testing"

func TestOpPlsPushesSameDO(t *testing.T) {
	task := NewTestTask()
	rp := NewTestRunProvider()
	RunOps(task, rp, OpPls, OpPls, OpStop)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2", rp.DataStack.Len())
	}
	a, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("TOS = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	b, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok || a != b {
		t.Fatalf("pls instances differ: %v vs %v", a, b)
	}
	if task.pls != a {
		t.Fatal("task.pls not sticky")
	}
}

func TestPlsCompoundStoreAndFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: plsDot
  99 pls.v!
  pls.v@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("plsDot")
	if w == nil {
		t.Fatal("missing plsDot")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 99 {
		t.Fatalf("TOS = %v, want 99", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (no pls DO left on DS)", rp.DataStack.Len())
	}
}

func TestPlsCompoundNestedWalk(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: plsNest
  DynamicObject.new^ pls.obj!
  7 pls.obj.x!
  pls.obj.x@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("plsNest")
	if w == nil {
		t.Fatal("missing plsNest")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("TOS = %v, want 7", got)
	}
}

func TestBarePlsStillPushesDO(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: plsBare
  pls
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("plsBare")
	if w == nil {
		t.Fatal("missing plsBare")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok || got == nil {
		t.Fatalf("TOS = %T, want *DynamicObject", rp.DataStack.Peek())
	}
}

func TestPlsCompoundHandlePathNoDS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(42))
	runHandleCompoundWithSegs(task, rp, []string{"pls", ".", "n", "!"})
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d after store, want 0", rp.DataStack.Len())
	}
	if got := task.GetPls().Get("n"); got != Int(42) {
		t.Fatalf("pls.n = %v, want 42", got)
	}
	runHandleCompoundWithSegs(task, rp, []string{"pls", ".", "n", "@"})
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want 42", got)
	}
}
