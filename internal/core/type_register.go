package core

import (
	"fmt"
	"sync"
)

var (
	typeRegistrar     *TypeRegistrar
	typeRegistrarOnce sync.Once
)

func GetTypeRegistrar() *TypeRegistrar {
	typeRegistrarOnce.Do(func() {
		typeRegistrar = NewTypeRegistrar()
	})
	return typeRegistrar
}

// GetCachedType returns a cached type by name
// The first access will populate the cache.
// Subsequent accesses will return the cached value.
// A good alternative is to use a *Type variable in your file,
// and assign it during init, since your Type now exists.
func GetCachedType(name string, cache **Type) *Type {
	if *cache != nil {
		return *cache
	}
	*cache = GetTypeRegistrar().Get(name)
	return *cache
}

// TypeResolutionKind identifies what metadata slot a resolution record patches (panic context).
type TypeResolutionKind int

const (
	TypeResolutionKindClass TypeResolutionKind = iota
	TypeResolutionKindProperty
	TypeResolutionKindItems
	TypeResolutionKindMethod
)

func (k TypeResolutionKind) String() string {
	switch k {
	case TypeResolutionKindClass:
		return "class"
	case TypeResolutionKindProperty:
		return "property"
	case TypeResolutionKindItems:
		return "items"
	case TypeResolutionKindMethod:
		return "method"
	default:
		return fmt.Sprintf("TypeResolutionKind(%d)", int(k))
	}
}

type TypeResolutionRecord struct {
	Kind           TypeResolutionKind
	ContainingType *Type // type whose metadata field is being filled (panic context)
	TypeName       string
	TypeSlot       **Type
}

type TypeResolutionRecordList struct {
	items []*TypeResolutionRecord
}

type TypeRegisterRecord struct {
	Type              *Type
	ResolutionRecords *TypeResolutionRecordList
}

func NewTypeRegisterRecord() *TypeRegisterRecord {
	return &TypeRegisterRecord{}
}

// AddResolution appends a resolution to this registration, lazily allocating the list.
func (r *TypeRegisterRecord) AddResolution(record *TypeResolutionRecord) {
	if r.ResolutionRecords == nil {
		r.ResolutionRecords = &TypeResolutionRecordList{}
	}
	r.ResolutionRecords.Add(record)
}

