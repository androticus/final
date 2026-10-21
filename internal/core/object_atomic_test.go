package core

import (
	"sync"
	"testing"
)

func TestAtomicDynamicObjectFinalTypeAndParent(t *testing.T) {
	ResetTestRuntime()
	o := NewAtomicDynamicObject()
	if o.FinalType() != atomicDynamicObjectType {
		t.Fatalf("FinalType = %v, want AtomicDynamicObject", o.FinalType())
	}
	if atomicDynamicObjectType.GetParent() != dynamicObjectType {
		t.Fatalf("parent = %v, want DynamicObject", atomicDynamicObjectType.GetParent())
	}
}

func TestAtomicDynamicObjectClassNew(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	mm := atomicDynamicObjectType.ClassMethodByName("new")
	f, ths := memberRunFrame(rp, atomicDynamicObjectType)
	mm.invokeCall(f, ths)
	got, ok := rp.DataStack.Pop().(*AtomicDynamicObject)
	if !ok || got == nil {
		t.Fatalf("new^ = %T", rp.DataStack.Peek())
	}
}

func TestAtomicDynamicObjectSetGet(t *testing.T) {
	o := NewAtomicDynamicObject()
	o.Set("x", Int(42))
	if o.Get("x") != Int(42) {
		t.Fatalf("Get = %v, want 42", o.Get("x"))
	}
	if !o.HasField("x") {
		t.Fatal("HasField x false")
	}
	if o.Len() != 1 {
		t.Fatalf("Len = %d, want 1", o.Len())
	}
}

func TestAtomicDynamicObjectConcurrentSet(t *testing.T) {
	o := NewAtomicDynamicObject()
	const n = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			o.Set("a", Int(i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			o.Set("b", String("x"))
		}
	}()
	wg.Wait()
	if o.Len() != 2 {
		t.Fatalf("Len = %d, want 2", o.Len())
	}
	_ = o.Get("a")
	_ = o.Get("b")
}

func TestAtomicDynamicObjectGetIterSnapshot(t *testing.T) {
	o := NewAtomicDynamicObject()
	o.Set("a", Int(1))
	o.Set("b", Int(2))
	it := o.GetIter()
	o.Set("c", Int(3))
	var keys []string
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		keys = append(keys, string(kv.Key.(String)))
	}
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("keys = %v, want [a b]", keys)
	}
	if o.Len() != 3 {
		t.Fatalf("live Len = %d, want 3", o.Len())
	}
}

func TestAtomicDynamicObjectItemsGetSet(t *testing.T) {
	ResetTestRuntime()
	o := NewAtomicDynamicObject()
	im := atomicDynamicObjectType.GetItemsMetadata()
	if im == nil {
		t.Fatal("missing items metadata")
	}
	im.SetByKey(o, String("k"), Int(9))
	if got := im.GetByKey(o, String("k")); got != Int(9) {
		t.Fatalf("GetByKey = %v, want 9", got)
	}
	if !im.HasByKey(o, String("k")) {
		t.Fatal("HasByKey false")
	}
}
