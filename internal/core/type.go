package core

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	TypeDispatchKindInstanceGet int = iota
	TypeDispatchKindInstanceSet
	TypeDispatchKindInstanceItemGetByOrdinal
	TypeDispatchKindInstanceItemSetByOrdinal
	TypeDispatchKindInstanceItemGetByKey
	TypeDispatchKindInstanceItemSetByKey
	TypeDispatchKindInstanceMethod
	TypeDispatchKindClassGet
	TypeDispatchKindClassSet
	TypeDispatchKindClassMethod
)

// NOTE: Value/Pointer distinction is largely unimportant, since Final typically
// uses value semantics for value types, include interfaces and go maps,
// and pointer semantics for pointer types, primarily for structs.
// these details are handled by the Final runtime, and the functors.
// TODO TBD what about pointers to value or pointer types?

// TypeFlags is a bitmask of flags that describe the type.
// These flags are a polymorphic union for Property and Method types
type TypeFlags = int

// TypeFlags constants
// The first set are for properties, the second set are for methods
/*const (
	// TypeIsSimpleType is a single value, such as a var DT slot
	// Simple types only support SimpleValueRef and @ ! accessors; no other accessors
	// However their values may be either be runtime values like Int or String,
	// or compound types, when can then be further dereferenced
	TypeIsSimpleType TypeFlags = 1 << iota
	// TypeIsCompoundType is a compound type such as List, Map, Object etc.
	// Compound types support all Ref accessors, including properties, methods, and
	// optionally default Ordinal and/or Named collections
	TypeIsCompoundType
	// The Compound value has a default ordinal collection, such as List or OrdinalMap
	TypeFlagHasOrdinalCollection
	// The Compound value has a default named collection, such as Map or OrdinalMap
	TypeFlagHasNamedCollection

	// The value is a method type, such as a function pointer
	// NOTE: currently methods have very little other metadata.
	// They get an MethodRef as their default accessor.
	TypeFlagIsMethod
)*/

type RuntimeDispatchKind = int

const (
	RuntimeDispatchKindInstanceItemsGetByOrdinal RuntimeDispatchKind = iota
	RuntimeDispatchKindInstanceItemsSetByOrdinal                     // if not read only
	RuntimeDispatchKindInstanceItemsGetByKey
	RuntimeDispatchKindInstanceItemsSetByKey // if not read only
	RuntimeDispatchKindInstanceItemsHasByKey
	RuntimeDispatchKindInstanceItemsGetByMultiIndex // @[] — N indexes (rank from ItemsMetadata)
	RuntimeDispatchKindInstanceItemsSetByMultiIndex // ![]
	RuntimeDispatchKindInstanceGetByName            // ( name -- value ) instance property get
	RuntimeDispatchKindInstanceSetByName            // ( value name -- ) instance property set
	RuntimeDispatchKindInstanceMethodByName         // ( name -- … ) instance method call
	RuntimeDispatchKindClassGetByName               // ( name -- value ) class facet property get
	RuntimeDispatchKindClassSetByName               // ( value name -- ) class facet property set
	RuntimeDispatchKindClassMethodByName            // ( name -- … ) class facet method call
)

type RuntimeDispatch interface {
	Invoke(f *RunFrame, ths *TypeObjPair, kind RuntimeDispatchKind)
}

// Typed compounds expose their sealed Type metadata.
type Typed interface {
	FinalType() *Type
}

// Type describes a sealed Final type (value or compound shim).
type Type struct {
	name    string
	typeIdx int // local, 0-based index within its owning registrar's catalog
	// registrar owns this type's catalog + deferred resolutions; nil = global GetTypeRegistrar().
	registrar *TypeRegistrar
	//not yet	flags            TypeFlags
	getZero                  func() any
	checkCoerceType          func(value any) any
	parent                   *Type   // explicit prototype; nil = implicit ProtoType via GetParent()
	mixins                   []*Type // mixin catalog types applied via ApplyMixin (doc/help only; not lookup)
	mixinRegistering         *Type   // non-nil only inside ApplyMixin; stamps mixinType on new members
	propertyMetadata         *SFM[*PropertyMetadata]
	propertyMetadataOverride PropertyMetadataOverrideFn
	itemsMetadata            *ItemsMetadata
	methodMetadata           *SFM[*MethodMetadata]
	classPropertyMetadata    *SFM[*PropertyMetadata]
	classMethodMetadata      *SFM[*MethodMetadata]
	classType                *Type                // we cache this for fast dispatch
	restrict                 SecurityRestrictions // SR bits; copied onto companion type Word.srFlags at mint
	// baseType is set on based TypedObjects (TypeDef baseType); nil otherwise.
	// The Go base is stored in schema field "base"; mint calls baseType._newWithDelegate.
	baseType *Type
}

// ReferenceTypeZeroFn is the shared getZero for nullable reference compounds
// (List, Object, Word, …). Produces untyped nil — the simple zero vocabulary.
var ReferenceTypeZeroFn = func() any { return nil }

// MakePointerTypeCheckCoerce returns checkCoerceType for a Final reference type
// whose runtime cells hold *T (T is the pointed-to struct, not the pointer).
//
// Nil policy:
//   - accept untyped nil (Final null, getZero)
//   - accept (*T)(nil)
//   - normalize any nil to untyped nil on store
//   - reject wrong types (including wrong typed nils)
func MakePointerTypeCheckCoerce[T any](finalTypeName string) func(value any) any {
	return func(value any) any {
		switch v := value.(type) {
		case nil:
			return nil
		case *T:
			if v == nil {
				return nil
			}
			return v
		default:
			panic(fmt.Errorf(
				"%s.checkCoerceType: expected %s or nil, got %T",
				finalTypeName, finalTypeName, value,
			))
		}
	}
}

// MakeHostTypeZero returns a getZero for a host type registered via NewTypeReg[T].
//
// Value types (small immutable structs like a geometry point or an RGB color) yield
// their Go zero value, so a fresh slot holds e.g. Point{0,0} rather than null. Reference
// types (pointers, and other nilable kinds whose zero is nil) yield Final null (untyped
// nil), matching MakeHostTypeCheckCoerce's nil normalization. The reference/value branch
// is decided once, at factory creation.
func MakeHostTypeZero[T any]() func() any {
	var z T
	if boxed := any(z); boxed == nil {
		return ReferenceTypeZeroFn
	} else if rv := reflect.ValueOf(boxed); rv.Kind() == reflect.Pointer && rv.IsNil() {
		return ReferenceTypeZeroFn
	}
	return func() any { var v T; return v }
}

