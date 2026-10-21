package core

import "fmt"

type CodeDataTable[T any] struct {
	items     []T
	maxSize   int // max number of cells; Append panics at overflow
	finalType *Type
}

type CodeTable = CodeDataTable[int]
type DataTable = CodeDataTable[any]

// NewCodeTable returns a new CodeTable. maxSize is the maximum number of cells
// (Append panics at overflow). Panics if maxSize is 0. cap is the initial slice
// capacity; cap=0 means default autogrow (no preallocation).
func NewCodeTable(maxSize int, cap int) *CodeTable {
	t := newCodeDataTable[int](maxSize, cap)
	t.finalType = GetTypeRegistrar().Get("CodeTable")
	return t
}

// NewDataTable returns a new DataTable. maxSize is the maximum number of cells
// (Append panics at overflow). Panics if maxSize is 0. cap is the initial slice
// capacity; cap=0 means default autogrow (no preallocation).
func NewDataTable(maxSize int, cap int) *DataTable {
	t := newCodeDataTable[any](maxSize, cap)
	t.finalType = GetTypeRegistrar().Get("DataTable")
	return t
}

func newCodeDataTable[T any](maxSize int, cap int) *CodeDataTable[T] {
	if maxSize == 0 {
		panic(fmt.Errorf("CodeDataTable: maxSize must be > 0"))
	}
	var items []T
	if cap > 0 {
		items = make([]T, 0, cap)
	} else {
		items = make([]T, 0)
	}
	return &CodeDataTable[T]{
		items:   items,
		maxSize: maxSize,
	}
}

// FinalType returns the sealed Final type set at table creation (see NewCodeTable / NewDataTable).
func (t *CodeDataTable[T]) FinalType() *Type {
	return t.finalType
}

// Read returns the cell at idx. Panics if idx is out of range.
func (t *CodeDataTable[T]) Read(idx int) T {
	// not path. we delegate bounds check to go runtime
	/*if idx < 0 || idx >= len(t.items) {
		panic(fmt.Errorf("CodeDataTable: read out of range: idx %d, len %d", idx, len(t.items)))
	}*/
	return t.items[idx]
}

// Get the address of a cell; **IMPORTANT**
func (t *CodeDataTable[T]) GetCellAddr(idx int) *T {
	if idx < 0 || idx >= len(t.items) {
		panic(fmt.Errorf("CodeDataTable: read out of range: idx %d, len %d", idx, len(t.items)))
	}
	return &(t.items[idx])
}

// Write stores val at an existing cell. Panics unless 0 <= idx < len.
func (t *CodeDataTable[T]) Write(idx int, val T) {
	if idx < 0 || idx >= len(t.items) {
		panic(fmt.Errorf("CodeDataTable: write out of range: idx %d, len %d", idx, len(t.items)))
	}
	t.items[idx] = val
}

// Append adds val at the next index and returns that index. Panics on overflow.
func (t *CodeDataTable[T]) Append(val T) int {
	if len(t.items) >= t.maxSize {
		panic(fmt.Errorf("CodeDataTable: overflow at len %d, maxSize %d", len(t.items), t.maxSize))
	}
	idx := len(t.items)
	t.items = append(t.items, val)
	return idx
}

// Len returns the number of cells currently stored.
func (t *CodeDataTable[T]) Len() int {
	return len(t.items)
}

// InsertAt inserts val at idx, shifting cells at idx and above right by one.
// Panics if idx is out of range [0, len].
func (t *CodeDataTable[T]) InsertAt(idx int, val T) int {
	if idx < 0 || idx > len(t.items) {
		panic(fmt.Errorf("CodeDataTable: insert out of range: idx %d, len %d", idx, len(t.items)))
	}
	if len(t.items) >= t.maxSize {
		panic(fmt.Errorf("CodeDataTable: overflow at len %d, maxSize %d", len(t.items), t.maxSize))
	}
	t.items = append(t.items, val)
	copy(t.items[idx+1:], t.items[idx:len(t.items)-1])
	t.items[idx] = val
	return idx
}

// Truncate drops cells so len becomes n. Panics if n is out of range.
func (t *CodeDataTable[T]) Truncate(n int) {
	if n < 0 || n > len(t.items) {
		panic(fmt.Errorf("CodeDataTable: truncate out of range: n %d, len %d", n, len(t.items)))
	}
	t.items = t.items[:n]
}

// Clear removes all cells but retains capacity.
func (t *CodeDataTable[T]) Clear() {
	t.items = t.items[:0]
}

func CreateRegisterCodeTableType() {
	t := &Type{
		name: "CodeTable",
		getZero: func() any {
			return 0
		},
		checkCoerceType: func(value any) any {
			ct, ok := value.(*CodeTable)
			if !ok {
				panic(fmt.Errorf("CodeTable.checkCoerceType: value %T is not *CodeTable", value))
			}
			return ct
		},
	}
	t.registerItemsData("Int",
		func(inst Any, idx Any) any {
			return Int(inst.(*CodeTable).Read(int(idx.(Int))))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*CodeTable).Write(int(idx.(Int)), int(value.(Int)))
		},
		nil, nil, nil,
	)
	t.registerNativeMethod("append", func(f *RunFrame, ths *TypeObjPair) {
		item := GetTypeRegistrar().Get("Int").CheckCoerceType(f.RP.DataStack.Pop()).(Int)
		idx := ths.Obj.(*CodeTable).Append(int(item))
		f.RP.DataStack.Push(Int(idx))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

func CreateRegisterDataTableType() {
	t := &Type{
		name: "DataTable",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			dt, ok := value.(*DataTable)
			if !ok {
				panic(fmt.Errorf("DataTable.checkCoerceType: value %T is not *DataTable", value))
			}
			return dt
		},
	}
	t.registerItemsData("Any",
		func(inst Any, idx Any) any {
			return inst.(*DataTable).Read(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*DataTable).Write(int(idx.(Int)), value)
		},
		nil, nil, nil,
	)
	t.registerNativeMethod("append", func(f *RunFrame, ths *TypeObjPair) {
		item := f.RP.DataStack.Pop()
		idx := ths.Obj.(*DataTable).Append(item)
		f.RP.DataStack.Push(Int(idx))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterCodeTableType()
	CreateRegisterDataTableType()
}
