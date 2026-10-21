package core

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Bootstrap compile + interpreter helpers and runtime dispatch.
// Spec: docs/003-design-notes/_archived/105-runtime-dispatch.md
//
// Phase 1: interpret REPL + interim Forth-style `: … ;` colon compile (global GCD only).

// --- Word kinds (bootstrap) ---

type WordKind int

const (
	WKAtomic WordKind = iota
	WKColon
	WKNative
)

// --- Runtime dispatch ---

type RuntimeDispatchID int

const (
	TrFind RuntimeDispatchID = iota
	TrNextToken
	TrSessionDispatch
	TrReplError
	TrWarmReset
	TrCheckResetByeReq
	TrBootColon
	TrBootSemicolon
	TrBootSubwordOpen
	TrBootSubwordClose
	FinalFind
	FinalCurrentDictHasWord
	FinalReplError
	FinalGetNextToken
	FinalReadIdentifier
	FinalValidatePlainIdent
	FinalHandleCompound
	FinalCompileFromFile
	FinalCompileFromManifest
	FinalSetCompileState
	FinalFmt
	FinalMintVarWord
	FinalMintConstWord
	FinalMintDeferWord
	FinalForgetMark
	runtimeDispatchCount
)

type RuntimeDispatchFn func(f *RunFrame)

// RunFrame is passed to every runtime-dispatch handler (WKNative / OpTransition).
// One instance per Run, initialized with Task and RunProvider at loop entry.
// Idempotent — no per-invoke mutable dispatch state; ths is passed separately to Type.Invoke paths.
type RunFrame struct {
	Task *Task
	RP   *RunProvider
	// SubCode, ForceOp arm the next run-loop op from a native handler
	// (SetInterrupt, SetExec); do not use them for other purposes.
	SubCode *int
	ForceOp *OpCode
}

// BreakRequested reports whether a cooperative break is armed on the active runLoop
// (task.forceOp non-nil and not zero). False outside runLoop or when no break is pending.
func (f *RunFrame) BreakRequested() bool {
	if f == nil || f.Task == nil || f.Task.forceOp == nil {
		return false
	}
	return *f.Task.forceOp != 0
}

// AdvisoryTimeoutS returns the current run's advisory budget in whole seconds (0 = none).
// See Task.advisoryTimeoutS; set by Session.Send before each batch.
func (f *RunFrame) AdvisoryTimeoutS() int {
	if f == nil || f.Task == nil {
		return 0
	}
	return f.Task.advisoryTimeoutS
}

// SetInterrupt arms OpInterrupt on the next run-loop iteration (moduleID:i entry:i --).
// Final-method OS cleanup is compiled odrop before OpUnnest, not an Interrupt operand.
func (f *RunFrame) SetInterrupt(moduleID ModuleID, entry int) {
	if f == nil || f.RP == nil || f.ForceOp == nil {
		panic("SetInterrupt: RunFrame missing run-loop fields")
	}
	f.RP.DataStack.Push(Int(moduleID))
	f.RP.DataStack.Push(Int(entry))
	*f.ForceOp = OpInterrupt
}

// SetExec arms OpExec on the next run-loop iteration (xt on DS).
// Veneer for Word.exec — same path as the exec axiom; no separate dispatch.
func (f *RunFrame) SetExec(xt XT) {
	if f == nil || f.RP == nil || f.ForceOp == nil {
		panic("SetExec: RunFrame missing run-loop fields")
	}
	if xt == 0 {
		panic(fmt.Errorf("SetExec: VirtXT was 0"))
	}
	f.RP.DataStack.Push(xt)
	*f.ForceOp = OpExec
}

var runtimeDispatchTable [runtimeDispatchCount]RuntimeDispatchFn

// bootstrapEmitRuntimeDispatch appends OpDoSubCode · id + OpTransition to compile-target CT.
func bootstrapEmitRuntimeDispatch(cs *CompilerState, id RuntimeDispatchID) VirtualAddr {
	start := bootstrapEmitOpcode(cs, OpDoSubCode)
	_ = cs.emitCT(int(id))
	bootstrapEmitOpcode(cs, OpTransition)
	return start
}

// bootstrapEmitSubCodePrefix appends OpDoSubCode · code before a subCode worker.
func bootstrapEmitSubCodePrefix(cs *CompilerState, code int) VirtualAddr {
	start := bootstrapEmitOpcode(cs, OpDoSubCode)
	_ = cs.emitCT(code)
	return start
}

