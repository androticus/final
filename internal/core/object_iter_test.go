package core

import "testing"

func collectDynamicObjectPairs(o *DynamicObject) []KeyValue[string, any] {
	out := make([]KeyValue[string, any], 0)
	it := o.GetIter()
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		out = append(out, KeyValue[string, any]{Key: string(kv.Key.(String)), Value: kv.Value})
	}
	return out
}

func TestDynamicObjectIterator(t *testing.T) {
	o := NewDynamicObject()
	o.Set("alpha", String("a"))
	o.Set("beta", Int(2))
	o.Set("gamma", Bool(true))

	got := collectDynamicObjectPairs(o)
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

func TestDynamicObjectIteratorEmpty(t *testing.T) {
	got := collectDynamicObjectPairs(NewDynamicObject())
	if len(got) != 0 {
		t.Fatalf("empty object iteration = %v, want none", got)
	}
}

func TestDynamicObjectIteratorCurrentStable(t *testing.T) {
	o := NewDynamicObject()
	o.Set("only", String("x"))
	it := o.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	first := it.Current()
	kv1 := first.(*AnyKeyValue)
	if kv1.Key.(String) != String("only") || kv1.Value != String("x") {
		t.Fatalf("Current() = %+v, want only/x", kv1)
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

func TestDynamicObjectGetIterFreshIterator(t *testing.T) {
	o := NewDynamicObject()
	o.Set("a", Int(1))
	o.Set("b", Int(2))
	got1 := collectDynamicObjectPairs(o)
	got2 := collectDynamicObjectPairs(o)
	if len(got1) != 2 || len(got2) != 2 {
		t.Fatalf("it1=%v it2=%v, want two pairs each", got1, got2)
	}
}

func TestDynamicObjectIteratorSnapshotIgnoresLaterSet(t *testing.T) {
	o := NewDynamicObject()
	o.Set("a", Int(1))
	o.Set("b", Int(2))
	it := o.GetIter()
	o.Set("c", Int(3))
	got := collectDynamicObjectPairsFromIter(it)
	if len(got) != 2 || got[0].Key != "a" || got[1].Key != "b" {
		t.Fatalf("snapshot = %v, want a,b only", got)
	}
}

func TestDynamicObjectIteratorOrderAfterSpill(t *testing.T) {
	n := sfmMaxIteratedFields + 6
	o := NewDynamicObject()
	fillDynamicObject(o, n)
	got := collectDynamicObjectPairs(o)
	if len(got) != n {
		t.Fatalf("pairs len = %d, want %d", len(got), n)
	}
	for i, kv := range got {
		if kv.Key != sfmTestKey(i) || kv.Value != Int(i) {
			t.Fatalf("pairs[%d] = {%q, %v}, want {%q, %v}", i, kv.Key, kv.Value, sfmTestKey(i), Int(i))
		}
	}
}

func collectDynamicObjectPairsFromIter(it Iterator) []KeyValue[string, any] {
	out := make([]KeyValue[string, any], 0)
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		out = append(out, KeyValue[string, any]{Key: string(kv.Key.(String)), Value: kv.Value})
	}
	return out
}
