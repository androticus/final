package core

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultRunDataStackMaxSize is the main user-visible computation stack.
	DefaultRunDataStackMaxSize = 1024 // note: also used for OStack, to handle recursive Final method calls
	// DefaultRunReturnStackMaxSize holds Nest/Unnest return addresses.
	DefaultRunReturnStackMaxSize = 1024
	// DefaultRunTempStackMaxSize is compiler/axiom scratch (-->ts / <--ts).
	DefaultRunTempStackMaxSize = 256

	// local caps are sizable, because bad user loops
	// shouldn't actually create words or emit code
	localTaskDictMaxSize = 4096
	localTaskDictCap     = 64
	localTaskCTMaxSize   = 1024 * 32
	localTaskCTCap       = 256
	localTaskDTMaxSize   = 1024 * 8
	localTaskDTCap       = 32
)

// Task is shared runtime shell for one Final program context: compile target, ip,
// and coordination for concurrent runs. Per-run stacks and stdio live in RunProvider.
type Task struct {
	name string // for error messages and debugging

	ip VirtualAddr // instruction pointer (VirtCT space)

	// Compile target: ModuleID for module-manager APIs (table slot or 0 = "this" sentinel).
	compileModID ModuleID

	dc            *DictionaryContext // Task compile/resolve DC. VirtAddr get/set use this task scope.
	compilerState *CompilerState     // compile session for compileModID; always non-nil

	// pendingExecWord runs one dictionary word at runLoop entry (Go-hosted dispatch outside in-VM repl).
	pendingExecWord *Word
	// pendingExecCS is the context-stack selector when pendingExecWord.RequiresStackContext(); ignored otherwise.
	pendingExecCS ContextStackIndex
	// pendingForceOp, if non-zero, is armed as forceOp at runLoop entry then cleared (no CT Append).
	// Used by runInterruptNest for interpret-time Final methods on a bare RunFrame.
	pendingForceOp OpCode
	forceOp        *OpCode // active only during runLoop; see Task.Break
	// pendingThrow is set by Throw() when OpThrow is armed; cleared by OpThrow and WarmBoot.
	pendingThrow bool
	// rsNestFloor is the RS depth a nested Run (runPendingExec / runInterruptNest) must not
	// unnest below. 0 means no nested Run: throw still leaves the last session floor frame.
	rsNestFloor int

	// doneChan receives nil on clean Run exit or *RunFault on panic; nil means re-panic after capture.
	// One active Run per Task (subTask model); leave nil when unused.
	doneChan chan error

	// advisoryTimeoutS is the run-budget hint for the current Send/batch (seconds; 0 = none).
	// Set by Session worker before interpretSession; cleared after the run completes.
	advisoryTimeoutS int

	// restrict is the active SecurityRestrictions disabled-set for this Task
	// (compiler and runtime SR checks). Send sets it to Session.restrict | sender extra.
	// Spawned pips copy this value at CreateSubTask; already-running pips are not updated.
	restrict SecurityRestrictions

	// systemMode allows Find/exec of WordFlagSystem words. True only on core-compile
	// Tasks (CreateCoreTask, compileBootstrap*). Local sessions and pips stay false.
	systemMode bool

	// Pip tree (Phase 2): Session root has parent==nil and pipId==0.
	parent     *Task
	pipId      int
	childrenMu sync.Mutex
	children   map[int]*Pip
	nextPipID  int // last allocated child id; next is nextPipID+1

	// descendantCount is live descendants (not self). Direct kids = len(children).
	descendantCount atomic.Int64
	// runFinished is true after this Task's Run returned (reap when also descendantCount==0).
	runFinished bool

	// Session-root fault snapshot (first fault wins).
	sessionFaultOnce  sync.Once
	sessionFault      *RunFault
	sessionFaultPipId int

	// pls is Pip Local Storage — sticky DynamicObject for this Task (pip).
	pls *DynamicObject

	// session is the owning Session when this Task was minted via NewSession / Pip spawn; nil for Session-less Tasks.
	session *Session
	// currentSearchOrder is the Task’s ModuleReference list (often shared with Session.modules until using forks).
	currentSearchOrder *ModuleReferenceList
}

