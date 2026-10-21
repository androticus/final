package core

import "fmt"

type TypedObject struct {
	typ    *Type
	fields []any
}

func (o *TypedObject) FinalType() *Type { return o.typ }

func (o *TypedObject) GetFieldByOrdinal(i int) any {
	return o.fields[i]
}

func (o *TypedObject) SetFieldByOrdinal(i int, v any) {
	o.fields[i] = v
}

// GetFieldByName returns the stored field value for name (schema field).
// Go-host helper for schema fields (e.g. based TypedObject mint wiring "base").
func (o *TypedObject) GetFieldByName(name string) any {
	pm := o.typ.GetPropertyMetadata(name)
	if pm == nil || pm.get == nil {
		panic(fmt.Errorf("TypedObject @{}: unknown name %q", name))
	}
	return pm.get(o)
}

// SetFieldByName stores v into the named schema field (with type coerce).
// Same Go-host role as GetFieldByName.
func (o *TypedObject) SetFieldByName(name string, v any) {
	pm := o.typ.GetPropertyMetadata(name)
	if pm == nil || pm.set == nil {
		panic(fmt.Errorf("TypedObject !{}: unknown name %q", name))
	}
	pm.set(o, v)
}

// GetPropertyNames returns the sealed type's property names in registration order.
func (o *TypedObject) GetPropertyNames() *StringList {
	return o.typ.GetPropertyNames()
}

type TypedObjectIterator struct {
	obj   *TypedObject
	index int
}

func (ti *TypedObjectIterator) GetNext() bool {
	if ti.index < len(ti.obj.fields)-1 {
		ti.index++
		return true
	}
	return false
}

func (ti *TypedObjectIterator) Current() any {
	return &AnyKeyValue{
		Key:   String(ti.obj.typ.propertyMetadata.ItemAt(ti.index).Name),
		Value: ti.obj.fields[ti.index],
	}
}

func (o *TypedObject) GetIter() Iterator {
	return &TypedObjectIterator{obj: o, index: -1}
}

// mintTypedObjectFieldGet returns a PropertyGetFn with ordinal baked in.
func mintTypedObjectFieldGet(ordinal int) PropertyGetFn {
	return func(inst Any) Any {
		return inst.(*TypedObject).fields[ordinal]
	}
}

// mintTypedObjectFieldSet returns a PropertySetFn with ordinal baked in.
// Store coerces via the field type's checkCoerceType before writing the slot.
func mintTypedObjectFieldSet(ordinal int, fieldType *Type) PropertySetFn {
	return func(inst Any, value Any) {
		o := inst.(*TypedObject)
		o.fields[ordinal] = fieldType.CheckCoerceType(value)
	}
}

func mintTypedObjectItemsGetByKey(t *Type) GetByKeyFn {
	return func(inst Any, key Any) any {
		name := string(key.(String))
		pm := t.GetPropertyMetadata(name)
		if pm == nil || pm.get == nil {
			panic(fmt.Errorf("TypedObject @{}: unknown name %q", name))
		}
		return pm.get(inst)
	}
}

func mintTypedObjectItemsSetByKey(t *Type) SetByKeyFn {
	return func(inst Any, key Any, val Any) {
		name := string(key.(String))
		pm := t.GetPropertyMetadata(name)
		if pm == nil || pm.set == nil {
			panic(fmt.Errorf("TypedObject !{}: unknown name %q", name))
		}
		pm.set(inst, val)
	}
}

func mintTypedObjectItemsHasByKey(t *Type) HasByKeyFn {
	return func(inst Any, key Any) bool {
		return t.GetPropertyMetadata(string(key.(String))) != nil
	}
}

func mintTypedObjectItemsGetByOrdinal(_ *Type) GetByOrdinalFn {
	return func(inst Any, idx Any) any {
		o := inst.(*TypedObject)
		i := int(idx.(Int))
		if i < 0 || i >= len(o.fields) {
			panic(fmt.Errorf("TypedObject @(): ordinal %d out of range [0,%d)", i, len(o.fields)))
		}
		return o.fields[i]
	}
}

func mintTypedObjectItemsSetByOrdinal(t *Type) SetByOrdinalFn {
	return func(inst Any, idx Any, value Any) {
		o := inst.(*TypedObject)
		i := int(idx.(Int))
		if i < 0 || i >= len(o.fields) {
			panic(fmt.Errorf("TypedObject !(): ordinal %d out of range [0,%d)", i, len(o.fields)))
		}
		pm := t.propertyMetadata.ItemAt(i).Value
		o.fields[i] = pm._type.CheckCoerceType(value)
	}
}

