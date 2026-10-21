package core

import "testing"

func TestMat2fNewZeroFilled(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2f(2, 3)
	if m.Rows() != 2 || m.Cols() != 3 {
		t.Fatalf("shape %d×%d, want 2×3", m.Rows(), m.Cols())
	}
	if m.Len() != 6 {
		t.Fatalf("len=%d, want 6", m.Len())
	}
	for i, v := range m.data {
		if v != 0 {
			t.Fatalf("data[%d]=%v, want 0", i, v)
		}
	}
}

func TestMat2fNewPanicsNonPositive(t *testing.T) {
	for _, tc := range []struct{ r, c int }{{0, 1}, {1, 0}, {-1, 2}, {2, -1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewMat2f(%d,%d): expected panic", tc.r, tc.c)
				}
			}()
			_ = NewMat2f(tc.r, tc.c)
		}()
	}
}

func TestMat2fAtPut(t *testing.T) {
	m := NewMat2f(2, 2)
	m.Put(1, 0, 3.5)
	if got := m.At_(1, 0); got != 3.5 {
		t.Fatalf("At(1,0)=%v, want 3.5", got)
	}
}

func TestMat2fFlatMatchesRowCol(t *testing.T) {
	// 3×5 row-major: (row, col) → row*cols+col; (1,1) is flat 6.
	m := NewMat2f(3, 5)
	m.Put(1, 1, 2.5)
	if got := m.AtFlat(6); got != 2.5 {
		t.Fatalf("AtFlat(6)=%v, want 2.5 (same cell as At(1,1))", got)
	}
	m.PutFlat(6, 7.25)
	if got := m.At_(1, 1); got != 7.25 {
		t.Fatalf("At(1,1)=%v, want 7.25 (same cell as atFlat(6))", got)
	}
}

func TestMat2fAtOutOfRangePanics(t *testing.T) {
	m := NewMat2f(2, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = m.At_(2, 0)
}

func TestMat2fClassNewViaStack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(2))
	rp.DataStack.Push(Int(3))
	f, ths := memberRunFrame(rp, mat2fType)
	mm := mat2fType.GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing Mat2f.new")
	}
	mm.invokeCall(f, ths)
	m, ok := rp.DataStack.Pop().(*Mat2f)
	if !ok || m.Rows() != 2 || m.Cols() != 3 {
		t.Fatalf("got %v, want 2×3 Mat2f", rp.DataStack.Peek())
	}
	_ = task
}

func TestMat2fItemsMetadataFlatAndMulti(t *testing.T) {
	ResetTestRuntime()
	im := mat2fType.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil || im.setByOrdinal == nil {
		t.Fatal("Mat2f must register flat @() / !()")
	}
	if im.getByMultiIndex == nil || im.setByMultiIndex == nil || im.multiIndexArity != 2 {
		t.Fatal("Mat2f must register @[] / ![] with arity 2")
	}
	// Final ItemsMetadata probes for @[] / ![]
	if !bool(im.FinalType().GetPropertyMetadata("hasGetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasGetByIndices want true for Mat2f")
	}
	if !bool(im.FinalType().GetPropertyMetadata("hasSetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasSetByIndices want true for Mat2f")
	}
	vIm := vecNfType.GetItemsMetadata()
	if bool(vIm.FinalType().GetPropertyMetadata("hasGetByIndices").GetValue(vIm).(Bool)) {
		t.Fatal("VecNf must not report hasGetByIndices")
	}
	m := NewMat2f(2, 2)
	im.SetByOrdinal(m, Int(3), Float(7))
	if m.At_(1, 1) != 7 {
		t.Fatalf("flat !() At(1,1)=%v, want 7", m.At_(1, 1))
	}
	im.SetByMultiIndex(m, Float(1.5), Int(0), Int(1))
	if got := im.GetByMultiIndex(m, Int(0), Int(1)).(Float); got != 1.5 {
		t.Fatalf("@[] = %v, want 1.5", got)
	}
}

