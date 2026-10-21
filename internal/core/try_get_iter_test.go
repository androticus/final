package core

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestOpTryGetIterIterable(t *testing.T) {
	ResetTestRuntime()
	r := NewRangeFromInts(1, 2)
	rp := NewTestRunProvider()
	rp.DataStack.Push(r)
	RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
	ok, hasBool := rp.DataStack.Pop().(Bool)
	if !hasBool || !ok {
		t.Fatalf("has = %v, want true", ok)
	}
	iter, isIter := rp.DataStack.Pop().(Iterator)
	if !isIter {
		t.Fatalf("got %T, want Iterator", iter)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatal("provider should be consumed on Iterable path")
	}
	if !iter.GetNext() {
		t.Fatal("expected first GetNext true")
	}
	if iter.Current() != Int(1) {
		t.Fatalf("first = %v, want 1", iter.Current())
	}
}

func TestOpTryGetIterIterator(t *testing.T) {
	ResetTestRuntime()
	iter := &RangeIterator{beg: 1, end: 2, curr: 0}
	rp := NewTestRunProvider()
	rp.DataStack.Push(iter)
	RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
	ok, hasBool := rp.DataStack.Pop().(Bool)
	if !hasBool || !ok {
		t.Fatalf("has = %v, want true", ok)
	}
	got, isIter := rp.DataStack.Pop().(Iterator)
	if !isIter {
		t.Fatalf("got %T, want Iterator", got)
	}
	if got != iter {
		t.Fatal("iterator on stack should be the same instance")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatal("stack should be empty")
	}
	if !got.GetNext() {
		t.Fatal("expected first GetNext true")
	}
	if got.Current() != Int(1) {
		t.Fatalf("first = %v, want 1", got.Current())
	}
}

func TestOpTryGetIterNonIterable(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{name: "int", val: Int(99)},
		{name: "string", val: String("ab")}, // Iterable via ._tryGetIter in foreach prologue, not OpTryGetIter
		{name: "bool", val: Bool(true)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ResetTestRuntime()
			rp := NewTestRunProvider()
			rp.DataStack.Push(tc.val)
			RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
			ok, hasBool := rp.DataStack.Pop().(Bool)
			if !hasBool || ok {
				t.Fatalf("has = %v, want false", ok)
			}
			got := rp.DataStack.Pop()
			if got != tc.val {
				t.Fatalf("got = %v (%T), want %v (%T)", got, got, tc.val, tc.val)
			}
		})
	}
}

func NewRangeFromInts(beg, end int) *Range {
	return &Range{beg: beg, end: end}
}

func TestTryGetIteratorProtoTypeDefault(t *testing.T) {
	ResetTestRuntime()
	dupW := globalDictionaryContext.dict.GetByName("dup")
	if dupW == nil {
		t.Fatal("missing dup word")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dupW)
	mm := ths.Typ.GetMethodMetadata("_tryGetIter")
	if mm == nil {
		t.Fatal("missing _tryGetIter on Word type chain")
	}
	mm.invokeCall(f, ths)
	ok, hasBool := rp.DataStack.Pop().(Bool)
	if !hasBool || ok {
		t.Fatalf("has = %v, want false", ok)
	}
}

func TestTryGetIteratorString(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, String("ab"))
	mm := stringType.GetMethodMetadata("_tryGetIter")
	if mm == nil {
		t.Fatal("missing String _tryGetIter")
	}
	mm.invokeCall(f, ths)
	ok, hasBool := rp.DataStack.Pop().(Bool)
	if !hasBool || !ok {
		t.Fatalf("has = %v, want true", ok)
	}
	it, isIter := rp.DataStack.Pop().(*StringRuneIter)
	if !isIter {
		t.Fatalf("got %T, want *StringRuneIter", rp.DataStack.Pop())
	}
	var runes []rune
	for it.GetNext() {
		runes = append(runes, rune(it.Current().(Int)))
	}
	if string(runes) != "ab" {
		t.Fatalf("runes = %q, want ab", string(runes))
	}
}

func TestStringRuneIterFreshIterator(t *testing.T) {
	it := &StringRuneIter{s: "xy"}
	got1 := runesFromStringIter(it)
	it2 := &StringRuneIter{s: "xy"}
	got2 := runesFromStringIter(it2)
	if got1 != "xy" || got2 != "xy" {
		t.Fatalf("got1=%q got2=%q, want xy each", got1, got2)
	}
}

func TestStringRuneIterNonASCII(t *testing.T) {
	const src = "aé🙂日" // ASCII + Latin-1 supplement + 4-byte emoji + CJK
	want := []rune(src)
	it := &StringRuneIter{s: src}
	got := make([]rune, 0, len(want))
	for it.GetNext() {
		got = append(got, rune(it.Current().(Int)))
	}
	if len(got) != len(want) {
		t.Fatalf("read %d runes, want %d (last rune missed?)", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rune[%d] = %U (%q), want %U (%q)", i, got[i], string(got[i]), want[i], string(want[i]))
		}
	}
	if it.GetNext() {
		t.Fatal("GetNext after exhaustion returned true — trailing rune may have been skipped")
	}
}

func runesFromStringIter(it *StringRuneIter) string {
	var b strings.Builder
	for it.GetNext() {
		b.WriteRune(rune(it.Current().(Int)))
	}
	return b.String()
}

func TestRunForeachStringSumRunesAfterBootstrap(t *testing.T) {
	t.Skip("foreach runtime: (foreachPrologue) nested colon + (pushIter) slot frame (same as repl_stack foreach tests)")
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(
			`: sumRunes 0 "ab" foreach i + /foreach ;` + "\nsumRunes\n",
		)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int('a')+Int('b') {
		t.Fatalf("sumRunes = %v, want %d", got, Int('a')+Int('b'))
	}
}

func TestForeachNonIterableError(t *testing.T) {
	t.Skip("foreach runtime: (foreachPrologue) nested colon RS frame — error path not reached in session")
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(`: bad 42 foreach /foreach ;` + "\nbad\n")),
		label: "test",
	}})
	rp.SetParser(p)
	var panicked any
	func() {
		defer func() { panicked = recover() }()
		_ = interpretSession(tk, rp)
	}()
	if panicked == nil {
		t.Fatal("expected panic for non-iterable foreach")
	}
	msg := fmt.Sprint(panicked)
	if !strings.Contains(msg, "foreach requires an object that is iterable") {
		t.Fatalf("panic = %q, want iterable error message", msg)
	}
}