// MakeHostTypeCheckCoerce returns a checkCoerceType for a host type registered via
// NewTypeReg[T]. Unlike MakePointerTypeCheckCoerce, T is the instance type itself
// (typically a pointer, e.g. *Lissajous), so no element unwrapping is needed.
//
// Nil policy: accept untyped nil (Final null, getZero); accept any value assignable to
// T; reject everything else.
func MakeHostTypeCheckCoerce[T any](finalTypeName string) func(value any) any {
	return func(value any) any {
		if value == nil {
			return nil
		}
		if v, ok := value.(T); ok {
			return v
		}
		panic(fmt.Errorf(
			"%s.checkCoerceType: expected %s or nil, got %T",
			finalTypeName, finalTypeName, value,
		))
	}
}

func (t *Type) CreateClassType() *Type {
	return &Type{
		name:    "Class",
		typeIdx: 0,
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			return value
		},
		propertyMetadata: t.classPropertyMetadata,
		methodMetadata:   t.classMethodMetadata,
		parent:           typeType, // introspection surface (propertyMetadata, …)
	}
}

// TypeMemberFn is a native member handler: RunFrame + ths + DS operands (110-objects-and-types.md).
type TypeMemberFn = func(f *RunFrame, ths *TypeObjPair)

// PropertyMetadataOverrideFn resolves a property on t when registered.
// Non-nil = property; nil = continue lookup (local static table, then prototype chain).
type PropertyMetadataOverrideFn func(t *Type, name string) *PropertyMetadata

type PropertyGetFn = func(inst Any) Any
type PropertySetFn = func(inst Any, value Any)

// PropertyMetadata is one named property on a Type (stored or computed).
type PropertyMetadata struct {
	name      string
	_type     *Type
	get       PropertyGetFn        // nil = no get
	set       PropertySetFn        // nil = read-only
	doc       string               // documentation string; usually set in Final
	restrict  SecurityRestrictions // 0 = unrestricted; set at registration via *Sec
	mixinType *Type                // set when registered inside ApplyMixin; nil = native to containing type
}

type GetByOrdinalFn func(inst Any, idx Any /*Int*/) Any
type SetByOrdinalFn func(inst Any, idx Any /*Int*/, value Any)
type GetByKeyFn func(inst Any, key Any) Any // panics if not found
type SetByKeyFn func(inst Any, key Any, value Any)
type HasByKeyFn func(inst Any, key Any) bool
type GetByMultiIndexFn func(inst Any, idxs ...Any) Any
type SetByMultiIndexFn func(inst Any, value Any, idxs ...Any)

// ItemsMetadata carries aggregate element access (List, Map, Mat2f, …) — not a property.
// A collection only includes the operations it supports.
// For example, a List only includes the @() and !() operations, but not the @{} and !{} operations.
// A Map only includes the @{} and !{} operations, but not the @() and !() operations.
// An OrdinalMap only includes the @() and @{} but not the !() and !{} operations,
// because it uses an Add method to add new items.
// By-name collections may also include ?{} (existence probe) when registered.
// Mat2f may also register multidimensional @[]/![] (multiIndexArity = rank).
type ItemsMetadata struct {
	_itemType       *Type
	getByOrdinal    GetByOrdinalFn    // the @() accessor, if supported
	setByOrdinal    SetByOrdinalFn    // the !() accessor, if supported
	getByKey        GetByKeyFn        // the @{} accessor, if supported
	setByKey        SetByKeyFn        // the !{} accessor, if supported
	hasByKey        HasByKeyFn        // the ?{} accessor, if supported
	getByMultiIndex GetByMultiIndexFn // the @[] accessor, if supported
	setByMultiIndex SetByMultiIndexFn // the ![] accessor, if supported
	multiIndexArity int               // N for @[] / ![]; 0 if unsupported
	doc             string            // documentation string; usually set in Final
}

// MethodMetadata is one named method on a Type facet.
type MethodMetadata struct {
	name           string
	call           TypeMemberFn
	isFinalMethod  bool
	finalModuleID  ModuleID
	finalColonCode int
	doc            string               // documentation string; usually set in Final
	restrict       SecurityRestrictions // 0 = unrestricted; set at registration via *Sec
	mixinType      *Type                // set when registered inside ApplyMixin; nil = native to containing type
	forgotten      bool                 // set when type leaves catalog (forget:); EventEscFn drops such handlers
}

// NativeMethodFn is easy native method code using ths and DS only.
type NativeMethodFn func(f *RunFrame, ths *TypeObjPair)

func wrapNativeMethod(call NativeMethodFn) TypeMemberFn {
	return func(f *RunFrame, ths *TypeObjPair) {
		if f == nil || f.RP == nil || call == nil {
			return
		}
		call(f, ths)
	}
}

// mintFinalMethodCall returns a TypeMemberFn that parks the receiver on o: and nests into a Final method word via OpInterrupt.
func mintFinalMethodCall(moduleID ModuleID, colonCode int) TypeMemberFn {
	return func(f *RunFrame, ths *TypeObjPair) {
		if f == nil || f.RP == nil || f.Task == nil {
			return
		}
		f.RP.OStack.Push(ths.Obj)
		if f.ForceOp != nil {
			f.SetInterrupt(moduleID, colonCode)
			return
		}
		// Interpret-time compound walk uses a bare RunFrame (no ForceOp). Nest synchronously.
		runInterruptNest(f.Task, f.RP, moduleID, colonCode)
	}
}

// forgottenFinalMethodCall replaces a Final method after its type is rolled off the catalog
// (forget: / compile-unit rollback). Outstanding MethodRefs still point at this metadata;
// the stub panics instead of nesting into truncated or reissued CT.
func forgottenFinalMethodCall(typeName, methodName string, isClass bool) TypeMemberFn {
	sep := "#"
	if isClass {
		sep = "##"
	}
	msg := fmt.Sprintf("MethodRef: type %s%s%s forgotten", typeName, sep, methodName)
	return func(_ *RunFrame, _ *TypeObjPair) {
		panic(fmt.Errorf("%s", msg))
	}
}

func poisonForgottenFinalMethodMap(typeName string, mm *MethodMetadataMap, isClass bool) {
	if mm == nil {
		return
	}
	for i := 0; i < mm.Len(); i++ {
		item := mm.ItemAt(i)
		m := item.Value
		if m == nil || !m.isFinalMethod {
			continue
		}
		name := item.Name
		if m.name != "" {
			name = m.name
		}
		m.call = forgottenFinalMethodCall(typeName, name, isClass)
		m.forgotten = true
	}
}

// poisonForgottenFinalMethods stubs Final method call slots so leftover MethodRefs
// cannot execute after this type leaves the catalog. Native methods are left as-is.
func (t *Type) poisonForgottenFinalMethods() {
	if t == nil {
		return
	}
	poisonForgottenFinalMethodMap(t.name, t.methodMetadata, false)
	poisonForgottenFinalMethodMap(t.name, t.classMethodMetadata, true)
}

