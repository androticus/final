package core

import "fmt"

type ReservedDTSlot = int

const (
	ReservedDTSlotColdBootTarget = 0 // XT of ColdBoot word
	ReservedDTSlotWarmBootTarget = 1 // XT of WarmBoot repl loop entry
	ReservedDTSlotByeReq         = 2 // Bool: one-shot leave quit (core DT only; OpBootstrapDoVar)
	ReservedDTSlots              = 3
)

// Reserved low CT cells: cold/warm OpCtImmDs·dtIdx·OpBootstrapDefer chains, OpStop (see InitializeBootstrapCT).
const (
	ReservedCTBootColdEntry   = 0 // OpCtImmDs; global entry (BootstrapGlobalColdEntry)
	ReservedCTBootColdSlot    = 1 // dtIdx operand: ReservedDTSlotColdBootTarget
	ReservedCTBootColdDefer   = 2 // OpBootstrapDefer
	ReservedCTBootWarmCtImmDs = 3 // OpCtImmDs
	ReservedCTBootWarmSlot    = 4 // dtIdx operand: ReservedDTSlotWarmBootTarget
	ReservedCTBootWarmDefer   = 5 // OpBootstrapDefer
	ReservedCTBootStop        = 6 // OpStop after warm defer (session exit)
	ReservedCTBootBodiesStart = 7 // repl loop and compiled vocabulary follow
)

const (
	bootWordColdBoot     = "ColdBoot"
	bootWordWarmBoot     = "WarmBoot"
	bootWordWarmReplBody = "_bootWarmRepl"
)

// InitializeCoreReservedDTSlots prepends bootstrap reserved cells on the sealed core module DT only.
// Local module DCs must not call this — bootstrap var/defer words use OpBootstrapDoVar/Defer (core DT).
func InitializeCoreReservedDTSlots(dt *DataTable) {
	dt.Append(nil)   // ReservedDTSlotColdBootTarget — wired in wireBootstrapBootTargets
	dt.Append(nil)   // ReservedDTSlotWarmBootTarget
	dt.Append(false) // ReservedDTSlotByeReq
}

// used for locking the DC once a child Task is created
// Task attaches this to the child map table count
type DictionaryContextLockedFn = func() bool

// DictionaryContext bundles a Dictionary with its CodeTable (CT) and DataTable (DT).
// Used for global, module, and task-local compilation targets.
type DictionaryContext struct {
	name     string // for error messages
	moduleId ModuleID
	dict     *Dictionary
	ct       *CodeTable
	dt       *DataTable
	tr       *TypeRegistrar // for core/global, this is the global one
	// sealed is sticky compile-finished: no further dict/CT/DT growth until Unseal.
	sealed bool
	// lockedFn is set for multithread (pip) exclusion; true while children live.
	lockedFn DictionaryContextLockedFn
}

// NewDictionaryContext creates a context with fresh Dictionary, CodeTable, and DataTable.
// Each maxSize/cap pair is passed to that constituent's factory; table cap=0 means
// default autogrow (see NewCodeTable / NewDataTable).
func NewDictionaryContext(
	name string,
	dictMaxSize int, dictCap int,
	ctMaxSize int, ctCap int,
	dtMaxSize int, dtCap int,
	tr *TypeRegistrar,
) *DictionaryContext {
	dc := &DictionaryContext{
		name: name,
		dict: NewDictionary(name, dictMaxSize, dictCap),
		ct:   NewCodeTable(ctMaxSize, ctCap),
		dt:   NewDataTable(dtMaxSize, dtCap),
		tr:   tr,
	}
	dc.dict.dc = dc
	return dc
}

// IsLockedOrSealed reports whether compile/define mutations must be refused.
func (dc *DictionaryContext) IsLockedOrSealed() bool {
	if dc == nil {
		return false
	}
	return dc.sealed || (dc.lockedFn != nil && dc.lockedFn())
}

func (dc *DictionaryContext) panicIfLockedOrSealed(op string) {
	if dc == nil || !dc.IsLockedOrSealed() {
		return
	}
	reason := "sealed"
	if !dc.sealed {
		reason = "locked"
	}
	panic(fmt.Errorf("DictionaryContext %q is %s: cannot %s", dc.name, reason, op))
}

// Seal marks the DC compile-finished (sticky until Unseal).
func (dc *DictionaryContext) Seal() {
	if dc != nil {
		dc.sealed = true
	}
}

// Unseal clears the sealed flag (dev: final-go -core; tests after ResetTestRuntime).
func (dc *DictionaryContext) Unseal() {
	if dc != nil {
		dc.sealed = false
	}
}

func (dc *DictionaryContext) AllotDTIndex() int {
	dc.panicIfLockedOrSealed("allot DT")
	return dc.dt.Append(nil)
}

// FindWordByName implements Scope — dictionary word lookup in this DC only.
func (dc *DictionaryContext) FindWordByName(name string) *Word {
	if dc == nil || dc.dict == nil {
		return nil
	}
	return dc.dict.GetByName(name)
}

func (dc *DictionaryContext) HasName(name string) bool {
	if dc == nil || dc.dict == nil {
		return false
	}
	return dc.dict.HasName(name)
}

// WordXT returns the execution token for w when w belongs to this DC's dictionary.
func (dc *DictionaryContext) WordXT(w *Word) (XT, bool) {
	if dc == nil || dc.dict == nil || w == nil {
		return 0, false
	}
	return dc.dict.XTOf(w)
}

func (dc *DictionaryContext) FinalType() *Type { return GetTypeRegistrar().Get("DictionaryContext") }

func CreateRegisterDictionaryContextType() {
	t := &Type{
		name: "DictionaryContext",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			dc, ok := value.(*DictionaryContext)
			if !ok {
				panic(fmt.Errorf("DictionaryContext.checkCoerceType: value %T is not *DictionaryContext", value))
			}
			return dc
		},
	}
	t.registerProperty("dict", "Dictionary",
		func(inst Any) Any { return inst.(*DictionaryContext).dict },
		nil,
	)
	t.registerProperty("ct", "CodeTable",
		func(inst Any) Any { return inst.(*DictionaryContext).ct },
		nil,
	)
	t.registerProperty("dt", "DataTable",
		func(inst Any) Any { return inst.(*DictionaryContext).dt },
		nil,
	)
	t.registerNativeMethod("allotDTIndex", func(f *RunFrame, ths *TypeObjPair) {
		idx := ths.Obj.(*DictionaryContext).AllotDTIndex()
		f.RP.DataStack.Push(Int(idx))
	})
	t.registerNativeMethod("findWordByName", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		word := ths.Obj.(*DictionaryContext).FindWordByName(name)
		if word != nil {
			assertWordAllowed(f.Task, word)
			f.RP.DataStack.Push(word)
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
	t.registerNativeMethod("hasName", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		has := ths.Obj.(*DictionaryContext).HasName(name)
		f.RP.DataStack.Push(Bool(has))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}
