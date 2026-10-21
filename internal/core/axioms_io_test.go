package core

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestOpS(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	b := NewStringBuilder()
	rp.Stdout = b
	rp.DataStack.Push(Int(42))
	RunOps(NewTestTask(), rp, OpS, OpStop)
	if got := b.ToString(); got != "42" {
		t.Fatalf("$ output = %q, want %q", got, "42")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("data stack len = %d, want 0 after $ pops ds tos", rp.DataStack.Len())
	}
}

func TestOpSS(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	outB := NewStringBuilder()
	errB := NewStringBuilder()
	outW := outB
	errW := errB
	rp.Stdout = outW
	rp.Stderr = errW
	rp.DataStack.Push(String("diag"))
	RunOps(NewTestTask(), rp, OpSS, OpStop)
	if got := errB.ToString(); got != "diag" {
		t.Fatalf("$$ stderr = %q, want %q", got, "diag")
	}
	if got := outB.ToString(); got != "" {
		t.Fatalf("$$ must not write stdout, got %q", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("data stack len = %d, want 0 after $$ pops ds tos", rp.DataStack.Len())
	}
}

func TestOpRtoS(t *testing.T) {
	t.Run("rune_to_string", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Int('A'))
		RunOps(NewTestTask(), rp, OpRtoS, OpStop)
		if rp.DataStack.Len() != 1 {
			t.Fatalf("data stack len = %d, want 1", rp.DataStack.Len())
		}
		got, ok := rp.DataStack.Pop().(String)
		if !ok || got != "A" {
			t.Fatalf("r>s = %#v, want String(\"A\")", got)
		}
	})

	t.Run("compose_with_dollar", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		b := NewStringBuilder()
		rp.Stdout = b
		rp.DataStack.Push(Int('é'))
		RunOps(NewTestTask(), rp, OpRtoS, OpS, OpStop)
		if got := b.ToString(); got != "é" {
			t.Fatalf("r>s $ output = %q, want %q", got, "é")
		}
	})

	t.Run("non_int_panics", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(String("A"))
		runOpsExpectPanic(t, rp, "is string, not int", OpRtoS, OpStop)
	})
}

func TestDotS(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	b := NewStringBuilder()
	rp.Stdout = b
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(NewTestTask(), rp, OpDsS, OpStop)
	if got := b.ToString(); got != "1 2 " {
		t.Fatalf(".s output = %q, want %q", got, "1 2 ")
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("data stack len = %d, want 2 unchanged", rp.DataStack.Len())
	}
}

func TestDotS_wordOnStack(t *testing.T) {
	ResetTestRuntime()
	word, _, ok := FindName("find?")
	if !ok || word == nil {
		t.Fatal("FindName(find?)")
	}
	rp := NewTestRunProvider()
	b := NewStringBuilder()
	rp.Stdout = b
	rp.DataStack.Push(word)
	RunOps(NewTestTask(), rp, OpDsS, OpStop)
	if got := b.ToString(); got != "Word: find? " {
		t.Fatalf("ds$ output = %q, want %q", got, "Word: find? ")
	}
}

