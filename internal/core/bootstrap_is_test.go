package core

import "testing"

func TestIsBindsViaAmpOfInterpret(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `
: isAmpOfTarget 42 ;
defer: isAmpOfDefer
&of isAmpOfTarget is isAmpOfDefer
`); err != nil {
		t.Fatal(err)
	}
	runInterpretScript(t, task, rp, "isAmpOfDefer\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("isAmpOfDefer => %v (%T), want 42", got, got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}

func TestIsBindsViaAmpOfInsideColon(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `
: isColonTarget 7 ;
defer: isColonDefer
: bindIs &of isColonTarget is isColonDefer ;
`); err != nil {
		t.Fatal(err)
	}
	runInterpretScript(t, task, rp, "bindIs\nisColonDefer\n")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("isColonDefer => %v (%T), want 7", got, got)
	}
}

func TestIsIsImmediateNotCompileOnly(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("is")
	if w == nil {
		t.Fatal("missing is")
	}
	if !w.IsImmediate() {
		t.Fatal("is should be Immediate")
	}
	if w.flags&WordFlagCompileOnly != 0 {
		t.Fatal("is should not be CompileOnly")
	}
}
