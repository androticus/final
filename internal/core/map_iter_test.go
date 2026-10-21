package core

import (
	"strings"
	"testing"
)

func collectStringMapPairs(m *StringMap) []KeyValue[string, any] {
	out := make([]KeyValue[string, any], 0)
	it := m.GetIter()
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		out = append(out, KeyValue[string, any]{Key: string(kv.Key.(String)), Value: kv.Value})
	}
	return out
}

func collectAnyMapPairs(m *AnyMap) []KeyValue[any, any] {
	out := make([]KeyValue[any, any], 0)
	it := m.GetIter()
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		out = append(out, *kv)
	}
	return out
}

func stringMapPairsMatch(m *StringMap, pairs []KeyValue[string, any]) bool {
	if len(pairs) != m.Len() {
		return false
	}
	seen := make(map[string]any, len(pairs))
	for _, kv := range pairs {
		if seen[kv.Key] != nil {
			return false
		}
		want := m.Get(String(kv.Key))
		if want != kv.Value {
			return false
		}
		seen[kv.Key] = kv.Value
	}
	return len(seen) == m.Len()
}

func anyMapPairsMatch(m *AnyMap, pairs []KeyValue[any, any]) bool {
	if len(pairs) != m.Len() {
		return false
	}
	seen := make(map[any]bool, len(pairs))
	for _, kv := range pairs {
		if seen[kv.Key] {
			return false
		}
		want := m.Get(kv.Key)
		if want != kv.Value {
			return false
		}
		seen[kv.Key] = true
	}
	return len(seen) == m.Len()
}

func TestStringMapIterator(t *testing.T) {
	m := NewStringMap()
	m.Put(String("alpha"), Int(1))
	m.Put(String("beta"), String("two"))
	m.Put(String("gamma"), Bool(true))

	got := collectStringMapPairs(m)
	if !stringMapPairsMatch(m, got) {
		t.Fatalf("iteration pairs %v do not match map contents", got)
	}
}

func TestStringMapIteratorEmpty(t *testing.T) {
	got := collectStringMapPairs(NewStringMap())
	if len(got) != 0 {
		t.Fatalf("empty map iteration = %v, want none", got)
	}
}

func TestStringMapIteratorCurrentStable(t *testing.T) {
	m := NewStringMap()
	m.Put(String("only"), Int(42))
	it := m.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	first := it.Current()
	kv1 := first.(*AnyKeyValue)
	if kv1.Key != String("only") || kv1.Value != Int(42) {
		t.Fatalf("Current() = %+v, want only/42", kv1)
	}
	second := it.Current()
	kv2 := second.(*AnyKeyValue)
	if kv2.Key != kv1.Key || kv2.Value != kv1.Value {
		t.Fatalf("Current() = %+v, want same key/value as %+v", kv2, kv1)
	}
	if it.GetNext() {
		t.Fatal("expected single-entry map to exhaust")
	}
}

func TestStringMapGetIterFreshIterator(t *testing.T) {
	m := NewStringMap()
	m.Put(String("a"), Int(1))
	m.Put(String("b"), Int(2))
	got1 := collectStringMapPairs(m)
	got2 := collectStringMapPairs(m)
	if !stringMapPairsMatch(m, got1) || !stringMapPairsMatch(m, got2) {
		t.Fatalf("it1=%v it2=%v, want full map each", got1, got2)
	}
}

func TestAnyMapIterator(t *testing.T) {
	m := NewAnyMap()
	m.Put(Int(10), String("ten"))
	m.Put(String("name"), String("Final"))
	m.Put(Bool(true), Int(0))

	got := collectAnyMapPairs(m)
	if !anyMapPairsMatch(m, got) {
		t.Fatalf("iteration pairs %v do not match map contents", got)
	}
}

func TestAnyMapIteratorEmpty(t *testing.T) {
	got := collectAnyMapPairs(NewAnyMap())
	if len(got) != 0 {
		t.Fatalf("empty map iteration = %v, want none", got)
	}
}

func TestMapIteratorCurrentPanicsBeforeFirstGetNext(t *testing.T) {
	it := NewStringMap().GetIter()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(mapIterPanicMsg(r), "undefined iterator state") {
			t.Fatalf("panic = %v", r)
		}
	}()
	it.Current()
}

func TestMapIteratorCurrentPanicsAfterExhaustion(t *testing.T) {
	m := NewStringMap()
	m.Put(String("k"), Int(1))
	it := m.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	for it.GetNext() {
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(mapIterPanicMsg(r), "undefined iterator state") {
			t.Fatalf("panic = %v", r)
		}
	}()
	it.Current()
}

func TestMapIterableInterface(t *testing.T) {
	var _ Iterable = (*StringMap)(nil)
	var _ Iterable = (*AnyMap)(nil)
	var _ Iterator = (*MapIterator[string, any])(nil)
	var _ Iterator = (*MapIterator[any, any])(nil)
}

func TestKeyValueTypeRegistration(t *testing.T) {
	ResetTestRuntime()
	kvT := GetTypeRegistrar().Get("KeyValue")
	if kvT == nil {
		t.Fatal("KeyValue type not registered")
	}
	keyPM := kvT.PropertyByName("key")
	if keyPM == nil || keyPM.set != nil {
		t.Fatal("KeyValue.key should be readonly")
	}
	valPM := kvT.PropertyByName("value")
	if valPM == nil || valPM.set != nil {
		t.Fatal("KeyValue.value should be readonly")
	}
	if kvT.HasClassMethod("new") {
		t.Fatal("KeyValue should not have class method new")
	}
}

func TestKeyValueConstPropertiesFromStringMapIter(t *testing.T) {
	ResetTestRuntime()
	m := NewStringMap()
	m.Put(String("brad"), String("is awesome"))
	it := m.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	kv := it.Current().(*AnyKeyValue)
	if kv.FinalType() != GetTypeRegistrar().Get("KeyValue") {
		t.Fatalf("FinalType() = %v, want KeyValue", kv.FinalType())
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, kv)
	keyPM := GetTypeRegistrar().Get("KeyValue").PropertyByName("key")
	keyPM.invokeGet(f, ths)
	gotKey, ok := rp.DataStack.Pop().(String)
	if !ok || gotKey != "brad" {
		t.Fatalf("key getter = %v, want brad", gotKey)
	}
	valPM := GetTypeRegistrar().Get("KeyValue").PropertyByName("value")
	valPM.invokeGet(f, ths)
	gotVal, ok := rp.DataStack.Pop().(String)
	if !ok || gotVal != "is awesome" {
		t.Fatalf("value getter = %v, want is awesome", gotVal)
	}
}

func mapIterPanicMsg(r any) string {
	if s, ok := r.(string); ok {
		return s
	}
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return ""
}
