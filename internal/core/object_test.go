package core

import (
	"strings"
	"testing"
)

func pushDynamicObjectOnDS(t *testing.T, rp *RunProvider, o *DynamicObject) {
	t.Helper()
	rp.DataStack.Push(o)
}

func TestDynamicObjectDotFieldImpliedRead(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	o := NewDynamicObject()
	o.Set("myField", Int(42))
	pushDynamicObjectOnDS(t, rp, o)
	runHandleCompoundWithSegs(task, rp, []string{".", "myField"})
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
}

func TestDynamicObjectDotFieldAddOnSet(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	o := NewDynamicObject()
	rp.TempStack.Push(o)
	rp.DataStack.Push(String("hello"))
	cp := CompoundPathWithSelectorForTest([]string{".", "greeting", "!"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, cp); err != nil {
		t.Fatal(err)
	}
	if !o.HasField("greeting") {
		t.Fatal("field greeting not created")
	}
	if got := o.Get("greeting"); got != String("hello") {
		t.Fatalf("greeting = %v, want hello", got)
	}
}

func TestDynamicObjectDotFieldStoreRejectedOnDS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	o := NewDynamicObject()
	pushDynamicObjectOnDS(t, rp, o)
	rp.DataStack.Push(String("hello"))
	segs := CompoundPathForTest([]string{".", "greeting", "!"})
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected error for DS-rooted store compound")
	}
	if !strings.Contains(err.Error(), "store path requires t:/x:/y:/z:/o: prefix") {
		t.Fatalf("err = %v", err)
	}
}

func TestDynamicObjectDotFieldStoreRejectedOnExplicitDSPrefix(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	o := NewDynamicObject()
	pushDynamicObjectOnDS(t, rp, o)
	rp.DataStack.Push(String("hello"))
	cp := CompoundPathWithSelectorForTest([]string{".", "greeting", "!"}, csDataStack)
	err := dispatchCompoundToken(p, task, rp, cp)
	if err == nil {
		t.Fatal("expected error for d:.greeting! store")
	}
	if !strings.Contains(err.Error(), "store path requires t:/x:/y:/z:/o: prefix") {
		t.Fatalf("err = %v", err)
	}
}

func TestDynamicObjectDotFieldMissingPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	o := NewDynamicObject()
	pushDynamicObjectOnDS(t, rp, o)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing field")
		}
		msg, ok := r.(error)
		if !ok || !strings.Contains(msg.Error(), "not found") {
			t.Fatalf("panic = %v, want missing-key error", r)
		}
	}()
	runHandleCompoundWithSegs(task, rp, []string{".", "missing"})
}

func TestDynamicObjectPropertyRefRMW(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	o := NewDynamicObject()
	o.Set("tag", String("old"))
	pushDynamicObjectOnDS(t, rp, o)
	runHandleCompoundWithSegs(task, rp, []string{".", "tag", "~"})
	RunOps(task, rp, OpDup, OpFetch, OpStop)
	rp.DataStack.Pop()
	rp.DataStack.Push(String("new"))
	RunOps(task, rp, OpSwap, OpStore, OpStop)
	if got := o.Get("tag"); got != String("new") {
		t.Fatalf("tag = %v, want new", got)
	}
}

func TestDynamicObjectOpAtIndexByFieldCreationOrder(t *testing.T) {
	ResetTestRuntime()
	o := NewDynamicObject()
	o.Set("mass", Float(1.5))
	o.Set("x", Float(10))
	o.Set("y", Float(20))

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Int(0))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != Float(1.5) {
		t.Fatalf("@() 0 = %v, want 1.5", got)
	}
	if rp.DataStack.Pop() != o {
		t.Fatal("receiver not preserved on DS after @()")
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Int(2))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != Float(20) {
		t.Fatalf("@() 2 = %v, want 20", got)
	}
}

func TestDynamicObjectOpBangIndexByFieldCreationOrder(t *testing.T) {
	ResetTestRuntime()
	o := NewDynamicObject()
	o.Set("mass", Float(1))
	o.Set("vx", Float(0))
	o.Set("vy", Float(0))

	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Float(99.5))
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpBangIndex, OpStop)
	if rp.DataStack.Pop() != o {
		t.Fatal("receiver not preserved on DS after !()")
	}
	if got := o.Get("vx"); got != Float(99.5) {
		t.Fatalf("vx after !() 1 = %v, want 99.5", got)
	}
	if !o.HasField("vx") || o.Len() != 3 {
		t.Fatalf("field names/count changed: has vx=%v len=%d", o.HasField("vx"), o.Len())
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Float(-3))
	rp.DataStack.Push(Int(0))
	RunOps(NewTestTask(), rp, OpBangIndex, OpStop)
	if got := o.Get("mass"); got != Float(-3) {
		t.Fatalf("mass after !() 0 = %v, want -3", got)
	}

	rp = NewTestRunProvider()
	pushTypedOnContextStack(rp, o)
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	if got := rp.DataStack.Pop(); got != Float(99.5) {
		t.Fatalf("@() 1 after store = %v, want 99.5", got)
	}
}

func fillDynamicObject(o *DynamicObject, n int) {
	for i := 0; i < n; i++ {
		o.Set(sfmTestKey(i), Int(i))
	}
}

func TestDynamicObjectCloneAfterSpill(t *testing.T) {
	n := sfmMaxIteratedFields + 6
	o := NewDynamicObject()
	fillDynamicObject(o, n)
	c := o.Clone()
	if c.Len() != n {
		t.Fatalf("clone Len = %d, want %d", c.Len(), n)
	}
	if c.fields.namesMap == nil {
		t.Fatal("clone should spill after copying past the scan limit")
	}
	if c.Get(sfmTestKey(0)) != Int(0) || c.Get(sfmTestKey(n-1)) != Int(n-1) {
		t.Fatal("clone field values mismatch")
	}
	c.Set(sfmTestKey(0), Int(99))
	if o.Get(sfmTestKey(0)) != Int(0) {
		t.Fatal("clone Set must not mutate original")
	}
}