func (l *TypeResolutionRecordList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

func (l *TypeResolutionRecordList) Add(record *TypeResolutionRecord) {
	if record == nil {
		panic(fmt.Errorf("TypeResolutionRecordList: nil record"))
	}
	if record.ContainingType == nil {
		panic(fmt.Errorf("TypeResolutionRecordList: nil containing type"))
	}
	if record.ContainingType.name == "" {
		panic(fmt.Errorf("TypeResolutionRecordList: empty containing type name"))
	}
	if record.TypeName == "" {
		panic(fmt.Errorf("TypeResolutionRecordList: empty dependent type name"))
	}
	l.items = append(l.items, record)
}

// TypeCatalog is a catalog of all types defined in a module, including core.
// It is used to lookup types by name. typeIdx is a local, 0-based index into
// this catalog; module base routing lives in the VirtTypeIdx (see NewTypeWord).
// Backed by OrdinalMap[*Type]: names are unique, and a duplicate/empty add panics.
type TypeCatalog struct {
	m *OrdinalMap[*Type]
}

func NewTypeCatalog() *TypeCatalog {
	return &TypeCatalog{m: NewOrdinalMap[*Type]()}
}

func (c *TypeCatalog) Add(name string, t *Type) {
	if t == nil {
		panic(fmt.Errorf("TypeCatalog: nil type"))
	}
	t.typeIdx = c.m.Len()
	c.m.Add(name, t) // panics on empty or duplicate name
}

// Get returns the Type with the given name, or nil if not found.
func (c *TypeCatalog) Get(name string) *Type {
	if t, ok := c.m.GetByName(name); ok {
		return t
	}
	return nil
}

// GetByTypeIdx returns the Type registered at local idx, or nil if out of range.
func (c *TypeCatalog) GetByTypeIdx(idx int) *Type {
	if idx < 0 || idx >= c.m.Len() {
		return nil
	}
	return c.m.GetByOrdinal(idx)
}

// Types returns all registered types in catalog registration order.
func (c *TypeCatalog) Types() []*Type {
	out := make([]*Type, c.m.Len())
	for i := range out {
		out[i] = c.m.GetByOrdinal(i)
	}
	return out
}

// Len returns the number of types in the catalog.
func (c *TypeCatalog) Len() int {
	if c == nil || c.m == nil {
		return 0
	}
	return c.m.Len()
}

// RollbackTo removes catalog entries with typeIdx >= start (inclusive cut of the tail).
// Types leaving the catalog have Final methods poisoned so leftover MethodRefs cannot
// nest into truncated or reissued CT.
func (c *TypeCatalog) RollbackTo(start int) {
	if c == nil || c.m == nil {
		panic(fmt.Errorf("TypeCatalog.RollbackTo: nil catalog"))
	}
	if start < 0 || start > c.m.Len() {
		panic(fmt.Errorf("TypeCatalog.RollbackTo: start %d out of range, len %d", start, c.m.Len()))
	}
	for i := c.m.Len() - 1; i >= start; i-- {
		t := c.m.GetByOrdinal(i)
		if t != nil {
			t.poisonForgottenFinalMethods()
			delete(c.m.byName, t.name)
		}
	}
	c.m.byOrdinal = c.m.byOrdinal[:start]
}

type TypeRegistrar struct {
	catalog *TypeCatalog
	// parent is the fallback registrar for by-name lookups (module registrar -> core global); nil for core.
	parent *TypeRegistrar
	// currentRegistration accumulates resolution records for the type being built until RegisterCurrentType.
	currentRegistration *TypeRegisterRecord
	// registrationsToResolve holds bootstrap batch registrations until ResolveRegistrations runs.
	registrationsToResolve []*TypeRegisterRecord
}

func NewTypeRegistrar() *TypeRegistrar {
	return &TypeRegistrar{
		catalog:                NewTypeCatalog(),
		currentRegistration:    NewTypeRegisterRecord(),
		registrationsToResolve: make([]*TypeRegisterRecord, 0, 128),
	}
}

// NewModuleTypeRegistrar creates a per-module registrar whose by-name lookups fall
// back to parent (typically the core global registrar) so module types can reference
// core field types (String, Int, …). typeIdx stays local; module base routing is in
// the VirtTypeIdx, so GetByTypeIdx is local-only.
func NewModuleTypeRegistrar(parent *TypeRegistrar) *TypeRegistrar {
	r := NewTypeRegistrar()
	r.parent = parent
	return r
}

// Get returns a registered type by CapCase name, falling back to the parent registrar.
func (r *TypeRegistrar) Get(name string) *Type {
	if t := r.catalog.Get(name); t != nil {
		return t
	}
	if r.parent != nil {
		return r.parent.Get(name)
	}
	return nil
}

// GetLocal returns a type from this registrar's catalog only — no parent fallback.
// Used for Final method tear-off ownership (decorate only types owned by the compiling DC).
func (r *TypeRegistrar) GetLocal(name string) *Type {
	if r == nil {
		return nil
	}
	return r.catalog.Get(name)
}

// GetByTypeIdx returns a registered type by local catalog index, or nil if out of range.
// No parent fallback: the module base in the VirtTypeIdx already selects the owning catalog.
func (r *TypeRegistrar) GetByTypeIdx(idx int) *Type {
	return r.catalog.GetByTypeIdx(idx)
}

// Types returns all registered types in catalog registration order.
func (r *TypeRegistrar) Types() []*Type {
	return r.catalog.Types()
}

// Len returns the local catalog length (no parent).
func (r *TypeRegistrar) Len() int {
	if r == nil || r.catalog == nil {
		return 0
	}
	return r.catalog.Len()
}

// RollbackTo truncates the local type catalog to typeKeep and drops pending
// resolve records for types that no longer exist in the catalog.
func (r *TypeRegistrar) RollbackTo(typeKeep int) {
	if r == nil || r.catalog == nil {
		panic(fmt.Errorf("TypeRegistrar.RollbackTo: nil registrar"))
	}
	r.catalog.RollbackTo(typeKeep)
	if len(r.registrationsToResolve) == 0 {
		return
	}
	kept := r.registrationsToResolve[:0]
	for _, rec := range r.registrationsToResolve {
		if rec == nil || rec.Type == nil {
			continue
		}
		if rec.Type.typeIdx < typeKeep && r.catalog.Get(rec.Type.name) == rec.Type {
			kept = append(kept, rec)
		}
	}
	r.registrationsToResolve = kept
}

// AddResolution appends a deferred type ref to the in-flight registration for the type under construction.
// Safe from any file init(); registration order across files is not relied upon.
func (r *TypeRegistrar) AddResolution(record *TypeResolutionRecord) {
	r.currentRegistration.AddResolution(record)
}

// RegisterCurrentType registers the in-flight record: catalog add, then queue metadata refs for resolve.
func (r *TypeRegistrar) RegisterCurrentType(typeBeingBuilt *Type) {
	if typeBeingBuilt == nil {
		panic(fmt.Errorf("TypeRegistrar.RegisterCurrentType: nil type"))
	}
	record := r.currentRegistration
	record.Type = typeBeingBuilt
	if record.Type.name == "" {
		panic(fmt.Errorf("TypeRegistrar.RegisterCurrentType: empty type name"))
	}
	r.catalog.Add(record.Type.name, record.Type)
	if record.ResolutionRecords.Len() > 0 {
		r.registrationsToResolve = append(r.registrationsToResolve, record)
	}
	r.currentRegistration = NewTypeRegisterRecord()
}

// ResolveRegistrations patches all bootstrap-queued metadata refs.
// Call exactly once from init, then call once again for a single or batch of dynamic registrations
func (r *TypeRegistrar) ResolveRegistrations() {
	for _, registration := range r.registrationsToResolve {
		r.resolveRegistrationRecord(registration)
	}
	// let GC collect the last batch of registrations
	// create a new slice, with just default capacity for dynamic registrations
	r.registrationsToResolve = make([]*TypeRegisterRecord, 0)
}

func (r *TypeRegistrar) resolveRegistrationRecord(registration *TypeRegisterRecord) {
	for _, record := range registration.ResolutionRecords.items {
		typeToResolve := r.Get(record.TypeName) // own catalog, then parent
		if typeToResolve == nil {
			panic(fmt.Errorf(
				"TypeRegistrar: %s on type %q needs unknown type %q",
				record.Kind, record.ContainingType.name, record.TypeName,
			))
		}
		*record.TypeSlot = typeToResolve
	}
}
