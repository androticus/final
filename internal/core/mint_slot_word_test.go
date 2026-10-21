package core

import "testing"

func TestMintVarConstDeferPrimitives(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()

	runInterpretScript(t, task, rp, "0 var: counter\n")
	w := task.dc.dict.GetByName("counter")
	if w == nil {
		t.Fatal("counter missing from local host")
	}
	if globalDictionaryContext.dict.GetByName("counter") != nil {
		t.Fatal("counter must not be in sealed core")
	}
	if w.flags&WordFlagDefined == 0 {
		t.Fatal("counter should be Defined")
	}
	runInterpretScript(t, task, rp, "counter @\n")
	if got, ok := rp.DataStack.Pop().(Int); !ok || got != 0 {
		t.Fatalf("counter @ = %v, want 0", rp.DataStack.Peek())
	}

	runInterpretScript(t, task, rp, "42 const: answer\n")
	c := task.dc.dict.GetByName("answer")
	if c == nil {
		t.Fatal("answer missing")
	}
	if c.dtIdx >= 0 {
		t.Fatalf("const dtIdx = %d, want -1 (OpDoLit embeds idx)", c.dtIdx)
	}
	runInterpretScript(t, task, rp, "answer\n")
	if got, ok := rp.DataStack.Pop().(Int); !ok || got != 42 {
		t.Fatalf("answer = %v, want 42", rp.DataStack.Peek())
	}

	runInterpretScript(t, task, rp, "defer: later\n")
	d := task.dc.dict.GetByName("later")
	if d == nil || d.flags&WordFlagDefer == 0 {
		t.Fatal("later missing or not Defer")
	}
}

func TestVarConstRefuseWhileCompiling(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad 0 var: x ;`)
	if err == nil {
		t.Fatal("expected compile error for var: inside colon")
	}
}