// GetPls returns Pip Local Storage, creating it if needed.
func (t *Task) GetPls() *DynamicObject {
	if t == nil {
		panic("GetPls: nil Task")
	}
	if t.pls == nil {
		t.pls = NewDynamicObject()
	}
	return t.pls
}

// removePip deletes a child pip handle from this Task's registry (thread-safe).
func (t *Task) removePip(id int) {
	if t == nil {
		return
	}
	t.childrenMu.Lock()
	defer t.childrenMu.Unlock()
	delete(t.children, id)
}

// hasPip reports whether id is still registered under this Task.
func (t *Task) hasPip(id int) bool {
	if t == nil {
		return false
	}
	t.childrenMu.Lock()
	defer t.childrenMu.Unlock()
	_, ok := t.children[id]
	return ok
}

// compilerForTarget returns the Task's CompilerState bound to dc.
func (t *Task) compilerForTarget() *CompilerState {
	return t.compilerState
}

// Break requests cooperative stop of an active runLoop by forcing OpStop on the next
// opcode boundary. No-op when runLoop is not active (forceOp nil).
func (t *Task) Break() {
	if t != nil && t.forceOp != nil {
		*t.forceOp = OpStop
	}
}

// TypeObjPair is a dispatch couplet: DS cell (Obj) and authoritative Type (Typ).
type TypeObjPair struct {
	Obj any
	Typ *Type
}

// NewTypeObjPair is the sole constructor for *TypeObjPair — do not allocate literals elsewhere.
func NewTypeObjPair(obj any) *TypeObjPair {
	if obj == nil {
		return &TypeObjPair{Obj: nil, Typ: nullType}
	}
	if objt, ok := obj.(Typed); ok {
		typ := objt.FinalType()
		if typ == nil {
			panic("NewTypeObjPair: object has no type")
		}
		return &TypeObjPair{Obj: obj, Typ: typ}
	} else {
		switch obj.(type) {
		case Bool:
			return &TypeObjPair{Obj: obj, Typ: boolType}
		case Int:
			return &TypeObjPair{Obj: obj, Typ: intType}
		case Uint:
			return &TypeObjPair{Obj: obj, Typ: uintType}
		case Float:
			return &TypeObjPair{Obj: obj, Typ: floatType}
		case String:
			return &TypeObjPair{Obj: obj, Typ: stringType}
		case time.Time:
			return &TypeObjPair{Obj: obj, Typ: dateTimeType}
		default:
			if typ := lookupHostType(obj); typ != nil {
				return &TypeObjPair{Obj: obj, Typ: typ}
			}
			panic("NewTypeObjPair: object is not typed")
		}
	}
}

// RunProvider holds execution stacks and stdio streams for one Run invocation.
// Run copies these to local variables; it does not dereference the provider per opcode.
type RunProvider struct {
	DataStack   *Stack[any]
	ReturnStack *ReturnStackType
	TempStack   *Stack[any]
	XStack      *Stack[any]
	YStack      *Stack[any]
	ZStack      *Stack[any]
	OStack      *Stack[any]
	FindScope   *FindScope // L0 name resolution chain; bound per run (global base + optional local layers)
	Parser      *Parser    // run-scoped lexer; always non-nil (GetNullParser by default)
	Stdin       TextRead
	Stdout      TextWrite
	Stderr      TextWrite

	// LastOp / LastIP / LastExecDC are published on panic for Session/REPL outer recoverers (see runLoop defer).
	LastOp     OpCode
	LastIP     IP
	LastExecDC *DictionaryContext
}

// NewRunProvider returns stacks and shared null stdio streams for a primary run.
func NewRunProvider() *RunProvider {
	return NewRunProviderWithLimits(
		DefaultRunDataStackMaxSize,
		DefaultRunReturnStackMaxSize,
		DefaultRunTempStackMaxSize,
	)
}
func NewRunProviderWithStdio(stdin TextRead, stdout TextWrite, stderr TextWrite) *RunProvider {
	rp := NewRunProvider()
	rp.Stdin = stdin
	rp.Stdout = stdout
	rp.Stderr = stderr
	return rp
}

