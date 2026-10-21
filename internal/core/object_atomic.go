package core

import "sync"

// AtomicDynamicObject is Final "AtomicDynamicObject": mutex-guarded DynamicObject
// for shared named slots across pips (string-key bag; v1 AtomicMap substitute).
type AtomicDynamicObject struct {
	mu sync.Mutex
	*DynamicObject
}

func NewAtomicDynamicObject() *AtomicDynamicObject {
	return &AtomicDynamicObject{DynamicObject: NewDynamicObject()}
}

func (o *AtomicDynamicObject) FinalType() *Type {
	return atomicDynamicObjectType
}

func (o *AtomicDynamicObject) Len() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.DynamicObject.Len()
}

func (o *AtomicDynamicObject) Get(name string) any {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.DynamicObject.Get(name)
}

func (o *AtomicDynamicObject) HasField(name string) Bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.DynamicObject.HasField(name)
}

func (o *AtomicDynamicObject) Set(name string, value any) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.DynamicObject.Set(name, value)
}

// ItemAt locks for the duration of the lookup only. Do not mutate (or use
// concurrently) while holding the returned *Item.
func (o *AtomicDynamicObject) ItemAt(i int) *Item[any] {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.DynamicObject.ItemAt(i)
}

// GetIter snapshots fields under the mutex (DynamicObject.GetIter clones).
func (o *AtomicDynamicObject) GetIter() Iterator {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.DynamicObject.GetIter()
}

func getAtomicDynamicObjectPropertyMetadataResolver(t *Type, name string) *PropertyMetadata {
	if t.GetMethodMetadata(name) != nil {
		return nil
	}
	return &PropertyMetadata{
		name: name,
		get: func(obj any) any {
			return obj.(*AtomicDynamicObject).Get(string(String(name)))
		},
		set: func(obj any, value any) {
			obj.(*AtomicDynamicObject).Set(string(String(name)), value)
		},
	}
}

func registerAtomicDynamicObjectMembers(t *Type) {
	t.setPropertyMetadataOverride(getAtomicDynamicObjectPropertyMetadataResolver)
	t.registerItemsData(
		"Any",
		nil, nil,
		func(inst Any, key Any) any {
			return inst.(*AtomicDynamicObject).Get(string(key.(String)))
		},
		func(inst Any, key Any, val Any) {
			inst.(*AtomicDynamicObject).Set(string(key.(String)), val)
		},
		func(inst Any, key Any) bool {
			return bool(inst.(*AtomicDynamicObject).HasField(string(key.(String))))
		},
	)
}

func CreateRegisterAtomicDynamicObjectType() {
	t := &Type{
		name:            "AtomicDynamicObject",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AtomicDynamicObject]("AtomicDynamicObject"),
		parent:          dynamicObjectType,
	}
	registerAtomicDynamicObjectMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAtomicDynamicObject())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	atomicDynamicObjectType = t
}

var atomicDynamicObjectType *Type

var (
	_ Typed    = (*AtomicDynamicObject)(nil)
	_ Iterable = (*AtomicDynamicObject)(nil)
)
