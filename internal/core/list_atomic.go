package core

import (
	"cmp"
	"sync"
)

// atomicList is a mutex-guarded List[T]. Go callers use shadowed methods (lock then
// delegate). Final natives register against *AtomicAnyList / *AtomicStringList and
// invoke those same shadowed methods.
type atomicList[T any] struct {
	mu sync.Mutex
	*List[T]
}

// AtomicAnyList is Final "AtomicList" (threadsafe List of Any).
type AtomicAnyList = atomicList[any]

// AtomicStringList is Final "AtomicStringList" (threadsafe StringList).
type AtomicStringList = atomicList[string]

func NewAtomicAnyList() *AtomicAnyList {
	return NewAtomicAnyListWithCapacity(0)
}

func NewAtomicAnyListWithCapacity(capacity int) *AtomicAnyList {
	if capacity < 0 {
		panic("AtomicList.NewAtomicAnyListWithCapacity: capacity must be >= 0")
	}
	inner := &List[any]{items: make([]any, 0, capacity)}
	inner.ListType = GetTypeRegistrar().Get("AtomicList")
	inner.CoerceGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	inner.CoerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return &AtomicAnyList{List: inner}
}

func NewAtomicAnyListWithLen(length int) *AtomicAnyList {
	if length < 0 {
		panic("AtomicList.NewAtomicAnyListWithLen: length must be >= 0")
	}
	inner := &List[any]{items: make([]any, length)}
	inner.ListType = GetTypeRegistrar().Get("AtomicList")
	inner.CoerceGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	inner.CoerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return &AtomicAnyList{List: inner}
}

func NewAtomicStringList() *AtomicStringList {
	return NewAtomicStringListWithCapacity(0)
}

func NewAtomicStringListWithCapacity(capacity int) *AtomicStringList {
	if capacity < 0 {
		panic("AtomicStringList.NewAtomicStringListWithCapacity: capacity must be >= 0")
	}
	inner := &List[string]{items: make([]string, 0, capacity)}
	inner.ListType = GetTypeRegistrar().Get("AtomicStringList")
	inner.CoerceGetAnyFn = MakeStringListCoerceGetFn()
	inner.CoerceSetFromAnyFn = MakeStringListCoerceSetFn()
	return &AtomicStringList{List: inner}
}

func NewAtomicStringListWithLen(length int) *AtomicStringList {
	if length < 0 {
		panic("AtomicStringList.NewAtomicStringListWithLen: length must be >= 0")
	}
	inner := &List[string]{items: make([]string, length)}
	inner.ListType = GetTypeRegistrar().Get("AtomicStringList")
	inner.CoerceGetAnyFn = MakeStringListCoerceGetFn()
	inner.CoerceSetFromAnyFn = MakeStringListCoerceSetFn()
	return &AtomicStringList{List: inner}
}

func (l *atomicList[T]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.Len()
}

func (l *atomicList[T]) At(i int) T {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.At(i)
}

func (l *atomicList[T]) AtAsAny(i int) any {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.AtAsAny(i)
}

func (l *atomicList[T]) Put(i int, v T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.Put(i, v)
}

func (l *atomicList[T]) PutFromAny(i int, v any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.PutFromAny(i, v)
}

func (l *atomicList[T]) Append(v T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.Append(v)
}

func (l *atomicList[T]) AppendFromAny(v any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.AppendFromAny(v)
}

func (l *atomicList[T]) DeleteAt(i int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.DeleteAt(i)
}

func (l *atomicList[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.Clear()
}

func (l *atomicList[T]) InsertAt(i int, v T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.InsertAt(i, v)
}

func (l *atomicList[T]) InsertAtFromAny(i int, v any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.InsertAtFromAny(i, v)
}

func (l *atomicList[T]) HasItemFromAny(v any) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.HasItemFromAny(v)
}

func (l *atomicList[T]) DeleteItemFromAny(v any) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.DeleteItemFromAny(v)
}

func (l *atomicList[T]) Enqueue(v T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.Enqueue(v)
}

func (l *atomicList[T]) Dequeue() T {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.Dequeue()
}

func (l *atomicList[T]) Peek() T {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.Peek()
}

func (l *atomicList[T]) Shuffle() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.List.Shuffle()
}

// GetIter snapshots items under the mutex (List.GetIter clones); foreach is stable.
func (l *atomicList[T]) GetIter() Iterator {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.List.GetIter()
}

