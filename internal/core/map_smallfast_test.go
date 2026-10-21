package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestSFMAddPreservesOrder(t *testing.T) {
	m := NewSFM[int](4)
	m.Add("alpha", 1)
	m.Add("beta", 2)
	m.Add("gamma", 3)
	if m.Len() != 3 {
		t.Fatalf("Len = %d, want 3", m.Len())
	}
	want := []struct {
		name  string
		value int
	}{
		{"alpha", 1},
		{"beta", 2},
		{"gamma", 3},
	}
	for i, w := range want {
		item := m.ItemAt(i)
		if item.Name != w.name || item.Value != w.value {
			t.Fatalf("ItemAt(%d) = {%q, %d}, want {%q, %d}", i, item.Name, item.Value, w.name, w.value)
		}
	}
	if m.Get("beta") != 2 {
		t.Fatalf("Get(beta) = %d, want 2", m.Get("beta"))
	}
	if m.IndexOf("alpha") != 0 || m.IndexOf("gamma") != 2 {
		t.Fatalf("IndexOf = %d,%d want 0,2", m.IndexOf("alpha"), m.IndexOf("gamma"))
	}
	if !m.HasKey("alpha") || m.HasKey("missing") {
		t.Fatal("HasKey mismatch")
	}
}

func TestSFMIndexOfMissingPanics(t *testing.T) {
	m := NewSFM[int](1)
	m.Add("x", 1)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on missing IndexOf")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "not found") {
			t.Fatalf("panic = %q, want not found", msg)
		}
	}()
	_ = m.IndexOf("missing")
}

func TestSFMAddDuplicatePanics(t *testing.T) {
	m := NewSFM[int](2)
	m.Add("x", 1)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate Add")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "duplicate") {
			t.Fatalf("panic = %q, want duplicate", msg)
		}
	}()
	m.Add("x", 2)
}

func TestSFMAddEmptyNamePanics(t *testing.T) {
	m := NewSFM[int](1)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on empty name")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "empty name") {
			t.Fatalf("panic = %q, want empty name", msg)
		}
	}()
	m.Add("", 1)
}

func TestSFMNilSafeLenAndHasKey(t *testing.T) {
	var m *SFM[int]
	if m.Len() != 0 {
		t.Fatalf("nil Len = %d, want 0", m.Len())
	}
	if m.HasKey("any") {
		t.Fatal("nil HasKey should be false")
	}
}

func TestSFMSetUpserts(t *testing.T) {
	m := NewSFM[int](2)
	m.Set("x", 1)
	m.Set("x", 2)
	if m.Get("x") != 2 {
		t.Fatalf("Get(x) = %d, want 2", m.Get("x"))
	}
}

func TestSFMIterator(t *testing.T) {
	m := NewSFM[any](4)
	m.Add("a", Int(1))
	m.Add("b", Int(2))
	it := m.GetIter()
	var keys []string
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		keys = append(keys, string(kv.Key.(String)))
	}
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("keys = %v, want [a b]", keys)
	}
}

func TestSFMNilGetIterReturnsNullIter(t *testing.T) {
	var m *PropertyMetadataMap
	it := m.GetIter()
	if _, ok := it.(*NullIter); !ok {
		t.Fatalf("GetIter(nil) = %T, want *NullIter", it)
	}
	if it.GetNext() {
		t.Fatal("NullIter.GetNext should be false")
	}
}

func TestMetadataMapTypesNotRegistered(t *testing.T) {
	for _, name := range []string{"SmallFastMap", "PropertyMetadataMap", "MethodMetadataMap"} {
		if GetTypeRegistrar().Get(name) != nil {
			t.Fatalf("%s must not be a Final catalog type", name)
		}
	}
}

func sfmTestKey(i int) string {
	return fmt.Sprintf("k%d", i)
}

func fillSFM(m *SFM[int], n int) {
	for i := 0; i < n; i++ {
		m.Add(sfmTestKey(i), i)
	}
}

