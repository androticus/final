package core

import "testing"

func TestEscFnLatchTakeClears(t *testing.T) {
	ClearEscFn()
	if _, ok := TakeEscFn(); ok {
		t.Fatal("empty latch should be false")
	}
	LatchEscFn(EscFnEsc)
	LatchEscFn(5) // last wins
	n, ok := TakeEscFn()
	if !ok || n != 5 {
		t.Fatalf("TakeEscFn = %d %v, want 5 true", n, ok)
	}
	if _, ok := TakeEscFn(); ok {
		t.Fatal("second Take should be empty")
	}
}

func TestEscFnLatchClear(t *testing.T) {
	ClearEscFn()
	LatchEscFn(EscFnF12)
	ClearEscFn()
	if _, ok := TakeEscFn(); ok {
		t.Fatal("ClearEscFn should empty latch")
	}
}

func TestEscFnLatchIgnoresOutOfRange(t *testing.T) {
	ClearEscFn()
	LatchEscFn(-1)
	LatchEscFn(13)
	if _, ok := TakeEscFn(); ok {
		t.Fatal("out-of-range LatchEscFn should not set latch")
	}
}