func exposedMethodNameFromWord(w *Word) string {
	name := w.name
	if i := strings.LastIndex(name, "#"); i >= 0 {
		name = name[i+1:]
	}
	//no: need to have _ name = strings.TrimPrefix(name, "_")
	if name == "" {
		panic(fmt.Errorf("exposedMethodNameFromWord: empty method name from %q", w.name))
	}
	return name
}

// RegisterFinalMethodFromWord registers a Final method backed by colon word w.
// Go/test helper; Final registration is via `: Type#… ;` / `##` tear-off.
func (t *Type) RegisterFinalMethodFromWord(w *Word) *MethodMetadata {
	if t == nil {
		panic("Type.RegisterFinalMethodFromWord: nil type")
	}
	if w == nil {
		panic(fmt.Errorf("%s.RegisterFinalMethodFromWord: nil word", t.name))
	}
	if !w.IsColon() {
		panic(fmt.Errorf("%s.RegisterFinalMethodFromWord: %q is not a colon word", t.name, w.name))
	}
	if w.colonCode < 0 {
		panic(fmt.Errorf("%s.RegisterFinalMethodFromWord: %q has no colonCode", t.name, w.name))
	}
	exposedName := exposedMethodNameFromWord(w)
	return t.registerFinalMethodMeta(exposedName, w.moduleId, int(w.colonCode), w.doc, false)
}

// registerFinalMethodMeta installs a Final method (instance or class) from CT coordinates.
func (t *Type) registerFinalMethodMeta(name string, moduleID ModuleID, colonCode int, doc string, isClass bool) *MethodMetadata {
	mm := &MethodMetadata{
		name:           name,
		doc:            doc,
		isFinalMethod:  true,
		finalModuleID:  moduleID,
		finalColonCode: colonCode,
	}
	mm.call = mintFinalMethodCall(mm.finalModuleID, mm.finalColonCode)
	if isClass {
		t.AddClassMethod(name, mm)
	} else {
		t.AddMethod(name, mm)
	}
	return mm
}

// ownMemberNameTaken reports whether name is already on this type's own flat surface
// (instance/class properties and methods) — no prototype walk.
func (t *Type) ownMemberNameTaken(name string) bool {
	if t == nil {
		return false
	}
	if t.propertyMetadata != nil && t.propertyMetadata.HasKey(name) {
		return true
	}
	if t.methodMetadata != nil && t.methodMetadata.HasKey(name) {
		return true
	}
	if t.classPropertyMetadata != nil && t.classPropertyMetadata.HasKey(name) {
		return true
	}
	if t.classMethodMetadata != nil && t.classMethodMetadata.HasKey(name) {
		return true
	}
	return false
}

// FinalType returns the class facet for type-literal dispatch (lazy; always includes standard _ members).
func (t *Type) FinalType() *Type {
	// lazily create the class type, but we cache it for fast dispatch
	if t.classType == nil {
		t.classType = t.CreateClassType()
	}
	return t.classType
}

// InvalidateClassTypeCache drops the lazy FinalType() companion so the next call rebuilds
// with live classMethodMetadata / classPropertyMetadata pointers.
func (t *Type) InvalidateClassTypeCache() {
	if t != nil {
		t.classType = nil
	}
}

func (t *Type) NaturalString() string {
	if t == nil {
		return "Type(nil)"
	}
	return "Type: " + t.name
}

func (t *Type) Name() string { return t.name }
func (t *Type) TypeIdx() int { return t.typeIdx }

// reg returns the registrar that owns this type's catalog and deferred resolutions.
// nil registrar (every core &Type{} literal) resolves to the global registrar, so
// core registration behavior is unchanged.
func (t *Type) reg() *TypeRegistrar {
	if t.registrar != nil {
		return t.registrar
	}
	return GetTypeRegistrar()
}

func (t *Type) Zero() any { return t.getZero() }

func (t *Type) CheckCoerceType(v any) any { return t.checkCoerceType(v) }

func (t *Type) setPropertyMetadataOverride(fn PropertyMetadataOverrideFn) {
	t.propertyMetadataOverride = fn
}

// SetParent links t to an explicit prototype. Nil parent panics; cycles panic.
func (t *Type) SetParent(parent *Type) {
	if t == nil {
		panic("Type.SetParent: nil type")
	}
	if parent == nil {
		panic(fmt.Errorf("%s.SetParent: nil parent", t.name))
	}
	if parent == t {
		panic(fmt.Errorf("%s.SetParent: self-parent", t.name))
	}
	for p := parent; p != nil; {
		if p == t {
			panic(fmt.Errorf("%s.SetParent: cycle via %s", t.name, parent.name))
		}
		if p == protoTypeType {
			break
		}
		p = p.parent
	}
	t.parent = parent
}

// ApplyMixin runs register on t's local member tables, then records mixinType in t.mixins
// for doc/help. Does not set parent or affect member lookup beyond register's additions.
// While register runs, t.mixinRegistering is set so AddMethod/AddProperty stamp mm.mixinType.
func ApplyMixin(t *Type, mixinType *Type, register func(*Type)) {
	if t == nil {
		panic("ApplyMixin: nil type")
	}
	if mixinType == nil {
		panic(fmt.Errorf("%s.ApplyMixin: nil mixin type", t.name))
	}
	if register == nil {
		panic(fmt.Errorf("%s.ApplyMixin: nil register", t.name))
	}
	if t.mixinRegistering != nil {
		panic(fmt.Errorf("%s.ApplyMixin: nested ApplyMixin (already registering %s)", t.name, t.mixinRegistering.name))
	}
	t.mixinRegistering = mixinType
	defer func() { t.mixinRegistering = nil }()
	register(t)
	t.mixins = append(t.mixins, mixinType)
}

// GetMixinTypes returns mixin catalog types applied to t (registration order).
func (t *Type) GetMixinTypes() []*Type {
	if t == nil || len(t.mixins) == 0 {
		return nil
	}
	out := make([]*Type, len(t.mixins))
	copy(out, t.mixins)
	return out
}

// GetParent returns the next type in the prototype chain, or nil at the root.
func (t *Type) GetParent() *Type {
	if t == nil {
		return nil
	}
	if t.parent == protoTypeType {
		return nil
	}
	if t.parent != nil {
		return t.parent
	}
	return protoTypeType
}

