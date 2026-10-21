package core

import (
	"fmt"
)

// StaticObject is a fixed-schema instance whose field names live on a
// prototype DynamicObject; the instance holds only values.
// Schema is frozen at construction — no new fields.
type StaticObject struct {
	prototype *DynamicObject
	values    []any
}

// NewStaticObject allocates an instance for prototype, shallow-copying
// current prototype field values into values[].
func NewStaticObject(prototype *DynamicObject) *StaticObject {
	n := prototype.Len()
	values := make([]any, n)
	for i := 0; i < n; i++ {
		values[i] = prototype.ItemAt(i).Value
	}
	return &StaticObject{
		prototype: prototype,
		values:    values,
	}
}

func (o *StaticObject) GetPrototype() *DynamicObject {
	return o.prototype
}

func (o *StaticObject) Len() int {
	return len(o.values)
}

func (o *StaticObject) Get(name string) any {
	return o.values[o.prototype.fields.IndexOf(name)]
}

func (o *StaticObject) HasField(name string) Bool {
	return o.prototype.HasField(name)
}

// GetPropertyNames returns prototype field names in registration order.
func (o *StaticObject) GetPropertyNames() *StringList {
	return o.prototype.GetPropertyNames()
}

// Set updates an existing schema field. Panics if name is not on the prototype
// (fixed schema — unlike DynamicObject, cannot grow).
func (o *StaticObject) Set(name string, value any) {
	o.values[o.prototype.fields.IndexOf(name)] = value
}

func (o *StaticObject) FinalType() *Type {
	return staticObjectType
}

func getStaticObjectPropertyMetadataResolver(t *Type, name string) *PropertyMetadata {
	if t.GetMethodMetadata(name) != nil {
		return nil
	}
	return &PropertyMetadata{
		name: name,
		get: func(obj any) any {
			return obj.(*StaticObject).Get(string(String(name)))
		},
		set: func(obj any, value any) {
			obj.(*StaticObject).Set(string(String(name)), value)
		},
	}
}

func registerStaticObjectMembers(t *Type) {
	t.setPropertyMetadataOverride(getStaticObjectPropertyMetadataResolver)
	t.registerItemsData(
		"Any",
		func(inst Any, idx Any) any {
			o := inst.(*StaticObject)
			i := int(idx.(Int))
			if i < 0 || i >= len(o.values) {
				panic(fmt.Errorf("StaticObject @(): ordinal %d out of range [0,%d)", i, len(o.values)))
			}
			return o.values[i]
		},
		func(inst Any, idx Any, value Any) {
			o := inst.(*StaticObject)
			i := int(idx.(Int))
			if i < 0 || i >= len(o.values) {
				panic(fmt.Errorf("StaticObject !(): ordinal %d out of range [0,%d)", i, len(o.values)))
			}
			o.values[i] = value
		},
		func(inst Any, key Any) any {
			return inst.(*StaticObject).Get(string(key.(String)))
		},
		func(inst Any, key Any, val Any) {
			inst.(*StaticObject).Set(string(key.(String)), val)
		},
		func(inst Any, key Any) bool {
			return bool(inst.(*StaticObject).HasField(string(key.(String))))
		},
	)
}

// CreateRegisterStaticObjectType registers StaticObject under ProtoMap.
func CreateRegisterStaticObjectType() {
	t := &Type{
		name:            "StaticObject",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[StaticObject]("StaticObject"),
	}
	registerStaticObjectMembers(t)
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	tr.AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       "ProtoMap",
		TypeSlot:       &t.parent,
	})
	staticObjectType = t
}

func init() {
	CreateRegisterStaticObjectType()
}

var staticObjectType *Type

// StaticObjectIterator walks a snapshot of (name, value) pairs (insertion order).
type StaticObjectIterator struct {
	items []Item[any]
	index int
}

func (si *StaticObjectIterator) GetNext() bool {
	if si.index < (len(si.items) - 1) {
		si.index++
		return true
	}
	return false
}

func (si *StaticObjectIterator) Current() any {
	item := si.items[si.index]
	return &AnyKeyValue{
		Key:   String(item.Name),
		Value: item.Value,
	}
}

func (o *StaticObject) GetIter() Iterator {
	items := make([]Item[any], len(o.values))
	for i := range o.values {
		items[i] = Item[any]{
			Name:  o.prototype.ItemAt(i).Name,
			Value: o.values[i],
		}
	}
	return &StaticObjectIterator{
		items: items,
		index: -1,
	}
}

var (
	_ Typed    = (*StaticObject)(nil)
	_ Iterable = (*StaticObject)(nil)
	_ Iterator = (*StaticObjectIterator)(nil)
)