// --- Compile session ---

// CompileTxnMark records DC high-water at beginTransaction for unit rollback
// (dict, CT, DT, and local type catalog — same axes forget: stamps).
type CompileTxnMark struct {
	dictStart int
	ctStart   int
	dtStart   int
	typeStart int
}

// CompilerState holds ephemeral compile-session state for one DictionaryContext.
// Forth STATE lives in state (any, Bool). Context-stack selection (CS) is a run-loop binding only;
// d:/t:/x:/y: compound prefixes and OpDoCSSelect set cs at runtime — not a DT variable.
// Unit-level compile transactions use txnActive + txnMark (single level; no nesting).
type CompilerState struct {
	DC           *DictionaryContext
	compileState CompileStateFlags
	txnActive    bool
	txnMark      CompileTxnMark
}

// EmittingCT reports whether tokens compile into CT (STATE true) vs interpret now.
func (cs *CompilerState) EmittingCT() bool {
	if cs == nil {
		return false
	}
	return GetCompileState(cs)
}

func NewCompilerState(dc *DictionaryContext) *CompilerState {
	if dc == nil {
		panic("NewCompilerState: nil DictionaryContext")
	}
	return &CompilerState{DC: dc}
}

func (cs *CompilerState) beginTransaction() {
	if cs == nil {
		panic("beginTransaction: nil CompilerState")
	}
	if cs.txnActive {
		panic("beginTransaction: transaction already active")
	}
	dictStart, ctStart, dtStart, typeStart := cs.DC.stampForgetKeeps()
	cs.txnMark = CompileTxnMark{
		dictStart: dictStart,
		ctStart:   ctStart,
		dtStart:   dtStart,
		typeStart: typeStart,
	}
	cs.txnActive = true
}

func (cs *CompilerState) endTransaction(commit bool) {
	if cs == nil || !cs.txnActive {
		return
	}
	if !commit {
		m := cs.txnMark
		if cs.DC != nil {
			cs.DC.refuseCompileUnitRollbackIfSurvivingFinalMethods(m.ctStart, m.typeStart)
		}
		if cs.DC.tr != nil {
			cs.DC.tr.RollbackTo(m.typeStart)
		}
		cs.DC.ct.Truncate(m.ctStart)
		cs.DC.dt.Truncate(m.dtStart)
		cs.DC.dict.ClearBuilding()
		cs.DC.dict.RollbackTo(m.dictStart)
		SetCompileState(cs, false)
	}
	cs.txnActive = false
}

// tryRollbackTransaction rolls the compile unit back, or refuses like forget:
// when a surviving type has a Final method past ctStart, tables are left intact
// and the txn is abandoned (STATE false, building cleared) so the session can continue.
func (cs *CompilerState) tryRollbackTransaction() (err error) {
	if cs == nil || !cs.txnActive {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			cs.abandonCompileTxn()
			err = fmt.Errorf("%v", r)
		}
	}()
	cs.endTransaction(false)
	return nil
}

func (cs *CompilerState) abandonCompileTxn() {
	if cs == nil {
		return
	}
	if cs.DC != nil && cs.DC.dict != nil {
		cs.DC.dict.ClearBuilding()
	}
	SetCompileState(cs, false)
	cs.txnActive = false
}

func (cs *CompilerState) rollbackTransaction() {
	if err := cs.tryRollbackTransaction(); err != nil {
		panic(err)
	}
}

func (cs *CompilerState) txnActiveForTest() bool {
	return cs != nil && cs.txnActive
}

// bootstrapDictCreateEntry adds a dictionary entry of the given kind.
func bootstrapDictCreateEntry(cs *CompilerState, name string, kind WordKind) *Word {
	if kind != WKColon {
		panic(fmt.Errorf("bootstrapDictCreateEntry(%q, %v): only WKColon supported", name, kind))
	}
	return bootstrapCreate(cs, name)
}

func bootstrapCreate(cs *CompilerState, name string) *Word {
	if GetCompileState(cs) {
		panic(fmt.Errorf("bootstrapCreate(%q): already compiling", name))
	}
	if cs.DC.dict.HasName(name) {
		panic(fmt.Errorf("bootstrapCreate(%q): name already exists", name))
	}
	w := cs.DC.dict.MakeNewWord(name, WordFlagColon)
	w.moduleId = cs.DC.moduleId
	w.colonCode = CTP(cs.ct().Len())
	cs.DC.dict.PushBuilding(w)
	SetCompileState(cs, true)
	SetInColonDef(cs, true)
	return w
}