// GetPropertyMetadata resolves an instance property on t, walking the prototype chain.
func (t *Type) GetPropertyMetadata(name string) *PropertyMetadata {
	if t == nil {
		return nil
	}
	if t.propertyMetadataOverride != nil {
		if pm := t.propertyMetadataOverride(t, name); pm != nil {
			return pm
		}
	}
	if t.propertyMetadata != nil && t.propertyMetadata.HasKey(name) {
		return t.propertyMetadata.Get(name)
	}
	if p := t.GetParent(); p != nil {
		return p.GetPropertyMetadata(name)
	}
	return nil
}

// GetMethodMetadata resolves an instance method on t, walking the prototype chain.
func (t *Type) GetMethodMetadata(name string) *MethodMetadata {
	if t == nil {
		return nil
	}
	if t.methodMetadata != nil && t.methodMetadata.HasKey(name) {
		return t.methodMetadata.Get(name)
	}
	if p := t.GetParent(); p != nil {
		return p.GetMethodMetadata(name)
	}
	return nil
}

// GetClassPropertyMetadata resolves a class property on t, walking the prototype chain.
func (t *Type) GetClassPropertyMetadata(name string) *PropertyMetadata {
	if t == nil {
		return nil
	}
	if t.classPropertyMetadata != nil && t.classPropertyMetadata.HasKey(name) {
		return t.classPropertyMetadata.Get(name)
	}
	if p := t.GetParent(); p != nil {
		return p.GetClassPropertyMetadata(name)
	}
	return nil
}

// GetClassMethodMetadata resolves a class method on t, walking the prototype chain.
func (t *Type) GetClassMethodMetadata(name string) *MethodMetadata {
	if t == nil {
		return nil
	}
	if t.classMethodMetadata != nil && t.classMethodMetadata.HasKey(name) {
		return t.classMethodMetadata.Get(name)
	}
	if p := t.GetParent(); p != nil {
		return p.GetClassMethodMetadata(name)
	}
	return nil
}

// GetItemsMetadata returns aggregate item access metadata, walking the prototype chain.
func (t *Type) GetItemsMetadata() *ItemsMetadata {
	if t == nil {
		return nil
	}
	if t.itemsMetadata != nil {
		return t.itemsMetadata
	}
	if p := t.GetParent(); p != nil {
		return p.GetItemsMetadata()
	}
	return nil
}

func (t *Type) HasProperty(name string) bool {
	return t.GetPropertyMetadata(name) != nil
}

// GetPropertyNames returns locally registered instance property names in
// registration (ordinal) order. Does not walk the parent chain. Types that
// only expose fields via propertyMetadataOverride (e.g. DynamicObject) yield
// an empty list.
func (t *Type) GetPropertyNames() *StringList {
	if t == nil || t.propertyMetadata == nil {
		return NewStringList()
	}
	n := t.propertyMetadata.Len()
	sl := NewStringListWithCapacity(n)
	for i := 0; i < n; i++ {
		sl.Append(t.propertyMetadata.ItemAt(i).Name)
	}
	return sl
}

func (t *Type) PropertyByName(name string) *PropertyMetadata {
	if t == nil {
		panic(fmt.Errorf("Type.PropertyByName: nil type"))
	}
	if pm := t.GetPropertyMetadata(name); pm != nil {
		return pm
	}
	panic(fmt.Errorf("%s: unknown property %q", t.name, name))
}

func (t *Type) HasMethod(name string) bool {
	return t.GetMethodMetadata(name) != nil
}

func (t *Type) MethodByName(name string) *MethodMetadata {
	if t == nil {
		panic(fmt.Errorf("Type.MethodByName: nil type"))
	}
	if mm := t.GetMethodMetadata(name); mm != nil {
		return mm
	}
	panic(fmt.Errorf("%s: unknown method %q", t.name, name))
}

func (t *Type) HasClassMethod(name string) bool {
	return t.GetClassMethodMetadata(name) != nil
}

func (t *Type) ClassMethodByName(name string) *MethodMetadata {
	if t == nil {
		panic(fmt.Errorf("Type.ClassMethodByName: nil type"))
	}
	if mm := t.GetClassMethodMetadata(name); mm != nil {
		return mm
	}
	panic(fmt.Errorf("%s: unknown class method %q", t.name, name))
}

// SetItemsMetadata assigns aggregate element access metadata once at init.
func (t *Type) SetItemsMetadata(im *ItemsMetadata) {
	if t.itemsMetadata != nil {
		panic(fmt.Errorf("%s: itemsMetadata already set", t.name))
	}
	t.itemsMetadata = im
}

// AddClassProperty registers a class property on this type.
func (t *Type) AddClassProperty(name string, pm *PropertyMetadata, typeName string) {
	if typeName == "" {
		panic(fmt.Errorf("Type.AddClassProperty: empty type name"))
	}
	if t.classMethodMetadata.HasKey(name) {
		panic(fmt.Errorf("%s: class property %q collides with class method", t.name, name))
	}
	if t.classPropertyMetadata == nil {
		t.classPropertyMetadata = NewSFM[*PropertyMetadata](4)
	}
	t.stampMixinTypeOnProperty(pm)
	t.classPropertyMetadata.Add(name, pm)
	t.reg().AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindProperty,
		ContainingType: t,
		TypeName:       typeName,
		TypeSlot:       &pm._type,
	})
}

// AddClassMethod registers a class method on this type.
func (t *Type) AddClassMethod(name string, mm *MethodMetadata) {
	if t.classPropertyMetadata.HasKey(name) {
		panic(fmt.Errorf("%s: class method %q collides with class property", t.name, name))
	}
	if t.classMethodMetadata == nil {
		t.classMethodMetadata = NewSFM[*MethodMetadata](4)
	}
	t.stampMixinTypeOnMethod(mm)
	t.classMethodMetadata.Add(name, mm)
	t.InvalidateClassTypeCache()
}

// registerProperty registers a proxy-native property (get required; set nil = read-only).
func (t *Type) registerProperty(name, typeName string, get PropertyGetFn, set PropertySetFn) {
	t.registerPropertySec(name, typeName, get, set, 0)
}

// registerPropertySec is registerProperty with a trailing SecurityRestrictions mask.
func (t *Type) registerPropertySec(name, typeName string, get PropertyGetFn, set PropertySetFn, restrict SecurityRestrictions) {
	pm := &PropertyMetadata{name: name, restrict: restrict}
	if get != nil {
		pm.get = get
	}
	if set != nil {
		pm.set = set
	}
	t.AddProperty(name, pm, typeName)
}

// registerItemsData registers aggregate element access without multi-index slots.
func (t *Type) registerItemsData(
	itemName string,
	getByOrdinal GetByOrdinalFn,
	setByOrdinal SetByOrdinalFn,
	getByKey GetByKeyFn,
	setByKey SetByKeyFn,
	hasByKey HasByKeyFn,
) {
	t.registerItemsData2(itemName, getByOrdinal, setByOrdinal, getByKey, setByKey, hasByKey, nil, nil, 0)
}

