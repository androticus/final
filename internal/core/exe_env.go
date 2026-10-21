package core

import (
	"fmt"
	"sync"
)

// This file contains the types, variables, and functions for the global Execution Environment
// The environments is shared by all tasks and modules

const (
	globalDictMaxSize = 5000
	globalDictCap     = 1024
	globalCTMaxSize   = globalDictMaxSize * 10 // guess of how many code cells will be needed for a typical program
	globalCTCap       = 2048
	globalDTMaxSize   = globalDictMaxSize / 5
	globalDTCap       = 128
)

var (
	// a persistent slice of all the axioms, indexed by their ordinal
	// loaded during int() and never modified thereafter
	// MakeGlobarDictionary loads these first, so they are available to all tasks and modules
	axioms                  []*Word = make([]*Word, 0, opAxiomLAST+1)
	globalDictionaryContext *DictionaryContext
)

var (
	// initGlobalEnvOnce ensures InitGlobalExecutionEnvironment runs at most once per process.
	// Test reset restores a snapshot captured after the one-time installGlobalCoreVocabulary.
	initGlobalEnvOnce sync.Once

	capturedGlobalEnvSnapshot globalEnvSnapshot
)

type globalEnvSnapshot struct {
	dictOrdinal []*Word
	dictByName  map[string]*Word
	ctCells     []int
	dtCells     []any
}

func init() {
	axioms = registerAxioms()
}

// axiomWordForOp returns the primary axiom Word registered for op.
// registerAxioms appends one Word per OpCode ordinal 0..opAxiomLAST before subcode aliases.
func axiomWordForOp(op OpCode) *Word {
	if op < 0 || op > opAxiomLAST {
		panic(fmt.Errorf("axiomWordForOp: opcode %d out of range 0..%d", op, opAxiomLAST))
	}
	if int(op) >= len(axioms) {
		panic(fmt.Errorf("axiomWordForOp: axioms slice too short for opcode %d", op))
	}
	w := axioms[op]
	if w == nil || w.opCode != op {
		panic(fmt.Errorf("axiomWordForOp: axioms[%d] = %q opCode %d", op, w.name, w.opCode))
	}
	return w
}

// InitGlobalExecutionEnvironment builds the global GCD once: axioms, bootstrap shell, boot CT,
// and core0/core1/core2 bootstrap compile (core3/core4 if EnableBootstrapTranche3/4 was called).
// Safe to call from multiple places; only the first call runs.
func InitGlobalExecutionEnvironment() {
	initGlobalEnvOnce.Do(initGlobalExecutionEnvironment)
}

func initGlobalExecutionEnvironment() {
	globalDictionaryContext = NewDictionaryContext(
		"global",
		globalDictMaxSize, globalDictCap,
		globalCTMaxSize, globalCTCap,
		globalDTMaxSize, globalDTCap,
		GetTypeRegistrar(),
	)
	InitializeCoreReservedDTSlots(globalDictionaryContext.dt)
	RegisterCoreModule(globalDictionaryContext)
	installGlobalCoreVocabulary(globalDictionaryContext)
	captureGlobalExecutionEnvironmentSnapshot()
}

// installGlobalCoreVocabulary registers the full core module vocabulary once per process:
// axioms, type words, bootstrap shell words, boot CT wiring, and core0/core1/core2 compile
// (core3/core4 if EnableBootstrapTranche3/4 was called).
func installGlobalCoreVocabulary(dc *DictionaryContext) {
	for _, w := range axioms {
		dc.dict.Add(w)
	}
	registerTypeWords(dc)
	registerBootstrapWords()
	initializeBootstrapEnvironment(dc)
	runBootstrapCoreCompile()
	dc.Seal()
}

func initializeBootstrapEnvironment(dc *DictionaryContext) {
	InitializeBootstrapCT(dc)
	t := &Task{
		dc:           dc,
		compileModID: CoreModuleSlot,
		ip:           BootstrapGlobalColdEntry(),
	}
	wireBootstrapBootTargets(t, dc)
}

// captureGlobalExecutionEnvironmentSnapshot records dict/CT/DT after installGlobalCoreVocabulary.
func captureGlobalExecutionEnvironmentSnapshot() {
	dc := globalDictionaryContext
	d := dc.dict
	capturedGlobalEnvSnapshot = globalEnvSnapshot{
		dictOrdinal: append([]*Word(nil), d.byOrdinal...),
		dictByName:  copyWordNameMap(d.byName),
		ctCells:     snapshotCTCells(dc.ct),
		dtCells:     snapshotDTCells(dc.dt),
	}
}

