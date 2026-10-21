package core

import "testing"

func TestOpEqual(t *testing.T) {
	comparePolyConversionMatrix(t, OpEqual,
		Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false),
	)
}

func TestOpNotEqual(t *testing.T) {
	comparePolyConversionMatrix(t, OpNotEqual,
		Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true),
	)
}

func TestOpZeroEqual(t *testing.T) {
	t.Run("zero_is_true", func(t *testing.T) { expectUnary(t, Int(0), OpZeroEqual, Bool(true)) })
	t.Run("nonzero_is_false", func(t *testing.T) { expectUnary(t, Int(1), OpZeroEqual, Bool(false)) })
	t.Run("uint_zero_is_true", func(t *testing.T) { expectUnary(t, Uint(0), OpZeroEqual, Bool(true)) })
	t.Run("uint_nonzero_is_false", func(t *testing.T) { expectUnary(t, Uint(1), OpZeroEqual, Bool(false)) })
	t.Run("float_zero_is_true", func(t *testing.T) { expectUnary(t, Float(0), OpZeroEqual, Bool(true)) })
}

func TestOpZeroNotEqual(t *testing.T) {
	t.Run("nonzero_is_true", func(t *testing.T) { expectUnary(t, Int(1), OpZeroNotEqual, Bool(true)) })
	t.Run("zero_is_false", func(t *testing.T) { expectUnary(t, Int(0), OpZeroNotEqual, Bool(false)) })
}