// NewRunProviderWithLimits builds a provider with explicit stack max depths and null stdio.
func NewRunProviderWithLimits(dataMax, returnMax, auxMax int) *RunProvider {
	rp := &RunProvider{
		DataStack:   NewStack[any]("dataStack", dataMax),
		ReturnStack: NewReturnStack("returnStack", returnMax),
		TempStack:   NewAuxStack[any]("tStack", auxMax),
		XStack:      NewAuxStack[any]("xStack", auxMax),
		YStack:      NewAuxStack[any]("yStack", auxMax),
		ZStack:      NewAuxStack[any]("zStack", auxMax),
		OStack:      NewAuxStack[any]("oStack", returnMax),
		Parser:      GetNullParser(),
		Stdin:       GetNullTextReader(),
		Stdout:      GetNullTextWriter(),
		Stderr:      GetNullTextWriter(),
	}
	if globalDictionaryContext != nil {
		rp.ResetFindScopeBase(globalDictionaryContext)
	}
	return rp
}

// CreateCoreTask returns a Task whose compile target is the sealed core module.
// Core is always CoreModuleSlot (1) in ModuleTable; slot 0 is reserved nil.
func CreateCoreTask() *Task {
	dc := ModuleTable.DCForModule(nil, CoreModuleSlot)
	t := newTaskForModule("core", CoreModuleSlot, dc)
	t.currentSearchOrder = newCoreOnlySearchOrder()
	t.systemMode = true
	return t
}

// CreateLocalTask registers a fresh module (own DictionaryContext) named local:<random int64>
// and returns a Task that compiles into that module.
//
// Each call allocates the next ModuleTable slot (≥2); the id is not predictable and must
// not be assumed. Many local modules may coexist — one per task/thread — and AddModule
// is safe for concurrent registration.
//
// Initializes currentSearchOrder / Module.requires to [local, core]. Session.modules is
// wired when NewSession wraps the Task.
func CreateLocalTask() *Task {
	if globalDictionaryContext == nil {
		panic("CreateLocalTask: global execution environment not initialized")
	}
	name := newLocalModuleName()
	dc := NewDictionaryContext(
		name,
		localTaskDictMaxSize, localTaskDictCap,
		localTaskCTMaxSize, localTaskCTCap,
		localTaskDTMaxSize, localTaskDTCap,
		NewModuleTypeRegistrar(GetTypeRegistrar()),
	)
	mod := NewModule(name, dc)
	mod.prefix = "local"
	ModuleTable.AddModule(mod)
	t := newTaskForModule(name, mod.moduleId, mod.dc)
	t.currentSearchOrder = newLocalCoreSearchOrder(mod)
	return t
}

// NewTask creates a Task with zero execution defaults. RunProvider supplies per-run
// stacks and stdio; this stub wires global compile target and ip only.
func NewTask(makeDoneChan bool) (*Task, <-chan error) {
	t := CreateCoreTask()
	if makeDoneChan {
		t.doneChan = make(chan error, 1)
	}
	return t, t.doneChan
}

func newTaskForModule(name string, modID ModuleID, dc *DictionaryContext) *Task {
	if dc == nil {
		panic(fmt.Sprintf("newTaskForModule(%q): nil dc", name))
	}
	return &Task{
		name:          name,
		ip:            EncodeVirt(ModuleIDThis, 0),
		compileModID:  modID,
		dc:            dc,
		compilerState: NewCompilerState(dc),
	}
}

func newLocalModuleName() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(fmt.Errorf("newLocalModuleName: rand: %w", err))
	}
	n := int64(binary.BigEndian.Uint64(buf[:]))
	return fmt.Sprintf("local:%d", n)
}

// SetParser binds the run-scoped session lexer. Panics on nil Parser.
func (rp *RunProvider) SetParser(p *Parser) {
	if p == nil {
		panic("SetParser: nil Parser")
	}
	rp.Parser = p
}

// Resolve materializes a dispatch couplet for inst via NewTypeObjPair.
func (rp *RunProvider) Resolve(inst any) *TypeObjPair {
	if rp == nil {
		panic("RunProvider.Resolve: nil RunProvider")
	}
	return NewTypeObjPair(inst)
}
