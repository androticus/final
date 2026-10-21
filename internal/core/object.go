package core

import "slices"

type DynamicObject struct {
	fields *SmallFastAnyMap
}

func NewDynamicObject() *DynamicObject {
	return &DynamicObject{
		fields: NewSFM[any](4), // initial capacity for 4 fields
	}
}

// Clone creates a new DynamicObject with the same fields as the original
// NOTE: this is a shallow copy, so the fields are not cloned.
// It is intended for use in prototyped-based cloning from a default main object.
func (o *DynamicObject) Clone() *DynamicObject {
	fields := NewSFM[any](o.fields.Len())
	for _, item := range o.fields.items {
		fields.Add(item.Name, item.Value)
	}
	return &DynamicObject{
		fields: fields,
	}
}

func (o *DynamicObject) Len() int {
	return o.fields.Len()
}

// ItemAt returns a pointer to the item at the given index
// NOTE: you must NOT mutate the DynamicObject while holding a pointer to an item
func (o *DynamicObject) ItemAt(i int) *Item[any] {
	return o.fields.ItemAt(i)
}

func (o *DynamicObject) Get(name string) any {
	return o.fields.Get(name)
}

func (o *DynamicObject) HasField(name string) Bool {
	return o.fields.HasKey(name)
}

// GetPropertyNames returns field names in insertion order at call time.
func (o *DynamicObject) GetPropertyNames() *StringList {
	n := o.Len()
	sl := NewStringListWithCapacity(n)
	for i := 0; i < n; i++ {
		sl.Append(o.ItemAt(i).Name)
	}
	return sl
}

func (o *DynamicObject) Set(name string, value any) {
	o.fields.Set(name, value)
}

func (o *DynamicObject) FinalType() *Type {
	return dynamicObjectType
}

func getDynamicObjectPropertyMetadataResolver(t *Type, name string) *PropertyMetadata {
	// check if have a method with this name, and return nil
	// so the invoke will call the method
	if t.GetMethodMetadata(name) != nil {
		return nil
	}
	return &PropertyMetadata{
		name: name,
		get: func(obj any) any {
			return obj.(*DynamicObject).Get(string(String(name)))
		},
		set: func(obj any, value any) {
			obj.(*DynamicObject).Set(string(String(name)), value)
		},
	}
}

// registerDynamicTypeMembers wires virtual instance field access via the embedded StringMap.
// DynamicType has no static property table — only named item get/set (see 121-final-native-objects.md).
func registerDynamicTypeMembers(t *Type) {
	t.setPropertyMetadataOverride(getDynamicObjectPropertyMetadataResolver)
	t.registerItemsData(
		"Any",
		func(inst Any, idx Any) any {
			return inst.(*DynamicObject).ItemAt(int(idx.(Int))).Value
		},
		func(inst Any, idx Any, value Any) {
			inst.(*DynamicObject).ItemAt(int(idx.(Int))).Value = value
		},
		func(inst Any, key Any) any {
			return inst.(*DynamicObject).Get(string(key.(String)))
		},
		func(inst Any, key Any, val Any) {
			inst.(*DynamicObject).Set(string(key.(String)), val)
		},
		func(inst Any, key Any) bool {
			return inst.(*DynamicObject).HasField(string(key.(String)))
		},
	)
}

// CreateRegisterDynamicType registers the shared DynamicType metadata for all DynamicObject instances.
func CreateRegisterDynamicType() {
	t := &Type{
		name:            "DynamicObject",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[DynamicObject]("DynamicType"),
	}
	registerDynamicTypeMembers(t)
	t.registerNativeMethod("clone", func(f *RunFrame, ths *TypeObjPair) {
		obj := ths.Obj.(*DynamicObject)
		f.RP.DataStack.Push(obj.Clone())
	})
	t.registerNativeMethod("newStatic", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewStaticObject(ths.Obj.(*DynamicObject)))
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewDynamicObject())
	})
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	tr.AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       "ProtoMap",
		TypeSlot:       &t.parent,
	})
	dynamicObjectType = t
}

func init() {
	CreateRegisterDynamicType()
	CreateRegisterAtomicDynamicObjectType()
}

var dynamicObjectType *Type // for caching

// DynamicObjectIterator walks a snapshot of field entries (insertion order).
// Safe vs later Set of new keys / value updates on the live object.
type DynamicObjectIterator struct {
	items []Item[any]
	index int
}

func (di *DynamicObjectIterator) GetNext() bool {
	if di.index < (len(di.items) - 1) {
		di.index++
		return true
	}
	return false
}

func (di *DynamicObjectIterator) Current() any {
	item := di.items[di.index]
	return &AnyKeyValue{
		Key:   String(item.Name),
		Value: item.Value,
	}
}

func (o *DynamicObject) GetIter() Iterator {
	return &DynamicObjectIterator{
		items: slices.Clone(o.fields.items),
		index: -1,
	}
}

var (
	_ Typed    = (*DynamicObject)(nil)
	_ Iterable = (*DynamicObject)(nil)
	_ Iterator = (*DynamicObjectIterator)(nil)
)
