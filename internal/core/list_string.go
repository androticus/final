package core

import "fmt"

// StringList stores []string. Bare Append/Put/InsertAt are strongly typed; only *FromAny
// paths run SetCoerceFromAnyFn to accept any[string] from Final stacks.
type StringList = List[string]

func MakeStringListCoerceGetFn() ValueCoerceGetAnyFn[string] {
	return func(typedValue string) any {
		return typedValue
	}
}
func MakeStringListCoerceSetFn() ValueCoerceSetFromAnyFn[string] {
	return func(anyValue any) string {
		switch v := anyValue.(type) {
		case string:
			return v
		default:
			panic(fmt.Errorf("StringList.setCoerceFromAnyFn: expected string, got %T", anyValue))
		}
	}
}

func NewStringList() *StringList {
	return NewStringListWithCapacity(0)
}

func InitNewStringList(l *StringList) {
	l.ListType = GetTypeRegistrar().Get("StringList")
	l.CoerceGetAnyFn = MakeStringListCoerceGetFn()
	l.CoerceSetFromAnyFn = MakeStringListCoerceSetFn()
}
func NewStringListFromSlice(strings []string) *StringList {
	l := &StringList{items: strings}
	InitNewStringList(l)
	return l
}

func NewStringListWithCapacity(capacity int) *StringList {
	if capacity < 0 {
		panic("StringList.NewStringListWithCapacity: capacity must be >= 0")
	}
	l := &StringList{items: make([]string, 0, capacity)}
	InitNewStringList(l)
	return l
}

func NewStringListWithLen(length int) *StringList {
	if length < 0 {
		panic("StringList.NewStringListWithLen: length must be >= 0")
	}
	l := &StringList{items: make([]string, length)}
	InitNewStringList(l)
	return l
}

func CreateRegisterStringListType() {
	t := &Type{
		name:            "StringList",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StringList]("StringList"),
		//parent:          anyListType,
		parent: listProtoType,
	}
	registerListMembers[string](t, "String")
	registerOrderedListSort[string](t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStringList())
	})
	// ( cap:i -- list:StringList )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewStringListWithCapacity(cap))
	})
	// ( len:i -- list:StringList )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewStringListWithLen(n))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

// all inits, in order, in list.go
