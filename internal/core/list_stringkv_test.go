package core

import "testing"

func TestStringKeyValueList_PutGetHas(t *testing.T) {
	l := NewStringKeyValueList()
	l.Put(String("format"), String("3"))
	if !l.Has(String("format")) {
		t.Fatal("Has(format) = false, want true")
	}
	if got := l.Get(String("format")); got != "3" {
		t.Fatalf("Get(format) = %q, want 3", got)
	}
	l.Put(String("format"), String("4"))
	if l.Len() != 1 {
		t.Fatalf("Len after upsert = %d, want 1", l.Len())
	}
	if got := l.Get(String("format")); got != "4" {
		t.Fatalf("Get(format) after upsert = %q, want 4", got)
	}
}

func TestStringKeyValueList_GetMissingPanics(t *testing.T) {
	l := NewStringKeyValueList()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Get missing key did not panic")
		}
	}()
	l.Get(String("missing"))
}

func TestStringKeyValueList_AddAppends(t *testing.T) {
	l := NewStringKeyValueList()
	l.Add("a", "1")
	l.Add("a", "2")
	if l.Len() != 2 {
		t.Fatalf("Len after two Add = %d, want 2", l.Len())
	}
	if got := l.Get(String("a")); got != "1" {
		t.Fatalf("Get(a) first-wins = %q, want 1", got)
	}
}
