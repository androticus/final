package core

import "testing"

// Bit ops are self-contained polys (no typed sibling opcodes). Distinctive masks
// catch And/Or/Xor/Set/Clear/Toggle mixups.

func TestOpBitAnd(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		expectDyadic(t, Uint(0xF0), Uint(0xFF), OpBitAnd, Uint(0xF0))
	})
	t.Run("int", func(t *testing.T) {
		expectDyadic(t, Int(0b1100), Int(0b1010), OpBitAnd, Int(0b1000))
	})
	t.Run("mixed_integral_to_int", func(t *testing.T) {
		expectDyadic(t, Uint(5), Int(3), OpBitAnd, Int(1))
	})
	t.Run("invalid_type_panics", func(t *testing.T) {
		expectDyadicPanic(t, Int(1), Float(1), OpBitAnd, "Invalid type pair")
	})
}

func TestOpBitOr(t *testing.T) {
	expectDyadic(t, Uint(0xF0), Uint(0x0F), OpBitOr, Uint(0xFF))
	expectDyadic(t, Int(0b1000), Int(0b0010), OpBitOr, Int(0b1010))
}

func TestOpBitXor(t *testing.T) {
	expectDyadic(t, Uint(0xFF), Uint(0x0F), OpBitXor, Uint(0xF0))
	expectDyadic(t, Int(0b1100), Int(0b1010), OpBitXor, Int(0b0110))
}

func TestOpBitNot(t *testing.T) {
	expectUnary(t, Int(0), OpBitNot, Int(-1))
	expectUnary(t, Uint(0), OpBitNot, ^Uint(0))
	t.Run("invalid_type_panics", func(t *testing.T) {
		expectUnaryPanic(t, Float(1), OpBitNot, "Invalid type")
	})
}

func TestOpBitSet(t *testing.T) {
	expectDyadic(t, Uint(0xF0), Uint(0x0F), OpBitSet, Uint(0xFF))
}

func TestOpBitClear(t *testing.T) {
	expectDyadic(t, Uint(0xFF), Uint(0x0F), OpBitClear, Uint(0xF0))
}

func TestOpBitToggle(t *testing.T) {
	expectDyadic(t, Uint(0xF0), Uint(0xFF), OpBitToggle, Uint(0x0F))
}