func bootstrapFinalizeColonWord(cs *CompilerState) {
	if !GetCompileState(cs) {
		panic("bootstrapFinalizeColonWord: not in compile state")
	}
	SetInColonDef(cs, false)
	w := cs.DC.dict.PeekBuilding()
	// Final methods: odrop then Unnest — pops the receiver pushed by mintFinalMethodCall.
	if w != nil && w.flags&WordFlagFinalMethod != 0 {
		bootstrapEmitOpcode(cs, OpODrop)
	}
	bootstrapEmitOpcode(cs, OpUnnest)
	if w != nil {
		w.clearSubwordCompileState()
	}
	if w != nil && w.flags&WordFlagFinalMethod != 0 {
		if err := bootstrapFinalizeFinalMethodTearOff(cs, w); err != nil {
			panic(err)
		}
		cs.DC.dict.PopDiscardBuilding()
	} else {
		cs.DC.dict.PopAddBuildingWordToParent()
	}
	SetCompileState(cs, false)
}

// instanceHostDC is the module where var:/const:/defer: instances are registered and
// allotted (compile-target / task.dc) — not sealed core when creating from a local Session.
func instanceHostDC(t *Task) *DictionaryContext {
	if t == nil {
		panic("instanceHostDC: nil Task")
	}
	cs := t.compilerForTarget()
	if cs == nil || cs.DC == nil {
		panic("instanceHostDC: nil compile-target DC")
	}
	return cs.DC
}

// runAllotDtSlotToWord allots the next DT slot in the building word's DC, stores def, sets w.dtIdx.
func runAllotDtSlotToWord(t *Task, rp *RunProvider) {
	def := rp.DataStack.Pop()
	cs := t.compilerForTarget()
	w := cs.DC.dict.PeekBuilding()
	if w == nil {
		panic("allotDtSlotToWord: outside colon definition")
	}
	idx := cs.DC.AllotDTIndex()
	cs.DC.dt.Write(idx, def)
	w.dtIdx = idx
}

// runDTRef pushes a mutable DT value ref (NewDTRef) for the exec DC slot at dtIdx.
func runDTRef(execDC *DictionaryContext, rp *RunProvider) {
	if execDC == nil {
		panic("dtRef: nil execDC")
	}
	dtIdx := int(rp.DataStack.Pop().(Int))
	rp.DataStack.Push(NewDTSlotRef(execDC.dt, dtIdx, false))
}

// runDTConst pushes DT[dtIdx] from execDC onto the data stack.
func runDTConst(execDC *DictionaryContext, rp *RunProvider) {
	if execDC == nil {
		panic("dtConst: nil execDC")
	}
	dtIdx := int(rp.DataStack.Pop().(Int))
	rp.DataStack.Push(execDC.dt.Read(dtIdx))
}

// bootstrapCoreWord resolves bootstrap machinery from the sealed core module.
func bootstrapCoreWord(name string) *Word {
	if globalDictionaryContext == nil {
		return nil
	}
	return globalDictionaryContext.dict.GetByName(name)
}

// compileModuleOperandForInsert is bootstrapEmitCompileModuleOperand policy for CT insert.
func compileModuleOperandForInsert(cs *CompilerState, targetModuleID ModuleID) int {
	if targetModuleID == cs.DC.moduleId {
		return int(ModuleIDThis)
	}
	return int(targetModuleID)
}

// bootstrapEmitOpcode appends one OpCode cell to CT.
func bootstrapEmitOpcode(cs *CompilerState, op OpCode) VirtualAddr {
	return cs.emitCT(int(op))
}

// bootstrapEmitLiteral pushes a literal value into DT and emits CT cells to load it at runtime.
func bootstrapEmitLiteral(cs *CompilerState, val any) VirtualAddr {
	if val == nil {
		return bootstrapEmitOpcode(cs, OpDoLitNull)
	}
	idx := cs.dt().Append(val)
	start := bootstrapEmitOpcode(cs, OpDoLit)
	_ = cs.emitCT(idx)
	return start
}

// bootstrapEmitCompoundLiteral freezes a compound path string in DT and emits OpDoLit + OpHandleCompound.
func bootstrapEmitCompoundLiteral(cs *CompilerState, val any) VirtualAddr {
	path, ok := val.(String)
	if !ok || path == "" {
		panic(fmt.Errorf("bootstrapEmitCompoundLiteral: want path String, got %T", val))
	}
	idx := cs.dt().Append(val)
	start := bootstrapEmitOpcode(cs, OpDoLit)
	_ = cs.emitCT(idx)
	bootstrapEmitOpcode(cs, OpHandleCompound)
	return start
}