func TestOpAtBangIndexMat2fFlat(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	m := NewMat2f(2, 3)
	// !(): ( val flatIdx -- ) flat row-major
	rp.DataStack.Push(m)
	rp.DataStack.Push(Float(1.25))
	rp.DataStack.Push(Int(5)) // (1,2)
	RunOps(task, rp, OpBangIndex, OpStop)
	if got := m.At_(1, 2); got != 1.25 {
		t.Fatalf("after !() At(1,2)=%v, want 1.25", got)
	}
	rp.DataStack.Push(Int(5))
	RunOps(task, rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 1.25 {
		t.Fatalf("@() = %v, want 1.25", got)
	}
}

func TestOpAtBangMultiMat2f(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	m := NewMat2f(2, 3)
	rp.DataStack.Push(m)
	rp.DataStack.Push(Float(2.5))
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(task, rp, OpBangMulti, OpStop)
	if got := m.At_(1, 2); got != 2.5 {
		t.Fatalf("after ![] At(1,2)=%v, want 2.5", got)
	}
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(task, rp, OpAtMulti, OpStop)
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 2.5 {
		t.Fatalf("@[] = %v, want 2.5", got)
	}
}

func TestOpAtMultiMat2fViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	m := NewMat2f(2, 2)
	m.Put(0, 1, 9)
	rp.DataStack.Push(Int(0))
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(m)
	segs := CompoundPathForTest([]string{segAtMulti})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 9 {
		t.Fatalf("got %v, want 9", got)
	}
}

func TestOpAtIndexMat2fFlatViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	m := NewMat2f(2, 2)
	m.Put(0, 1, 9) // flat index 1
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(m)
	segs := CompoundPathForTest([]string{segAtIndex})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Float)
	if !ok || got != 9 {
		t.Fatalf("got %v, want 9", got)
	}
}

func TestMat2fIterableFlat(t *testing.T) {
	m := NewMat2f(2, 2)
	m.Put(0, 0, 1)
	m.Put(0, 1, 2)
	m.Put(1, 0, 3)
	m.Put(1, 1, 4)
	it := m.GetIter()
	var vals []Float
	for it.GetNext() {
		vals = append(vals, it.Current().(Float))
	}
	want := []Float{1, 2, 3, 4}
	if len(vals) != len(want) {
		t.Fatalf("vals=%v, want %v", vals, want)
	}
	for i := range want {
		if vals[i] != want[i] {
			t.Fatalf("vals=%v, want %v", vals, want)
		}
	}
}

func TestOpTryGetIterOnMat2f(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2f(1, 2)
	rp := NewTestRunProvider()
	rp.DataStack.Push(m)
	RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
	ok := rp.DataStack.Pop().(Bool)
	if !ok {
		t.Fatal("want true")
	}
	it, is := rp.DataStack.Pop().(*Mat2Iter[float64])
	if !is || it.mat != m {
		t.Fatalf("want *Mat2Iter[float64], got %T", rp.DataStack.Peek())
	}
}

func TestOpTryGetIterOnMat2RowRange(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2f(2, 2)
	rowRange := m.GetRowIter()
	rp := NewTestRunProvider()
	rp.DataStack.Push(rowRange)
	RunOps(NewTestTask(), rp, OpTryGetIter, OpStop)
	ok := rp.DataStack.Pop().(Bool)
	if !ok {
		t.Fatal("want true")
	}
	if _, is := rp.DataStack.Pop().(*RangeIterator); !is {
		t.Fatalf("Range is Iterable; want *RangeIterator, got %T", rp.DataStack.Peek())
	}
}

func TestMat2fRowColRanges(t *testing.T) {
	m := NewMat2f(2, 3)
	rows := intsFromIter(m.GetRowIter().GetIter())
	cols := intsFromIter(m.GetColIter().GetIter())
	if len(rows) != 2 || rows[0] != 0 || rows[1] != 1 {
		t.Fatalf("rows=%v, want 0,1", rows)
	}
	if len(cols) != 3 || cols[0] != 0 || cols[2] != 2 {
		t.Fatalf("cols=%v, want 0..2", cols)
	}
}

