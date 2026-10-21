package core

import "testing"

func TestOpBoolAnd(t *testing.T) {
	t.Run("bool", func(t *testing.T) { expectDyadic(t, Bool(true), Bool(false), OpBoolAnd, Bool(false)) })
	t.Run("int", func(t *testing.T) { expectDyadic(t, Int(0), Int(1), OpBoolAnd, Bool(false)) })
	t.Run("uint", func(t *testing.T) { expectDyadic(t, Uint(1), Uint(1), OpBoolAnd, Bool(true)) })
	t.Run("float", func(t *testing.T) { expectDyadic(t, Float(1), Float(0), OpBoolAnd, Bool(false)) })
	t.Run("string", func(t *testing.T) { expectDyadic(t, String("x"), String(""), OpBoolAnd, Bool(false)) })
	t.Run("non_truthy_panics", func(t *testing.T) {
		expectDyadicPanic(t, &AnyList{}, Int(1), OpBoolAnd, "Truthy")
	})
}

func TestOpBoolOr(t *testing.T) {
	t.Run("bool", func(t *testing.T) { expectDyadic(t, Bool(false), Bool(true), OpBoolOr, Bool(true)) })
	t.Run("int", func(t *testing.T) { expectDyadic(t, Int(0), Int(1), OpBoolOr, Bool(true)) })
	t.Run("uint", func(t *testing.T) { expectDyadic(t, Uint(0), Uint(0), OpBoolOr, Bool(false)) })
	t.Run("float", func(t *testing.T) { expectDyadic(t, Float(0), Float(2), OpBoolOr, Bool(true)) })
	t.Run("string", func(t *testing.T) { expectDyadic(t, String(""), String("y"), OpBoolOr, Bool(true)) })
	t.Run("non_truthy_panics", func(t *testing.T) {
		expectDyadicPanic(t, Int(1), &AnyList{}, OpBoolOr, "Truthy")
	})
}

func TestOpBoolNot(t *testing.T) {
	t.Run("bool", func(t *testing.T) { expectUnary(t, Bool(true), OpBoolNot, Bool(false)) })
	t.Run("int_zero", func(t *testing.T) { expectUnary(t, Int(0), OpBoolNot, Bool(true)) })
	t.Run("uint_nonzero", func(t *testing.T) { expectUnary(t, Uint(1), OpBoolNot, Bool(false)) })
	t.Run("float_zero", func(t *testing.T) { expectUnary(t, Float(0), OpBoolNot, Bool(true)) })
	t.Run("string_empty", func(t *testing.T) { expectUnary(t, String(""), OpBoolNot, Bool(true)) })
	t.Run("non_truthy_panics", func(t *testing.T) {
		expectUnaryPanic(t, &AnyList{}, OpBoolNot, "Truthy")
	})
}