// bootstrapEmitRecurse emits OpNest to the open colon word's colonCode (compile-time self-call).
func bootstrapEmitRecurse(cs *CompilerState) error {
	if cs == nil || !GetCompileState(cs) {
		return fmt.Errorf("recurse: outside compile state")
	}
	w := cs.DC.dict.PeekBuilding()
	if w == nil {
		return fmt.Errorf("recurse: outside colon definition")
	}
	if w.colonCode < 0 {
		return fmt.Errorf("recurse: building word %q has no colonCode", w.name)
	}
	bootstrapEmitColonCall(cs, w)
	return nil
}

// bootstrapEmitColonCall emits OpNest to a colon word's CT entry (compile time).
// Same-module targets emit moduleId 0 (sentinel "this") + plain entry offset.
func bootstrapEmitColonCall(cs *CompilerState, w *Word) VirtualAddr {
	start := bootstrapEmitOpcode(cs, OpNest)
	bootstrapEmitCompileModuleOperand(cs, w.moduleId)
	_ = cs.emitCT(int(w.colonCode))
	return start
}

// bootstrapEmitCompileModuleOperand emits 0 when target shares the compile DC, else absolute slot.
func bootstrapEmitCompileModuleOperand(cs *CompilerState, targetModuleID ModuleID) {
	if targetModuleID == cs.DC.moduleId {
		_ = cs.emitCT(int(ModuleIDThis))
	} else {
		_ = cs.emitCT(int(targetModuleID))
	}
}

// bootstrapEmitDtIdxPrefix emits OpCtImmDs + dtIdx before a DT-worker axiom.
func bootstrapEmitDtIdxPrefix(cs *CompilerState, dtIdx int) VirtualAddr {
	start := bootstrapEmitOpcode(cs, OpCtImmDs)
	_ = cs.emitCT(dtIdx)
	return start
}

// bootstrapEmitDoCSSelectPrefix appends OpDoCSSelect · idx to the compile target CT.
func bootstrapEmitDoCSSelectPrefix(cs *CompilerState, idx ContextStackIndex) {
	_ = cs.emitCT(int(OpDoCSSelect))
	_ = cs.emitCT(int(idx))
}

// bootstrapEmitWordReference compiles a dictionary reference to w into the open colon body.
// Words with dtIdx >= 0 prefix (dtIdx) / OpCtImmDs + operand before the runtime entry (worker axiom or Nest).
// Words with subCode >= 0 prefix OpDoSubCode + operand.
// RequiresStackContext words prefix OpDoCSSelect · selector (default csDataStack; override via cssel).
func bootstrapEmitWordReference(cs *CompilerState, w *Word, cssel ...ContextStackIndex) error {
	if w == nil {
		return fmt.Errorf("bootstrapEmitWordReference: nil word")
	}
	if b := cs.DC.dict.PeekBuilding(); b != nil {
		accumulateSRFrom(b, w)
	}
	sel := csDataStack
	if len(cssel) > 0 {
		sel = cssel[0]
	}
	if w.RequiresStackContext() {
		if err := checkOStackExecAllowed(sel, w); err != nil {
			return err
		}
		bootstrapEmitDoCSSelectPrefix(cs, sel)
	}
	if w.subCode >= 0 {
		bootstrapEmitSubCodePrefix(cs, w.subCode)
	}
	if w.dtIdx >= 0 {
		bootstrapEmitDtIdxPrefix(cs, w.dtIdx)
	}
	if w.IsAxiom() {
		bootstrapEmitOpcode(cs, w.opCode)
		return nil
	}
	if w.flags&WordFlagCompilerOnly != 0 {
		return fmt.Errorf("word %q is compiler-only", w.name)
	}
	if w.flags&WordFlagInterpretOnly != 0 {
		return fmt.Errorf("word %q is interpret-only", w.name)
	}
	if w.colonCode < 0 {
		return fmt.Errorf("word %q has no entry point", w.name)
	}
	bootstrapEmitColonCall(cs, w)
	return nil
}

