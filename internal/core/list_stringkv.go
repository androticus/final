package core

import "fmt"

/*
StringKeyValue is a key-value pair with string keys and string values.
StringKeyValueList is a list of StringKeyValue.
This is a convenience type for contexts such as HttpReq headers
and query parameters, and HttpRes headers.
Note that the items are stored as values, not pointers; this is much
more efficient than pointers, and enables the caller to keep reusing
the same StringKeyValue instances.

Key access (Final !{} @{} ?{}) uses map-like semantics with first-match wins:
  !{}  upsert — append if key absent, else update first matching key
  @{}  get — first match, panic if missing (same as Map)
  ?{}  has — true if any pair has the key
  add  append-only (value key --); use for duplicate keys on the wire
*/

type StringKeyValue = KeyValue[string, string]

type StringKeyValueList struct {
	List[StringKeyValue]
}

func MakeStringKeyValueCoerceGetFn() ValueCoerceGetAnyFn[StringKeyValue] {
	return func(kv StringKeyValue) any {
		return kv
	}
}

func MakeStringKeyValueCoerceSetFn() ValueCoerceSetFromAnyFn[StringKeyValue] {
	return func(anyValue any) StringKeyValue {
		switch v := anyValue.(type) {
		case StringKeyValue:
			return v
		case *StringKeyValue:
			return *v
		default:
			panic(fmt.Errorf("StringKeyValueList.setCoerceFromAnyFn: expected StringKeyValue, got %T", anyValue))
		}
	}
}

func NewStringKeyValueList() *StringKeyValueList {
	return NewStringKeyValueListWithCapacity(0)
}

func NewStringKeyValueListWithCapacity(capacity int) *StringKeyValueList {
	if capacity < 0 {
		panic("StringKeyValueList.NewStringKeyValueListWithCapacity: capacity must be >= 0")
	}
	l := &StringKeyValueList{}
	l.ListType = GetTypeRegistrar().Get("StringKeyValueList")
	l.items = make([]StringKeyValue, 0, capacity)
	l.CoerceGetAnyFn = MakeStringKeyValueCoerceGetFn()
	l.CoerceSetFromAnyFn = MakeStringKeyValueCoerceSetFn()
	return l
}

func NewStringKeyValueListWithLen(length int) *StringKeyValueList {
	if length < 0 {
		panic("StringKeyValueList.NewStringKeyValueListWithLen: length must be >= 0")
	}
	l := &StringKeyValueList{}
	l.ListType = GetTypeRegistrar().Get("StringKeyValueList")
	l.items = make([]StringKeyValue, length)
	l.CoerceGetAnyFn = MakeStringKeyValueCoerceGetFn()
	l.CoerceSetFromAnyFn = MakeStringKeyValueCoerceSetFn()
	return l
}

func (l *StringKeyValueList) FinalType() *Type {
	return l.ListType
}

func (l *StringKeyValueList) Add(key, value string) {
	l.Append(StringKeyValue{Key: key, Value: value})
}

func (l *StringKeyValueList) AddFromAny(key, val any) {
	l.Add(coerceStringForSKV(key), coerceStringForSKV(val))
}

// Put upserts by key: updates the first matching pair, or appends if absent.
func (l *StringKeyValueList) Put(key, val any) {
	k := coerceStringForSKV(key)
	v := coerceStringForSKV(val)
	for i := range l.items {
		if l.items[i].Key == k {
			l.items[i].Value = v
			return
		}
	}
	l.Append(StringKeyValue{Key: k, Value: v})
}

// Get returns the value for the first matching key; panics if not found.
func (l *StringKeyValueList) Get(key any) String {
	k := coerceStringForSKV(key)
	for _, kv := range l.items {
		if kv.Key == k {
			return String(kv.Value)
		}
	}
	panic(fmt.Errorf("StringKeyValueList.Get: key %q not found", k))
}

// Has reports whether any pair has the given key.
func (l *StringKeyValueList) Has(key any) bool {
	k := coerceStringForSKV(key)
	for _, kv := range l.items {
		if kv.Key == k {
			return true
		}
	}
	return false
}

func coerceStringForSKV(v any) string {
	switch s := v.(type) {
	case String:
		return string(s)
	default:
		panic(fmt.Errorf("StringKeyValueList: expected String, got %T", v))
	}
}

func registerStringKeyValueListMembers(t *Type) {
	t.registerItemsData(
		"String",
		nil, nil,
		func(inst Any, key Any) any {
			return inst.(*StringKeyValueList).Get(key)
		},
		func(inst Any, key Any, val any) {
			inst.(*StringKeyValueList).Put(key, val)
		},
		func(inst Any, key Any) bool {
			return inst.(*StringKeyValueList).Has(key)
		},
	)
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		if ds.Len() < 2 {
			panic(fmt.Errorf("StringKeyValueList.add: need value and name on DS"))
		}
		key := ds.Pop()
		val := ds.Pop()
		ths.Obj.(*StringKeyValueList).AddFromAny(key, val)
	})
}

func CreateRegisterStringKeyValueListType() {
	t := &Type{
		name:            "StringKeyValueList",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StringKeyValueList]("StringKeyValueList"),
		parent:          listProtoType,
	}
	registerStringKeyValueListMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStringKeyValueList())
	})
	// ( cap:i -- list:StringKeyValueList )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, ths *TypeObjPair) {
		cap := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewStringKeyValueListWithCapacity(cap))
	})
	// ( len:i -- list:StringKeyValueList )
	t.registerNativeClassMethod("newWithLen", func(f *RunFrame, ths *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewStringKeyValueListWithLen(n))
	})
	tr :=GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	stringKeyValueListType = t
}

var stringKeyValueListType *Type

// all inits, in order, in list.go

var (
	_ Iterable = (*StringKeyValueList)(nil)
	_ Typed    = (*StringKeyValueList)(nil)
)
