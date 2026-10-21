package core

import (
	"sync"
	"testing"
)

func TestAtomicListFinalTypeAndNew(t *testing.T) {
	ResetTestRuntime()
	l := NewAtomicAnyList()
	if l.FinalType() != atomicAnyListType {
		t.Fatalf("FinalType = %v, want AtomicList", l.FinalType())
	}
	if GetTypeRegistrar().Get("AtomicList") != atomicAnyListType {
		t.Fatal("AtomicList not registered")
	}
	if atomicAnyListType.GetParent() != anyListType {
		t.Fatalf("AtomicList.parent = %v, want List", atomicAnyListType.GetParent())
	}

	sl := NewAtomicStringList()
	if sl.FinalType() != atomicStringListType {
		t.Fatalf("FinalType = %v, want AtomicStringList", sl.FinalType())
	}
	if atomicStringListType.GetParent() != GetTypeRegistrar().Get("StringList") {
		t.Fatalf("AtomicStringList.parent = %v, want StringList", atomicStringListType.GetParent())
	}
}

func TestAtomicListClassNew(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	mm := atomicAnyListType.ClassMethodByName("new")
	f, ths := memberRunFrame(rp, atomicAnyListType)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(*AtomicAnyList)
	if !ok || got == nil {
		t.Fatalf("AtomicList.new^ = %T", rp.DataStack.Peek())
	}
}

func TestAtomicListAppendLockedRoundTrip(t *testing.T) {
	ResetTestRuntime()
	l := NewAtomicAnyList()
	rp := NewTestRunProvider()
	mm := atomicAnyListType.MethodByName("append")
	f, ths := memberRunFrame(rp, l)
	rp.DataStack.Push(Int(7))
	mm.invokeCall(f, ths)
	if l.Len() != 1 || l.At(0) != Int(7) {
		t.Fatalf("after append: len=%d at0=%v", l.Len(), l.At(0))
	}
}

func TestListIteratorSnapshotIgnoresLaterAppend(t *testing.T) {
	l := NewAnyList()
	l.Append(Int(1))
	l.Append(Int(2))
	it := l.GetIter()
	l.Append(Int(3))
	var got []any
	for it.GetNext() {
		got = append(got, it.Current())
	}
	if len(got) != 2 || got[0] != Int(1) || got[1] != Int(2) {
		t.Fatalf("snapshot iter = %v, want [1 2]", got)
	}
}

func TestAtomicListConcurrentAppend(t *testing.T) {
	l := NewAtomicAnyList()
	const n = 200
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			l.Append(Int(i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			l.Append(String("x"))
		}
	}()
	wg.Wait()
	if l.Len() != 2*n {
		t.Fatalf("Len = %d, want %d", l.Len(), 2*n)
	}
}

func TestAtomicListGetIterSnapshotUnderLoad(t *testing.T) {
	l := NewAtomicStringList()
	l.Append("a")
	l.Append("b")
	it := l.GetIter()
	l.Append("c")
	var got []string
	for it.GetNext() {
		got = append(got, string(it.Current().(string)))
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("iter = %v, want [a b]", got)
	}
	if l.Len() != 3 {
		t.Fatalf("live Len = %d, want 3", l.Len())
	}
}

func TestAtomicListDeleteItemFromAny(t *testing.T) {
	type rec struct{ n int }
	a := &rec{1}
	b := &rec{2}
	l := NewAtomicAnyList()
	l.Append(a)
	l.Append(b)
	if !l.DeleteItemFromAny(a) {
		t.Fatal("want deleteItem a")
	}
	if l.Len() != 1 || l.At(0) != b {
		t.Fatalf("after deleteItem: len=%d at0=%v, want b", l.Len(), l.At(0))
	}
	if l.DeleteItemFromAny(a) {
		t.Fatal("second deleteItem a should be false")
	}
}

func TestAtomicListDeleteItemDuringSnapshotIter(t *testing.T) {
	type rec struct{ n int }
	a := &rec{1}
	b := &rec{2}
	l := NewAtomicAnyList()
	l.Append(a)
	l.Append(b)
	it := l.GetIter()
	for it.GetNext() {
		l.DeleteItemFromAny(it.Current())
	}
	if l.Len() != 0 {
		t.Fatalf("after snapshot deleteItem: Len=%d, want 0", l.Len())
	}
}

func TestAtomicListNativeMethodDeleteItemViaRunFrame(t *testing.T) {
	ResetTestRuntime()
	l := NewAtomicAnyList()
	l.Append(Int(7))
	l.Append(Int(8))
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(7))
	mm := atomicAnyListType.MethodByName("deleteItem")
	f, ths := memberRunFrame(rp, l)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || !bool(got) {
		t.Fatalf("deleteItem = %v, want true", got)
	}
	if l.Len() != 1 || l.At(0) != Int(8) {
		t.Fatalf("after deleteItem: len=%d at0=%v", l.Len(), l.At(0))
	}
}
