package core

import (
	"fmt"
	"iter"
)

// Map is a mutable key-value compound. Final-facing access uses any keys and values
// on the DS; Coerce* functors specialize storage per sealed Type (Map, StringMap, …).
type Map[K comparable, V any] struct {
	MapType                 *Type
	CoerceKeyFromAnyFn      ValueCoerceSetFromAnyFn[K]
	CoerceValueGetAnyFn     ValueCoerceGetAnyFn[V]
	CoerceValueSetFromAnyFn ValueCoerceSetFromAnyFn[V]
	items                   map[K]V
}

type AnyMap = Map[any, any]

type StringMap = Map[string, any]

func (m *Map[K, V]) Len() int {
	return len(m.items)
}

func (m *Map[K, V]) FinalType() *Type {
	return m.MapType
}

// Get coerces key, looks up storage, and widens the value to a DS cell.
// Missing keys panic (same posture as list index out of range).
func (m *Map[K, V]) Get(key any) any {
	k := m.CoerceKeyFromAnyFn(key)
	v, ok := m.items[k]
	if !ok {
		panic(fmt.Errorf("Map.Get: key %v not found", key))
	}
	return m.CoerceValueGetAnyFn(v)
}

func (m *Map[K, V]) Put(key, val any) {
	k := m.CoerceKeyFromAnyFn(key)
	m.items[k] = m.CoerceValueSetFromAnyFn(val)
}

func (m *Map[K, V]) Has(key any) Bool {
	k := m.CoerceKeyFromAnyFn(key)
	_, ok := m.items[k]
	return Bool(ok)
}

func (m *Map[K, V]) Delete(key any) {
	k := m.CoerceKeyFromAnyFn(key)
	delete(m.items, k)
}

func (m *Map[K, V]) Clear() {
	clear(m.items)
}

func NewAnyMap() *AnyMap {
	m := &AnyMap{items: make(map[any]any)}
	m.MapType = anyMapType
	m.CoerceKeyFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	m.CoerceValueGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	m.CoerceValueSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return m
}

func NewStringMap() *StringMap {
	m := &StringMap{items: make(map[string]any)}
	m.MapType = stringMapType
	m.CoerceKeyFromAnyFn = MakeStringListCoerceSetFn()
	m.CoerceValueGetAnyFn = MakeDefaultValueCoerceGetFn[any]()
	m.CoerceValueSetFromAnyFn = MakeDefaultValueCoerceSetFn[any]()
	return m
}

// NewStringMapFromGoMap creates a StringMap from a map[string]string.
// This is useful e.g. for creating a StringMap from a map of headers or query parameters.
func NewStringMapFromGoMap(m map[string]string) *StringMap {
	sm := NewStringMap()
	for k, v := range m {
		sm.Put(String(k), String(v))
	}
	return sm
}

// registerCommonMapMembers registers the common members for all Map types.
func registerCommonMapMembers[K comparable, V any](t *Type) {
	t.registerNativeMethod("clear", func(f *RunFrame, ths *TypeObjPair) {
		(ths.Obj.(*Map[K, V])).Clear()
	})
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Map[K, V]).Len()) },
		nil,
	)
}

func CreateRegisterMapType() {
	t := &Type{
		name:            "Map",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AnyMap]("Map"),
	}
	t.registerItemsData(
		"Any",
		nil, nil,
		func(inst Any, key Any) any {
			return inst.(*AnyMap).Get(key)
		},
		func(inst Any, key Any, val any) {
			inst.(*AnyMap).Put(key, val)
		},
		func(inst Any, key Any) bool {
			return inst.(*AnyMap).Has(key)
		},
	)
	t.registerNativeMethod("hasKey", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push((ths.Obj.(*AnyMap)).Has(f.RP.DataStack.Pop()))
	})
	t.registerNativeMethod("deleteKey", func(f *RunFrame, ths *TypeObjPair) {
		(ths.Obj.(*AnyMap)).Delete(f.RP.DataStack.Pop())
	})
	registerCommonMapMembers[any, any](t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAnyMap())
	})
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	tr.AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       "ProtoMap",
		TypeSlot:       &t.parent,
	})
	anyMapType = t
}

