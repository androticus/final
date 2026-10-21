package core

import (
	"math"
	"testing"
)

func helperDS(vals ...any) *Stack[any] {
	ds := NewStack[any]("helper", DefaultDSStackSize)
	for _, v := range vals {
		ds.Push(v)
	}
	return ds
}

func expectHelperDyadic(t *testing.T, a, b any, op OpCode, fn func(*Stack[any], OpCode), want any) {
	t.Helper()
	ds := helperDS(a, b)
	fn(ds, op)
	got := ds.Pop()
	if got != want {
		t.Fatalf("%s(%v,%v) = %v (%T), want %v (%T)", OpCodeName(op), a, b, got, got, want, want)
	}
}

func expectHelperUnary(t *testing.T, a any, op OpCode, fn func(*Stack[any], OpCode), want any) {
	t.Helper()
	ds := helperDS(a)
	fn(ds, op)
	got := ds.Pop()
	if got != want {
		t.Fatalf("%s(%v) = %v (%T), want %v (%T)", OpCodeName(op), a, got, got, want, want)
	}
}

func expectHelperDyadicPanic(t *testing.T, a, b any, op OpCode, fn func(*Stack[any], OpCode), wantSubstr string) {
	t.Helper()
	expectPanicSubstr(t, wantSubstr, func() {
		fn(helperDS(a, b), op)
	})
}

func expectHelperUnaryPanic(t *testing.T, a any, op OpCode, fn func(*Stack[any], OpCode), wantSubstr string) {
	t.Helper()
	expectPanicSubstr(t, wantSubstr, func() {
		fn(helperDS(a), op)
	})
}

// --- arith ---

func TestDoDyadicArithOp(t *testing.T) {
	t.Run("add_ii", func(t *testing.T) { expectHelperDyadic(t, Int(20), Int(22), OpAdd, doDyadicArithOp, Int(42)) })
	t.Run("add_uu", func(t *testing.T) { expectHelperDyadic(t, Uint(20), Uint(22), OpAdd, doDyadicArithOp, Uint(42)) })
	t.Run("add_ff", func(t *testing.T) { expectHelperDyadic(t, Float(20), Float(22), OpAdd, doDyadicArithOp, Float(42)) })
	t.Run("add_if", func(t *testing.T) { expectHelperDyadic(t, Int(20), Float(3), OpAdd, doDyadicArithOp, Float(23)) })
	t.Run("sub_ii", func(t *testing.T) { expectHelperDyadic(t, Int(20), Int(3), OpSub, doDyadicArithOp, Int(17)) })
	t.Run("mul_uu", func(t *testing.T) { expectHelperDyadic(t, Uint(6), Uint(7), OpMul, doDyadicArithOp, Uint(42)) })
	t.Run("div_ii", func(t *testing.T) { expectHelperDyadic(t, Int(20), Int(4), OpDiv, doDyadicArithOp, Int(5)) })
	t.Run("div_zero_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, Int(1), Int(0), OpDiv, doDyadicArithOp, "divide by zero")
	})
	t.Run("div_f_inf", func(t *testing.T) {
		ds := helperDS(Float(1), Float(0))
		doDyadicArithOp(ds, OpDiv)
		got := ds.Pop().(Float)
		if !math.IsInf(float64(got), 1) {
			t.Fatalf("got %v, want +Inf", got)
		}
	})
	t.Run("iu_mixed_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, Int(20), Uint(3), OpAdd, doDyadicArithOp, "mixed Int/Uint")
	})
	t.Run("ui_mixed_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, Uint(20), Int(3), OpSub, doDyadicArithOp, "mixed Int/Uint")
	})
	t.Run("non_numeric_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, String("a"), String("b"), OpMul, doDyadicArithOp, "Invalid type pair")
	})
}

func TestDoMonadicNumericOp(t *testing.T) {
	t.Run("abs_int_neg", func(t *testing.T) { expectHelperUnary(t, Int(-5), OpAbs, doMonadicNumericOp, Int(5)) })
	t.Run("abs_uint_nop", func(t *testing.T) { expectHelperUnary(t, Uint(5), OpAbs, doMonadicNumericOp, Uint(5)) })
	t.Run("abs_float_neg_zero", func(t *testing.T) {
		ds := helperDS(Float(math.Copysign(0, -1)))
		doMonadicNumericOp(ds, OpAbs)
		f := ds.Pop().(Float)
		if float64(f) != 0 || math.Signbit(float64(f)) {
			t.Fatalf("Abs(-0) = %v signbit=%v", f, math.Signbit(float64(f)))
		}
	})
	t.Run("neg_int", func(t *testing.T) { expectHelperUnary(t, Int(5), OpNeg, doMonadicNumericOp, Int(-5)) })
	t.Run("neg_float", func(t *testing.T) { expectHelperUnary(t, Float(2.5), OpNeg, doMonadicNumericOp, Float(-2.5)) })
	t.Run("neg_uint_panics", func(t *testing.T) {
		expectHelperUnaryPanic(t, Uint(1), OpNeg, doMonadicNumericOp, "neg not allowed on Uint")
	})
	t.Run("abs_non_numeric_panics", func(t *testing.T) {
		expectHelperUnaryPanic(t, String("x"), OpAbs, doMonadicNumericOp, "Invalid type")
	})
}

// --- boolean ---

