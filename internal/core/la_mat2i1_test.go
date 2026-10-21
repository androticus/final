package core

import (
	"math/bits"
	"testing"
)

func TestMat2i1NewZeroFilled(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2i1(2, 3)
	if m.Rows() != 2 || m.Cols() != 3 {
		t.Fatalf("shape %d×%d, want 2×3", m.Rows(), m.Cols())
	}
	if m.Len() != 6 {
		t.Fatalf("len=%d, want 6", m.Len())
	}
	for r := 0; r < m.Rows(); r++ {
		for c := 0; c < m.Cols(); c++ {
			if got := m.At(r, c); got != 0 {
				t.Fatalf("At(%d,%d)=%v, want 0", r, c, got)
			}
		}
	}
	for i, w := range m.data {
		if w != 0 {
			t.Fatalf("data[%d]=%v, want 0", i, w)
		}
	}
}

func TestMat2i1NewPanicsNonPositive(t *testing.T) {
	for _, tc := range []struct{ r, c int }{{0, 1}, {1, 0}, {-1, 2}, {2, -1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewMat2i1(%d,%d): expected panic", tc.r, tc.c)
				}
			}()
			_ = NewMat2i1(tc.r, tc.c)
		}()
	}
}

func TestMat2i1AtPut(t *testing.T) {
	m := NewMat2i1(2, 2)
	m.Put(1, 0, 1)
	if got := m.At(1, 0); got != 1 {
		t.Fatalf("At(1,0)=%v, want 1", got)
	}
	if got := m.At(0, 0); got != 0 {
		t.Fatalf("At(0,0)=%v, want 0", got)
	}
	m.Put(1, 0, 0)
	if got := m.At(1, 0); got != 0 {
		t.Fatalf("after clear At(1,0)=%v, want 0", got)
	}
}

func TestMat2i1FlatMatchesRowCol(t *testing.T) {
	// 3×5 row-major: (row, col) → row*cols+col; (1,1) is flat 6.
	m := NewMat2i1(3, 5)
	m.Put(1, 1, 1)
	if got := m.AtFlat(6); got != 1 {
		t.Fatalf("AtFlat(6)=%v, want 1 (same cell as At(1,1))", got)
	}
	m.Put(1, 1, 0)
	m.PutFlat(6, 1)
	if got := m.At(1, 1); got != 1 {
		t.Fatalf("At(1,1)=%v, want 1 (same cell as AtFlat(6))", got)
	}
}

func TestMat2i1PutNonZeroStoresOne(t *testing.T) {
	m := NewMat2i1(1, 3)
	for _, v := range []int{1, 2, -1, 99} {
		m.Put(0, 1, 0)
		m.Put(0, 1, v)
		if got := m.At(0, 1); got != 1 {
			t.Fatalf("Put(0,1,%d) At(0,1)=%v, want 1", v, got)
		}
	}
}

func TestMat2i1AtReturnsZeroOrOne(t *testing.T) {
	// Shift ≠ 0 so a raw bit-mask readout would not equal 1.
	m := NewMat2i1(2, 4)
	m.Put(0, 3, 1) // flat 3
	m.Put(1, 2, 1) // flat 6
	if got := m.At(0, 3); got != 1 {
		t.Fatalf("At(0,3)=%v, want 1", got)
	}
	if got := m.At(1, 2); got != 1 {
		t.Fatalf("At(1,2)=%v, want 1", got)
	}
	if got := m.AtFlat(3); got != 1 {
		t.Fatalf("AtFlat(3)=%v, want 1", got)
	}
	if got := m.AtFlat(6); got != 1 {
		t.Fatalf("AtFlat(6)=%v, want 1", got)
	}
}

func TestMat2i1PackedWordBoundary(t *testing.T) {
	// Spans a native uint word boundary (32 on arm, 64 on amd64/arm64).
	bpw := bits.UintSize
	m := NewMat2i1(1, bpw+6)
	lo, hi := bpw-1, bpw
	m.Put(0, 0, 1)
	m.Put(0, lo, 1)
	m.Put(0, hi, 1)
	if m.At(0, 0) != 1 || m.At(0, lo) != 1 || m.At(0, hi) != 1 {
		t.Fatalf("At(0,{0,%d,%d})=%v,%v,%v, want 1,1,1", lo, hi, m.At(0, 0), m.At(0, lo), m.At(0, hi))
	}
	if m.At(0, 1) != 0 || m.At(0, lo-1) != 0 || m.At(0, hi+1) != 0 {
		t.Fatalf("neighbors not zero: At(0,1)=%v At(0,%d)=%v At(0,%d)=%v",
			m.At(0, 1), lo-1, m.At(0, lo-1), hi+1, m.At(0, hi+1))
	}
	m.Put(0, lo, 0)
	if m.At(0, 0) != 1 || m.At(0, lo) != 0 || m.At(0, hi) != 1 {
		t.Fatalf("after clear %d: At(0,{0,%d,%d})=%v,%v,%v, want 1,0,1",
			lo, lo, hi, m.At(0, 0), m.At(0, lo), m.At(0, hi))
	}
}

