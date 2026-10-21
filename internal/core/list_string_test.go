package core

import (
	"strings"
	"testing"
)

func stringAt(t *testing.T, l *StringList, i int) string {
	t.Helper()
	return l.At(i)
}

// boxString is an any[string] operand as Final/DS would supply it.
func boxString(s string) any {
	return String(s)
}

func TestNewStringListEmpty(t *testing.T) {
	l := NewStringList()
	if l == nil {
		t.Fatal("nil list")
	}
	if l.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", l.Len())
	}
	stringListT := GetTypeRegistrar().Get("StringList")
	if l.ListType != stringListT {
		t.Fatalf("ListType = %v, want StringList", l.ListType)
	}
	if l.FinalType() != stringListT {
		t.Fatalf("FinalType() = %v, want StringList", l.FinalType())
	}
}

// Bare Append/Put/InsertAt are strongly typed (string); no runtime cell coercion.
func TestStringListTypedMutators(t *testing.T) {
	l := NewStringList()
	l.Append("a")
	l.Append("c")
	l.InsertAt(1, "b")
	if l.Len() != 3 || stringAt(t, l, 0) != "a" || stringAt(t, l, 1) != "b" || stringAt(t, l, 2) != "c" {
		t.Fatalf("after append/insert: %q %q %q", stringAt(t, l, 0), stringAt(t, l, 1), stringAt(t, l, 2))
	}
	l.Put(1, "B")
	if stringAt(t, l, 1) != "B" {
		t.Fatalf("Put = %q, want B", stringAt(t, l, 1))
	}
	l.DeleteAt(0)
	if l.Len() != 2 || stringAt(t, l, 0) != "B" || stringAt(t, l, 1) != "c" {
		t.Fatalf("after delete: len=%d [%q %q]", l.Len(), stringAt(t, l, 0), stringAt(t, l, 1))
	}
	l.Clear()
	if l.Len() != 0 {
		t.Fatalf("Len() after clear = %d", l.Len())
	}
}

func TestStringListInsertAtAppendViaLen(t *testing.T) {
	l := NewStringList()
	l.Append("first")
	l.InsertAt(l.Len(), "second")
	if l.Len() != 2 || stringAt(t, l, 1) != "second" {
		t.Fatalf("list = %v, want [first second]", []string{stringAt(t, l, 0), stringAt(t, l, 1)})
	}
}

func TestStringListAtAsAnyReturnsStringBox(t *testing.T) {
	l := NewStringList()
	l.Append("seg")
	if l.At(0) != "seg" {
		t.Fatalf("At(0) = %q, want seg", l.At(0))
	}
	if l.AtAsAny(0) != "seg" {
		t.Fatalf("AtAsAny(0) = %v, want seg", l.AtAsAny(0))
	}
}

// FromAny write paths accept only any[string]; empty string is valid.
func TestStringListFromAnyAcceptsAnyString(t *testing.T) {
	l := NewStringList()
	l.AppendFromAny(boxString("alpha"))
	l.AppendFromAny(boxString("beta"))
	l.PutFromAny(0, boxString("Alpha"))
	l.InsertAtFromAny(1, boxString("gamma"))
	l.AppendFromAny(boxString(""))
	if l.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", l.Len())
	}
	want := []string{"Alpha", "gamma", "beta", ""}
	for i, w := range want {
		if stringAt(t, l, i) != w {
			t.Fatalf("At(%d) = %q, want %q", i, stringAt(t, l, i), w)
		}
	}
}

// FromAny write paths reject non-string any cells (nil, Int, …).
func TestStringListFromAnyRejectsNonString(t *testing.T) {
	bad := []struct {
		name string
		val  any
	}{
		{"nil", nil},
		{"Int", Int(1)},
		{"Bool", Bool(true)},
		{"AnyList", NewAnyList()},
		{"Rune", Rune('x')},
	}
	for _, tc := range bad {
		t.Run("AppendFromAny_"+tc.name, func(t *testing.T) {
			expectStringListFromAnyPanic(t, func() { NewStringList().AppendFromAny(tc.val) })
		})
		t.Run("PutFromAny_"+tc.name, func(t *testing.T) {
			l := NewStringList()
			l.Append("slot")
			expectStringListFromAnyPanic(t, func() { l.PutFromAny(0, tc.val) })
		})
		t.Run("InsertAtFromAny_"+tc.name, func(t *testing.T) {
			l := NewStringList()
			l.Append("slot")
			expectStringListFromAnyPanic(t, func() { l.InsertAtFromAny(0, tc.val) })
		})
		t.Run("InsertAtFromAny_appendIndex_"+tc.name, func(t *testing.T) {
			l := NewStringList()
			l.Append("slot")
			expectStringListFromAnyPanic(t, func() { l.InsertAtFromAny(l.Len(), tc.val) })
		})
	}
}

func TestNewCompoundPathProducesPathString(t *testing.T) {
	cp := newCompoundPath([]string{".", "name"}, compoundPathNoSelector)
	if cp.path != String(". name") {
		t.Fatalf("path = %q, want %q", cp.path, ". name")
	}
}

func TestStringListOutOfRangePanics(t *testing.T) {
	l := NewStringList()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		} else if !strings.Contains(stringListPanicMsg(r), "out of range") {
			t.Fatalf("panic = %v", r)
		}
	}()
	l.At(0)
}

func expectStringListFromAnyPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(stringListPanicMsg(r), "expected string") {
			t.Fatalf("panic = %v, want expected string", r)
		}
	}()
	fn()
}

func stringListPanicMsg(r any) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return ""
}
