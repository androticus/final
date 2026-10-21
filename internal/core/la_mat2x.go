package core

import "fmt"

// Series of Mat2 matrices (numeric types and Color).

type Mat2[T any] struct {
	typ *Type
	RowCol
	data []T
}

type Mat2u8 = Mat2[uint8]
type Mat2u16 = Mat2[uint16]
type Mat2u32 = Mat2[uint32]
type Mat2u = Mat2[uint]
type Mat2i8 = Mat2[int8]
type Mat2i16 = Mat2[int16]
type Mat2i32 = Mat2[int32]
type Mat2i = Mat2[int]
type Mat2f32 = Mat2[float32]
type Mat2f = Mat2[float64]

// AnyMat is the generic Final Mat (Any cells). Final type name is "Mat".
type AnyMat = Mat2[any]

func NewAnyMat(rows, cols int) *AnyMat {
	m := NewMat2[any](rows, cols)
	m.typ = anyMatType
	return m
}

func NewMat2[T any](rows, cols int) *Mat2[T] {
	if rows <= 0 || cols <= 0 {
		panic(fmt.Errorf("Mat2: rows and cols must be > 0, got %d×%d", rows, cols))
	}
	return &Mat2[T]{
		RowCol: RowCol{rows: rows, cols: cols},
		data:   make([]T, rows*cols),
	}
}

func NewMat2u8(rows, cols int) *Mat2u8   { return newMat2Typed[uint8](mat2u8Type, rows, cols) }
func NewMat2u16(rows, cols int) *Mat2u16 { return newMat2Typed[uint16](mat2u16Type, rows, cols) }
func NewMat2u32(rows, cols int) *Mat2u32 { return newMat2Typed[uint32](mat2u32Type, rows, cols) }
func NewMat2u(rows, cols int) *Mat2u     { return newMat2Typed[uint](mat2uType, rows, cols) }
func NewMat2i8(rows, cols int) *Mat2i8   { return newMat2Typed[int8](mat2i8Type, rows, cols) }
func NewMat2i16(rows, cols int) *Mat2i16 { return newMat2Typed[int16](mat2i16Type, rows, cols) }
func NewMat2i32(rows, cols int) *Mat2i32 { return newMat2Typed[int32](mat2i32Type, rows, cols) }
func NewMat2i(rows, cols int) *Mat2i     { return newMat2Typed[int](mat2iType, rows, cols) }
func NewMat2f32(rows, cols int) *Mat2f32 { return newMat2Typed[float32](mat2f32Type, rows, cols) }
func NewMat2f(rows, cols int) *Mat2f     { return newMat2Typed[float64](mat2fType, rows, cols) }

func newMat2Typed[T any](typ *Type, rows, cols int) *Mat2[T] {
	m := NewMat2[T](rows, cols)
	m.typ = typ
	return m
}

func (m *Mat2[T]) FinalType() *Type { return m.typ }

func (m *Mat2[T]) Len() int {
	return len(m.data)
}

func (m *Mat2[T]) index(row, col int) int {
	if row < 0 || row >= m.rows || col < 0 || col >= m.cols {
		panic(fmt.Errorf("Mat2: index (%d,%d) out of range [0,%d)×[0,%d)", row, col, m.rows, m.cols))
	}
	return row*m.cols + col
}

func mat2uFromCell(c any) uint {
	switch v := c.(type) {
	case Uint:
		return uint(v)
	case Int:
		return uint(v)
	default:
		panic(fmt.Errorf("Mat2u: expected Int or Uint, got %T", c))
	}
}

func (m *Mat2[T]) At_(row, col int) T {
	return m.data[m.index(row, col)]
}

func (m *Mat2[T]) Put(row, col int, v T) {
	m.data[m.index(row, col)] = v
}

func (m *Mat2[T]) AtFlat(i int) T {
	// we use the runtime bounds check for speed
	return m.data[i]
}

func (m *Mat2[T]) PutFlat(i int, v T) {
	// we use the runtime bounds check for speed
	m.data[i] = v
}

// GetIter returns a flat cell iterator (Mat2 is Iterable). Current is the Final cell.
func (m *Mat2[T]) GetIter() Iterator {
	return &Mat2Iter[T]{mat: m, i: -1}
}

// Mat2Iter walks Mat2 cells in flat row-major order.
type Mat2Iter[T any] struct {
	mat *Mat2[T]
	i   int // last fetched flat index; -1 before first GetNext
}

func (it *Mat2Iter[T]) GetNext() bool {
	next := it.i + 1
	if next >= len(it.mat.data) {
		return false
	}
	it.i = next
	return true
}

func (it *Mat2Iter[T]) Current() any {
	return mat2CellOf(it.mat.data[it.i])
}

func mat2CellOf[T any](v T) any {
	switch x := any(v).(type) {
	case uint8:
		return Int(x)
	case uint16:
		return Int(x)
	case uint32:
		return Int(x)
	case uint:
		return Uint(x)
	case int8:
		return Int(x)
	case int16:
		return Int(x)
	case int32:
		return Int(x)
	case int:
		return Int(x)
	case float32:
		return Float(x)
	case float64:
		return Float(x)
	case Color:
		return x
	default:
		return any(v) // Mat (AnyMat) and future cell types
	}
}

