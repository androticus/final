package core

import "testing"

// Smoke: primitive receiver pairs at OO boundary and invokes instance method via compound path.
func TestSmokePrimitiveToStringViaCompound(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(String("hi"))
	runHandleCompoundWithSegs(task, rp, []string{".", "_toString", "^"})
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + result)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hi" {
		t.Fatalf("TOS = %v, want hi", got)
	}
}

// Smoke: _renderJson on Int writes fragment to StringBuilder on z: (handler path).
func TestSmokePrimitiveRenderJson(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.ZStack.Push(sb)
	ths := rp.Resolve(Int(42))
	f := &RunFrame{RP: rp}
	GetTypeRegistrar().Get("Int").InvokeNamed(f, rp.ZStack, ths, RuntimeDispatchKindInstanceMethodByName, "_renderJson")
	if got := sb.ToString(); got != "42" {
		t.Fatalf("sb = %q, want 42", got)
	}
}

// Smoke: PropertyRef fetch/store via compound + opcode path uses ref.Parent (no re-Resolve).
func TestSmokePropertyRefFetchStore(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name", "~"})
	RunOps(task, rp, OpDup, OpFetch, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("fetch = %v, want hello", got)
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want receiver + parked ref on TOS", rp.DataStack.Len())
	}
	if _, ok := rp.DataStack.Peek().(*PropertyRef); !ok {
		t.Fatalf("TOS = %T, want PropertyRef parked for !", rp.DataStack.Peek())
	}
	rp.DataStack.Push(String("fresh"))
	RunOps(task, rp, OpSwap, OpStore, OpStop)
	if w.name != "fresh" {
		t.Fatalf("w.name = %q, want fresh", w.name)
	}
}

// Smoke: DTSlotRef @ fetch pairs primitive ref cell at boundary.
func TestSmokeDTSlotRefFetch(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	dt := globalDictionaryContext.dt
	idx := dt.Append(Bool(false))
	rp.DataStack.Push(NewDTSlotRef(dt, idx, false))
	RunOps(NewTestTask(), rp, OpFetch, OpStop)
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got {
		t.Fatalf("fetch = %v, want Bool(false)", got)
	}
}

// Smoke: checkResetByeReq reads DTSlotRef without stack leak after clear.
func TestSmokeCheckResetByeReq(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	SetByeReq(globalDictionaryContext, true)
	runDeferWordByName(t, task, rp, "checkResetByeReq")
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || !got {
		t.Fatalf("checkResetByeReq = %v, want Bool(true)", got)
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq should be cleared")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after byeReq clear", rp.DataStack.Len())
	}
}