// registerItemsData2 is registerItemsData plus optional @[] / ![] (multiIndexArity = N).
func (t *Type) registerItemsData2(
	itemName string,
	getByOrdinal GetByOrdinalFn,
	setByOrdinal SetByOrdinalFn,
	getByKey GetByKeyFn,
	setByKey SetByKeyFn,
	hasByKey HasByKeyFn,
	getByMultiIndex GetByMultiIndexFn,
	setByMultiIndex SetByMultiIndexFn,
	multiIndexArity int,
) {
	if (getByMultiIndex != nil || setByMultiIndex != nil) && multiIndexArity < 1 {
		panic(fmt.Errorf("%s: registerItemsData2: multi-index requires multiIndexArity >= 1", t.name))
	}
	im := &ItemsMetadata{}
	if getByOrdinal != nil {
		im.getByOrdinal = getByOrdinal
	}
	if setByOrdinal != nil {
		im.setByOrdinal = setByOrdinal
	}
	if getByKey != nil {
		im.getByKey = getByKey
	}
	if setByKey != nil {
		im.setByKey = setByKey
	}
	if hasByKey != nil {
		im.hasByKey = hasByKey
	}
	if getByMultiIndex != nil {
		im.getByMultiIndex = getByMultiIndex
	}
	if setByMultiIndex != nil {
		im.setByMultiIndex = setByMultiIndex
	}
	im.multiIndexArity = multiIndexArity
	t.SetItemsMetadata(im)
	t.reg().AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindItems,
		ContainingType: t,
		TypeName:       itemName,
		TypeSlot:       &im._itemType,
	})
}

// registerNativeMethod registers a proxy-native instance method.
func (t *Type) registerNativeMethod(name string, call NativeMethodFn) {
	t.registerNativeMethodSec(name, call, 0)
}

// registerNativeMethodSec is registerNativeMethod with a trailing SecurityRestrictions mask.
func (t *Type) registerNativeMethodSec(name string, call NativeMethodFn, restrict SecurityRestrictions) {
	t.AddMethod(name, &MethodMetadata{
		name:     name,
		call:     wrapNativeMethod(call),
		restrict: restrict,
	})
}

// registerNativeClassMethod registers a proxy-native class method on this type.
func (t *Type) registerNativeClassMethod(name string, call NativeMethodFn) {
	t.registerNativeClassMethodSec(name, call, 0)
}

// registerNativeClassMethodSec is registerNativeClassMethod with a trailing SecurityRestrictions mask.
func (t *Type) registerNativeClassMethodSec(name string, call NativeMethodFn, restrict SecurityRestrictions) {
	t.AddClassMethod(name, &MethodMetadata{
		name:     name,
		call:     wrapNativeMethod(call),
		restrict: restrict,
	})
}

// registerClassProperty registers a read-only or read-write class property on this type.
func (t *Type) registerClassProperty(name, typeName string, get PropertyGetFn, set PropertySetFn) {
	t.registerClassPropertySec(name, typeName, get, set, 0)
}

// registerClassPropertySec is registerClassProperty with a trailing SecurityRestrictions mask.
func (t *Type) registerClassPropertySec(name, typeName string, get PropertyGetFn, set PropertySetFn, restrict SecurityRestrictions) {
	pm := &PropertyMetadata{name: name, restrict: restrict}
	if get != nil {
		pm.get = get
	}
	if set != nil {
		pm.set = set
	}
	t.AddClassProperty(name, pm, typeName)
}

// NewRegisteringType constructs a catalog Type before RegisterCurrentType (name, getZero, checkCoerceType).
// Set parent with SetParent, or queue deferred parent link with AddClassResolution.
func NewRegisteringType(name string, getZero func() any, checkCoerce func(any) any) *Type {
	if name == "" {
		panic("NewRegisteringType: empty name")
	}
	if getZero == nil {
		panic("NewRegisteringType: nil getZero")
	}
	if checkCoerce == nil {
		panic("NewRegisteringType: nil checkCoerce")
	}
	return &Type{
		name:            name,
		getZero:         getZero,
		checkCoerceType: checkCoerce,
	}
}

// AddClassResolution queues deferred prototype link (TypeResolutionKindClass) into t.parent.
func (t *Type) AddClassResolution(protoTypeName string) {
	if t == nil {
		panic("Type.AddClassResolution: nil type")
	}
	if protoTypeName == "" {
		panic(fmt.Errorf("Type.AddClassResolution: empty proto type name"))
	}
	t.reg().AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindClass,
		ContainingType: t,
		TypeName:       protoTypeName,
		TypeSlot:       &t.parent,
	})
}

// RegisterProperty registers a proxy-native property (get required; set nil = read-only).
func (t *Type) RegisterProperty(name, typeName string, get PropertyGetFn, set PropertySetFn) {
	t.RegisterPropertySec(name, typeName, get, set, 0)
}

// RegisterPropertySec is RegisterProperty with a trailing SecurityRestrictions mask.
func (t *Type) RegisterPropertySec(name, typeName string, get PropertyGetFn, set PropertySetFn, restrict SecurityRestrictions) {
	t.registerPropertySec(name, typeName, get, set, restrict)
}

// RegisterItemsData registers aggregate element access functors (assign-once item-type resolution).
func (t *Type) RegisterItemsData(
	itemName string,
	getByOrdinal GetByOrdinalFn,
	setByOrdinal SetByOrdinalFn,
	getByKey GetByKeyFn,
	setByKey SetByKeyFn,
	hasByKey HasByKeyFn,
) {
	t.registerItemsData(itemName, getByOrdinal, setByOrdinal, getByKey, setByKey, hasByKey)
}

// RegisterNativeMethod registers a proxy-native instance method.
func (t *Type) RegisterNativeMethod(name string, call NativeMethodFn) {
	t.RegisterNativeMethodSec(name, call, 0)
}

// RegisterNativeMethodSec is RegisterNativeMethod with a trailing SecurityRestrictions mask.
func (t *Type) RegisterNativeMethodSec(name string, call NativeMethodFn, restrict SecurityRestrictions) {
	t.registerNativeMethodSec(name, call, restrict)
}

// RegisterNativeClassMethod registers a proxy-native class method on this type.
func (t *Type) RegisterNativeClassMethod(name string, call NativeMethodFn) {
	t.RegisterNativeClassMethodSec(name, call, 0)
}

// RegisterNativeClassMethodSec is RegisterNativeClassMethod with a trailing SecurityRestrictions mask.
func (t *Type) RegisterNativeClassMethodSec(name string, call NativeMethodFn, restrict SecurityRestrictions) {
	t.registerNativeClassMethodSec(name, call, restrict)
}

