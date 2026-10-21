package core

import (
	"strings"
	"testing"
)

func TestNewStringMapEmpty(t *testing.T) {
	m := NewStringMap()
	if m == nil {
		t.Fatal("nil map")
	}
	if m.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", m.Len())
	}
	stringMapT := GetTypeRegistrar().Get("StringMap")
	if m.MapType != stringMapT {
		t.Fatalf("MapType = %v, want StringMap", m.MapType)
	}
	if m.FinalType() != stringMapT {
		t.Fatalf("FinalType() = %v, want StringMap", m.FinalType())
	}
}

func TestStringMapPutGetRoundTrip(t *testing.T) {
	m := NewStringMap()
	m.Put(String("alpha"), Int(1))
	m.Put(String("beta"), String("two"))
	if got := m.Get(String("alpha")); got != Int(1) {
		t.Fatalf("Get(alpha) = %v, want Int(1)", got)
	}
	if got := m.Get(String("beta")); got != String("two") {
		t.Fatalf("Get(beta) = %q, want two", got)
	}
	if m.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", m.Len())
	}
}

func TestStringMapNamedAccessViaStringKey(t *testing.T) {
	m := NewStringMap()
	m.Put("slot", Bool(true))
	if !m.Get("slot").(Bool) {
		t.Fatal("Get(slot) want true")
	}
}

func TestStringMapHasDeleteClear(t *testing.T) {
	m := NewStringMap()
	m.Put(String("k"), Int(9))
	if !m.Has(String("k")) {
		t.Fatal("Has(k) want true")
	}
	if m.Has(String("missing")) {
		t.Fatal("Has(missing) want false")
	}
	m.Delete(String("k"))
	if m.Has(String("k")) {
		t.Fatal("Has(k) after delete want false")
	}
	m.Put(String("a"), Int(1))
	m.Put(String("b"), Int(2))
	m.Clear()
	if m.Len() != 0 {
		t.Fatalf("Len() after clear = %d", m.Len())
	}
}

func TestStringMapGetMissingKeyPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
		 t.Fatal("expected panic")
		}
		if !strings.Contains(mapPanicMsg(r), "not found") {
			t.Fatalf("panic = %v", r)
		}
	}()
	NewStringMap().Get(String("nosuch"))
}

func TestStringMapRejectsNonStringKey(t *testing.T) {
	expectMapKeyPanic(t, func() { NewStringMap().Put(Int(1), Int(0)) })
	expectMapKeyPanic(t, func() { NewStringMap().Has(Int(1)) })
}

func TestNewAnyMapEmpty(t *testing.T) {
	m := NewAnyMap()
	mapT := GetTypeRegistrar().Get("Map")
	if m.FinalType() != mapT {
		t.Fatalf("FinalType() = %v, want Map", m.FinalType())
	}
}

func TestAnyMapAnyComparableKeys(t *testing.T) {
	m := NewAnyMap()
	m.Put(Int(42), String("answer"))
	m.Put(String("name"), String("Final"))
	if got := m.Get(Int(42)); got != String("answer") {
		t.Fatalf("Get(42) = %v", got)
	}
	if !m.Has(Int(42)) || !m.Has(String("name")) {
		t.Fatal("Has want true for stored keys")
	}
	m.Delete(Int(42))
	if m.Has(Int(42)) {
		t.Fatal("Has(42) after delete want false")
	}
}

func TestMapTypeRegistration(t *testing.T) {
	cases := []struct {
		name     string
		newMap   func() Typed
		typeName string
	}{
		{"Map", func() Typed { return NewAnyMap() }, "Map"},
		{"StringMap", func() Typed { return NewStringMap() }, "StringMap"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.newMap()
			wantT := GetTypeRegistrar().Get(tc.typeName)
			if wantT == nil {
				t.Fatalf("type %q not registered", tc.typeName)
			}
			if m.FinalType() != wantT {
				t.Fatalf("FinalType() = %v, want %q", m.FinalType(), tc.typeName)
			}
		})
	}
}

func expectMapKeyPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(mapPanicMsg(r), "expected string") {
			t.Fatalf("panic = %v, want expected string", r)
		}
	}()
	fn()
}

func mapPanicMsg(r any) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return ""
}

func TestInterpretMapTsBangBrace(t *testing.T) {
	ResetTestRuntime()
	task, err := newSessionTask(SessionKindLocal)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTextReaderFromString("Map.new^\n>ts\n\"is awesome\" \"brad\" t:!{}\n")
	if err != nil {
		t.Fatal(err)
	}
	rp := NewTestRunProvider()
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (store operands consumed)", rp.DataStack.Len())
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (map handle kept)", rp.TempStack.Len())
	}
	m, ok := rp.TempStack.Peek().(*AnyMap)
	if !ok {
		t.Fatalf("TS TOS = %T, want *AnyMap", rp.TempStack.Peek())
	}
	if got := m.Get(String("brad")); got != String("is awesome") {
		t.Fatalf("map[brad] = %v, want \"is awesome\"", got)
	}
}

func TestWalkCompoundTsBangBraceWithoutLeadingDot(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	m := NewAnyMap()
	rp.TempStack.Push(m)
	rp.DataStack.Push(String("is awesome"))
	rp.DataStack.Push(String("brad"))
	cp := CompoundPathWithSelectorForTest([]string{segBangBrace}, csTStack)
	if err := walkCompoundSegments(p, task, rp, cp); err != nil {
		t.Fatal(err)
	}
	if got := m.Get(String("brad")); got != String("is awesome") {
		t.Fatalf("map[brad] = %v, want \"is awesome\"", got)
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1", rp.TempStack.Len())
	}
}
