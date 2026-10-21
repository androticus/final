package core

import "testing"

func TestNewRunProviderAllocatesZAndOStacks(t *testing.T) {
	rp := NewTestRunProvider()
	if rp.ZStack == nil || rp.OStack == nil {
		t.Fatalf("ZStack/OStack = %v %v, want non-nil", rp.ZStack, rp.OStack)
	}
	rp.ZStack.Push(Int(1))
	rp.OStack.Push(Int(2))
	if rp.ZStack.Len() != 1 || rp.OStack.Len() != 1 {
		t.Fatalf("ZStack len = %d OStack len = %d, want 1 each", rp.ZStack.Len(), rp.OStack.Len())
	}
}

func TestNewRunProviderOStackMatchesReturnStackLimit(t *testing.T) {
	const returnMax = 4
	rp := NewRunProviderWithLimits(10, returnMax, 8)
	for i := 0; i < returnMax; i++ {
		rp.OStack.Push(Int(i))
	}
	if rp.OStack.Len() != returnMax {
		t.Fatalf("OStack len = %d, want %d", rp.OStack.Len(), returnMax)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected OStack overflow when exceeding returnMax")
		}
	}()
	rp.OStack.Push(Int(returnMax))
}