func TestDoBooleanOp(t *testing.T) {
	t.Run("and", func(t *testing.T) { expectHelperDyadic(t, Int(0), Int(1), OpBoolAnd, doDyadicBooleanOp, Bool(false)) })
	t.Run("or", func(t *testing.T) { expectHelperDyadic(t, Bool(false), Bool(true), OpBoolOr, doDyadicBooleanOp, Bool(true)) })
	t.Run("not", func(t *testing.T) { expectHelperUnary(t, String(""), OpBoolNot, doMonadicBooleanOp, Bool(true)) })
	t.Run("non_truthy_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, &AnyList{}, Int(1), OpBoolAnd, doDyadicBooleanOp, "Truthy")
	})
}

// --- equate / order / compare ---

func TestDoDyadicEquateOp(t *testing.T) {
	t.Run("ii_eq", func(t *testing.T) { expectHelperDyadic(t, Int(2), Int(2), OpEqual, doDyadicEquateOp, Bool(true)) })
	t.Run("ss_ne", func(t *testing.T) { expectHelperDyadic(t, String("x"), String("y"), OpNotEqual, doDyadicEquateOp, Bool(true)) })
	t.Run("truthy", func(t *testing.T) { expectHelperDyadic(t, Int(0), Bool(true), OpEqual, doDyadicEquateOp, Bool(false)) })
	t.Run("nan_eq_false", func(t *testing.T) {
		expectHelperDyadic(t, Float(math.NaN()), Float(1), OpEqual, doDyadicEquateOp, Bool(false))
	})
	t.Run("nan_ne_true", func(t *testing.T) {
		expectHelperDyadic(t, Float(math.NaN()), Float(math.NaN()), OpNotEqual, doDyadicEquateOp, Bool(true))
	})
	t.Run("iu_mixed_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, Int(3), Uint(5), OpEqual, doDyadicEquateOp, "mixed Int/Uint")
	})
	t.Run("string_int_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, String("3"), Int(3), OpEqual, doDyadicEquateOp, "compare op")
	})
}

func TestDoDyadicCompareOp(t *testing.T) {
	t.Run("lt_ii", func(t *testing.T) { expectHelperDyadic(t, Int(3), Int(5), OpLessThan, doDyadicCompareOp, Bool(true)) })
	t.Run("lt_bool", func(t *testing.T) { expectHelperDyadic(t, Bool(false), Bool(true), OpLessThan, doDyadicCompareOp, Bool(true)) })
	t.Run("nan_lt_false", func(t *testing.T) {
		expectHelperDyadic(t, Float(math.NaN()), Float(1), OpLessThan, doDyadicCompareOp, Bool(false))
	})
	t.Run("nan_ge_false", func(t *testing.T) {
		expectHelperDyadic(t, Float(1), Float(math.NaN()), OpGreaterThanOrEqual, doDyadicCompareOp, Bool(false))
	})
	t.Run("iu_mixed_panics", func(t *testing.T) {
		expectHelperDyadicPanic(t, Int(3), Uint(5), OpLessThan, doDyadicCompareOp, "mixed Int/Uint")
	})
}

func TestDoCompareOp(t *testing.T) {
	cmp := func(a, b any) Int {
		ds := helperDS(a, b)
		doCompareOp(ds)
		return ds.Pop().(Int)
	}
	t.Run("int_lt", func(t *testing.T) {
		if got := cmp(Int(3), Int(5)); got != -1 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("int_eq", func(t *testing.T) {
		if got := cmp(Int(5), Int(5)); got != 0 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("int_gt", func(t *testing.T) {
		if got := cmp(Int(5), Int(3)); got != 1 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("bool_order", func(t *testing.T) {
		if got := cmp(Bool(false), Bool(true)); got != -1 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("nan_vs_nan", func(t *testing.T) {
		if got := cmp(Float(math.NaN()), Float(math.NaN())); got != 0 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("nan_always_less", func(t *testing.T) {
		if got := cmp(Float(math.NaN()), Float(1)); got != -1 {
			t.Fatalf("got %d", got)
		}
		if got := cmp(Float(1), Float(math.NaN())); got != 1 {
			t.Fatalf("got %d", got)
		}
	})
	t.Run("string", func(t *testing.T) {
		if got := cmp(String("a"), String("b")); got != -1 {
			t.Fatalf("got %d", got)
		}
	})
}

func TestDoMonadicCompare0Op(t *testing.T) {
	t.Run("eq0_int", func(t *testing.T) { expectHelperUnary(t, Int(0), OpZeroEqual, doMonadicCompare0Op, Bool(true)) })
	t.Run("eq0_uint", func(t *testing.T) { expectHelperUnary(t, Uint(0), OpZeroEqual, doMonadicCompare0Op, Bool(true)) })
	t.Run("lt0_uint", func(t *testing.T) { expectHelperUnary(t, Uint(0), OpZeroLessThan, doMonadicCompare0Op, Bool(false)) })
	t.Run("gt0_int", func(t *testing.T) { expectHelperUnary(t, Int(1), OpZeroGreaterThan, doMonadicCompare0Op, Bool(true)) })
}

func TestDoIntegralBitOp(t *testing.T) {
	t.Run("and_uu", func(t *testing.T) {
		expectHelperDyadic(t, Uint(0b1100), Uint(0b1010), OpBitAnd, doIntegralBitDyadic, Uint(0b1000))
	})
	t.Run("not_i", func(t *testing.T) {
		expectHelperUnary(t, Int(0), OpBitNot, doIntegralBitMonadic, Int(^0))
	})
}