func copyWordNameMap(src map[string]*Word) map[string]*Word {
	dst := make(map[string]*Word, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func snapshotCTCells(ct *CodeTable) []int {
	if ct.Len() == 0 {
		return nil
	}
	out := make([]int, ct.Len())
	for i := 0; i < ct.Len(); i++ {
		out[i] = ct.Read(i)
	}
	return out
}

func snapshotDTCells(dt *DataTable) []any {
	if dt.Len() == 0 {
		return nil
	}
	out := make([]any, dt.Len())
	for i := 0; i < dt.Len(); i++ {
		out[i] = dt.Read(i)
	}
	return out
}

func restoreGlobalExecutionEnvironmentSnapshot() {
	if globalDictionaryContext == nil {
		panic(fmt.Errorf("global execution environment not initialized"))
	}
	snap := capturedGlobalEnvSnapshot
	if snap.dictByName == nil {
		panic(fmt.Errorf("global execution environment snapshot missing (InitGlobalExecutionEnvironment not run)"))
	}
	d := globalDictionaryContext.dict
	d.byOrdinal = append(d.byOrdinal[:0], snap.dictOrdinal...)
	d.byName = copyWordNameMap(snap.dictByName)
	d.ClearBuilding()
	restoreCTCells(globalDictionaryContext.ct, snap.ctCells)
	restoreDTCells(globalDictionaryContext.dt, snap.dtCells)
	// Tests compile colon words into core; leave unsealed after restore.
	globalDictionaryContext.Unseal()
}

func restoreCTCells(ct *CodeTable, cells []int) {
	ct.Clear()
	for _, v := range cells {
		ct.Append(v)
	}
}

func restoreDTCells(dt *DataTable, cells []any) {
	dt.Clear()
	for _, v := range cells {
		dt.Append(v)
	}
}

// registerBootstrapWords adds boot and REPL compile driver words after axioms.
// Caller must ensure bootstrap names are not already in the dictionary (see ResetGlobalExecutionEnvironmentForTest).
func registerBootstrapWords() {
	d := globalDictionaryContext.dict
	const hiddenCompiler = WordFlagHidden | WordFlagCompilerOnly
	colonH := registerBootstrapTransitionWord("_boot_colon", TrBootColon, hiddenCompiler)
	registerBootDeferWord(bootWordColdBoot, WordFlagHidden|WordFlagAxiom, ReservedDTSlotColdBootTarget)
	registerBootDeferWord(bootWordWarmBoot, WordFlagHidden|WordFlagAxiom, ReservedDTSlotWarmBootTarget)
	registerDeferWord(":", WordFlagHidden, colonH)
	semiH := registerBootstrapTransitionWord("_boot_semicolon", TrBootSemicolon, hiddenCompiler)
	registerDeferWord(";", WordFlagHidden|WordFlagImmediate|WordFlagCompileOnly, semiH)
	sOpenH := registerBootstrapTransitionWord("_boot_s_open", TrBootSubwordOpen, hiddenCompiler)
	registerDeferWord(":s", WordFlagHidden|WordFlagImmediate|WordFlagCompileOnly, sOpenH)
	sCloseH := registerBootstrapTransitionWord("_boot_s_close", TrBootSubwordClose, hiddenCompiler)
	registerDeferWord(";s", WordFlagHidden|WordFlagImmediate|WordFlagCompileOnly, sCloseH)
	findH := registerBootstrapTransitionWord("_boot_find", FinalFind, hiddenCompiler)
	registerDeferWord("find?", 0, findH)
	cdH := registerBootstrapTransitionWord("_boot_currentDictHasWord", FinalCurrentDictHasWord, hiddenCompiler)
	registerDeferWord("currentDictHasWord", 0, cdH)
	tokH := registerBootstrapTransitionWord("_boot_getNextToken", FinalGetNextToken, hiddenCompiler)
	registerDeferWord("getNextToken", 0, tokH)
	errH := registerBootstrapTransitionWord("_boot_replError", FinalReplError, hiddenCompiler)
	registerDeferWord("replError", 0, errH)
	registerBootstrapTransitionWord("readIdentifier", FinalReadIdentifier, 0)
	registerBootstrapTransitionWord("validatePlainIdent", FinalValidatePlainIdent, 0)
	registerBootstrapTransitionWord("mintVarWord", FinalMintVarWord, WordFlagHidden|WordFlagCompilerOnly)
	registerBootstrapTransitionWord("mintConstWord", FinalMintConstWord, WordFlagHidden|WordFlagCompilerOnly)
	registerBootstrapTransitionWord("mintDeferWord", FinalMintDeferWord, WordFlagHidden|WordFlagCompilerOnly)
	compileH := registerBootstrapTransitionWord("_boot_compileFromFile", FinalCompileFromFile, hiddenCompiler)
	registerDeferWord("compileFromFile", 0, compileH)
	manifestH := registerBootstrapTransitionWord("_boot_compileFromManifest", FinalCompileFromManifest, hiddenCompiler)
	registerDeferWord("compileFromManifest", 0, manifestH)
	compileStateH := registerBootstrapTransitionWord("_boot_setCompileState", FinalSetCompileState, hiddenCompiler)
	_ = compileStateH
	registerBootstrapTransitionWord("fmt", FinalFmt, 0)
	registerBootstrapTransitionWord("(forget:)", FinalForgetMark, WordFlagHidden)
	d.Add(NewReservedVarWord("coldBootTarget", ReservedDTSlotColdBootTarget))
	d.Add(NewReservedVarWord("warmBootTarget", ReservedDTSlotWarmBootTarget))
	d.Add(NewReservedVarWord("byeReq", ReservedDTSlotByeReq))
}

// GetDictionaryContext returns the DictionaryContext (Dict/CT/DT bundle) for modID.
// Core lives at CoreModuleSlot (1). ModuleIDThis (0) is the "this" sentinel — resolves to core until multi-module compile targets exist.
func GetDictionaryContext(modID ModuleID) *DictionaryContext {
	switch modID {
	case ModuleIDThis, CoreModuleSlot:
		if globalDictionaryContext == nil {
			panic(fmt.Errorf("global execution environment not initialized"))
		}
		return globalDictionaryContext
	default:
		return ModuleTable.DCForModule(nil, modID)
	}
}

// ResetGlobalExecutionEnvironmentForTest restores global Dict/CT/DT to the post-init snapshot
// captured after installGlobalCoreVocabulary (core0, core1, core2, and core3 if enabled).
// Does not recompile bootstrap.
func ResetGlobalExecutionEnvironmentForTest() {
	restoreGlobalExecutionEnvironmentSnapshot()
	clear(hostTypeRegistry)
}