func registerTypedObjectItemsMetadata(t *Type) {
	t.registerItemsData(
		"Any",
		mintTypedObjectItemsGetByOrdinal(t),
		mintTypedObjectItemsSetByOrdinal(t),
		mintTypedObjectItemsGetByKey(t),
		mintTypedObjectItemsSetByKey(t),
		mintTypedObjectItemsHasByKey(t),
	)
}

// Factory methods

// NewTypeDef returns an empty TypeDef shell: name, fields, and constants DynamicObjects.
// In Final, each fields entry must be a live catalog Type (e.g. String), not a type name
// string — see docs/003-design-notes/_archived/121-final-native-objects.md.
// Each constants entry is a value (.name!) sealed as a readonly class const (Any).
func NewTypeDef() *DynamicObject {
	do := NewDynamicObject()
	do.Set("name", String(""))
	do.Set("fields", NewDynamicObject())
	do.Set("constants", NewDynamicObject())
	return do
}

func MakeNewTypedObject(t *Type) *TypedObject {
	propCount := t.propertyMetadata.Len()
	to := &TypedObject{
		typ:    t,
		fields: make([]any, propCount),
	}
	for i := 0; i < propCount; i++ {
		pm := t.propertyMetadata.ItemAt(i).Value
		to.fields[i] = pm._type.getZero()
	}
	return to
}

const typedObjectBaseFieldName = "base"

// FinishNewTypedObject wires a based TypedObject: calls baseType._newWithDelegate
// with the child as del, then stores the result in schema field "base".
// No-op when t.baseType is nil. ProtoTypedObject.new always calls this after mint.
func FinishNewTypedObject(f *RunFrame, to *TypedObject) {
	if to == nil || to.typ == nil || to.typ.baseType == nil {
		return
	}
	if f == nil || f.RP == nil {
		panic("FinishNewTypedObject: nil RunFrame for based TypedObject")
	}
	bt := to.typ.baseType
	mm := bt.GetClassMethodMetadata("_newWithDelegate")
	if mm == nil {
		panic(fmt.Errorf("FinishNewTypedObject(%s): base type %q has no _newWithDelegate", to.typ.name, bt.name))
	}
	ds := f.RP.DataStack
	ds.Push(to)
	mm.invokeCall(f, NewTypeObjPair(bt))
	base := ds.Pop()
	to.SetFieldByName(typedObjectBaseFieldName, base)
}

// addTypedObjectBaseField appends the auto schema field "base" typed as baseType.
func addTypedObjectBaseField(t *Type, baseType *Type) {
	i := 0
	if t.propertyMetadata != nil {
		i = t.propertyMetadata.Len()
	}
	pm := &PropertyMetadata{
		name: typedObjectBaseFieldName,
		get:  mintTypedObjectFieldGet(i),
		set:  mintTypedObjectFieldSet(i, baseType),
	}
	t.AddResolvedProperty(typedObjectBaseFieldName, pm, baseType)
	t.baseType = baseType
}

func CreateProperties(t *Type, fields *DynamicObject) {
	for i := 0; i < fields.Len(); i++ {
		item := fields.ItemAt(i)
		name := item.Name
		typ, ok := item.Value.(*Type)
		if !ok {
			panic(fmt.Errorf("CreateProperties(%s): field %q value is %T, want *Type", t.name, name, item.Value))
		}
		if typ == nil {
			panic(fmt.Errorf("CreateProperties(%s): field %q has nil type", t.name, name))
		}
		pm := &PropertyMetadata{
			name: name,
			get:  mintTypedObjectFieldGet(i),
			set:  mintTypedObjectFieldSet(i, typ),
		}
		t.AddResolvedProperty(name, pm, typ)
	}
}

// CreateConstants seals TypeDef .constants (value .name!) as readonly class const properties.
func CreateConstants(t *Type, constants *DynamicObject) {
	for i := 0; i < constants.Len(); i++ {
		item := constants.ItemAt(i)
		name := item.Name
		if t.ownMemberNameTaken(name) {
			panic(fmt.Errorf("CreateConstants(%s): %q clashes with an existing member", t.name, name))
		}
		val := item.Value
		pm := &PropertyMetadata{
			name: name,
			get:  func(_ Any) Any { return val },
		}
		t.AddClassProperty(name, pm, "Any")
	}
}