func registerMat2Members[T any](t *Type, itemTypeName string, toCell func(T) any, fromCell func(any) T) {
	t.registerProperty("rows", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2[T]).rows) },
		nil,
	)
	t.registerProperty("cols", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2[T]).cols) },
		nil,
	)
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Mat2[T]).Len()) },
		nil,
	)
	t.registerItemsData2(itemTypeName,
		func(inst Any, idx Any) any {
			return toCell(inst.(*Mat2[T]).AtFlat(int(idx.(Int))))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*Mat2[T]).PutFlat(int(idx.(Int)), fromCell(value))
		},
		nil, nil, nil,
		func(inst Any, idxs ...Any) any {
			if len(idxs) != 2 {
				panic(fmt.Errorf("%s @[]: need 2 indexes, got %d", t.name, len(idxs)))
			}
			return toCell(inst.(*Mat2[T]).At_(int(idxs[0].(Int)), int(idxs[1].(Int))))
		},
		func(inst Any, value Any, idxs ...Any) {
			if len(idxs) != 2 {
				panic(fmt.Errorf("%s ![]: need 2 indexes, got %d", t.name, len(idxs)))
			}
			inst.(*Mat2[T]).Put(int(idxs[0].(Int)), int(idxs[1].(Int)), fromCell(value))
		},
		2,
	)
	// ( rows:i cols:i -- mat )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		cols := int(f.RP.DataStack.Pop().(Int))
		rows := int(f.RP.DataStack.Pop().(Int))
		m := NewMat2[T](rows, cols)
		m.typ = t
		f.RP.DataStack.Push(m)
	})
	t.registerNativeMethod("getRowIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2[T]).GetRowIter())
	})
	t.registerNativeMethod("getColIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Mat2[T]).GetColIter())
	})
	t.registerNativeMethod("wrap", func(f *RunFrame, ths *TypeObjPair) {
		m := ths.Obj.(*Mat2[T])
		wrapMat2dOnFrame(f, m.rows, m.cols)
	})
}

func registerMat2Type[T any](finalName, itemTypeName string, typSlot **Type, toCell func(T) any, fromCell func(any) T) {
	t := &Type{
		name:            finalName,
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Mat2[T]](finalName),
	}
	*typSlot = t
	registerMat2Members[T](t, itemTypeName, toCell, fromCell)
	GetTypeRegistrar().RegisterCurrentType(t)
	t.SetParent(matProtoType)
}

func CreateRegisterMat2Types() {
	registerMat2Type[uint8]("Mat2u8", "Int", &mat2u8Type,
		func(v uint8) any { return Int(v) },
		func(c any) uint8 { return uint8(c.(Int)) })
	registerMat2Type[uint16]("Mat2u16", "Int", &mat2u16Type,
		func(v uint16) any { return Int(v) },
		func(c any) uint16 { return uint16(c.(Int)) })
	registerMat2Type[uint32]("Mat2u32", "Int", &mat2u32Type,
		func(v uint32) any { return Int(v) },
		func(c any) uint32 { return uint32(c.(Int)) })
	registerMat2Type[uint]("Mat2u", "Uint", &mat2uType,
		func(v uint) any { return Uint(v) },
		mat2uFromCell)
	registerMat2Type[int8]("Mat2i8", "Int", &mat2i8Type,
		func(v int8) any { return Int(v) },
		func(c any) int8 { return int8(c.(Int)) })
	registerMat2Type[int16]("Mat2i16", "Int", &mat2i16Type,
		func(v int16) any { return Int(v) },
		func(c any) int16 { return int16(c.(Int)) })
	registerMat2Type[int32]("Mat2i32", "Int", &mat2i32Type,
		func(v int32) any { return Int(v) },
		func(c any) int32 { return int32(c.(Int)) })
	registerMat2Type[int]("Mat2i", "Int", &mat2iType,
		func(v int) any { return Int(v) },
		func(c any) int { return int(c.(Int)) })
	registerMat2Type[float32]("Mat2f32", "Float", &mat2f32Type,
		func(v float32) any { return Float(v) },
		func(c any) float32 { return float32(c.(Float)) })
	registerMat2Type[float64]("Mat2f", "Float", &mat2fType,
		func(v float64) any { return Float(v) },
		func(c any) float64 { return float64(c.(Float)) })
	registerMat2Type[any]("Mat", "Any", &anyMatType,
		func(v any) any { return v },
		func(c any) any { return c })
	registerMat2CType()
}

var (
	mat2u8Type  *Type
	mat2u16Type *Type
	mat2u32Type *Type
	mat2uType   *Type
	mat2i8Type  *Type
	mat2i16Type *Type
	mat2i32Type *Type
	mat2iType   *Type
	mat2f32Type *Type
	mat2fType   *Type
	anyMatType  *Type
	mat2cType   *Type
)

var (
	_ Typed    = (*Mat2f)(nil)
	_ Iterable = (*Mat2f)(nil)
	_ Iterator = (*Mat2Iter[float64])(nil)
)
