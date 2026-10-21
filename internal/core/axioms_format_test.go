package core

import "testing"

func TestBootstrapFinalFmt(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	args := NewAnyList()
	args.Append(42)
	args.Append("hi")
	rp.DataStack.Push(args)
	rp.DataStack.Push("%d %s")
	f := &RunFrame{Task: NewTestTask(), RP: rp}
	bootstrapFinalFmt(f)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "42 hi" {
		t.Fatalf("fmt result = %q ok=%v, want %q", got, ok, "42 hi")
	}
}