// CreateTypeFromDef creates and registers a new Type from a TypeDef shell in dc.tr
// and adds a type proxy word to dc.dict (same path as bootstrap registerTypeWords).
func CreateTypeFromDef(dc *DictionaryContext, def *DynamicObject) *Type {
	if dc == nil || dc.tr == nil || dc.dict == nil {
		panic("CreateTypeFromDef: nil dictionary context, TypeRegistrar, or dictionary")
	}
	if !def.HasField("name") {
		panic("CreateTypeFromDef: name is required")
	}
	name, ok := def.Get("name").(String)
	if !ok || name == "" {
		panic("CreateTypeFromDef: empty type name")
	}
	if dc.tr.Get(name) != nil {
		panic(fmt.Errorf("CreateTypeFromDef: type name %q already used", name))
	}
	if !def.HasField("fields") {
		panic("CreateTypeFromDef: fields is required")
	}
	fields, ok := def.Get("fields").(*DynamicObject)
	if !ok {
		panic(fmt.Errorf("CreateTypeFromDef: fields is %T, want *DynamicObject", def.Get("fields")))
	}

	t := &Type{
		name:            name,
		registrar:       dc.tr,
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[TypedObject](name),
	}

	var baseType *Type
	if def.HasField("baseType") {
		raw := def.Get("baseType")
		bt, ok := raw.(*Type)
		if !ok || bt == nil {
			panic(fmt.Errorf("CreateTypeFromDef(%s): baseType is %T, want catalog *Type", name, raw))
		}
		if bt.GetClassMethodMetadata("_newWithDelegate") == nil {
			panic(fmt.Errorf("CreateTypeFromDef(%s): %q is not a Base type (missing _newWithDelegate)", name, bt.Name()))
		}
		if bool(fields.HasField(typedObjectBaseFieldName)) {
			panic(fmt.Errorf("CreateTypeFromDef(%s): fields must not declare %q (reserved for baseType)", name, typedObjectBaseFieldName))
		}
		baseType = bt
	}

	CreateProperties(t, fields)
	if baseType != nil {
		addTypedObjectBaseField(t, baseType)
	}
	// Optional constants
	if def.HasField("constants") {
		constants, ok := def.Get("constants").(*DynamicObject)
		if !ok {
			panic(fmt.Errorf("CreateTypeFromDef: constants is %T, want *DynamicObject", def.Get("constants")))
		}
		CreateConstants(t, constants)
	}
	registerTypedObjectItemsMetadata(t)
	t.SetParent(typedObjectType)

	dc.tr.RegisterCurrentType(t)
	dc.tr.ResolveRegistrations()
	addTypeDictionaryWord(dc, t)
	return t
}

var (
	protoTypedObjectType *Type
	typedObjectType      *Type
)

func CreateRegisterProtoTypedObject() {
	t := &Type{
		name: "ProtoTypedObject",
		getZero: func() any {
			panic("ProtoTypedObject: no zero value")
		},
		checkCoerceType: func(value any) any {
			typ, ok := value.(*Type)
			if !ok {
				panic(fmt.Errorf("ProtoTypedObject.checkCoerceType: expected *Type, got %T", value))
			}
			return typ
		},
	}
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		to := MakeNewTypedObject(ths.Obj.(*Type))
		FinishNewTypedObject(f, to)
		f.RP.DataStack.Push(to)
	})
	tr := GetTypeRegistrar()
	tr.RegisterCurrentType(t)
	tr.AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       "ProtoMap",
		TypeSlot:       &t.parent,
	})
	protoTypedObjectType = t
}

func CreateRegisterTypedObject() {
	t := &Type{
		name:            "TypedObject",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[TypedObject]("TypedObject"),
		parent:          protoTypedObjectType,
	}
	t.registerNativeClassMethod("newTypeDef", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewTypeDef())
	})
	t.registerNativeClassMethod("makeType", func(f *RunFrame, ths *TypeObjPair) {
		// ( def:DynamicObject -- )
		def := f.RP.DataStack.Pop().(*DynamicObject)
		CreateTypeFromDef(f.Task.dc, def)
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	typedObjectType = t
}

func init() {
	CreateRegisterProtoTypedObject() // MUST be first
	CreateRegisterTypedObject()
}

var (
	_ Typed    = (*TypedObject)(nil)
	_ Iterable = (*TypedObject)(nil)
	_ Iterator = (*TypedObjectIterator)(nil)
)
