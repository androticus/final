package core

import "testing"

// Operands for matrix: false?true, 3?5, "a"?"b", 0?true (truthy).

func TestOpLessThan(t *testing.T) {
	comparePolyConversionMatrix(t, OpLessThan,
		Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true),
	)
}

func TestOpLessThanOrEqual(t *testing.T) {
	comparePolyConversionMatrix(t, OpLessThanOrEqual,
		Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true), Bool(true),
	)
}

func TestOpGreaterThan(t *testing.T) {
	comparePolyConversionMatrix(t, OpGreaterThan,
		Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false),
	)
}

func TestOpGreaterThanOrEqual(t *testing.T) {
	comparePolyConversionMatrix(t, OpGreaterThanOrEqual,
		Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false), Bool(false),
	)
}

func TestOpCompare(t *testing.T) {
	t.Run("int_lt", func(t *testing.T) { expectDyadic(t, Int(3), Int(5), OpCompare, Int(-1)) })
	t.Run("int_eq", func(t *testing.T) { expectDyadic(t, Int(5), Int(5), OpCompare, Int(0)) })
	t.Run("int_gt", func(t *testing.T) { expectDyadic(t, Int(5), Int(3), OpCompare, Int(1)) })
	t.Run("bool_order", func(t *testing.T) { expectDyadic(t, Bool(false), Bool(true), OpCompare, Int(-1)) })
	t.Run("string", func(t *testing.T) { expectDyadic(t, String("a"), String("b"), OpCompare, Int(-1)) })
}

func TestOpZeroLessThan(t *testing.T) {
	t.Run("negative_is_true", func(t *testing.T) { expectUnary(t, Int(-1), OpZeroLessThan, Bool(true)) })
	t.Run("zero_is_false", func(t *testing.T) { expectUnary(t, Int(0), OpZeroLessThan, Bool(false)) })
	t.Run("uint_never_less_than_zero", func(t *testing.T) { expectUnary(t, Uint(0), OpZeroLessThan, Bool(false)) })
}

func TestOpZeroLessThanOrEqual(t *testing.T) {
	t.Run("zero_is_true", func(t *testing.T) { expectUnary(t, Int(0), OpZeroLessThanOrEqual, Bool(true)) })
	t.Run("positive_is_false", func(t *testing.T) { expectUnary(t, Int(1), OpZeroLessThanOrEqual, Bool(false)) })
	t.Run("uint_zero_is_true", func(t *testing.T) { expectUnary(t, Uint(0), OpZeroLessThanOrEqual, Bool(true)) })
}

func TestOpZeroGreaterThan(t *testing.T) {
	t.Run("positive_is_true", func(t *testing.T) { expectUnary(t, Int(1), OpZeroGreaterThan, Bool(true)) })
	t.Run("zero_is_false", func(t *testing.T) { expectUnary(t, Int(0), OpZeroGreaterThan, Bool(false)) })
}

func TestOpZeroGreaterThanOrEqual(t *testing.T) {
	t.Run("zero_is_true", func(t *testing.T) { expectUnary(t, Int(0), OpZeroGreaterThanOrEqual, Bool(true)) })
	t.Run("negative_is_false", func(t *testing.T) { expectUnary(t, Int(-1), OpZeroGreaterThanOrEqual, Bool(false)) })
}
