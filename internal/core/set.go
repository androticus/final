package core

import (
	//"fmt"
	"iter"
)

type Set[T comparable] struct {
	setType            *Type
	coerceSetFromAnyFn ValueCoerceSetFromAnyFn[T]
	items              map[T]bool
}
type AnySet = Set[any]
type StringSet = Set[string]

func NewAnySet() *AnySet {
	s := &AnySet{items: make(map[any]bool)}
	s.setType = anySetType
	s.coerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return s
}

func NewStringSet() *StringSet {
	s := &StringSet{items: make(map[string]bool)}
	s.setType = stringSetType
	s.coerceSetFromAnyFn = MakeStringListCoerceSetFn()
	return s
}

func (s *Set[T]) Put(item T) {
	s.items[s.coerceSetFromAnyFn(item)] = true
}

func (s *Set[T]) Has(item T) bool {
	return s.items[item]
}

func (s *Set[T]) Len() int {
	return len(s.items)
}

func (s *Set[T]) FinalType() *Type {
	return s.setType
}

// MapIterator implements a stateful pull-based iterator over a map.
type SetIterator[T comparable] struct {
	nextFunc func() (T, bool, bool) // The underlying pull function from Go's runtime
	currentK T
	hasValue bool
}

// NewMapIterator initializes the iterator for the given map.
func NewSetIterator[T comparable](m map[T]bool) *SetIterator[T] {
	// 1. Create a standard range-over-func sequence
	seq := func(yield func(T, bool) bool) {
		for k := range m {
			if !yield(k, true) {
				return
			}
		}
	}

	// 2. iter.Pull2 transforms the push sequence into a stateful pull function
	next, _ := iter.Pull2(seq)

	return &SetIterator[T]{
		nextFunc: next,
	}
}

// GetNext advances to the next item. Returns false if the collection is empty
// or if we have reached the end of the map.
func (it *SetIterator[T]) GetNext() bool {
	k, _, ok := it.nextFunc()
	if !ok {
		it.hasValue = false
		return false
	}

	it.currentK = k
	it.hasValue = true
	return true
}

// Current returns the current key and value.
// It panics if called before GetNext() or after GetNext() has returned false.
func (it *SetIterator[T]) Current() any {
	if !it.hasValue {
		panic("Current() called on undefined iterator state. Call GetNext() successfully first.")
	}
	return it.currentK
}

func (m *Set[T]) GetIter() Iterator {
	return NewSetIterator(m.items)
}

func registerSetMembers[T comparable](t *Type, itemTypeName string) {
	// ?{} only — Sets have no @{} / !{} (those take a value in addition to the key).
	t.registerItemsData(
		itemTypeName,
		nil, nil,
		nil, nil,
		func(inst Any, key Any) bool {
			s := inst.(*Set[T])
			return s.Has(s.coerceSetFromAnyFn(key))
		},
	)
	t.registerNativeMethod("put", func(f *RunFrame, ths *TypeObjPair) {
		s := ths.Obj.(*Set[T])
		s.Put(s.coerceSetFromAnyFn(f.RP.DataStack.Pop()))
	})
	t.registerNativeMethod("has", func(f *RunFrame, ths *TypeObjPair) {
		s := ths.Obj.(*Set[T])
		f.RP.DataStack.Push(Bool(s.Has(s.coerceSetFromAnyFn(f.RP.DataStack.Pop()))))
	})
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Set[T]).Len()) },
		nil,
	)
}

func CreateRegisterSetType() {
	t := &Type{
		name:            "Set",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AnySet]("Set"),
		parent:          protoSetType,
	}
	registerSetMembers[any](t, "Any")
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAnySet())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	anySetType = t
}

func CreateRegisterStringSetType() {
	t := &Type{
		name:            "StringSet",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StringSet]("StringSet"),
		parent:          protoSetType,
	}
	registerSetMembers[string](t, "String")
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStringSet())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	stringSetType = t
}

var (
	anySetType    *Type
	stringSetType *Type
)

func init() {
	CreateRegisterProtoSet()
	CreateRegisterSetType()
	CreateRegisterStringSetType()
}

// Validate Interfaces

var (
	_ Iterator = (*SetIterator[any])(nil)
	_ Iterable = (*Set[any])(nil)
	_ Typed    = (*Set[any])(nil)
)
