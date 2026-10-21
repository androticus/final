package core

import "testing"

func TestOpLocalsPushesSameDO(t *testing.T) {
	task := NewTestTask()
	rp := NewTestRunProvider()
	cm := rp.ReturnStack.PushNew(0)
	RunOps(task, rp, OpLocals, OpLocals, OpStop)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2", rp.DataStack.Len())
	}
	a, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("TOS = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	b, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok || a != b {
		t.Fatalf("locals instances differ: %v vs %v", a, b)
	}
	if cm.locals != a {
		t.Fatal("locals not sticky on ColonFrameMark")
	}
}

func TestOpLocalsViaColonWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: locProbe
  l >ts
  99 t:.v!
  t:.v@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("locProbe")
	if w == nil {
		t.Fatal("missing locProbe")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 99 {
		t.Fatalf("TOS = %v, want 99", got)
	}
}

func TestLocalsCompoundStoreAndFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: locLDot
  99 l.v!
  l.v@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("locLDot")
	if w == nil {
		t.Fatal("missing locLDot")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 99 {
		t.Fatalf("TOS = %v, want 99", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (no locals DO left on DS)", rp.DataStack.Len())
	}
}

func TestLocalsCompoundNestedWalk(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: locNest
  DynamicObject.new^ l.obj!
  7 l.obj.x!
  l.obj.x@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("locNest")
	if w == nil {
		t.Fatal("missing locNest")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("TOS = %v, want 7", got)
	}
}

func TestLocalsCompoundNestedViaMidAt(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: locMidAt
  DynamicObject.new^ l.obj!
  3 l.obj.x!
  l.obj@.x@
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("locMidAt")
	if w == nil {
		t.Fatal("missing locMidAt")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("TOS = %v, want 3", got)
	}
}

func TestBareLStillPushesLocalsDO(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: locBare
  l
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := globalDictionaryContext.dict.GetByName("locBare")
	if w == nil {
		t.Fatal("missing locBare")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok || got == nil {
		t.Fatalf("TOS = %T, want *DynamicObject", rp.DataStack.Peek())
	}
}

func TestLocalsCompoundHandlePathNoDS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cm := rp.ReturnStack.PushNew(0)
	rp.DataStack.Push(Int(42))
	runHandleCompoundWithSegs(task, rp, []string{"l", ".", "n", "!"})
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d after store, want 0", rp.DataStack.Len())
	}
	if got := cm.GetLocals().Get("n"); got != Int(42) {
		t.Fatalf("locals.n = %v, want 42", got)
	}
	runHandleCompoundWithSegs(task, rp, []string{"l", ".", "n", "@"})
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want 42", got)
	}
}
