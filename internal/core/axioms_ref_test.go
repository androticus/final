package core

import "testing"

func TestNewDTRefFetchStore(t *testing.T) {
	dt := NewDataTable(16, 4)
	idx := dt.Append(Int(42))
	ref := NewDTSlotRef(dt, idx, false)
	got, ok := ref.Fetch().(Int)
	if !ok || got != 42 {
		t.Fatalf("Fetch = %v, want Int(42)", got)
	}
	ref.Store(Int(99))
	if got := dt.Read(idx).(Int); got != 99 {
		t.Fatalf("DT[%d] = %v, want 99", idx, got)
	}
}

func TestNewDTRefReadOnlyPanicsOnStore(t *testing.T) {
	dt := NewDataTable(16, 4)
	idx := dt.Append(Int(1))
	ref := NewDTSlotRef(dt, idx, true)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic storing read-only DT ref")
		}
	}()
	ref.Store(Int(2))
}

func TestOpFetchDTRef(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	dt := globalDictionaryContext.dt
	idx := dt.Append(Int(42))
	rp.DataStack.Push(NewDTSlotRef(dt, idx, false))
	RunOps(NewTestTask(), rp, OpFetch, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("fetch = %v, want Int(42)", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want empty after var fetch", rp.DataStack.Len())
	}
}

func TestOpStoreDTRef(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	dt := globalDictionaryContext.dt
	idx := dt.Append(Int(0))
	rp.DataStack.Push(Int(99))
	rp.DataStack.Push(NewDTSlotRef(dt, idx, false))
	RunOps(NewTestTask(), rp, OpStore, OpStop)
	if got := dt.Read(idx).(Int); got != 99 {
		t.Fatalf("DT[%d] = %v, want 99", idx, got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want empty after var store", rp.DataStack.Len())
	}
}

func TestCompileReservedVarByeReq(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: touchbyeReq byeReq ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("touchbyeReq")
	if w == nil {
		t.Fatal("touchbyeReq missing")
	}
	ct := globalDictionaryContext.ct
	body := int(w.colonCode)
	if ct.Read(body) != int(OpDoCSSelect) {
		t.Fatalf("CT[%d] = %d, want OpDoCSSelect", body, ct.Read(body))
	}
	if ct.Read(body+1) != int(csDataStack) {
		t.Fatalf("CT[%d] = %d, want csDataStack", body+1, ct.Read(body+1))
	}
	if ct.Read(body+2) != int(OpCtImmDs) {
		t.Fatalf("CT[%d] = %d, want OpCtImmDs", body+2, ct.Read(body+2))
	}
	if ct.Read(body+3) != ReservedDTSlotByeReq {
		t.Fatalf("CT[%d] = %d, want ReservedDTSlotByeReq %d", body+3, ct.Read(body+3), ReservedDTSlotByeReq)
	}
	if ct.Read(body+4) != int(OpBootstrapDoVar) {
		t.Fatalf("CT[%d] = %d, want OpBootstrapDoVar", body+4, ct.Read(body+4))
	}
}

func TestBracketBlockRunsInterpretableWords(t *testing.T) {
	t.Skip("Go [ ] defer removed; Final [ ] words deferred to slice-007 deliverable 4")
}
