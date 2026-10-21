package core

import (
	"fmt"
	"math/bits"
)

// Mat2i1 is a dense row-major Bit matrix.
// Bits are packed into native uint words (bits.UintSize: 32 or 64).
type Mat2i1 struct {
	RowCol
	len  int
	data []uint
}

func (m *Mat2i1) FinalType() *Type { return mat2i1Type }

func (m *Mat2i1) Len() int { return m.len }

// NewMat2i1 allocates a zero-filled rows×cols matrix. rows and cols must both be > 0.
func NewMat2i1(rows, cols int) *Mat2i1 {
	if rows <= 0 || cols <= 0 {
		panic(fmt.Errorf("Mat2i1.new: rows and cols must be > 0, got %d×%d", rows, cols))
	}
	len := rows * cols
	bpw := bits.UintSize
	cells := (len + bpw - 1) / bpw
	return &Mat2i1{
		RowCol: RowCol{rows: rows, cols: cols},
		len:    len,
		data:   make([]uint, cells),
	}
}

func (m *Mat2i1) indexFlat(idx int) (cell int, shift uint) {
	if idx < 0 || idx >= m.len {
		panic(fmt.Errorf("Mat2i1: index (%d) out of range [0,%d))", idx, m.len))
	}
	bpw := bits.UintSize
	cell = idx / bpw
	shift = uint(idx % bpw)
	return
}

func (m *Mat2i1) cellBitIndex(row, col int) (cell int, shift uint) {
	if row < 0 || row >= m.rows || col < 0 || col >= m.cols {
		panic(fmt.Errorf("Mat2i1: index (%d,%d) out of range [0,%d)×[0,%d)", row, col, m.rows, m.cols))
	}
	bitIndex := row*m.cols + col
	bpw := bits.UintSize
	cell = bitIndex / bpw
	shift = uint(bitIndex % bpw)
	return
}

func BitToInt(v uint) Int {
	if v != 0 {
		return 1
	}
	return 0
}

func (m *Mat2i1) At(row, col int) Int {
	cell, shift := m.cellBitIndex(row, col)
	return BitToInt(m.data[cell] & (1 << shift))
}

func (m *Mat2i1) AtFlat(idx int) Int {
	cell, shift := m.indexFlat(idx)
	return BitToInt(m.data[cell] & (1 << shift))
}

func (m *Mat2i1) Put(row, col int, v int) {
	cell, shift := m.cellBitIndex(row, col)
	if v != 0 {
		m.data[cell] |= 1 << shift
	} else {
		m.data[cell] &= ^(1 << shift)
	}
}

func (m *Mat2i1) PutFlat(idx int, v int) {
	cell, shift := m.indexFlat(idx)
	if v != 0 {
		m.data[cell] |= 1 << shift
	} else {
		m.data[cell] &= ^(1 << shift)
	}
}

func registerMat2i1Members(t *Type) {
	t.registerProperty("rows", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i1).rows) },
		nil,
	)
	t.registerProperty("cols", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i1).cols) },
		nil,
	)
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2i1).Len()) },
		nil,
	)
	// Flat @()/!(); multi-index @[]/![] with arity 2 (row, col).
	t.registerItemsData2("Int",
		func(inst Any, idx Any) any {
			return inst.(*Mat2i1).AtFlat(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Mat2i1).PutFlat(int(idx.(Int)), value.(Int))
		},
		nil, nil, nil,
		func(inst Any, idxs ...Any) any {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2i1 @[]: need 2 indexes, got %d", len(idxs)))
			}
			return inst.(*Mat2i1).At(int(idxs[0].(Int)), int(idxs[1].(Int)))
		},
		func(inst Any, value Any, idxs ...Any) {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Mat2i1 ![]: need 2 indexes, got %d", len(idxs)))
			}
			inst.(*Mat2i1).Put(int(idxs[0].(Int)), int(idxs[1].(Int)), value.(Int))
		},
		2,
	)
	t.registerNativeMethod("getRowIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2i1).GetRowIter())
	})
	t.registerNativeMethod("getColIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2i1).GetColIter())
	})
	t.registerNativeMethod("wrap", func(f *RunFrame, ths *TypeObjPair) {
		m := ths.Obj.(*Mat2i1)
		wrapMat2dOnFrame(f, m.rows, m.cols)
	})

	// ( rows:i cols:i -- mat:Mat2i1 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		cols := int(f.RP.DataStack.Pop().(Int))
		rows := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewMat2i1(rows, cols))
	})
}

func CreateRegisterMat2i1Type() {
	t := &Type{
		name:            "Mat2i1",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Mat2i1]("Mat2i1"),
	}
	registerMat2i1Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	t.SetParent(matProtoType)
	mat2i1Type = t
}

var (
	mat2i1Type *Type
	_          Typed = (*Mat2i1)(nil)
)
