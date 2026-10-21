package core

import (
	"fmt"
	"math/bits"
)

const (
	mat2i4BitsPerCell = 4
	mat2i4NibbleMask  = uint(0xF)
)

// Mat2i4 is a dense row-major grid of 4-bit indexes (0..15).
// Values are packed into native uint words (bits.UintSize), same strategy as Mat2i1.
// Final cells are Int. Put panics if v is outside 0..15.
type Mat2i4 struct {
	RowCol
	len  int
	data []uint
}

func (m *Mat2i4) FinalType() *Type { return mat2i4Type }

// NewMat2i4 allocates a zero-filled rows×cols matrix. rows and cols must both be > 0.
func NewMat2i4(rows, cols int) *Mat2i4 {
	if rows <= 0 || cols <= 0 {
		panic(fmt.Errorf("Mat2i4.new: rows and cols must be > 0, got %d×%d", rows, cols))
	}
	n := rows * cols
	spw := mat2i4SlotsPerWord()
	words := (n + spw - 1) / spw
	return &Mat2i4{
		RowCol: RowCol{rows: rows, cols: cols},
		len:    n,
		data:   make([]uint, words),
	}
}

func mat2i4SlotsPerWord() int {
	return bits.UintSize / mat2i4BitsPerCell
}

func requireI4(v int) {
	if v < 0 || v > 15 {
		panic(fmt.Errorf("Mat2i4: value %d out of range 0..15", v))
	}
}

func i4FromCell(c any) uint8 {
	n := int(c.(Int))
	requireI4(n)
	return uint8(n)
}

func (m *Mat2i4) Len() int { return m.len }

func (m *Mat2i4) indexFlat(idx int) (word int, shift uint) {
	if idx < 0 || idx >= m.len {
		panic(fmt.Errorf("Mat2i4: index (%d) out of range [0,%d))", idx, m.len))
	}
	spw := mat2i4SlotsPerWord()
	word = idx / spw
	shift = uint(idx%spw) * mat2i4BitsPerCell
	return
}

func (m *Mat2i4) cellNibbleIndex(row, col int) (word int, shift uint) {
	if row < 0 || row >= m.rows || col < 0 || col >= m.cols {
		panic(fmt.Errorf("Mat2i4: index (%d,%d) out of range [0,%d)×[0,%d)", row, col, m.rows, m.cols))
	}
	return m.indexFlat(row*m.cols + col)
}

func (m *Mat2i4) At(row, col int) uint8 {
	word, shift := m.cellNibbleIndex(row, col)
	return uint8((m.data[word] >> shift) & mat2i4NibbleMask)
}

func (m *Mat2i4) AtFlat(i int) uint8 {
	word, shift := m.indexFlat(i)
	return uint8((m.data[word] >> shift) & mat2i4NibbleMask)
}

func (m *Mat2i4) Put(row, col int, v uint8) {
	requireI4(int(v))
	word, shift := m.cellNibbleIndex(row, col)
	m.data[word] = (m.data[word] & ^(mat2i4NibbleMask << shift)) | (uint(v) << shift)
}

func (m *Mat2i4) PutFlat(i int, v uint8) {
	requireI4(int(v))
	word, shift := m.indexFlat(i)
	m.data[word] = (m.data[word] & ^(mat2i4NibbleMask << shift)) | (uint(v) << shift)
}

// GetIter returns a flat cell iterator (Mat2i4 is Iterable). Current is Int 0..15.
func (m *Mat2i4) GetIter() Iterator {
	return &Mat2i4Iter{mat: m, i: -1}
}

// Mat2i4Iter walks Mat2i4 cells in flat row-major order.
type Mat2i4Iter struct {
	mat *Mat2i4
	i   int // last fetched flat index; -1 before first GetNext
}

func (it *Mat2i4Iter) GetNext() bool {
	next := it.i + 1
	if next >= it.mat.len {
		return false
	}
	it.i = next
	return true
}

func (it *Mat2i4Iter) Current() any {
	return Int(it.mat.AtFlat(it.i))
}

func registerMat2i4Members(t *Type) {
	t.registerProperty("rows", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i4).Rows()) },
		nil,
	)
	t.registerProperty("cols", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i4).Cols()) },
		nil,
	)
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i4).Len()) },
		nil,
	)
	t.registerItemsData2("Int",
		func(inst Any, idx Any) any {
			return Int(inst.(*Mat2i4).AtFlat(int(idx.(Int))))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Mat2i4).PutFlat(int(idx.(Int)), i4FromCell(value))
		},
		nil, nil, nil,
		func(inst Any, idxs ...Any) any {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2i4 @[]: need 2 indexes, got %d", len(idxs)))
			}
			return Int(inst.(*Mat2i4).At(int(idxs[0].(Int)), int(idxs[1].(Int))))
		},
		func(inst Any, value Any, idxs ...Any) {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2i4 ![]: need 2 indexes, got %d", len(idxs)))
			}
			inst.(*Mat2i4).Put(int(idxs[0].(Int)), int(idxs[1].(Int)), i4FromCell(value))
		},
		2,
	)
	// ( rows:i cols:i -- mat:Mat2i4 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		cols := int(f.RP.DataStack.Pop().(Int))
		rows := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewMat2i4(rows, cols))
	})
	t.registerNativeMethod("getRowIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2i4).GetRowIter())
	})
	t.registerNativeMethod("getColIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2i4).GetColIter())
	})
	t.registerNativeMethod("wrap", func(f *RunFrame, ths *TypeObjPair) {
		m := ths.Obj.(*Mat2i4)
		wrapMat2dOnFrame(f, m.Rows(), m.Cols())
	})
}

func CreateRegisterMat2i4Type() {
	t := &Type{
		name:            "Mat2i4",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Mat2i4]("Mat2i4"),
	}
	registerMat2i4Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	t.SetParent(matProtoType)
	mat2i4Type = t
}

var (
	mat2i4Type *Type
	_          Typed    = (*Mat2i4)(nil)
	_          Iterable = (*Mat2i4)(nil)
	_          Iterator = (*Mat2i4Iter)(nil)
)
