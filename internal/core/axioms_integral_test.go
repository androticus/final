package core

import "testing"

func runIntegralOp(t *testing.T, rp *RunProvider, op OpCode, by, value any) {
	t.Helper()
	ResetTestRuntime()
	rp.DataStack.Push(value)
	rp.DataStack.Push(by)
	RunOps(NewTestTask(), rp, op, OpStop)
}

func runModRemOp(t *testing.T, rp *RunProvider, dividend, divisor any) (quot, rem any) {
	t.Helper()
	ResetTestRuntime()
	rp.DataStack.Push(dividend)
	rp.DataStack.Push(divisor)
	RunOps(NewTestTask(), rp, OpModRem, OpStop)
	rem = rp.DataStack.Pop()
	quot = rp.DataStack.Pop()
	return quot, rem
}

func shiftRotateByMatrix(t *testing.T, op OpCode, value any, wantWithBy1 any) {
	t.Helper()
	t.Run("by_int", func(t *testing.T) {
		rp := NewTestRunProvider()
		runIntegralOp(t, rp, op, Int(1), value)
		got := rp.DataStack.Pop()
		if got != wantWithBy1 {
			t.Fatalf("by Int: got %v (%T), want %v", got, got, wantWithBy1)
		}
	})
	t.Run("by_uint", func(t *testing.T) {
		rp := NewTestRunProvider()
		runIntegralOp(t, rp, op, Uint(1), value)
		got := rp.DataStack.Pop()
		if got != wantWithBy1 {
			t.Fatalf("by Uint: got %v (%T), want %v", got, got, wantWithBy1)
		}
	})
	t.Run("negative_by_panics", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.DataStack.Push(value)
		rp.DataStack.Push(Int(-1))
		runOpsExpectPanic(t, rp, "non-negative", op, OpStop)
	})
}

// --- OpShiftLeft ---

func TestOpShiftLeft(t *testing.T) {
	t.Run("int_value", func(t *testing.T) {
		shiftRotateByMatrix(t, OpShiftLeft, Int(0b1011), Int(0b10110))
	})
	t.Run("uint_value", func(t *testing.T) {
		shiftRotateByMatrix(t, OpShiftLeft, Uint(3), Uint(6))
	})
}

// --- OpShiftRight ---

func TestOpShiftRight(t *testing.T) {
	t.Run("int_sign_extended", func(t *testing.T) {
		// -4 >> 1 == -2 (arithmetic / sign-extending)
		shiftRotateByMatrix(t, OpShiftRight, Int(-4), Int(-2))
	})
	t.Run("uint_zero_extended", func(t *testing.T) {
		want := Uint(0x4000000000000000)
		shiftRotateByMatrix(t, OpShiftRight, Uint(0x8000000000000000), want)
	})
}

// --- OpRotateLeft ---

func TestOpRotateLeft(t *testing.T) {
	t.Run("uint_value", func(t *testing.T) {
		// 0x1234… <<1 rotate; by-4 case covered below for width mask
		want := Uint(0x2468ACF121579BDE)
		shiftRotateByMatrix(t, OpRotateLeft, Uint(0x1234567890ABCDEF), want)
	})
	t.Run("int_value", func(t *testing.T) {
		shiftRotateByMatrix(t, OpRotateLeft, Int(0b0001), Int(0b0010))
	})
	t.Run("uint_rotate_left_by_4", func(t *testing.T) {
		rp := NewTestRunProvider()
		runIntegralOp(t, rp, OpRotateLeft, Uint(4), Uint(0x1234567890ABCDEF))
		want := Uint(0x234567890ABCDEF1)
		got, ok := rp.DataStack.Pop().(Uint)
		if !ok || got != want {
			t.Fatalf("got %#x, want %#x", got, want)
		}
	})
	t.Run("count_masked_mod_width", func(t *testing.T) {
		rp := NewTestRunProvider()
		runIntegralOp(t, rp, OpRotateLeft, Uint(integralBitWidth), Uint(0xF))
		got, ok := rp.DataStack.Pop().(Uint)
		if !ok || got != 0xF {
			t.Fatalf("got %v, want identity rotate", got)
		}
	})
}

// --- OpRotateRight ---

func TestOpRotateRight(t *testing.T) {
	t.Run("int_value", func(t *testing.T) {
		shiftRotateByMatrix(t, OpRotateRight, Int(0b1010), Int(0b0101))
	})
	t.Run("uint_value", func(t *testing.T) {
		shiftRotateByMatrix(t, OpRotateRight, Uint(0b1010), Uint(0b0101))
	})
}

// --- OpRem ---

func TestOpRem(t *testing.T) {
	t.Run("truncated_remainder_int", func(t *testing.T) {
		expectDyadic(t, Int(-7), Int(3), OpRem, Int(-1))
	})
	t.Run("mixed_int_uint_panics", func(t *testing.T) {
		expectDyadicPanic(t, Uint(7), Int(3), OpRem, "mixed Int/Uint")
	})
	t.Run("zero_divisor_panics", func(t *testing.T) {
		expectDyadicPanic(t, Int(1), Int(0), OpRem, "divide by zero")
	})
}

// --- OpModRem ---

func TestOpModRem(t *testing.T) {
	t.Run("int_quot_and_rem", func(t *testing.T) {
		rp := NewTestRunProvider()
		quot, rem := runModRemOp(t, rp, Int(144), Int(7))
		if quot != Int(20) || rem != Int(4) {
			t.Fatalf("got quot=%v rem=%v, want Int(20) Int(4)", quot, rem)
		}
	})
	t.Run("int_negative_matches_div_and_percent", func(t *testing.T) {
		rp := NewTestRunProvider()
		quot, rem := runModRemOp(t, rp, Int(-7), Int(3))
		if quot != Int(-2) || rem != Int(-1) {
			t.Fatalf("got quot=%v rem=%v, want Int(-2) Int(-1)", quot, rem)
		}
	})
	t.Run("uint_quot_and_rem", func(t *testing.T) {
		rp := NewTestRunProvider()
		quot, rem := runModRemOp(t, rp, Uint(10), Uint(3))
		if quot != Uint(3) || rem != Uint(1) {
			t.Fatalf("got quot=%v rem=%v, want Uint(3) Uint(1)", quot, rem)
		}
	})
	t.Run("mixed_int_uint_panics", func(t *testing.T) {
		expectDyadicPanic(t, Int(-7), Uint(3), OpModRem, "mixed Int/Uint")
	})
}