// RegisterClassProperty registers a read-only or read-write class property on this type.
func (t *Type) RegisterClassProperty(name, typeName string, get PropertyGetFn, set PropertySetFn) {
	t.RegisterClassPropertySec(name, typeName, get, set, 0)
}

// RegisterClassPropertySec is RegisterClassProperty with a trailing SecurityRestrictions mask.
func (t *Type) RegisterClassPropertySec(name, typeName string, get PropertyGetFn, set PropertySetFn, restrict SecurityRestrictions) {
	t.registerClassPropertySec(name, typeName, get, set, restrict)
}

// AddProperty registers a property; panics if name collides with a method.
func (t *Type) AddProperty(name string, pm *PropertyMetadata, typeName string) {
	if typeName == "" {
		panic(fmt.Errorf("Type.AddProperty: empty type name"))
	}
	if t.methodMetadata.HasKey(name) {
		panic(fmt.Errorf("%s: property %q collides with method", t.name, name))
	}
	if t.propertyMetadata == nil {
		t.propertyMetadata = NewSFM[*PropertyMetadata](4)
	}
	t.stampMixinTypeOnProperty(pm)
	t.propertyMetadata.Add(name, pm)
	t.reg().AddResolution(&TypeResolutionRecord{
		Kind:           TypeResolutionKindProperty,
		ContainingType: t,
		TypeName:       typeName,
		TypeSlot:       &pm._type,
	})
}

// AddResolvedProperty registers a property whose field type is already a *Type.
// pm._type is set directly; no deferred TypeResolutionRecord is queued.
func (t *Type) AddResolvedProperty(name string, pm *PropertyMetadata, typ *Type) {
	if typ == nil {
		panic(fmt.Errorf("Type.AddResolvedProperty: nil field type"))
	}
	if typ.name == "" {
		panic(fmt.Errorf("Type.AddResolvedProperty: empty field type name"))
	}
	if t.methodMetadata.HasKey(name) {
		panic(fmt.Errorf("%s: property %q collides with method", t.name, name))
	}
	if t.propertyMetadata == nil {
		t.propertyMetadata = NewSFM[*PropertyMetadata](4)
	}
	pm._type = typ
	t.stampMixinTypeOnProperty(pm)
	t.propertyMetadata.Add(name, pm)
}

// AddMethod registers a method; panics if name collides with a property.
func (t *Type) AddMethod(name string, mm *MethodMetadata) {
	if t.propertyMetadata.HasKey(name) {
		panic(fmt.Errorf("%s: method %q collides with property", t.name, name))
	}
	if t.methodMetadata == nil {
		t.methodMetadata = NewSFM[*MethodMetadata](4)
	}
	t.stampMixinTypeOnMethod(mm)
	t.methodMetadata.Add(name, mm)
}

// stampMixinTypeOnMethod sets mm.mixinType from t.mixinRegistering when inside ApplyMixin.
func (t *Type) stampMixinTypeOnMethod(mm *MethodMetadata) {
	if t == nil || mm == nil || t.mixinRegistering == nil {
		return
	}
	mm.mixinType = t.mixinRegistering
}

// stampMixinTypeOnProperty sets pm.mixinType from t.mixinRegistering when inside ApplyMixin.
func (t *Type) stampMixinTypeOnProperty(pm *PropertyMetadata) {
	if t == nil || pm == nil || t.mixinRegistering == nil {
		return
	}
	pm.mixinType = t.mixinRegistering
}

func (pm *PropertyMetadata) PropertyName() string { return pm.name }

func (pm *PropertyMetadata) HasGet() bool { return pm != nil && pm.get != nil }

func (pm *PropertyMetadata) HasSet() bool { return pm != nil && pm.set != nil }

// Type returns the declared property type, or nil if unset.
func (pm *PropertyMetadata) Type() *Type {
	if pm == nil {
		return nil
	}
	return pm._type
}

// MixinType returns the mixin Type that contributed this property, or nil.
func (pm *PropertyMetadata) MixinType() *Type {
	if pm == nil {
		return nil
	}
	return pm.mixinType
}

func (pm *PropertyMetadata) Doc() string {
	if pm == nil {
		return ""
	}
	return pm.doc
}

func (pm *PropertyMetadata) GetValue(obj any) any {
	if pm == nil {
		panic(fmt.Errorf("PropertyMetadata.GetValue: nil metadata"))
	}
	if pm.get == nil {
		panic(fmt.Errorf("PropertyMetadata.GetValue: no get for %q", pm.name))
	}
	return pm.get(obj)
}

func (pm *PropertyMetadata) SetValue(obj any, newValue any) {
	if pm == nil {
		panic(fmt.Errorf("PropertyMetadata.SetValue: nil metadata"))
	}
	if pm.set == nil {
		panic(fmt.Errorf("PropertyMetadata.SetValue: read-only %q", pm.name))
	}
	if pm._type != nil {
		newValue = pm._type.checkCoerceType(newValue)
	}
	pm.set(obj, newValue)
}

func (pm *PropertyMetadata) invokeGet(f *RunFrame, ths *TypeObjPair) {
	assertMemberRestrict(f, pm.restrict)
	pushFetchResult(f, pm.GetValue(ths.Obj))
}

func (pm *PropertyMetadata) invokeSet(f *RunFrame, ths *TypeObjPair) {
	if f == nil || f.RP == nil {
		panic("PropertyMetadata.invokeSet: nil frame")
	}
	assertMemberRestrict(f, pm.restrict)
	ds := f.RP.DataStack
	if ds.Len() < 1 {
		panic(fmt.Errorf("PropertyMetadata.invokeSet: empty DS"))
	}
	pm.SetValue(ths.Obj, ds.Pop())
}

func (im *ItemsMetadata) GetByOrdinal(obj any, idx any) any {
	if im == nil {
		panic("ItemsMetadata.GetByOrdinal: nil metadata")
	}
	if im.getByOrdinal == nil {
		panic("ItemsMetadata.GetByOrdinal: no getByOrdinal")
	}
	return im.getByOrdinal(obj, idx)
}

func (im *ItemsMetadata) SetByOrdinal(obj any, idx any, value any) {
	if im == nil {
		panic("ItemsMetadata.SetByOrdinal: nil metadata")
	}
	if im.setByOrdinal == nil {
		panic("ItemsMetadata.SetByOrdinal: no setByOrdinal")
	}
	if im._itemType != nil {
		value = im._itemType.checkCoerceType(value)
	}
	im.setByOrdinal(obj, idx, value)
}

