package core

import (
	"fmt"
	"strings"
	"sync"
)

// CoreModuleSlot is the module-table slot for the sealed core module (id 1).
// Slot 0 is reserved (nil) in the table list.
// ModuleIDThis (0) remains the compile/run "this" sentinel — not a table row.
const CoreModuleSlot ModuleID = 1

// Initialized early in global Init.
var ModuleTable = NewModules()

var coreModuleRegistered bool

type Module struct {
	name     string
	version  string // concrete _MODULE_.version when installed; "" for core/local shells
	moduleId ModuleID
	dc       *DictionaryContext
	prefix   string               // catalog default: "core", "local", or last path segment
	requires *ModuleReferenceList // sealed: parents; Session-local: also Session.modules ([local, core, …imports])
}

// ModuleReference is one search-order / requires attach: catalog Module + Session-effective prefix.
type ModuleReference struct {
	module *Module
	prefix string
}

type ModuleList = List[*Module]
type ModuleReferenceList = List[*ModuleReference]

func NewModuleList() *ModuleList {
	l := &ModuleList{items: make([]*Module, 1, 16)}
	//l.ListType = GetTypeRegistrar().Get("ModuleList")
	//l.CoerceGetAnyFn = MakeDefaultValueCoerceGetFn[*Module]()
	//l.CoerceSetFromAnyFn = MakeDefaultValueCoerceSetFn[*Module]()
	return l
}

func NewModuleReferenceList() *ModuleReferenceList {
	return &ModuleReferenceList{items: make([]*ModuleReference, 0, 8)}
}

func NewModule(name string, dc *DictionaryContext) *Module {
	return &Module{name: name, dc: dc}
}

func NewModuleReference(mod *Module, prefix string) *ModuleReference {
	if mod == nil {
		panic("NewModuleReference: nil module")
	}
	if prefix == "" {
		prefix = mod.prefix
	}
	return &ModuleReference{module: mod, prefix: prefix}
}

// RegisterCoreModule adds the global core DC as module "core" at CoreModuleSlot (1).
// Idempotent — safe across test resets that reuse the same globalDictionaryContext.
func RegisterCoreModule(dc *DictionaryContext) {
	if dc == nil {
		panic("RegisterCoreModule: nil dc")
	}
	if coreModuleRegistered {
		if m := ModuleTable.moduleByName("core"); m == nil || m.dc != dc {
			panic("RegisterCoreModule: core already registered with a different DC")
		}
		if m := ModuleTable.moduleByName("core"); m != nil && m.prefix == "" {
			m.prefix = "core"
		}
		return
	}
	m := NewModule("core", dc)
	m.prefix = "core"
	ModuleTable.AddModule(m)
	coreModuleRegistered = true
}

// newLocalCoreSearchOrder builds [localRef, coreRef] and stores it on local.requires.
func newLocalCoreSearchOrder(local *Module) *ModuleReferenceList {
	if local == nil {
		panic("newLocalCoreSearchOrder: nil local module")
	}
	core := ModuleTable.moduleByName("core")
	if core == nil {
		panic("newLocalCoreSearchOrder: core module not registered")
	}
	if local.prefix == "" {
		local.prefix = "local"
	}
	if core.prefix == "" {
		core.prefix = "core"
	}
	list := NewModuleReferenceList()
	list.Append(NewModuleReference(local, "local"))
	list.Append(NewModuleReference(core, "core"))
	local.requires = list
	return list
}

func newCoreOnlySearchOrder() *ModuleReferenceList {
	core := ModuleTable.moduleByName("core")
	if core == nil {
		panic("newCoreOnlySearchOrder: core module not registered")
	}
	if core.prefix == "" {
		core.prefix = "core"
	}
	list := NewModuleReferenceList()
	list.Append(NewModuleReference(core, "core"))
	return list
}

func findModuleRefByPrefix(list *ModuleReferenceList, prefix string) *ModuleReference {
	if list == nil {
		return nil
	}
	for i := 0; i < list.Len(); i++ {
		ref := list.At(i)
		if ref != nil && ref.prefix == prefix {
			return ref
		}
	}
	return nil
}

func (mt *Modules) moduleByName(name string) *Module {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.byName[name]
}

func (mt *Modules) moduleByID(id ModuleID) *Module {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.moduleAtLocked(id)
}

type Modules struct {
	mu      sync.RWMutex
	modules *ModuleList
	byName  map[string]*Module
}

