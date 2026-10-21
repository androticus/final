package core

import (
	"math/bits"
	"testing"
)

func TestMat2i4ItemsAndRange(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2i4(2, 2)
	if m.FinalType() != mat2i4Type {
		t.Fatalf("FinalType = %v, want Mat2i4", m.FinalType())
	}
	if mat2i4Type.GetParent() != matProtoType {
		t.Fatalf("parent = %v, want ProtoMat", mat2i4Type.GetParent())
	}
	im := mat2i4Type.GetItemsMetadata()
	im.SetByMultiIndex(m, Int(15), Int(1), Int(0))
	if m.At(1, 0) != 15 {
		t.Fatalf("At(1,0)=%v, want 15", m.At(1, 0))
	}
	got := im.GetByMultiIndex(m, Int(1), Int(0)).(Int)
	if got != 15 {
		t.Fatalf("@[] = %v, want 15", got)
	}
}

func TestMat2i4PutRejectsOver15(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2i4(1, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("want panic for value 16")
		}
	}()
	m.Put(0, 0, 16)
}

func TestMat2i4ItemsRejectsOver15(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2i4(1, 1)
	im := mat2i4Type.GetItemsMetadata()
	defer func() {
		if recover() == nil {
			t.Fatal("want panic for ![] 16")
		}
	}()
	im.SetByMultiIndex(m, Int(16), Int(0), Int(0))
}

func TestMat2i4ItemsRejectsNegative(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2i4(1, 1)
	im := mat2i4Type.GetItemsMetadata()
	defer func() {
		if recover() == nil {
			t.Fatal("want panic for ![] -1")
		}
	}()
	im.SetByMultiIndex(m, Int(-1), Int(0), Int(0))
}

func TestMat2i4PackedNibbleWordBoundary(t *testing.T) {
	// Spans a native uint word of 4-bit slots (8 on arm, 16 on amd64/arm64).
	spw := bits.UintSize / 4
	m := NewMat2i4(1, spw+2)
	lo, hi := spw-1, spw
	m.Put(0, 0, 1)
	m.Put(0, lo, 15)
	m.Put(0, hi, 7)
	if m.At(0, 0) != 1 || m.At(0, lo) != 15 || m.At(0, hi) != 7 {
		t.Fatalf("At(0,{0,%d,%d})=%v,%v,%v, want 1,15,7", lo, hi, m.At(0, 0), m.At(0, lo), m.At(0, hi))
	}
	if m.At(0, 1) != 0 || m.At(0, lo-1) != 0 || m.At(0, hi+1) != 0 {
		t.Fatalf("neighbors not zero: At(0,1)=%v At(0,%d)=%v At(0,%d)=%v",
			m.At(0, 1), lo-1, m.At(0, lo-1), hi+1, m.At(0, hi+1))
	}
	m.Put(0, lo, 0)
	if m.At(0, 0) != 1 || m.At(0, lo) != 0 || m.At(0, hi) != 7 {
		t.Fatalf("after clear %d: At(0,{0,%d,%d})=%v,%v,%v, want 1,0,7",
			lo, lo, hi, m.At(0, 0), m.At(0, lo), m.At(0, hi))
	}
}

func TestMat2i4AdjacentNibblesIndependent(t *testing.T) {
	m := NewMat2i4(1, 4)
	m.Put(0, 0, 0xA)
	m.Put(0, 1, 0x5)
	m.Put(0, 2, 0xC)
	m.Put(0, 3, 0x3)
	if m.At(0, 0) != 0xA || m.At(0, 1) != 0x5 || m.At(0, 2) != 0xC || m.At(0, 3) != 0x3 {
		t.Fatalf("got %v %v %v %v", m.At(0, 0), m.At(0, 1), m.At(0, 2), m.At(0, 3))
	}
	m.Put(0, 1, 0xF)
	if m.At(0, 0) != 0xA || m.At(0, 1) != 0xF || m.At(0, 2) != 0xC {
		t.Fatalf("after mid put: %v %v %v", m.At(0, 0), m.At(0, 1), m.At(0, 2))
	}
}

func TestMat2i4GetIter(t *testing.T) {
	m := NewMat2i4(1, 3)
	m.PutFlat(0, 1)
	m.PutFlat(1, 2)
	m.PutFlat(2, 3)
	it := m.GetIter()
	var got []Int
	for it.GetNext() {
		got = append(got, it.Current().(Int))
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("iter = %v, want [1 2 3]", got)
	}
}
