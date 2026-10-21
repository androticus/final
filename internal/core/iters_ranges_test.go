package core

import (
	"fmt"
	"strings"
	"testing"
)

func intsFromIter(it Iterator) []int {
	out := make([]int, 0)
	for it.GetNext() {
		out = append(out, int(it.Current().(Int)))
	}
	return out
}

func anysFromIter(it Iterator) []any {
	out := make([]any, 0)
	for it.GetNext() {
		out = append(out, it.Current())
	}
	return out
}

func newRangeFromDS(beg, end int) *Range {
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(beg))
	rp.DataStack.Push(Int(end))
	return NewRange(rp)
}

func newRangeByFromDS(beg, end, by Any) *RangeBy {
	rp := NewTestRunProvider()
	rp.DataStack.Push(beg)
	rp.DataStack.Push(end)
	rp.DataStack.Push(by)
	return NewRangeBy(rp)
}

func TestRangeNewFromStack(t *testing.T) {
	r := newRangeFromDS(0, 4)
	if r.beg != 0 || r.end != 4 {
		t.Fatalf("Range = {%d,%d}, want {0,4}", r.beg, r.end)
	}
}

func TestRangeByNewFromStack(t *testing.T) {
	rb := newRangeByFromDS(Int(0), Int(8), Int(2))
	if rb.beg != Int(0) || rb.end != Int(8) || rb.by != Int(2) {
		t.Fatalf("RangeBy = {%v,%v,%v}, want {0,8,2}", rb.beg, rb.end, rb.by)
	}
}