func CreateRegisterStringMapType() {
	t := &Type{
		name:            "StringMap",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StringMap]("StringMap"),
		parent:          anyMapType, // anyMapType is registered first
	}
	t.registerItemsData(
		"Any",
		nil, nil,
		func(inst Any, key Any /*String*/) any {
			return inst.(*StringMap).Get(key.(String))
		},
		func(inst Any, key Any /*String*/, val any) {
			inst.(*StringMap).Put(key.(String), val)
		},
		func(inst Any, key Any /*String*/) bool {
			return inst.(*StringMap).Has(key.(String))
		},
	)
	t.registerNativeMethod("hasKey", func(f *RunFrame, ths *TypeObjPair) {
		inst := ths.Obj.(*StringMap)
		f.RP.DataStack.Push(inst.Has(string(f.RP.DataStack.Pop().(String))))
	})
	t.registerNativeMethod("deleteKey", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*StringMap).Delete(f.RP.DataStack.Pop().(String))
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStringMap())
	})
	//registerCommonMapMembers[string, any](t)
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	stringMapType = t
}

var anyMapType *Type
var stringMapType *Type

// Iteration

type AnyKeyValue = KeyValue[any, any]

var keyValueType *Type

type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

func (kv *KeyValue[K, V]) FinalType() *Type {
	return keyValueType
}

func registerKeyValueMembers(t *Type) {
	t.registerProperty("key", "Any",
		func(inst Any) Any { return inst.(*KeyValue[any, any]).Key },
		nil,
	)
	t.registerProperty("value", "Any",
		func(inst Any) Any { return inst.(*KeyValue[any, any]).Value },
		nil,
	)
}

func CreateRegisterKeyValueType() {
	t := &Type{
		name:            "KeyValue",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[*AnyKeyValue]("KeyValue"),
	}
	registerKeyValueMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	keyValueType = t
}

// MapIterator implements a stateful pull-based iterator over a map.
type MapIterator[K comparable, V any] struct {
	nextFunc func() (K, V, bool) // The underlying pull function from Go's runtime
	currentK K
	currentV V
	hasValue bool
}

// NewMapIterator initializes the iterator for the given map.
func NewMapIterator[K comparable, V any](m map[K]V) *MapIterator[K, V] {
	// 1. Create a standard range-over-func sequence
	seq := func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}

	// 2. iter.Pull2 transforms the push sequence into a stateful pull function
	next, _ := iter.Pull2(seq)

	return &MapIterator[K, V]{
		nextFunc: next,
	}
}

// GetNext advances to the next item. Returns false if the collection is empty
// or if we have reached the end of the map.
func (it *MapIterator[K, V]) GetNext() bool {
	k, v, ok := it.nextFunc()
	if !ok {
		it.hasValue = false
		return false
	}

	it.currentK = k
	it.currentV = v
	it.hasValue = true
	return true
}

// Current returns the current key and value.
// It panics if called before GetNext() or after GetNext() has returned false.
func (it *MapIterator[K, V]) Current() any {
	if !it.hasValue {
		panic("Current() called on undefined iterator state. Call GetNext() successfully first.")
	}
	return &AnyKeyValue{
		Key:   it.currentK,
		Value: it.currentV,
	}
}

func (m *Map[K, V]) GetIter() Iterator {
	return NewMapIterator(m.items)
}

func init() {
	CreateRegisterKeyValueType()
	// note: these are registered in reverse order of dependency
	CreateRegisterMapType()
	CreateRegisterStringMapType()
}

// Validate Interfaces

var (
	_ Iterator = (*MapIterator[any, any])(nil)
	_ Iterable = (*Map[any, any])(nil)
	_ Typed    = (*KeyValue[any, any])(nil)
)