func TestMat2i1AtOutOfRangePanics(t *testing.T) {
	m := NewMat2i1(2, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = m.At(2, 0)
}

func TestMat2i1PutOutOfRangePanics(t *testing.T) {
	m := NewMat2i1(2, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	m.Put(0, 2, 1)
}

func TestMat2i1AtFlatOutOfRangePanics(t *testing.T) {
	m := NewMat2i1(2, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = m.AtFlat(4)
}

func TestMat2i1ClassNewViaStack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(2))
	rp.DataStack.Push(Int(3))
	f, ths := memberRunFrame(rp, mat2i1Type)
	mm := mat2i1Type.GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing Mat2i1.new")
	}
	mm.invokeCall(f, ths)
	m, ok := rp.DataStack.Pop().(*Mat2i1)
	if !ok || m.Rows() != 2 || m.Cols() != 3 {
		t.Fatalf("got %v, want 2×3 Mat2i1", rp.DataStack.Peek())
	}
	_ = task
}

func TestMat2i1ItemsMetadataFlatAndMulti(t *testing.T) {
	ResetTestRuntime()
	im := mat2i1Type.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("Mat2i1 must register flat @() / !()")
	}
	if im.getByMultiIndex == nil || im.setByMultiIndex == nil || im.multiIndexArity != 2 {
		t.Fatal("Mat2i1 must register @[] / ![] with arity 2")
	}
	if !bool(im.FinalType().GetPropertyMetadata("hasGetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasGetByIndices want true for Mat2i1")
	}
	if !bool(im.FinalType().GetPropertyMetadata("hasSetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasSetByIndices want true for Mat2i1")
	}
	m := NewMat2i1(2, 2)
	im.SetByOrdinal(m, Int(3), Int(1))
	if m.At(1, 1) != 1 {
		t.Fatalf("flat !() At(1,1)=%v, want 1", m.At(1, 1))
	}
	im.SetByMultiIndex(m, Int(1), Int(0), Int(1))
	if got := im.GetByMultiIndex(m, Int(0), Int(1)).(Int); got != 1 {
		t.Fatalf("@[] = %v, want 1", got)
	}
}

func TestOpAtBangIndexMat2i1Flat(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	m := NewMat2i1(2, 3)
	rp.DataStack.Push(m)
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(5)) // (1,2)
	RunOps(task, rp, OpBangIndex, OpStop)
	if got := m.At(1, 2); got != 1 {
		t.Fatalf("after !() At(1,2)=%v, want 1", got)
	}
	rp.DataStack.Push(Int(5))
	RunOps(task, rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("@() = %v, want 1", got)
	}
}

func TestOpAtBangMultiMat2i1(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	m := NewMat2i1(2, 3)
	rp.DataStack.Push(m)
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(task, rp, OpBangMulti, OpStop)
	if got := m.At(1, 2); got != 1 {
		t.Fatalf("after ![] At(1,2)=%v, want 1", got)
	}
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(task, rp, OpAtMulti, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("@[] = %v, want 1", got)
	}
}

func TestOpAtMultiMat2i1ViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	m := NewMat2i1(2, 2)
	m.Put(0, 1, 1)
	rp.DataStack.Push(Int(0))
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(m)
	segs := CompoundPathForTest([]string{segAtMulti})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
}

func TestMat2i1GetRowColIterMethods(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	m := NewMat2i1(2, 4)
	f, ths := memberRunFrame(rp, m)
	mm := mat2i1Type.GetMethodMetadata("getRowIter")
	if mm == nil {
		t.Fatal("missing getRowIter")
	}
	mm.invokeCall(f, ths)
	rows, ok := rp.DataStack.Pop().(*Range)
	if !ok || rows.beg != 0 || rows.end != 1 {
		t.Fatalf("getRowIter = %#v, want Range 0..1", rp.DataStack.Peek())
	}
	mm = mat2i1Type.GetMethodMetadata("getColIter")
	if mm == nil {
		t.Fatal("missing getColIter")
	}
	mm.invokeCall(f, ths)
	cols, ok := rp.DataStack.Pop().(*Range)
	if !ok || cols.beg != 0 || cols.end != 3 {
		t.Fatalf("getColIter = %#v, want Range 0..3", rp.DataStack.Peek())
	}
}

func TestOpAtIndexMat2i1FlatViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	m := NewMat2i1(2, 2)
	m.Put(0, 1, 1) // flat index 1
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(m)
	segs := CompoundPathForTest([]string{segAtIndex})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
}
