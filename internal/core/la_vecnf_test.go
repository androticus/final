package core

import "testing"

func TestVecNfNewZeroFilled(t *testing.T) {
	ResetTestRuntime()
	v := NewVecNf(4)
	if v.Len() != 4 {
		t.Fatalf("len=%d, want 4", v.Len())
	}
	for i, x := range v.data {
		if x != 0 {
			t.Fatalf("data[%d]=%v, want 0", i, x)
		}
	}
}

func TestVecNfNewPanicsNonPositive(t *testing.T) {
	for _, n := range []int{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewVecNf(%d): expected panic", n)
				}
			}()
			_ = NewVecNf(n)
		}()
	}
}

func TestVecNfAtPut(t *testing.T) {
	v := NewVecNf(3)
	v.Put(1, 3.5)
	if got := v.At(1); got != 3.5 {
		t.Fatalf("At(1)=%v, want 3.5", got)
	}
}

func TestVecNfAtOutOfRangePanics(t *testing.T) {
	v := NewVecNf(2)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = v.At(2)
}

func TestVecNfClassNewViaStack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(5))
	f, ths := memberRunFrame(rp, vecNfType)
	mm := vecNfType.GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing VecNf.new")
	}
	mm.invokeCall(f, ths)
	v, ok := rp.DataStack.Pop().(*VecNf)
	if !ok || v.Len() != 5 {
		t.Fatalf("got %v, want length-5 VecNf", rp.DataStack.Peek())
	}
	_ = task
}

func TestVecNfItemsMetadataOrdinal(t *testing.T) {
	ResetTestRuntime()
	im := vecNfType.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("VecNf must register flat @() / !()")
	}
	if im.getByMultiIndex != nil || im.setByMultiIndex != nil {
		t.Fatal("VecNf should not register @[] / ![]")
	}
	v := NewVecNf(3)
	im.SetByOrdinal(v, Int(2), Float(7))
	if v.At(2) != 7 {
		t.Fatalf("At(2)=%v, want 7", v.At(2))
	}
	if got := im.GetByOrdinal(v, Int(2)).(Float); got != 7 {
		t.Fatalf("@() = %v, want 7", got)
	}
}

func TestOpAtBangIndexVecNf(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	v := NewVecNf(4)
	rp.DataStack.Push(v)
	rp.DataStack.Push(Float(1.25))
	rp.DataStack.Push(Int(3))
	RunOps(task, rp, OpBangIndex, OpStop)
	if got := v.At(3); got != 1.25 {
		t.Fatalf("after !() At(3)=%v, want 1.25", got)
	}
	rp.DataStack.Push(Int(3))
	RunOps(task, rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 1.25 {
		t.Fatalf("@() = %v, want 1.25", got)
	}
}

func TestOpAtIndexVecNfViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	v := NewVecNf(2)
	v.Put(1, 9)
	rp.DataStack.Push(v)
	// compound: .@() with TOS index — use OpAtIndex after pushing index
	rp.DataStack.Push(Int(1))
	RunOps(NewTestTask(), rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 9 {
		t.Fatalf("got %v, want 9", got)
	}
}

func TestVecNfIterable(t *testing.T) {
	v := NewVecNf(3)
	v.Put(0, 1)
	v.Put(1, 2)
	v.Put(2, 3)
	it := v.GetIter()
	var got []Float
	for it.GetNext() {
		got = append(got, it.Current().(Float))
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("iter got %v, want [1 2 3]", got)
	}
}

func TestOpTryGetIterOnVecNf(t *testing.T) {
	ResetTestRuntime()
	v := NewVecNf(2)
	rp := NewTestRunProvider()
	rp.DataStack.Push(v)
	RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
	ok := rp.DataStack.Pop().(Bool)
	if !ok {
		t.Fatal("want true")
	}
	it, is := rp.DataStack.Pop().(*VecNfIter)
	if !is || it.vec != v {
		t.Fatalf("want *VecNfIter, got %T", rp.DataStack.Peek())
	}
}