func NewModules() *Modules {
	return &Modules{modules: NewModuleList(), byName: make(map[string]*Module)}
}

// AddModule registers a new module. Holds the table write lock.
// Panics on duplicate name. Assigns the next table slot (≥ CoreModuleSlot).
func (mt *Modules) AddModule(module *Module) {
	if module == nil {
		panic("AddModule: nil module")
	}
	if module.name == "" {
		panic("AddModule: empty name")
	}
	if module.dc == nil {
		panic("AddModule: nil dc")
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.byName[module.name] != nil {
		panic(fmt.Errorf("AddModule: duplicate module name %q", module.name))
	}
	module.moduleId = ModuleID(mt.modules.Len())
	module.dc.moduleId = module.moduleId
	if module.dc.dict != nil {
		module.dc.dict.moduleId = module.moduleId
	}
	mt.modules.Append(module)
	mt.byName[module.name] = module
}

// FindOnPath resolves a name along the task search path. Not implemented yet.
// Find is lexical (search order → *Word with moduleId); not VirtAddr-based. See search-system work.
func (mt *Modules) FindOnPath(t *Task, name string) *Word {
	panic("FindOnPath: not implemented")
}

// DCForModule returns the DictionaryContext for an absolute table slot (≥ 1).
// moduleId 0 (execute DC) is not implemented yet. t is unused until manifest exists.
func (mt *Modules) DCForModule(t *Task, slotID ModuleID) *DictionaryContext {
	_ = t
	if slotID == ModuleIDThis {
		panic("DCForModule: moduleId 0 not implemented")
	}
	mt.mu.RLock()
	mod := mt.moduleAtLocked(slotID)
	mt.mu.RUnlock()
	if mod == nil {
		panic(fmt.Errorf("DCForModule: unknown module slot %d", slotID))
	}
	return mod.dc
}

// ResolveDT resolves a module DT backing for moduleId and plain slot index.
// moduleId 0 uses execDC.dt; ≥1 uses the module table. No table lock is held after return.
func (mt *Modules) ResolveDT(t *Task, execDC *DictionaryContext, moduleId ModuleID, idx int) (*DataTable, int) {
	if moduleId == ModuleIDThis {
		if execDC == nil || execDC.dt == nil {
			panic(fmt.Errorf("ResolveDT: nil execute DC DT for moduleId 0"))
		}
		return execDC.dt, idx
	}
	dc := mt.DCForModule(t, moduleId)
	if dc == nil || dc.dt == nil {
		panic(fmt.Errorf("ResolveDT: nil DT for moduleId %d", moduleId))
	}
	return dc.dt, idx
}

func (mt *Modules) moduleAtLocked(id ModuleID) *Module {
	if id <= 0 {
		return nil
	}
	i := int(id)
	if i >= mt.modules.Len() {
		return nil
	}
	return mt.modules.items[i]
}

// finishInstall renames a compile-Session local module into a sealed loadable row.
// Caller must Seal the DC before calling. Holds the table write lock.
func (mt *Modules) finishInstall(mod *Module, name, version, prefix string) {
	if mod == nil {
		panic("finishInstall: nil module")
	}
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	prefix = strings.TrimSpace(prefix)
	if name == "" {
		panic("finishInstall: empty name")
	}
	if prefix == "" {
		panic("finishInstall: empty prefix")
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.byName[name] != nil && mt.byName[name] != mod {
		panic(fmt.Errorf("finishInstall: duplicate module name %q", name))
	}
	if old := mod.name; old != "" && mt.byName[old] == mod {
		delete(mt.byName, old)
	}
	mod.name = name
	mod.version = version
	mod.prefix = prefix
	mod.requires = NewModuleReferenceList() // leaf: no parent attaches on the catalog Module
	if mod.dc != nil {
		mod.dc.name = name
	}
	mt.byName[name] = mod
}

// abandonModule removes a failed compile-Session module from the table (tombstone slot).
func (mt *Modules) abandonModule(mod *Module) {
	if mod == nil {
		return
	}
	mt.mu.Lock()
	defer mt.mu.Unlock()
	if mt.byName[mod.name] == mod {
		delete(mt.byName, mod.name)
	}
	i := int(mod.moduleId)
	if i > 0 && i < mt.modules.Len() && mt.modules.items[i] == mod {
		mt.modules.items[i] = nil
	}
}
