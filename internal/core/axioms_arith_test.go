package core

import (
	"math"
	"testing"
)

func TestOpAdd(t *testing.T) {
	arithPolyConversionMatrix(t, OpAdd,
		Int(23), Uint(23), Float(23),
		Float(23), Float(23), Float(23), Float(23),
	)
}

func TestOpSub(t *testing.T) {
	arithPolyConversionMatrix(t, OpSub,
		Int(17), Uint(17), Float(17),
		Float(17), Float(17), Float(17), Float(17),
	)
}

func TestOpMul(t *testing.T) {
	arithPolyConversionMatrix(t, OpMul,
		Int(60), Uint(60), Float(60),
		Float(60), Float(60), Float(60), Float(60),
	)
}

func TestOpDiv(t *testing.T) {
	arithPolyConversionMatrix(t, OpDiv,
		Int(6), Uint(6), Float(20.0/3.0),
		Float(20.0/3.0), Float(20.0/3.0), Float(20.0/3.0), Float(20.0/3.0),
	)
	t.Run("poly_integral_zero_divisor_panics", func(t *testing.T) {
		expectDyadicPanic(t, Int(9), Int(0), OpDiv, "divide by zero")
	})
	t.Run("float_zero_divisor_allowed", func(t *testing.T) {
		got := runDyadicOp(t, Float(1), Float(0), OpDiv)
		f, ok := got.(Float)
		if !ok || !math.IsInf(float64(f), 1) {
			t.Fatalf("got %v, want +Inf", got)
		}
	})
}

func TestOpAbs(t *testing.T) {
	t.Run("int_negative", func(t *testing.T) { expectUnary(t, Int(-5), OpAbs, Int(5)) })
	t.Run("int_nonneg", func(t *testing.T) { expectUnary(t, Int(5), OpAbs, Int(5)) })
	t.Run("uint", func(t *testing.T) { expectUnary(t, Uint(5), OpAbs, Uint(5)) })
	t.Run("float_negative", func(t *testing.T) { expectUnary(t, Float(-2.5), OpAbs, Float(2.5)) })
	t.Run("float_nonneg", func(t *testing.T) { expectUnary(t, Float(2.5), OpAbs, Float(2.5)) })
	t.Run("float_neg_zero", func(t *testing.T) {
		got := runUnaryOp(t, Float(math.Copysign(0, -1)), OpAbs)
		f, ok := got.(Float)
		if !ok || float64(f) != 0 || math.Signbit(float64(f)) {
			t.Fatalf("OpAbs(-0) = %v (signbit=%v), want +0", got, ok && math.Signbit(float64(f)))
		}
	})
	t.Run("float_nan", func(t *testing.T) {
		got := runUnaryOp(t, Float(math.NaN()), OpAbs)
		f, ok := got.(Float)
		if !ok || !math.IsNaN(float64(f)) {
			t.Fatalf("OpAbs(NaN) = %v (%T), want NaN Float", got, got)
		}
	})
	t.Run("non_numeric_panics", func(t *testing.T) {
		expectUnaryPanic(t, String("x"), OpAbs, "Invalid type")
	})
}

func TestOpNeg(t *testing.T) {
	t.Run("int", func(t *testing.T) { expectUnary(t, Int(5), OpNeg, Int(-5)) })
	t.Run("float", func(t *testing.T) { expectUnary(t, Float(2.5), OpNeg, Float(-2.5)) })
	t.Run("uint_panics", func(t *testing.T) { expectUnaryPanic(t, Uint(1), OpNeg, "neg not allowed on Uint") })
}