func TestRangeIteratorInclusiveAscending(t *testing.T) {
	got := intsFromIter(newRangeFromDS(0, 4).GetIter())
	want := []int{0, 1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeIteratorSingleton(t *testing.T) {
	got := intsFromIter(newRangeFromDS(0, 0).GetIter())
	want := []int{0}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("values = %v, want %v", got, want)
	}
}

func TestRangeIteratorEmptyInclusive(t *testing.T) {
	got := intsFromIter(newRangeFromDS(0, -1).GetIter())
	if len(got) != 0 {
		t.Fatalf("values = %v, want empty", got)
	}
}

func TestRangeByIteratorStep(t *testing.T) {
	got := intsFromIter(newRangeByFromDS(Int(0), Int(8), Int(2)).GetIter())
	want := []int{0, 2, 4, 6, 8}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeByIteratorDescendingStep(t *testing.T) {
	got := intsFromIter(newRangeByFromDS(Int(10), Int(0), Int(-2)).GetIter())
	want := []int{10, 8, 6, 4, 2, 0}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeByIteratorInexactEnd(t *testing.T) {
	got := intsFromIter(newRangeByFromDS(Int(0), Int(9), Int(2)).GetIter())
	want := []int{0, 2, 4, 6, 8}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeByIteratorFloat(t *testing.T) {
	got := anysFromIter(newRangeByFromDS(Float(0), Float(1), Float(0.25)).GetIter())
	want := []any{Float(0), Float(0.25), Float(0.5), Float(0.75), Float(1)}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeByIteratorUint(t *testing.T) {
	got := anysFromIter(newRangeByFromDS(Uint(0), Uint(6), Uint(2)).GetIter())
	want := []any{Uint(0), Uint(2), Uint(4), Uint(6)}
	if len(got) != len(want) {
		t.Fatalf("values = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("values = %v, want %v", got, want)
		}
	}
}

func TestRangeIteratorCurrentStableUntilNext(t *testing.T) {
	it := newRangeFromDS(1, 3).GetIter()
	if !it.GetNext() {
		t.Fatal("expected first GetNext true")
	}
	first := it.Current()
	if int(first.(Int)) != 1 {
		t.Fatalf("Current() = %v, want 1", first)
	}
	if it.Current() != first {
		t.Fatal("Current() changed before next GetNext")
	}
	if !it.GetNext() {
		t.Fatal("expected second GetNext true")
	}
	if int(it.Current().(Int)) != 2 {
		t.Fatalf("Current() = %v, want 2", it.Current())
	}
}

func TestRangeGetIterFreshIterator(t *testing.T) {
	r := newRangeFromDS(1, 2)
	got1 := intsFromIter(r.GetIter())
	got2 := intsFromIter(r.GetIter())
	want := []int{1, 2}
	if len(got1) != len(want) || len(got2) != len(want) {
		t.Fatalf("it1=%v it2=%v, want %v each", got1, got2, want)
	}
	for i := range want {
		if got1[i] != want[i] || got2[i] != want[i] {
			t.Fatalf("it1=%v it2=%v, want %v each", got1, got2, want)
		}
	}
}

func TestRangeTypeRegistration(t *testing.T) {
	rangeT := GetTypeRegistrar().Get("Range")
	rangeByT := GetTypeRegistrar().Get("RangeBy")
	if rangeT == nil || rangeByT == nil {
		t.Fatal("Range and RangeBy must be registered")
	}
	if newRangeFromDS(0, 1).FinalType() != rangeT {
		t.Fatal("Range.FinalType mismatch")
	}
	if newRangeByFromDS(Int(0), Int(1), Int(1)).FinalType() != rangeByT {
		t.Fatal("RangeBy.FinalType mismatch")
	}
}

func TestRangeIterableInterface(t *testing.T) {
	var _ Iterable = (*Range)(nil)
	var _ Iterable = (*RangeBy)(nil)
	var _ Iterator = (*RangeIterator)(nil)
	var _ Iterator = (*RangeByIterator)(nil)
}

func rangeExpectPanic(t *testing.T, wantSubstr string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if wantSubstr != "" {
			msg := ""
			switch v := r.(type) {
			case error:
				msg = v.Error()
			case string:
				msg = v
			default:
				msg = fmt.Sprint(v)
			}
			if !strings.Contains(msg, wantSubstr) {
				t.Fatalf("panic = %v, want substring %q", r, wantSubstr)
			}
		}
	}()
	fn()
}

func TestAssertRangeValidRejectsInvalid(t *testing.T) {
	cases := []struct {
		name       string
		beg        int
		end        int
		wantSubstr string
	}{
		{name: "forward overshoot", beg: 6, end: 4, wantSubstr: "range is invalid"},
		{name: "forward wrong order", beg: 3, end: 0, wantSubstr: "range is invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rangeExpectPanic(t, tc.wantSubstr, func() {
				assertRangeValid(tc.beg, tc.end)
			})
		})
	}
}

func TestAssertRangeValidAcceptsEmpty(t *testing.T) {
	assertRangeValid(0, -1)
	got := intsFromIter(newRangeFromDS(0, -1).GetIter())
	if len(got) != 0 {
		t.Fatalf("forward empty values = %v, want empty", got)
	}
}

func TestAssertRangeByValidRejectsInvalid(t *testing.T) {
	cases := []struct {
		name       string
		beg        Any
		end        Any
		by         Any
		wantSubstr string
	}{
		{name: "by zero", beg: Int(0), end: Int(4), by: Int(0), wantSubstr: "by must not be 0"},
		{name: "forward wrong sign", beg: Int(3), end: Int(0), by: Int(1), wantSubstr: "range is invalid"},
		{name: "forward indices negative by", beg: Int(0), end: Int(4), by: Int(-1), wantSubstr: "range is invalid"},
		{name: "reverse overshoot", beg: Int(1), end: Int(3), by: Int(-1), wantSubstr: "range is invalid"},
		{name: "reverse indices positive by", beg: Int(10), end: Int(0), by: Int(2), wantSubstr: "range is invalid"},
		{name: "mixed int float", beg: Int(0), end: Float(4), by: Int(1), wantSubstr: "same numeric type"},
		{name: "non numeric", beg: String("a"), end: Int(1), by: Int(1), wantSubstr: "numeric"},
		{name: "float by zero", beg: Float(0), end: Float(1), by: Float(0), wantSubstr: "by must not be 0"},
		{name: "uint wrong direction", beg: Uint(5), end: Uint(1), by: Uint(1), wantSubstr: "range is invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rangeExpectPanic(t, tc.wantSubstr, func() {
				assertRangeByValid(tc.beg, tc.end, tc.by)
			})
		})
	}
}

func TestAssertRangeByValidAcceptsEmptyEdges(t *testing.T) {
	got := intsFromIter(newRangeByFromDS(Int(0), Int(1), Int(-1)).GetIter())
	if len(got) != 0 {
		t.Fatalf("reverse empty values = %v, want empty", got)
	}
	got = intsFromIter(newRangeByFromDS(Int(4), Int(3), Int(1)).GetIter())
	if len(got) != 0 {
		t.Fatalf("forward empty one-step values = %v, want empty", got)
	}
}
