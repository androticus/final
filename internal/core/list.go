package core

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
)

type ValueCoerceSetFromAnyFn[T any] func(anyValue any) T
type ValueCoerceGetAnyFn[T any] func(typedValue T) any

// List is a mutable array-like compound. The struct wraps the slice header so
// append can replace items without breaking handles (see
// docs/003-design-notes/_archived/100-list-compound.md).
type List[T any] struct {
	ListType           *Type
	CoerceGetAnyFn     ValueCoerceGetAnyFn[T]
	CoerceSetFromAnyFn ValueCoerceSetFromAnyFn[T]
	items              []T
}

func MakeDefaultValueCoerceGetFn[T any]() ValueCoerceGetAnyFn[T] {
	return func(typedValue T) any {
		return typedValue
	}
}
func MakeDefaultValueCoerceSetFn[T any]() ValueCoerceSetFromAnyFn[T] {
	return func(anyValue any) T {
		if anyValue == nil {
			var zero T
			return zero
		}
		return anyValue.(T)
	}
}

// AnyList is the generic Final List (Cell / any elements).
// List[T] exposes two write paths: bare names (Append, Put, InsertAt) take T directly with no
// runtime cell check; *FromAny methods coerce via SetCoerceFromAnyFn (Final any[string] etc.).
type AnyList = List[any]

func NewAnyList() *List[any] {
	return NewAnyListWithCapacity(0)
}

func NewAnyListWithCapacity(capacity int) *List[any] {
	if capacity < 0 {
		panic("List.NewAnyListWithCapacity: capacity must be >= 0")
	}
	l := &List[any]{items: make([]any, 0, capacity)}
	l.ListType = GetTypeRegistrar().Get("List")
	l.CoerceGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	l.CoerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return l
}

func NewAnyListWithLen(length int) *List[any] {
	if length < 0 {
		panic("List.NewAnyListWithLen: length must be >= 0")
	}
	l := &List[any]{items: make([]any, length)}
	l.ListType = GetTypeRegistrar().Get("List")
	l.CoerceGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	l.CoerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return l
}

func (l *List[T]) Len() int {
	return len(l.items)
}

func (l *List[T]) At(i int) T {
	l.checkIndexGet(i)
	return l.items[i]
}

func (l *List[T]) AtAsAny(i int) any {
	l.checkIndexGet(i)
	return l.CoerceGetAnyFn(l.items[i])
}

func (l *List[T]) Put(i int, v T) {
	l.checkIndexGet(i)
	l.items[i] = v
}

// PutFromAny coerces one any operand via SetCoerceFromAnyFn (e.g. any[string] for StringList).
func (l *List[T]) PutFromAny(i int, v any) {
	l.checkIndexGet(i)
	l.items[i] = l.CoerceSetFromAnyFn(v)
}

func (l *List[T]) Append(v T) {
	l.items = append(l.items, v)
}

// AppendFromAny coerces one any operand via SetCoerceFromAnyFn.
func (l *List[T]) AppendFromAny(v any) {
	l.items = append(l.items, l.CoerceSetFromAnyFn(v))
}

func (l *List[T]) DeleteAt(i int) {
	n := len(l.items)
	if i < 0 || i >= n {
		panic(fmt.Errorf("List.DeleteAt: index %d out of range for len %d", i, n))
	}
	copy(l.items[i:], l.items[i+1:])
	l.items = l.items[:n-1]
}

func (l *List[T]) Clear() {
	l.items = l.items[:0]
}

func (l *List[T]) InsertAt(i int, v T) {
	if i == len(l.items) {
		l.Append(v)
		return
	}
	l.checkIndexGet(i)
	l.items = slices.Insert(l.items, i, v)
}

// InsertAtFromAny coerces one any operand via SetCoerceFromAnyFn.
func (l *List[T]) InsertAtFromAny(i int, v any) {
	if i == len(l.items) {
		l.AppendFromAny(v)
		return
	}
	l.checkIndexGet(i)
	l.items = slices.Insert(l.items, i, l.CoerceSetFromAnyFn(v))
}