func TestStarS_allStacks(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	b := NewStringBuilder()
	rp.Stdout = b
	rp.DataStack.Push(Int(10))
	rp.DataStack.Push(Int(20))
	rp.TempStack.Push(Int(99))
	rp.XStack.Push(Int(7))
	rp.YStack.Push(Int(8))
	if !strings.Contains(FormatStackSnapshotReport(rp, task.dc, 0, OpDup), "1: 20") {
		t.Fatal("FormatStackSnapshotReport should list DS TOS-first")
	}
	RunOps(task, rp, OpStarS, OpStop)
	out := b.ToString()
	if !strings.Contains(out, "Return Stack (tos first, * = current):") {
		t.Fatalf("want *$ RS heading with * = current:\n%s", out)
	}
	if !strings.Contains(out, "Data Stack (tos first):") || !strings.Contains(out, "Temp Stack (tos first):") {
		t.Fatalf("want stack headings:\n%s", out)
	}
	if !strings.Contains(out, "X Stack (tos first):") || !strings.Contains(out, "Y Stack (tos first):") {
		t.Fatalf("want X/Y stack headings:\n%s", out)
	}
	if !strings.Contains(out, "1: 20") || !strings.Contains(out, "0: 10") || !strings.Contains(out, "0: 99") {
		t.Fatalf("want stack cells in *$ output:\n%s", out)
	}
	if !strings.Contains(out, "0: 7") || !strings.Contains(out, "0: 8") {
		t.Fatalf("want X/Y cells in *$ output:\n%s", out)
	}
	if rp.DataStack.Len() != 2 || rp.TempStack.Len() != 1 || rp.XStack.Len() != 1 || rp.YStack.Len() != 1 {
		t.Fatal("stacks should be unchanged after *$")
	}
}

func TestDotS_typeLiteralOnStack(t *testing.T) {
	ResetTestRuntime()
	tt := GetTypeRegistrar().Get("TokenType")
	if tt == nil {
		t.Fatal("missing TokenType")
	}
	rp := NewTestRunProvider()
	b := NewStringBuilder()
	rp.Stdout = b
	rp.DataStack.Push(tt)
	rp.DataStack.Push(Int(TokenEOF))
	RunOps(NewTestTask(), rp, OpDsS, OpStop)
	if got := b.ToString(); got != "Type: TokenType 0 " {
		t.Fatalf("ds$ output = %q, want %q", got, "Type: TokenType 0 ")
	}
}

func TestFormatNaturalString(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{String("hi"), "hi"},
		{String("a\nb"), "a\nb"},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{Int(-3), "-3"},
		{Uint(42), "42u"},
		{Float(1), "1.0"},
		{Float(math.Copysign(0, -1)), "-0.0"},
		{Float(1.5), "1.5"},
		{Float(1e20), "1e+20"},
		{Float(math.Inf(1)), "+Inf"},
		{Float(math.NaN()), "NaN"},
		{Interval{time.Hour + 30 * time.Minute}, "1h30m0s"},
		{time.Date(2026, 7, 4, 14, 30, 0, 250 * int(time.Millisecond), time.UTC), "2026-07-04T14:30:00.250"},
		{(*Word)(nil), "Word(nil)"},
	}
	for _, tc := range cases {
		if got := formatNaturalString(tc.v); got != tc.want {
			t.Fatalf("formatNaturalString(%#v) = %q, want %q", tc.v, got, tc.want)
		}
	}
	word, _, ok := FindName("find?")
	if !ok || word == nil {
		t.Fatal("FindName(find?)")
	}
	if got := formatNaturalString(word); got != "Word: find?" {
		t.Fatalf("formatNaturalString(find?) = %q, want Word: find?", got)
	}
	tt := GetTypeRegistrar().Get("TokenType")
	if tt == nil {
		t.Fatal("missing TokenType")
	}
	if got := formatNaturalString(tt); got != "Type: TokenType" {
		t.Fatalf("formatNaturalString(TokenType) = %q, want Type: TokenType", got)
	}
	if got := FormatErrorCell(tt); got != "Type: TokenType" {
		t.Fatalf("FormatErrorCell(TokenType) = %q, want Type: TokenType", got)
	}
	if got := FormatErrorCell(word); got != "Word: find?" {
		t.Fatalf("FormatErrorCell(find?) = %q, want Word: find?", got)
	}
	ai := NewAtomicInt()
	ai.Set(Int(-3))
	if got := formatNaturalString(ai); got != "-3" {
		t.Fatalf("formatNaturalString(AtomicInt(-3)) = %q, want -3", got)
	}
	if got := formatNaturalString((*AtomicInt)(nil)); got != "AtomicInt(nil)" {
		t.Fatalf("formatNaturalString(nil AtomicInt) = %q, want AtomicInt(nil)", got)
	}
}
