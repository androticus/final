package core

import (
	"fmt"
)

// PropertyRef is a focused property accessor (Parent compound + metadata).
type PropertyRef struct {
	Parent *TypeObjPair
	PM     *PropertyMetadata
}

func (PropertyRef) FinalType() *Type { return GetTypeRegistrar().Get("PropertyRef") }

func (r *PropertyRef) Fetch() any {
	if r == nil || r.PM == nil || r.Parent == nil {
		panic("PropertyRef.Fetch: nil ref")
	}
	return r.PM.GetValue(r.Parent.Obj)
}

func (r *PropertyRef) Store(v any) {
	if r == nil || r.PM == nil || r.Parent == nil {
		panic("PropertyRef.Store: nil ref")
	}
	r.PM.SetValue(r.Parent.Obj, v)
}

// MethodRef is a focused method accessor (Parent compound + metadata).
type MethodRef struct {
	Parent *TypeObjPair
	MM     *MethodMetadata
}

func (MethodRef) FinalType() *Type { return GetTypeRegistrar().Get("MethodRef") }

// NewMethodRef binds obj's instance method named methodName.
// Panics if obj is untyped or the method is missing.
func NewMethodRef(obj any, methodName string) *MethodRef {
	if obj == nil {
		panic("NewMethodRef: obj is nil")
	}
	if methodName == "" {
		panic("NewMethodRef: methodName is empty")
	}
	ths := NewTypeObjPair(obj)
	mm := ths.Typ.GetMethodMetadata(methodName)
	if mm == nil {
		panic(fmt.Errorf("NewMethodRef: type %s has no #%s", ths.Typ.Name(), methodName))
	}
	return &MethodRef{Parent: ths, MM: mm}
}

// Invoke runs the bound method with Parent as the receiver.
// Stack args for the method must already be on the data stack.
func (r *MethodRef) Invoke(f *RunFrame) {
	if r == nil || r.MM == nil || r.Parent == nil {
		panic("MethodRef.Invoke: nil ref")
	}
	r.MM.invokeCall(f, r.Parent)
}

type DTSlotRef struct {
	dt       *DataTable
	dtIdx    int
	readonly bool
}

func (r *DTSlotRef) Fetch() any {
	return r.dt.Read(r.dtIdx)
}

func (r *DTSlotRef) Store(v any) {
	if r.readonly {
		panic(fmt.Errorf("DTRef: store not allowed"))
	}
	r.dt.Write(r.dtIdx, v)
}

// NewDTRef makes a DT ref from a DataTable and index; index is a base, not VirtualAddr.
// If the value is constant, set readonly to true.
func NewDTSlotRef(dt *DataTable, dtIdx int, readonly bool) ValueRef {
	return &DTSlotRef{
		dt:       dt,
		dtIdx:    dtIdx,
		readonly: readonly,
	}
}

func (r *DTSlotRef) FinalType() *Type {
	return dtSlotRefType
}

func CreateRegisterPropertyRefType() {
	t := &Type{
		name: "PropertyRef",
		getZero: func() any {
			panic("PropertyRef: no zero value")
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
}

func CreateRegisterMethodRefType() {
	t := &Type{
		name: "MethodRef",
		getZero: func() any {
			panic("MethodRef: no zero value")
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
}

func CreateRegisterDTSlotRefType() {
	t := &Type{
		name: "DTSlotRef",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	dtSlotRefType = t
}

var dtSlotRefType *Type

var _ ValueRef = (*PropertyRef)(nil)
var _ ValueRef = (*DTSlotRef)(nil)

func init() {
	CreateRegisterPropertyRefType()
	CreateRegisterMethodRefType()
	CreateRegisterDTSlotRefType()
}