func TestSFMNoSpillAtMaxIterated(t *testing.T) {
	m := NewSFM[int](sfmMaxIteratedFields)
	fillSFM(m, sfmMaxIteratedFields)
	if m.namesMap != nil {
		t.Fatal("namesMap should stay nil at the scan limit")
	}
	if m.Len() != sfmMaxIteratedFields {
		t.Fatalf("Len = %d, want %d", m.Len(), sfmMaxIteratedFields)
	}
	last := sfmMaxIteratedFields - 1
	if m.IndexOf(sfmTestKey(0)) != 0 || m.IndexOf(sfmTestKey(last)) != last {
		t.Fatal("IndexOf mismatch at scan limit")
	}
	if m.Get(sfmTestKey(last)) != last {
		t.Fatalf("Get last = %d, want %d", m.Get(sfmTestKey(last)), last)
	}
	if !m.HasKey(sfmTestKey(0)) || m.HasKey("missing") {
		t.Fatal("HasKey mismatch at scan limit")
	}
}

func TestSFMSpilloverNameLookup(t *testing.T) {
	m := NewSFM[int](sfmMaxIteratedFields + 2)
	fillSFM(m, sfmMaxIteratedFields)
	m.Set(sfmTestKey(0), 100)
	if m.namesMap != nil {
		t.Fatal("existing-key Set must not spill")
	}
	m.Add(sfmTestKey(sfmMaxIteratedFields), sfmMaxIteratedFields)
	if m.namesMap == nil {
		t.Fatal("namesMap should be filled after crossing the scan limit")
	}
	if len(m.namesMap) != sfmMaxIteratedFields+1 {
		t.Fatalf("namesMap len = %d, want %d", len(m.namesMap), sfmMaxIteratedFields+1)
	}
	n := sfmMaxIteratedFields
	if m.IndexOf(sfmTestKey(0)) != 0 || m.IndexOf(sfmTestKey(n)) != n {
		t.Fatalf("IndexOf after spill = %d,%d want 0,%d", m.IndexOf(sfmTestKey(0)), m.IndexOf(sfmTestKey(n)), n)
	}
	if m.Get(sfmTestKey(0)) != 100 {
		t.Fatalf("Get k0 after spill = %d, want 100", m.Get(sfmTestKey(0)))
	}
	m.Set(sfmTestKey(0), 200)
	if m.Get(sfmTestKey(0)) != 200 || m.Len() != n+1 {
		t.Fatal("existing-key Set after spill should update in place")
	}
	if m.Get(sfmTestKey(n)) != n {
		t.Fatalf("Get spilled key = %d, want %d", m.Get(sfmTestKey(n)), n)
	}
	if !m.HasKey(sfmTestKey(n)) || m.HasKey("missing") {
		t.Fatal("HasKey mismatch after spill")
	}
	m.Set(sfmTestKey(n+1), n+1)
	if m.Get(sfmTestKey(n+1)) != n+1 || m.IndexOf(sfmTestKey(n+1)) != n+1 {
		t.Fatal("Set of new key after spill failed")
	}
	if len(m.namesMap) != n+2 {
		t.Fatalf("namesMap len after append = %d, want %d", len(m.namesMap), n+2)
	}
}

func TestSFMAddDuplicateAfterSpillPanics(t *testing.T) {
	m := NewSFM[int](sfmMaxIteratedFields + 1)
	fillSFM(m, sfmMaxIteratedFields+1)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate Add after spill")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "duplicate") {
			t.Fatalf("panic = %q, want duplicate", msg)
		}
	}()
	m.Add(sfmTestKey(0), 99)
}

func TestSFMLargeCapDoesNotPreSpill(t *testing.T) {
	m := NewSFM[int](128)
	fillSFM(m, 10)
	if m.namesMap != nil {
		t.Fatal("namesMap must stay nil when len is under the scan limit")
	}
}

func TestSFMIteratorOrderAfterSpill(t *testing.T) {
	n := sfmMaxIteratedFields + 6
	m := NewSFM[any](n)
	for i := 0; i < n; i++ {
		m.Add(sfmTestKey(i), Int(i))
	}
	it := m.GetIter()
	var keys []string
	for it.GetNext() {
		kv := it.Current().(*AnyKeyValue)
		keys = append(keys, string(kv.Key.(String)))
	}
	if len(keys) != n {
		t.Fatalf("iter len = %d, want %d", len(keys), n)
	}
	for i, k := range keys {
		if k != sfmTestKey(i) {
			t.Fatalf("iter[%d] = %q, want %q", i, k, sfmTestKey(i))
		}
	}
}
