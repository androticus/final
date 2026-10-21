package core

import "testing"

func TestMat2cNewZeroFilled(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2C(2, 3)
	if m.Rows() != 2 || m.Cols() != 3 {
		t.Fatalf("shape %d×%d, want 2×3", m.Rows(), m.Cols())
	}
	if m.Len() != 6 {
		t.Fatalf("len=%d, want 6", m.Len())
	}
	if m.FinalType() != mat2cType {
		t.Fatalf("FinalType = %v, want Mat2c", m.FinalType())
	}
	z := Color{}
	for i, v := range m.data {
		if v != z {
			t.Fatalf("data[%d]=%v, want zero Color", i, v)
		}
	}
}

func TestMat2cAtPut(t *testing.T) {
	m := NewMat2C(2, 3)
	ink := NewColor(255, 128, 0, 200)
	m.Put(1, 2, ink)
	if got := m.At_(1, 2); got != ink {
		t.Fatalf("At(1,2)=%v, want %v", got, ink)
	}
	if got := m.AtFlat(5); got != ink {
		t.Fatalf("AtFlat(5)=%v, want %v", got, ink)
	}
	m.PutFlat(0, ink)
	if got := m.At_(0, 0); got != ink {
		t.Fatalf("PutFlat(0) At(0,0)=%v, want %v", got, ink)
	}
}

func TestMat2cItemsMetadata(t *testing.T) {
	ResetTestRuntime()
	im := mat2cType.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("Mat2c must register flat @() / !()")
	}
	if im.getByMultiIndex == nil || im.setByMultiIndex == nil || im.multiIndexArity != 2 {
		t.Fatal("Mat2c must register @[] / ![] with arity 2")
	}
	m := NewMat2C(2, 2)
	ink := NewColor(1, 2, 3, 255)
	im.SetByOrdinal(m, Int(3), ink)
	if m.At_(1, 1) != ink {
		t.Fatalf("flat !() At(1,1)=%v, want %v", m.At_(1, 1), ink)
	}
	ink2 := NewColor(9, 8, 7, 255)
	im.SetByMultiIndex(m, ink2, Int(0), Int(1))
	if got := im.GetByMultiIndex(m, Int(0), Int(1)).(Color); got != ink2 {
		t.Fatalf("@[] = %v, want %v", got, ink2)
	}
}

func TestMat2cWrap(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2C(2, 3)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, m)
	mm := mat2cType.GetMethodMetadata("wrap")
	if mm == nil {
		t.Fatal("missing Mat2c.wrap")
	}
	rp.DataStack.Push(Int(-1))
	rp.DataStack.Push(Int(3))
	mm.invokeCall(f, ths)
	col := rp.DataStack.Pop().(Int)
	row := rp.DataStack.Pop().(Int)
	if row != 1 || col != 0 {
		t.Fatalf("wrap(-1, 3) on 2×3 = (%d, %d), want (1, 0)", row, col)
	}
}

func TestMat2cIterYieldsColor(t *testing.T) {
	m := NewMat2C(1, 2)
	ink := NewColor(10, 20, 30, 255)
	m.PutFlat(1, ink)
	it := m.GetIter()
	if !it.GetNext() {
		t.Fatal("first GetNext")
	}
	if got := it.Current().(Color); got != (Color{}) {
		t.Fatalf("cell 0 = %v, want zero", got)
	}
	if !it.GetNext() {
		t.Fatal("second GetNext")
	}
	if got := it.Current().(Color); got != ink {
		t.Fatalf("cell 1 = %v, want %v", got, ink)
	}
}

func TestMat2cNoNamedAtPutMethods(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"at", "atFlat", "put", "putFlat"} {
		if mat2cType.GetMethodMetadata(name) != nil {
			t.Fatalf("Mat2c must not expose named method %q", name)
		}
	}
}
