package core

import (
	"strings"
	"testing"
)

// Shared helpers for polymorphic axiom conversion-matrix tests.
// Keep matrices to Numeric / String / Truthy families — not an open type×type product.

func runOpsExpectPanic(t *testing.T, rp *RunProvider, wantSubstr string, ops ...OpCode) {
	t.Helper()
	task := NewTestTask()
	ct := &VirtCTResolver{}
	ip := appendOpsWithCSArming(ct, csDataStack, ops...)
	task.ip = ip
	done := make(chan error, 1)
	task.doneChan = done
	Run(task, rp)
	err := <-done
	if err == nil {
		t.Fatal("expected panic")
	}
	if wantSubstr != "" && !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("panic error = %q, want substring %q", err.Error(), wantSubstr)
	}
}

func runDyadicOp(t *testing.T, a, b any, op OpCode) any {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(a)
	rp.DataStack.Push(b)
	RunOps(NewTestTask(), rp, op, OpStop)
	return rp.DataStack.Pop()
}

func runUnaryOp(t *testing.T, a any, op OpCode) any {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(a)
	RunOps(NewTestTask(), rp, op, OpStop)
	return rp.DataStack.Pop()
}

func expectDyadic(t *testing.T, a, b any, op OpCode, want any) {
	t.Helper()
	got := runDyadicOp(t, a, b, op)
	if got != want {
		t.Fatalf("%s(%v,%v) = %v (%T), want %v (%T)", OpCodeName(op), a, b, got, got, want, want)
	}
}

func expectUnary(t *testing.T, a any, op OpCode, want any) {
	t.Helper()
	got := runUnaryOp(t, a, op)
	if got != want {
		t.Fatalf("%s(%v) = %v (%T), want %v (%T)", OpCodeName(op), a, got, got, want, want)
	}
}

func expectDyadicPanic(t *testing.T, a, b any, op OpCode, wantSubstr string) {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(a)
	rp.DataStack.Push(b)
	runOpsExpectPanic(t, rp, wantSubstr, op, OpStop)
}

func expectUnaryPanic(t *testing.T, a any, op OpCode, wantSubstr string) {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(a)
	runOpsExpectPanic(t, rp, wantSubstr, op, OpStop)
}

// arithPolyConversionMatrix exercises Numeric promotion families for one poly op.
// wants are hard-coded per op (not recomputed via Go +/-).
func arithPolyConversionMatrix(t *testing.T, op OpCode, wantII, wantUU, wantFF, wantIF, wantFI, wantUF, wantFU any) {
	t.Helper()
	t.Run("ii", func(t *testing.T) { expectDyadic(t, Int(20), Int(3), op, wantII) })
	t.Run("uu", func(t *testing.T) { expectDyadic(t, Uint(20), Uint(3), op, wantUU) })
	t.Run("ff", func(t *testing.T) { expectDyadic(t, Float(20), Float(3), op, wantFF) })
	t.Run("if_promote", func(t *testing.T) { expectDyadic(t, Int(20), Float(3), op, wantIF) })
	t.Run("fi_promote", func(t *testing.T) { expectDyadic(t, Float(20), Int(3), op, wantFI) })
	t.Run("uf_promote", func(t *testing.T) { expectDyadic(t, Uint(20), Float(3), op, wantUF) })
	t.Run("fu_promote", func(t *testing.T) { expectDyadic(t, Float(20), Uint(3), op, wantFU) })
	t.Run("iu_mixed_panics", func(t *testing.T) { expectDyadicPanic(t, Int(20), Uint(3), op, "mixed Int/Uint") })
	t.Run("ui_mixed_panics", func(t *testing.T) { expectDyadicPanic(t, Uint(20), Int(3), op, "mixed Int/Uint") })
	t.Run("non_numeric_panics", func(t *testing.T) {
		expectDyadicPanic(t, String("a"), String("b"), op, "Invalid type pair")
	})
}

// comparePolyConversionMatrix exercises equate/order family promote for one poly op.
func comparePolyConversionMatrix(t *testing.T, op OpCode, wantBB, wantII, wantUU, wantFF, wantSS, wantIF, wantTruthy any) {
	t.Helper()
	t.Run("bb", func(t *testing.T) { expectDyadic(t, Bool(false), Bool(true), op, wantBB) })
	t.Run("ii", func(t *testing.T) { expectDyadic(t, Int(3), Int(5), op, wantII) })
	t.Run("uu", func(t *testing.T) { expectDyadic(t, Uint(3), Uint(5), op, wantUU) })
	t.Run("ff", func(t *testing.T) { expectDyadic(t, Float(3), Float(5), op, wantFF) })
	t.Run("ss", func(t *testing.T) { expectDyadic(t, String("a"), String("b"), op, wantSS) })
	t.Run("if_promote", func(t *testing.T) { expectDyadic(t, Int(3), Float(5), op, wantIF) })
	t.Run("truthy_bool_int", func(t *testing.T) { expectDyadic(t, Int(0), Bool(true), op, wantTruthy) })
	t.Run("iu_mixed_panics", func(t *testing.T) { expectDyadicPanic(t, Int(3), Uint(5), op, "mixed Int/Uint") })
	t.Run("string_int_panics", func(t *testing.T) { expectDyadicPanic(t, String("3"), Int(3), op, "compare op") })
	t.Run("non_truthy_panics", func(t *testing.T) {
		expectDyadicPanic(t, &AnyList{}, Int(1), op, "compare op")
	})
}

func expectPanicSubstr(t *testing.T, wantSubstr string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg := panicMsg(r)
		if wantSubstr != "" && !strings.Contains(msg, wantSubstr) {
			t.Fatalf("panic = %v, want substring %q", r, wantSubstr)
		}
	}()
	fn()
}