func (im *ItemsMetadata) GetByMultiIndex(obj any, idxs ...any) any {
	if im == nil {
		panic("ItemsMetadata.GetByMultiIndex: nil metadata")
	}
	if im.getByMultiIndex == nil {
		panic("ItemsMetadata.GetByMultiIndex: no getByMultiIndex")
	}
	if len(idxs) != im.multiIndexArity {
		panic(fmt.Errorf("ItemsMetadata.GetByMultiIndex: got %d indexes, want %d", len(idxs), im.multiIndexArity))
	}
	return im.getByMultiIndex(obj, idxs...)
}

func (im *ItemsMetadata) SetByMultiIndex(obj any, value any, idxs ...any) {
	if im == nil {
		panic("ItemsMetadata.SetByMultiIndex: nil metadata")
	}
	if im.setByMultiIndex == nil {
		panic("ItemsMetadata.SetByMultiIndex: no setByMultiIndex")
	}
	if len(idxs) != im.multiIndexArity {
		panic(fmt.Errorf("ItemsMetadata.SetByMultiIndex: got %d indexes, want %d", len(idxs), im.multiIndexArity))
	}
	if im._itemType != nil {
		value = im._itemType.checkCoerceType(value)
	}
	im.setByMultiIndex(obj, value, idxs...)
}

func (im *ItemsMetadata) GetByKey(obj any, key any) any {
	if im == nil {
		panic("ItemsMetadata.GetByKey: nil metadata")
	}
	if im.getByKey == nil {
		panic("ItemsMetadata.GetByKey: no getByKey")
	}
	return im.getByKey(obj, key)
}

func (im *ItemsMetadata) SetByKey(obj any, key any, value any) {
	if im == nil {
		panic("ItemsMetadata.SetByKey: nil metadata")
	}
	if im.setByKey == nil {
		panic("ItemsMetadata.SetByKey: no setByKey")
	}
	if im._itemType != nil {
		value = im._itemType.checkCoerceType(value)
	}
	im.setByKey(obj, key, value)
}

func (im *ItemsMetadata) HasByKey(obj any, key any) bool {
	if im == nil {
		panic("ItemsMetadata.HasByKey: nil metadata")
	}
	if im.hasByKey == nil {
		panic("ItemsMetadata.HasByKey: no hasByKey")
	}
	return im.hasByKey(obj, key)
}

func (mm *MethodMetadata) invokeCall(f *RunFrame, ths *TypeObjPair) {
	if mm == nil {
		panic(fmt.Errorf("MethodMetadata.invokeCall: nil metadata"))
	}
	if mm.call == nil {
		panic(fmt.Errorf("MethodMetadata.invokeCall: nil call for %q", mm.name))
	}
	assertMemberRestrict(f, mm.restrict)
	mm.call(f, ths)
}

// Name is the method's registered name (e.g. "handleRender").
func (mm *MethodMetadata) Name() string {
	if mm == nil {
		return ""
	}
	return mm.name
}

func (mm *MethodMetadata) Doc() string {
	if mm == nil {
		return ""
	}
	return mm.doc
}

func (mm *MethodMetadata) IsFinalMethod() bool {
	return mm != nil && mm.isFinalMethod
}

// MixinType returns the mixin Type that contributed this method, or nil.
func (mm *MethodMetadata) MixinType() *Type {
	if mm == nil {
		return nil
	}
	return mm.mixinType
}

// Invoke runs this method with ths as the receiver (same as MethodRef.Invoke).
func (mm *MethodMetadata) Invoke(f *RunFrame, ths *TypeObjPair) {
	mm.invokeCall(f, ths)
}

// tryPushMetadataMap pushes ( map n ) when m is non-nil and non-empty, else Int(0).
func tryPushMetadataMap[V any](ds *Stack[any], m *SFM[V]) {
	if m == nil || m.Len() == 0 {
		ds.Push(Int(0))
		return
	}
	ds.Push(m)
	ds.Push(Int(m.Len()))
}

// --- Final-visible metadata types (introspection surface) ---

var (
	typeType             *Type // catalog type for *Type values (property slots / coerce)
	propertyMetadataType *Type
	itemsMetadataType    *Type
	methodMetadataType   *Type
)

func (pm *PropertyMetadata) FinalType() *Type { return propertyMetadataType }
func (im *ItemsMetadata) FinalType() *Type    { return itemsMetadataType }
func (mm *MethodMetadata) FinalType() *Type   { return methodMetadataType }

