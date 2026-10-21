package core

import (
	"cmp"
	"fmt"
	"reflect"
)

type Int8List = List[int8]
type Int16List = List[int16]
type Int32List = List[int32]
type IntList = List[int] // Go int, not int64
type Uint16List = List[uint16]
type Uint32List = List[uint32]
type UintList = List[uint] // Go uint, not uint64

// ByteList is a strongly typed list of unsigned 8-bit values (0…255).
// Prefer this over a general List for byte buffers (String, streams, DB BLOBs).
type ByteList = List[uint8]

func MakeIntegralListCoerceGetFn[T any]() ValueCoerceGetAnyFn[T] {
	return func(v T) any {
		switch t := any(v).(type) {
		case int8:
			return Int(int(t))
		case int16:
			return Int(int(t))
		case int32:
			return Int(int(t))
		case int:
			return Int(t)
		case uint8:
			return Int(int(t))
		case uint16:
			return Int(int(t))
		case uint32:
			return Int(int(t))
		case uint:
			return Int(int(t))
		default:
			panic(fmt.Errorf("MakeNumericListCoerceGetFn: expected integral, got %T", v))
		}
	}
}

// MakeNumericListCoerceSetFn converts runtime Int/Uint cells into narrow/wide storage T.
func MakeIntegralListCoerceSetFn[T any]() ValueCoerceSetFromAnyFn[T] {
	var zero T
	targetType := reflect.TypeOf(zero)

	return func(anyValue any) T {
		switch v := anyValue.(type) {
		case Int:
			return reflect.ValueOf(int(v)).Convert(targetType).Interface().(T)
		case Uint:
			return reflect.ValueOf(uint(v)).Convert(targetType).Interface().(T)
		default:
			panic(fmt.Errorf("MakeNumericListCoerceSetFn: expected integral, got %T", anyValue))
		}
	}
}

func newIntegralList[T any](finalTypeName string, capacity int) *List[T] {
	if capacity < 0 {
		panic(fmt.Sprintf("%s: capacity must be >= 0", finalTypeName))
	}
	l := &List[T]{items: make([]T, 0, capacity)}
	l.ListType = GetTypeRegistrar().Get(finalTypeName)
	l.CoerceGetAnyFn = MakeIntegralListCoerceGetFn[T]()
	l.CoerceSetFromAnyFn = MakeIntegralListCoerceSetFn[T]()
	return l
}

func newIntegralListWithLen[T any](finalTypeName string, length int) *List[T] {
	if length < 0 {
		panic(fmt.Sprintf("%s: length must be >= 0", finalTypeName))
	}
	l := &List[T]{items: make([]T, length)}
	l.ListType = GetTypeRegistrar().Get(finalTypeName)
	l.CoerceGetAnyFn = MakeIntegralListCoerceGetFn[T]()
	l.CoerceSetFromAnyFn = MakeIntegralListCoerceSetFn[T]()
	return l
}

func registerIntegralListType[T cmp.Ordered](finalName, itemTypeName string, newWithCap, newWithLen func(int) *List[T]) {
	t := &Type{
		name:            finalName,
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[List[T]](finalName),
		parent:          listProtoType,
	}
	registerListMembers[T](t, itemTypeName)
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
	if finalName == "ByteList" {
		// ( bytes -- s ) UTF-8 decode via Go string([]byte) copy
		t.registerNativeMethod("asString", func(f *RunFrame, ths *TypeObjPair) {
			l := ths.Obj.(*ByteList)
			f.RP.DataStack.Push(String(string(l.items)))
		})
	}
	GetTypeRegistrar().RegisterCurrentType(t)
}

func NewInt8List() *Int8List     { return NewInt8ListWithCapacity(0) }
func NewInt16List() *Int16List   { return NewInt16ListWithCapacity(0) }
func NewInt32List() *Int32List   { return NewInt32ListWithCapacity(0) }
func NewIntList() *IntList       { return NewIntListWithCapacity(0) }
func NewByteList() *ByteList     { return NewByteListWithCapacity(0) }
func NewUint16List() *Uint16List { return NewUint16ListWithCapacity(0) }
func NewUint32List() *Uint32List { return NewUint32ListWithCapacity(0) }
func NewUintList() *UintList     { return NewUintListWithCapacity(0) }

func NewInt8ListWithCapacity(capacity int) *Int8List {
	return newIntegralList[int8]("Int8List", capacity)
}
func NewInt16ListWithCapacity(capacity int) *Int16List {
	return newIntegralList[int16]("Int16List", capacity)
}
func NewInt32ListWithCapacity(capacity int) *Int32List {
	return newIntegralList[int32]("Int32List", capacity)
}
func NewIntListWithCapacity(capacity int) *IntList {
	return newIntegralList[int]("IntList", capacity)
}
func NewByteListWithCapacity(capacity int) *ByteList {
	return newIntegralList[uint8]("ByteList", capacity)
}
func NewUint16ListWithCapacity(capacity int) *Uint16List {
	return newIntegralList[uint16]("Uint16List", capacity)
}
func NewUint32ListWithCapacity(capacity int) *Uint32List {
	return newIntegralList[uint32]("Uint32List", capacity)
}
func NewUintListWithCapacity(capacity int) *UintList {
	return newIntegralList[uint]("UintList", capacity)
}

func NewInt8ListWithLen(length int) *Int8List {
	return newIntegralListWithLen[int8]("Int8List", length)
}
func NewInt16ListWithLen(length int) *Int16List {
	return newIntegralListWithLen[int16]("Int16List", length)
}
func NewInt32ListWithLen(length int) *Int32List {
	return newIntegralListWithLen[int32]("Int32List", length)
}
func NewIntListWithLen(length int) *IntList {
	return newIntegralListWithLen[int]("IntList", length)
}
func NewByteListWithLen(length int) *ByteList {
	return newIntegralListWithLen[uint8]("ByteList", length)
}
func NewUint16ListWithLen(length int) *Uint16List {
	return newIntegralListWithLen[uint16]("Uint16List", length)
}
func NewUint32ListWithLen(length int) *Uint32List {
	return newIntegralListWithLen[uint32]("Uint32List", length)
}
func NewUintListWithLen(length int) *UintList {
	return newIntegralListWithLen[uint]("UintList", length)
}

// Create a ByteList from a slice of bytes.
// The list takes ownership of the bytes slice.
func NewByteListFromBytes(bytes []byte) *ByteList {
	// note: we have to create a list properly, to get the other list members
	// set, then we just replace the items with the bytes slice
	l := NewByteListWithCapacity(0)
	l.items = bytes
	return l
}

// NewByteListFromString copies the UTF-8 bytes of s into a new ByteList.
func NewByteListFromString(s string) *ByteList {
	b := make([]byte, len(s))
	copy(b, s)
	return NewByteListFromBytes(b)
}

// all inits, in order, in list.go