func TestMat2fNestedMultiIndexFillPattern(t *testing.T) {
	// Mirrors Final: getRowIter foreach … getColIter foreach j i ![]
	m := NewMat2f(2, 3)
	rowIt := m.GetRowIter().GetIter()
	for rowIt.GetNext() {
		j := int(rowIt.Current().(Int))
		colIt := m.GetColIter().GetIter()
		for colIt.GetNext() {
			i := int(colIt.Current().(Int))
			m.Put(j, i, float64(j*10+i))
		}
	}
	if m.At_(1, 2) != 12 {
		t.Fatalf("At(1,2)=%v, want 12", m.At_(1, 2))
	}
}

func TestMat2fGetRowIterMethod(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	m := NewMat2f(1, 1)
	f, ths := memberRunFrame(rp, m)
	mm := mat2fType.GetMethodMetadata("getRowIter")
	if mm == nil {
		t.Fatal("missing getRowIter")
	}
	mm.invokeCall(f, ths)
	it, ok := rp.DataStack.Pop().(*Range)
	if !ok || it.beg != 0 || it.end != 0 {
		t.Fatalf("got %#v, want Range 0..0", rp.DataStack.Peek())
	}
}

func TestMat2fGetColIterMethod(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	m := NewMat2f(2, 4)
	f, ths := memberRunFrame(rp, m)
	mm := mat2fType.GetMethodMetadata("getColIter")
	if mm == nil {
		t.Fatal("missing getColIter")
	}
	mm.invokeCall(f, ths)
	it, ok := rp.DataStack.Pop().(*Range)
	if !ok || it.beg != 0 || it.end != 3 {
		t.Fatalf("got %#v, want Range 0..3", rp.DataStack.Peek())
	}
}

func TestMat2u8ItemsInt(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2u8(2, 2)
	if m.FinalType() != mat2u8Type {
		t.Fatalf("FinalType = %v, want Mat2u8", m.FinalType())
	}
	im := mat2u8Type.GetItemsMetadata()
	im.SetByMultiIndex(m, Int(255), Int(1), Int(0))
	if m.At_(1, 0) != 255 {
		t.Fatalf("At(1,0)=%v, want 255", m.At_(1, 0))
	}
	got := im.GetByMultiIndex(m, Int(1), Int(0)).(Int)
	if got != 255 {
		t.Fatalf("@[] = %v, want 255", got)
	}
}

func TestMat2uItemsUint(t *testing.T) {
	ResetTestRuntime()
	m := NewMat2u(1, 2)
	im := mat2uType.GetItemsMetadata()
	im.SetByOrdinal(m, Int(0), Uint(7))
	im.SetByOrdinal(m, Int(1), Int(9))
	if m.AtFlat(0) != 7 || m.AtFlat(1) != 9 {
		t.Fatalf("store = %v %v, want 7 9", m.AtFlat(0), m.AtFlat(1))
	}
	got0 := im.GetByOrdinal(m, Int(0)).(Uint)
	got1 := im.GetByOrdinal(m, Int(1)).(Uint)
	if got0 != 7 || got1 != 9 {
		t.Fatalf("@() = %v %v, want Uint 7 9", got0, got1)
	}
}

func TestMat2FamilyParents(t *testing.T) {
	ResetTestRuntime()
	for _, typ := range []*Type{
		mat2u8Type, mat2u16Type, mat2u32Type, mat2uType,
		mat2i8Type, mat2i16Type, mat2i32Type, mat2iType,
		mat2f32Type, mat2fType, mat2cType,
	} {
		if typ == nil || typ.GetParent() != matProtoType {
			t.Fatalf("%v parent = %v, want ProtoMat", typ, typ.GetParent())
		}
	}
	if NewMat2f(1, 1).FinalType() != mat2fType {
		t.Fatal("NewMat2f FinalType want Mat2f")
	}
}