func sortAtomicOrderedList[T cmp.Ordered](l *atomicList[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	sortOrderedList(l.List)
}

func registerAtomicListMembers[T any](t *Type, itemTypeName string, cast func(Any) *atomicList[T]) {
	t.registerItemsData(itemTypeName,
		func(inst Any, idx Any) any {
			return cast(inst).AtAsAny(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			cast(inst).PutFromAny(int(idx.(Int)), value)
		},
		nil, nil, nil,
	)
	t.registerNativeMethod("append", func(f *RunFrame, ths *TypeObjPair) {
		cast(ths.Obj).AppendFromAny(f.RP.DataStack.Pop())
	})
	t.registerNativeMethod("insertAt", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		i := int(popDSInt(ds))
		v := ds.Pop()
		cast(ths.Obj).InsertAtFromAny(i, v)
	})
	t.registerNativeMethod("hasItem", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Bool(cast(ths.Obj).HasItemFromAny(f.RP.DataStack.Pop())))
	})
	t.registerNativeMethod("deleteItem", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Bool(cast(ths.Obj).DeleteItemFromAny(f.RP.DataStack.Pop())))
	})
	t.registerNativeMethod("deleteAt", func(f *RunFrame, ths *TypeObjPair) {
		cast(ths.Obj).DeleteAt(int(popDSInt(f.RP.DataStack)))
	})
	t.registerNativeMethod("clear", func(f *RunFrame, ths *TypeObjPair) {
		cast(ths.Obj).Clear()
	})
	t.registerNativeMethod("enqueue", func(f *RunFrame, ths *TypeObjPair) {
		l := cast(ths.Obj)
		l.Enqueue(l.CoerceSetFromAnyFn(f.RP.DataStack.Pop()))
	})
	t.registerNativeMethod("dequeue", func(f *RunFrame, ths *TypeObjPair) {
		l := cast(ths.Obj)
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Dequeue()))
	})
	t.registerNativeMethod("push", func(f *RunFrame, ths *TypeObjPair) {
		cast(ths.Obj).AppendFromAny(f.RP.DataStack.Pop())
	})
	t.registerNativeMethod("pop", func(f *RunFrame, ths *TypeObjPair) {
		l := cast(ths.Obj)
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Dequeue()))
	})
	t.registerNativeMethod("peek", func(f *RunFrame, ths *TypeObjPair) {
		l := cast(ths.Obj)
		f.RP.DataStack.Push(l.CoerceGetAnyFn(l.Peek()))
	})
	t.registerNativeMethod("depth", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Int(cast(ths.Obj).Len()))
	})
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(cast(inst).Len()) },
		nil,
	)
	t.registerNativeMethod("shuffle", func(f *RunFrame, ths *TypeObjPair) {
		cast(ths.Obj).Shuffle()
	})
}

func CreateRegisterAtomicAnyListType() {
	t := &Type{
		name:            "AtomicList",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AtomicAnyList]("AtomicList"),
		parent:          anyListType,
	}
	registerAtomicListMembers[any](t, "Any", func(inst Any) *atomicList[any] {
		return inst.(*AtomicAnyList)
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAtomicAnyList())
	})
	// ( cap:i -- list:AtomicList )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAtomicAnyListWithCapacity(cap))
	})
	// ( len:i -- list:AtomicList )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAtomicAnyListWithLen(n))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	atomicAnyListType = t
}

func CreateRegisterAtomicStringListType() {
	t := &Type{
		name:            "AtomicStringList",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AtomicStringList]("AtomicStringList"),
		parent:          GetTypeRegistrar().Get("StringList"),
	}
	registerAtomicListMembers[string](t, "String", func(inst Any) *atomicList[string] {
		return inst.(*AtomicStringList)
	})
	t.registerNativeMethod("sort", func(f *RunFrame, ths *TypeObjPair) {
		sortAtomicOrderedList(ths.Obj.(*AtomicStringList))
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAtomicStringList())
	})
	// ( cap:i -- list:AtomicStringList )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAtomicStringListWithCapacity(cap))
	})
	// ( len:i -- list:AtomicStringList )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewAtomicStringListWithLen(n))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	atomicStringListType = t
}

var (
	atomicAnyListType    *Type
	atomicStringListType *Type
)

var (
	_ Iterable = (*AtomicAnyList)(nil)
	_ Iterable = (*AtomicStringList)(nil)
	_ Typed    = (*AtomicAnyList)(nil)
	_ Typed    = (*AtomicStringList)(nil)
)
