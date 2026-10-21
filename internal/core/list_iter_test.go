package core

import "testing"

func collectStringListIter(l *StringList) []string {
	out := make([]string, 0)
	it := l.GetIter()
	for it.GetNext() {
		out = append(out, it.Current().(String))
	}
	return out
}

func collectInt32ListIter(l *Int32List) []int {
	out := make([]int, 0)
	it := l.GetIter()
	for it.GetNext() {
		out = append(out, int(it.Current().(Int)))
	}
	return out
}

func TestStringListIterator(t *testing.T) {
	l := NewStringList()
	l.Append("alpha")
	l.Append("beta")
	l.Append("gamma")

	got := collectStringListIter(l)
	want := []string{"alpha", "beta", "gamma"}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestStringListIteratorEmpty(t *testing.T) {
	got := collectStringListIter(NewStringList())
	if len(got) != 0 {
		t.Fatalf("empty list iteration = %v, want none", got)
	}
}

func TestStringListIteratorCurrentStable(t *testing.T) {
	l := NewStringList()
	l.Append("only")
	it := l.GetIter()
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	first := it.Current()
	if first != String("only") {
		t.Fatalf("Current() = %v, want only", first)
	}
	if it.Current() != first {
		t.Fatal("Current() changed before next GetNext")
	}
	if it.GetNext() {
		t.Fatal("expected single-element list to exhaust")
	}
}

func TestStringListGetIterFreshIterator(t *testing.T) {
	l := NewStringList()
	l.Append("a")
	l.Append("b")
	got1 := collectStringListIter(l)
	got2 := collectStringListIter(l)
	if len(got1) != 2 || len(got2) != 2 {
		t.Fatalf("it1=%v it2=%v, want [a b] each", got1, got2)
	}
}

func TestInt32ListIterator(t *testing.T) {
	l := NewInt32List()
	l.AppendFromAny(boxInt(10))
	l.AppendFromAny(boxInt(-3))
	l.AppendFromAny(boxUint(200))

	got := collectInt32ListIter(l)
	want := []int{
		goWidenInt32(goNarrow[int32](boxInt(10))),
		goWidenInt32(goNarrow[int32](boxInt(-3))),
		goWidenInt32(goNarrow[int32](boxUint(200))),
	}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestInt32ListIteratorEmpty(t *testing.T) {
	got := collectInt32ListIter(NewInt32List())
	if len(got) != 0 {
		t.Fatalf("empty list iteration = %v, want none", got)
	}
}

func TestListIterableInterface(t *testing.T) {
	var _ Iterable = (*StringList)(nil)
	var _ Iterable = (*Int32List)(nil)
	var _ Iterator = (*ListIterator[string])(nil)
	var _ Iterator = (*ListIterator[int32])(nil)
}
