package core

import (
	"fmt"
	"testing"
)

func collectTypedObjectPairs(o *TypedObject) []KeyValue[string, any] {
	out := make([]KeyValue[string, any], 0)
	it := o.GetIter()
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		out = append(out, KeyValue[string, any]{Key: string(kv.Key.(String)), Value: kv.Value})
	}
	return out
}

func TestTypedObjectIterator(t *testing.T) {
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("IterHost_%s", t.Name()), "alpha", "beta", "gamma")
	inst := MakeNewTypedObject(typ)
	inst.SetFieldByOrdinal(0, String("a"))
	inst.SetFieldByOrdinal(1, Int(2))
	inst.SetFieldByOrdinal(2, Bool(true))

	got := collectTypedObjectPairs(inst)
	want := []KeyValue[string, any]{
		{Key: "alpha", Value: String("a")},
		{Key: "beta", Value: Int(2)},
		{Key: "gamma", Value: Bool(true)},
	}
	if len(got) != len(want) {
		t.Fatalf("pairs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].Key != want[i].Key || got[i].Value != want[i].Value {
			t.Fatalf("pairs = %v, want %v", got, want)
		}
	}
}

func TestTypedObjectIteratorEmpty(t *testing.T) {
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("IterEmpty_%s", t.Name()))
	got := collectTypedObjectPairs(MakeNewTypedObject(typ))
	if len(got) != 0 {
		t.Fatalf("empty typed object iteration = %v, want none", got)
	}
}

func TestTypedObjectIteratorCurrentStable(t *testing.T) {
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("IterStable_%s", t.Name()), "alpha")
	inst := MakeNewTypedObject(typ)
	inst.SetFieldByOrdinal(0, String("x"))

	it := inst.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	first := it.Current()
	kv1 := first.(*AnyKeyValue)
	if kv1.Key.(String) != String("alpha") || kv1.Value != String("x") {
		t.Fatalf("Current() = %+v, want alpha/x", kv1)
	}
	second := it.Current()
	kv2 := second.(*AnyKeyValue)
	if kv2.Key != kv1.Key || kv2.Value != kv1.Value {
		t.Fatalf("Current() = %+v, want same key/value as %+v", kv2, kv1)
	}
	if it.GetNext() {
		t.Fatal("expected single-field object to exhaust")
	}
}

func TestTypedObjectGetIterFreshIterator(t *testing.T) {
	task := NewTestTask()
	typ := makeTypedObjectTypeForTest(t, task.dc, fmt.Sprintf("IterFresh_%s", t.Name()), "alpha", "beta")
	inst := MakeNewTypedObject(typ)
	inst.SetFieldByOrdinal(0, Int(1))
	inst.SetFieldByOrdinal(1, Int(2))

	got1 := collectTypedObjectPairs(inst)
	got2 := collectTypedObjectPairs(inst)
	if len(got1) != 2 || len(got2) != 2 {
		t.Fatalf("it1=%v it2=%v, want two pairs each", got1, got2)
	}
}

func TestTypedObjectIterableInterface(t *testing.T) {
	var _ Iterable = (*TypedObject)(nil)
	var _ Iterator = (*TypedObjectIterator)(nil)
}