func (l *List[T]) HasItemFromAny(v any) bool {
	for i := 0; i < len(l.items); i++ {
		if l.AtAsAny(i) == v {
			return true
		}
	}
	return false
}

// DeleteItemFromAny removes the first element for which AtAsAny(i) == v
// (same Go == as HasItemFromAny). Returns false if absent; does not panic.
func (l *List[T]) DeleteItemFromAny(v any) bool {
	for i := 0; i < len(l.items); i++ {
		if l.AtAsAny(i) == v {
			l.DeleteAt(i)
			return true
		}
	}
	return false
}

// FinalType returns ListType for the generic List shim (v1: *List[any] only).
func (l *List[T]) FinalType() *Type {
	return l.ListType
}

func (l *List[T]) checkIndexGet(i int) {
	n := len(l.items)
	if i < 0 || i >= n {
		panic(fmt.Errorf("List: index %d out of range for len %d", i, n))
	}
}

func popDSInt(ds *Stack[any]) Int {
	v := ds.Pop()
	i, ok := v.(Int)
	if !ok {
		panic(fmt.Errorf("List: expected Int index, got %T", v))
	}
	return i
}

// Queue API
func (l *List[T]) Enqueue(v T) {
	l.InsertAt(0, v)
}

func (l *List[T]) Dequeue() T {
	if len(l.items) == 0 {
		panic(fmt.Errorf("List.Dequeue: list is empty"))
	}
	itemIdx := len(l.items) - 1
	item := l.items[itemIdx]
	l.DeleteAt(itemIdx)
	return item
}

// Stack API
// Push is the same as Append
// Pop is the same as Dequeue
// Depth is the same as Len

// Peek returns the top item without removing it
func (l *List[T]) Peek() T {
	if len(l.items) == 0 {
		panic(fmt.Errorf("List.Peek: list is empty"))
	}
	return l.items[len(l.items)-1]
}

func (l *List[T]) swapAt(i, j int) {
	l.items[i], l.items[j] = l.items[j], l.items[i]
}

// Shuffle randomizes element order in place (math/rand global source).
func (l *List[T]) Shuffle() {
	rand.Shuffle(len(l.items), l.swapAt)
}

func sortOrderedList[T cmp.Ordered](l *List[T]) {
	slices.Sort(l.items)
}

func registerListShuffle[T any](t *Type) {
	t.registerNativeMethod("shuffle", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*List[T]).Shuffle()
	})
}

func registerOrderedListSort[T cmp.Ordered](t *Type) {
	t.registerNativeMethod("sort", func(f *RunFrame, ths *TypeObjPair) {
		sortOrderedList(ths.Obj.(*List[T]))
	})
}

func registerListMembers[T any](t *Type, itemTypeName string) {
	t.registerItemsData(itemTypeName,
		func(inst Any, idx Any) any {
			return inst.(*List[T]).AtAsAny(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*List[T]).PutFromAny(int(idx.(Int)), value)
		},
		nil, nil, nil,
	)
	t.registerNativeMethod("append", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*List[T]).AppendFromAny(f.RP.DataStack.Pop())
	})

	t.registerNativeMethod("insertAt", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		i := int(popDSInt(ds))
		v := ds.Pop()
		ths.Obj.(*List[T]).InsertAtFromAny(i, v)
	})
	t.registerNativeMethod("hasItem", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Bool(ths.Obj.(*List[T]).HasItemFromAny(f.RP.DataStack.Pop())))
	})
	t.registerNativeMethod("deleteItem", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Bool(ths.Obj.(*List[T]).DeleteItemFromAny(f.RP.DataStack.Pop())))
	})

	t.registerNativeMethod("deleteAt", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*List[T]).DeleteAt(int(popDSInt(f.RP.DataStack)))
	})
	t.registerNativeMethod("clear", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*List[T]).Clear()
	})
	// Queue API
	t.registerNativeMethod("enqueue", func(f *RunFrame, ths *TypeObjPair) {
		l := ths.Obj.(*List[T])
		l.Enqueue(l.CoerceSetFromAnyFn(f.RP.DataStack.Pop()))
	})
	t.registerNativeMethod("dequeue", func(f *RunFrame, ths *TypeObjPair) {
		l := ths.Obj.(*List[T])
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Dequeue()))
	})
	// Stack API
	t.registerNativeMethod("push", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*List[T]).AppendFromAny(f.RP.DataStack.Pop())
	})
	t.registerNativeMethod("pop", func(f *RunFrame, ths *TypeObjPair) {
		l := ths.Obj.(*List[T])
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Dequeue()))
	})
	t.registerNativeMethod("peek", func(f *RunFrame, ths *TypeObjPair) {
		l := ths.Obj.(*List[T])
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Peek()))
	})
	t.registerNativeMethod("depth", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Int(ths.Obj.(*List[T]).Len()))
	})

	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*List[T]).Len()) },
		nil,
	)
	registerListShuffle[T](t)
}