// CreateRegisterTypeType registers the catalog type for *Type cells (e.g. PropertyMetadata.type).
// Class facets parent here (CreateClassType) so type literals expose introspection properties.
func CreateRegisterTypeType() {
	t := &Type{
		name:    "Type",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			typ, ok := value.(*Type)
			if !ok {
				panic(fmt.Errorf("Type.checkCoerceType: expected *Type, got %T", value))
			}
			return typ
		},
	}
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*Type).name) },
		nil,
	)
	// Metadata tables: tryGet* methods (nil-safe); not exposed as properties.
	t.registerNativeMethod("tryGetPropertyMetadata", func(f *RunFrame, ths *TypeObjPair) {
		tryPushMetadataMap(f.RP.DataStack, ths.Obj.(*Type).propertyMetadata)
	})
	t.registerNativeMethod("tryGetMethodMetadata", func(f *RunFrame, ths *TypeObjPair) {
		tryPushMetadataMap(f.RP.DataStack, ths.Obj.(*Type).methodMetadata)
	})
	t.registerNativeMethod("tryGetClassPropertyMetadata", func(f *RunFrame, ths *TypeObjPair) {
		tryPushMetadataMap(f.RP.DataStack, ths.Obj.(*Type).classPropertyMetadata)
	})
	t.registerNativeMethod("tryGetClassMethodMetadata", func(f *RunFrame, ths *TypeObjPair) {
		tryPushMetadataMap(f.RP.DataStack, ths.Obj.(*Type).classMethodMetadata)
	})
	// ( -- im true | false ) — this type's own itemsMetadata only (no prototype walk).
	t.registerNativeMethod("tryGetItemsMetadata", func(f *RunFrame, ths *TypeObjPair) {
		if im := ths.Obj.(*Type).itemsMetadata; im != nil {
			f.RP.DataStack.Push(im)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	// ( name:s -- metadata true | false )
	t.registerNativeMethod("getProperty", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		if pm := ths.Obj.(*Type).GetPropertyMetadata(name); pm != nil {
			f.RP.DataStack.Push(pm)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	t.registerNativeMethod("getMethod", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		if mm := ths.Obj.(*Type).GetMethodMetadata(name); mm != nil {
			f.RP.DataStack.Push(mm)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	t.registerNativeMethod("getClassProperty", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		if pm := ths.Obj.(*Type).GetClassPropertyMetadata(name); pm != nil {
			f.RP.DataStack.Push(pm)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	t.registerNativeMethod("getClassMethod", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		if mm := ths.Obj.(*Type).GetClassMethodMetadata(name); mm != nil {
			f.RP.DataStack.Push(mm)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	// ( -- list:AnyList true | false ) — mixin catalog types from ApplyMixin (empty → false).
	t.registerNativeMethod("tryGetMixinTypes", func(f *RunFrame, ths *TypeObjPair) {
		mixins := ths.Obj.(*Type).mixins
		if len(mixins) == 0 {
			f.RP.DataStack.Push(Bool(false))
			return
		}
		list := NewAnyListWithCapacity(len(mixins))
		for _, m := range mixins {
			list.Append(m)
		}
		f.RP.DataStack.Push(list)
		f.RP.DataStack.Push(Bool(true))
	})
	// ( -- t:Type true | false ) — false at ProtoType sentinel / chain end (GetParent).
	t.registerNativeMethod("getParentType", func(f *RunFrame, ths *TypeObjPair) {
		if p := ths.Obj.(*Type).GetParent(); p != nil {
			f.RP.DataStack.Push(p)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	// ( name:s -- member:prop|meth|items true | false ) — local members only (no prototype walk).
	// "__items" returns this type's own itemsMetadata when present.
	t.registerNativeMethod("getMemberForDoc", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		typ := ths.Obj.(*Type)
		ds := f.RP.DataStack
		if name == "__items" {
			if im := typ.itemsMetadata; im != nil {
				ds.Push(im)
				ds.Push(Bool(true))
				return
			}
			ds.Push(Bool(false))
			return
		}
		if typ.propertyMetadata.HasKey(name) {
			ds.Push(typ.propertyMetadata.Get(name))
			ds.Push(Bool(true))
			return
		}
		if typ.methodMetadata.HasKey(name) {
			ds.Push(typ.methodMetadata.Get(name))
			ds.Push(Bool(true))
			return
		}
		if typ.classPropertyMetadata.HasKey(name) {
			ds.Push(typ.classPropertyMetadata.Get(name))
			ds.Push(Bool(true))
			return
		}
		if typ.classMethodMetadata.HasKey(name) {
			ds.Push(typ.classMethodMetadata.Get(name))
			ds.Push(Bool(true))
			return
		}
		ds.Push(Bool(false))
	})
	t.registerNativeMethod("getClearLastComment", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		comment := f.RP.Parser.GetClearLastComment()
		ds.Push(String(comment))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	typeType = t
}

func CreateRegisterPropertyMetadataType() {
	t := &Type{
		name:            "PropertyMetadata",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[PropertyMetadata]("PropertyMetadata"),
	}
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*PropertyMetadata).name) },
		nil,
	)
	t.registerProperty("type", "Type",
		func(inst Any) Any { return inst.(*PropertyMetadata)._type },
		nil,
	)
	t.registerProperty("hasGet", "Bool",
		func(inst Any) Any { return Bool(inst.(*PropertyMetadata).get != nil) },
		nil,
	)
	t.registerProperty("hasSet", "Bool",
		func(inst Any) Any { return Bool(inst.(*PropertyMetadata).set != nil) },
		nil,
	)
	t.registerProperty("doc", "String",
		func(inst Any) Any { return String(inst.(*PropertyMetadata).doc) },
		func(inst Any, value Any) { inst.(*PropertyMetadata).doc = string(value.(String)) },
	)
	// ( -- mixin:Type true | false )
	t.registerNativeMethod("tryGetMixinType", func(f *RunFrame, ths *TypeObjPair) {
		if m := ths.Obj.(*PropertyMetadata).mixinType; m != nil {
			f.RP.DataStack.Push(m)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	propertyMetadataType = t
}

func CreateRegisterItemsMetadataType() {
	t := &Type{
		name:            "ItemsMetadata",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[ItemsMetadata]("ItemsMetadata"),
	}
	t.registerProperty("itemType", "Type",
		func(inst Any) Any { return inst.(*ItemsMetadata)._itemType },
		nil,
	)
	t.registerProperty("hasGetByOrdinal", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).getByOrdinal != nil) },
		nil,
	)
	t.registerProperty("hasSetByOrdinal", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).setByOrdinal != nil) },
		nil,
	)
	t.registerProperty("hasGetByKey", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).getByKey != nil) },
		nil,
	)
	t.registerProperty("hasSetByKey", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).setByKey != nil) },
		nil,
	)
	t.registerProperty("hasByKey", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).hasByKey != nil) },
		nil,
	)
	t.registerProperty("hasGetByIndices", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).getByMultiIndex != nil) },
		nil,
	)
	t.registerProperty("hasSetByIndices", "Bool",
		func(inst Any) Any { return Bool(inst.(*ItemsMetadata).setByMultiIndex != nil) },
		nil,
	)
	t.registerProperty("multiIndexArity", "Int",
		func(inst Any) Any { return Int(inst.(*ItemsMetadata).multiIndexArity) },
		nil,
	)
	t.registerProperty("doc", "String",
		func(inst Any) Any { return String(inst.(*ItemsMetadata).doc) },
		func(inst Any, value Any) { inst.(*ItemsMetadata).doc = string(value.(String)) },
	)
	GetTypeRegistrar().RegisterCurrentType(t)
	itemsMetadataType = t
}

func CreateRegisterMethodMetadataType() {
	t := &Type{
		name:            "MethodMetadata",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[MethodMetadata]("MethodMetadata"),
	}
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*MethodMetadata).name) },
		nil,
	)
	t.registerProperty("isFinalMethod", "Bool",
		func(inst Any) Any { return Bool(inst.(*MethodMetadata).isFinalMethod) },
		nil,
	)
	t.registerProperty("doc", "String",
		func(inst Any) Any { return String(inst.(*MethodMetadata).doc) },
		func(inst Any, value Any) { inst.(*MethodMetadata).doc = string(value.(String)) },
	)
	// ( -- mixin:Type true | false )
	t.registerNativeMethod("tryGetMixinType", func(f *RunFrame, ths *TypeObjPair) {
		if m := ths.Obj.(*MethodMetadata).mixinType; m != nil {
			f.RP.DataStack.Push(m)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	methodMetadataType = t
}

func init() {
	CreateRegisterTypeType() // before metadata types that reference "Type"
	CreateRegisterPropertyMetadataType()
	CreateRegisterItemsMetadataType()
	CreateRegisterMethodMetadataType()
}

var (
	_ Typed = (*PropertyMetadata)(nil)
	_ Typed = (*ItemsMetadata)(nil)
	_ Typed = (*MethodMetadata)(nil)
)