// bootstrapReserveDTSlot appends a zeroed DT cell and returns its VirtualAddr for Lit emission.
func bootstrapReserveDTSlot(cs *CompilerState, zero any) VirtualAddr {
	cs.DC.panicIfLockedOrSealed("allot DT literal")
	dt := cs.dt()
	idx := dt.Append(zero)
	return EncodeVirt(ModuleIDThis, idx)
}

func (cs *CompilerState) emitCT(val int) VirtualAddr {
	cs.maybeResolveSubwordSkip()
	cs.DC.panicIfLockedOrSealed("emit CT")
	idx := cs.ct().Append(val)
	return EncodeVirt(ModuleIDThis, idx)
}

func (cs *CompilerState) ct() *CodeTable {
	return cs.DC.ct
}

func (cs *CompilerState) dt() *DataTable {
	return cs.DC.dt
}

func decodeCTOffset(v VirtualAddr) int {
	_, off := DecodeVirt(v)
	return off
}

// bootstrapColonStart handles interpret-time `:` — reads def name and opens compile mode.
func bootstrapColonStart(p *Parser, t *Task, rp *RunProvider) error {
	cs := t.compilerForTarget()
	if GetCompileState(cs) {
		return fmt.Errorf(": while already compiling")
	}
	val, typ, err := p.Next()
	if err != nil {
		return err
	}
	if !IsPlainIdentToken(typ) {
		return fmt.Errorf(": expected definition name, got %v", typ)
	}
	name, ok := val.(String)
	if !ok {
		return fmt.Errorf(": expected String name, got %T", val)
	}
	nameStr := string(name)
	typeName, methodName, isClass, isFinal, perr := parseFinalMethodHeader(nameStr)
	if perr != nil {
		return perr
	}
	if isFinal {
		if err := prepareFinalMethodTearOff(cs, rp, nameStr, typeName, methodName, isClass); err != nil {
			return err
		}
	}
	w := bootstrapCreate(cs, nameStr)
	if isFinal {
		w.flags |= WordFlagFinalMethod
	}
	w.doc = p.GetClearLastComment()
	rp.TempStack.Push(NewBlockToken(BlockKindColon))
	return nil
}

// --- Phase 1: interpret-only REPL (no compiling) ---

// RunInteractiveREPL creates a persistent Session for kind and runs its REPL until stdin EOF.
func RunInteractiveREPL(kind SessionKind) error {
	s, err := NewSession(&SessionOptions{Kind: kind, Persist: true})
	if err != nil {
		return err
	}
	defer s.Close()
	return s.RunREPL()
}

// BootstrapREPL runs the interactive REPL until stdin EOF.
// Uses the Go interpret/compile loop (colon definitions, immediates, [[ … ]] emit).
// The in-VM ColdBoot CT driver (bootstrapRunColdBoot) remains for boot-vector tests.
func BootstrapREPL(t *Task, rp *RunProvider) error {
	if t == nil || rp == nil {
		return fmt.Errorf("BootstrapREPL: nil Task or RunProvider")
	}
	if rp.Stdin == nil {
		return fmt.Errorf("BootstrapREPL: nil Stdin")
	}
	runInterpretLoop(t, rp)
	return nil
}

func bootstrapWriteREPLError(rp *RunProvider, err error) {
	if rp.Stderr == nil {
		return
	}
	msg := err.Error()
	if rp.Parser != nil {
		msg = formatFault(err, rp.Parser)
		rp.Parser.clearParserLocation()
	}
	rp.Stderr.WriteString(msg)
	if !strings.HasSuffix(msg, "\n") {
		rp.Stderr.WriteString("\n")
	}
	flushTextWrite(rp.Stderr)
}

// BootstrapREPLFromReader is a test/helper entry using any io.Reader as stdin.
func BootstrapREPLFromReader(t *Task, r io.Reader) error {
	rp := NewRunProvider()
	rp.Stdin = &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(r),
		label: "repl",
	}}
	return BootstrapREPL(t, rp)
}

type bufioReaderProvider struct {
	r     *bufio.Reader
	label string
}

func (p *bufioReaderProvider) GetNextRune() (rune, bool) {
	return readNextRuneFromBufio(p.r, p.label)
}

func (p *bufioReaderProvider) GetNextByte() (byte, bool) {
	return readNextByteFromBufio(p.r, p.label)
}

func (p *bufioReaderProvider) ReadBytes(n int) []byte {
	return readBytesFromBufio(p.r, p.label, n)
}

var (
	_ TextReadProvider = (*bufioReaderProvider)(nil)
	_ ByteReadProvider = (*bufioReaderProvider)(nil)
)
