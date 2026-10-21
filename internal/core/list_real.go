package core

import (
	"cmp"
	"fmt"
	"reflect"
)

// Float32List and FloatList store narrow/wide float storage. Bare Append/Put/InsertAt
// are strongly typed; *FromAny paths coerce Int, Uint, and Float runtime cells.
type Float32List = List[float32]
type FloatList = List[float64]

func MakeRealListCoerceGetFn[T any]() ValueCoerceGetAnyFn[T] {
	return func(v T) any {
		switch t := any(v).(type) {
		case float32:
			return Float(t)
		case float64:
			return Float(t)
		default:
			panic(fmt.Errorf("MakeRealListCoerceGetFn: expected float, got %T", v))
		}
	}
}

// MakeRealListCoerceSetFn converts runtime Int, Uint, and Float cells into storage T.
func MakeRealListCoerceSetFn[T any]() ValueCoerceSetFromAnyFn[T] {
	var zero T
	targetType := reflect.TypeOf(zero)

	return func(anyValue any) T {
		switch v := anyValue.(type) {
		case Int:
			return reflect.ValueOf(int(v)).Convert(targetType).Interface().(T)
		case Uint:
			return reflect.ValueOf(uint(v)).Convert(targetType).Interface().(T)
		case Float:
			return reflect.ValueOf(float64(v)).Convert(targetType).Interface().(T)
		default:
			panic(fmt.Errorf("MakeRealListCoerceSetFn: expected a number, got %T", anyValue))
		}
	}
}

func newRealList[T any](finalTypeName string, capacity int) *List[T] {
	if capacity < 0 {
		panic(fmt.Sprintf("%s: capacity must be >= 0", finalTypeName))
	}
	l := &List[T]{items: make([]T, 0, capacity)}
	l.ListType = GetTypeRegistrar().Get(finalTypeName)
	l.CoerceGetAnyFn = MakeRealListCoerceGetFn[T]()
	l.CoerceSetFromAnyFn = MakeRealListCoerceSetFn[T]()
	return l
}

func newRealListWithLen[T any](finalTypeName string, length int) *List[T] {
	if length < 0 {
		panic(fmt.Sprintf("%s: length must be >= 0", finalTypeName))
	}
	l := &List[T]{items: make([]T, length)}
	l.ListType = GetTypeRegistrar().Get(finalTypeName)
	l.CoerceGetAnyFn = MakeRealListCoerceGetFn[T]()
	l.CoerceSetFromAnyFn = MakeRealListCoerceSetFn[T]()
	return l
}

func registerRealListType[T cmp.Ordered](finalName string, newWithCap, newWithLen func(int) *List[T]) {
	t := &Type{
		name:            finalName,
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[List[T]](finalName),
		parent:          listProtoType,
	}
	registerListMembers[T](t, "Float")
	registerOrderedListSort[T](t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(newWithCap(0))
	})
	// ( cap:i -- list )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(newWithCap(cap))
	})
	// ( len:i -- list )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(newWithLen(n))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

func NewFloat32List() *Float32List { return NewFloat32ListWithCapacity(0) }
func NewFloatList() *FloatList     { return NewFloatListWithCapacity(0) }

func NewFloat32ListWithCapacity(capacity int) *Float32List {
	return newRealList[float32]("Float32List", capacity)
}
func NewFloatListWithCapacity(capacity int) *FloatList {
	return newRealList[float64]("FloatList", capacity)
}

func NewFloat32ListWithLen(length int) *Float32List {
	return newRealListWithLen[float32]("Float32List", length)
}
func NewFloatListWithLen(length int) *FloatList {
	return newRealListWithLen[float64]("FloatList", length)
}

// all inits, in order, in list.go
