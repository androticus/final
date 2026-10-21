package core

import "testing"

func TestAllotDtSlotToWordSetsDtIdxAndValue(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	w := bootstrapCreate(cs, "myvar")
	rp.DataStack.Push(Int(42))
	runAllotDtSlotToWord(task, rp)
	if w.dtIdx < 0 {
		t.Fatalf("dtIdx = %d, want >= 0", w.dtIdx)
	}
	got, ok := globalDictionaryContext.dt.Read(w.dtIdx).(Int)
	if !ok || got != 42 {
		t.Fatalf("DT[%d] = %v, want Int(42)", w.dtIdx, globalDictionaryContext.dt.Read(w.dtIdx))
	}
	cs.DC.dict.ClearBuilding()
}

func TestAllotDtSlotToWordOutsideColonPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(1))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic outside colon definition")
		}
	}()
	runAllotDtSlotToWord(task, rp)
}

func TestOpDTRefPushesMutableRef(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	dt := globalDictionaryContext.dt
	idx := dt.Append(Int(7))
	rp.DataStack.Push(Int(idx))
	RunOps(NewTestTask(), rp, OpDTRef, OpStop)
	ref, ok := rp.DataStack.Pop().(ValueRef)
	if !ok {
		t.Fatalf("stack top = %T, want ValueRef", rp.DataStack.Peek())
	}
	got, ok := ref.Fetch().(Int)
	if !ok || got != 7 {
		t.Fatalf("Fetch = %v, want Int(7)", ref.Fetch())
	}
	ref.Store(Int(99))
	if dt.Read(idx).(Int) != 99 {
		t.Fatalf("DT[%d] = %v, want 99", idx, dt.Read(idx))
	}
}

func TestOpDTConstPushesValue(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	dt := globalDictionaryContext.dt
	idx := dt.Append(String("hello"))
	rp.DataStack.Push(Int(idx))
	RunOps(NewTestTask(), rp, OpDTConst, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("stack = %v, want String(hello)", rp.DataStack.Peek())
	}
}