type ListIterator[T any] struct {
	items  []T
	coerce ValueCoerceGetAnyFn[T]
	index  int
}

func (li *ListIterator[T]) GetNext() bool {
	if li.index < (len(li.items) - 1) {
		li.index++
		return true
	}
	return false
}

func (li *ListIterator[T]) Current() any {
	return li.coerce(li.items[li.index])
}

func (l *List[T]) GetIter() Iterator {
	return &ListIterator[T]{
		items:  slices.Clone(l.items),
		coerce: l.CoerceGetAnyFn,
		index:  -1,
	}
}

func CreateRegisterListType() {
	t := &Type{
		name:            "List",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AnyList]("List"),
	}
	registerListMembers[any](t, "Any")
	t.RegisterClassProperty("empty", "List",
		func(_ any) any {
			return GetEmptyAnyList()
		},
		nil,
	)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAnyList())
	})
	// ( cap:i -- list:List )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAnyListWithCapacity(cap))
	})
	// ( len:i -- list:List )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAnyListWithLen(n))
	})
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	tr.AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       "ProtoList",
		TypeSlot:       &t.parent,
	})
	anyListType = t
}

func GetEmptyAnyList() *AnyList {
	if emptyAnyList == nil {
		emptyAnyList = NewAnyList()
	}
	return emptyAnyList
}

var anyListType *Type
var emptyAnyList *AnyList

func init() {
	CreateRegisterListProto()
	CreateRegisterListType()

	CreateRegisterStringListType()

	CreateRegisterAtomicAnyListType()
	CreateRegisterAtomicStringListType()

	registerIntegralListType[int8]("Int8List", "Int", NewInt8ListWithCapacity, NewInt8ListWithLen)
	registerIntegralListType[int16]("Int16List", "Int", NewInt16ListWithCapacity, NewInt16ListWithLen)
	registerIntegralListType[int32]("Int32List", "Int", NewInt32ListWithCapacity, NewInt32ListWithLen)
	registerIntegralListType[int]("IntList", "Int", NewIntListWithCapacity, NewIntListWithLen)
	registerIntegralListType[uint8]("ByteList", "Int", NewByteListWithCapacity, NewByteListWithLen)
	registerIntegralListType[uint16]("Uint16List", "Int", NewUint16ListWithCapacity, NewUint16ListWithLen)
	registerIntegralListType[uint32]("Uint32List", "Int", NewUint32ListWithCapacity, NewUint32ListWithLen)
	registerIntegralListType[uint]("UintList", "Int", NewUintListWithCapacity, NewUintListWithLen)

	registerRealListType[float32]("Float32List", NewFloat32ListWithCapacity, NewFloat32ListWithLen)
	registerRealListType[float64]("FloatList", NewFloatListWithCapacity, NewFloatListWithLen)

	CreateRegisterStringKeyValueListType()
}

// Validate Interfaces

var (
	_ Iterator = (*ListIterator[any])(nil)
	_ Iterable = (*List[any])(nil)
)
