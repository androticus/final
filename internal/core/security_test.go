package core

import (
	"strings"
	"testing"
)

func TestSRClassConstantsRegistered(t *testing.T) {
	ResetTestRuntime()
	srType := GetTypeRegistrar().Get("SR")
	if srType == nil {
		t.Fatal("SR type not registered")
	}
	want := map[string]SecurityRestrictions{
		"ModuleInstall":  SR_ModuleInstall,
		"ModuleImport":   SR_ModuleImport,
		"FileRead":       SR_FileRead,
		"FileWrite":      SR_FileWrite,
		"FileSystemRead": SR_FileSystemRead,
		"HttpDial":       SR_HttpDial,
		"HttpListen":     SR_HttpListen,
		"Threads":        SR_Threads,
		"Db":             SR_Db,
		"Crypto":         SR_Crypto,
		"Mask":           SR_Mask,
	}
	for name, w := range want {
		got := invokeClassConst(t, srType, name)
		if got != Uint(w) {
			t.Fatalf("SR.%s = %v, want %v", name, got, Uint(w))
		}
	}
}

func TestWordFlagsHasNoSR_Mask(t *testing.T) {
	ResetTestRuntime()
	wf := GetTypeRegistrar().Get("WordFlags")
	if wf == nil {
		t.Fatal("WordFlags type not registered")
	}
	if pm := wf.GetClassPropertyMetadata("SR_Mask"); pm != nil {
		t.Fatal("WordFlags.SR_Mask must not exist; SR lives on Word.srFlags / type SR")
	}
	if got := invokeClassConst(t, wf, "Colon"); got != Uint(WordFlagColon) {
		t.Fatalf("WordFlags.Colon = %v, want %v", got, Uint(WordFlagColon))
	}
}

func TestSRGroupSpareBits(t *testing.T) {
	// Active bits must sit in the mask and not collide across groups (remapped uint32 layout).
	active := []SecurityRestrictions{SR_ModuleInstall, SR_ModuleImport, SR_System, SR_FileRead, SR_FileWrite, SR_FileSystemRead, SR_HttpDial, SR_HttpListen, SR_Threads, SR_Db, SR_Crypto}
	var or SecurityRestrictions
	for _, b := range active {
		if b&SR_Mask != b {
			t.Fatalf("%#x outside SR_Mask", b)
		}
		if or&b != 0 {
			t.Fatalf("overlap involving %#x", b)
		}
		or |= b
	}
	// File spare bit 15; Http spare 19; Threads spare 22; Db spare 24; Crypto spare 26.
	for _, spare := range []SecurityRestrictions{1 << 15, 1 << 19, 1 << 22, 1 << 24, 1 << 26} {
		if or&spare != 0 {
			t.Fatalf("spare %#x already used", spare)
		}
		if spare&SR_Mask != spare {
			t.Fatalf("spare %#x outside mask", spare)
		}
	}
}

func TestSRAccumulateOnColonCompile(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	cs := task.compilerForTarget()

	spawnAx := globalDictionaryContext.dict.GetByName("(spawnPip)")
	if spawnAx == nil {
		t.Fatal("missing (spawnPip)")
	}
	if spawnAx.SRFlags()&SR_Threads == 0 {
		t.Fatal("(spawnPip) should carry SR_Threads")
	}

	nop := globalDictionaryContext.dict.GetByName("nop")
	if nop == nil {
		t.Fatal("missing nop")
	}
	saved := nop.srFlags
	nop.srFlags |= uint32(SR_FileRead)
	defer func() { nop.srFlags = saved }()

	bootstrapCreate(cs, "srOuter")
	if err := bootstrapEmitWordReference(cs, spawnAx); err != nil {
		t.Fatal(err)
	}
	if err := bootstrapEmitWordReference(cs, nop); err != nil {
		t.Fatal(err)
	}
	bootstrapFinalizeColonWord(cs)

	outer := globalDictionaryContext.dict.GetByName("srOuter")
	if outer == nil {
		t.Fatal("missing srOuter")
	}
	want := SR_Threads | SR_FileRead
	if got := outer.SRFlags(); got != want {
		t.Fatalf("srOuter.SRFlags() = %#x, want %#x", got, want)
	}
}

func TestRegisterNativeMethodSecStoresRestrict(t *testing.T) {
	ResetTestRuntime()
	typ := NewRegisteringType("SRMethodProbe", ReferenceTypeZeroFn, MakePointerTypeCheckCoerce[struct{}]("SRMethodProbe"))
	typ.registerNativeMethodSec("gated", func(_ *RunFrame, _ *TypeObjPair) {}, SR_FileRead)
	mm := typ.GetMethodMetadata("gated")
	if mm == nil {
		t.Fatal("missing method gated")
	}
	if mm.restrict != SR_FileRead {
		t.Fatalf("restrict = %#x, want SR_FileRead", mm.restrict)
	}
	typ.registerNativeMethod("open", func(_ *RunFrame, _ *TypeObjPair) {})
	if open := typ.GetMethodMetadata("open"); open == nil || open.restrict != 0 {
		t.Fatalf("legacy registerNativeMethod should leave restrict 0, got %#v", open)
	}
}

func TestNewTypeWordORsTypeRestrict(t *testing.T) {
	ResetTestRuntime()
	typ := GetTypeRegistrar().Get("Int")
	if typ == nil {
		t.Fatal("missing Int")
	}
	saved := typ.restrict
	typ.restrict = SR_HttpDial
	defer func() { typ.restrict = saved }()
	w := NewTypeWord(typ, CoreModuleSlot)
	if w.SRFlags()&SR_HttpDial == 0 {
		t.Fatalf("type Word srFlags = %#x, want SR_HttpDial from Type.restrict", w.srFlags)
	}
	if w.flags&WordFlags(SR_HttpDial) != 0 {
		t.Fatalf("Type.restrict must not pack into Word.flags; flags=%#x", w.flags)
	}
}

func TestAssertSRPanicsOpaque(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	AssertSR(SR_Threads, SR_Threads)
}

func TestFindNameInTaskRejectsRestrictedWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Threads
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	FindNameInTask(task, rp, "(spawnPip)")
}

func TestArmExecWordRejectsRestrictedWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Threads
	w := globalDictionaryContext.dict.GetByName("(spawnPip)")
	if w == nil {
		t.Fatal("missing (spawnPip)")
	}
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	var runSubCode int
	armExecWord(task, w, &task.dc, nil, &runSubCode, nil, nil)
}

func TestCompoundWalkAssertsMethodRestrict(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_FileRead
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	tr, err := NewTextReaderFromString("\"/no/such\" TextReader.newFromFile^\n")
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	_ = interpretSession(task, rp)
	t.Fatal("expected panic from restricted class method")
}

func TestCompoundWalkAllowsUnrestrictedFactory(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_FileRead
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp, "\"hi\" TextReader.newFromString^\ndrop\n")
}

func TestFileFileExistsRestrict(t *testing.T) {
	ResetTestRuntime()
	mm := GetTypeRegistrar().Get("File").GetClassMethodMetadata("fileExists?")
	if mm == nil || mm.restrict != SR_FileSystemRead {
		t.Fatalf("File.fileExists? restrict = %#v, want SR_FileSystemRead", mm)
	}
}

func TestHttpDialTypesRestrict(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"HttpDialer", "HttpReq", "HttpRes"} {
		typ := GetTypeRegistrar().Get(name)
		if typ == nil {
			t.Fatalf("missing type %s", name)
		}
		if typ.restrict != SR_HttpDial {
			t.Fatalf("%s.restrict = %#x, want SR_HttpDial", name, typ.restrict)
		}
		w := NewTypeWord(typ, CoreModuleSlot)
		if w.SRFlags()&SR_HttpDial == 0 {
			t.Fatalf("%s type Word srFlags = %#x, want SR_HttpDial", name, w.srFlags)
		}
	}
}

func TestHttpListenTypesRestrict(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"HttpListener", "HttpInReq", "HttpOutRes", "HttpServeJob", "HttpRoute"} {
		typ := GetTypeRegistrar().Get(name)
		if typ == nil {
			t.Fatalf("missing type %s", name)
		}
		if typ.restrict != SR_HttpListen {
			t.Fatalf("%s.restrict = %#x, want SR_HttpListen", name, typ.restrict)
		}
		w := NewTypeWord(typ, CoreModuleSlot)
		if w.SRFlags()&SR_HttpListen == 0 {
			t.Fatalf("%s type Word srFlags = %#x, want SR_HttpListen", name, w.srFlags)
		}
	}
}

func TestInvokeCallAssertsMethodRestrict(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_FileRead
	rp := NewTestRunProvider()
	mm := GetTypeRegistrar().Get("TextReader").GetClassMethodMetadata("newFromFile")
	if mm == nil {
		t.Fatal("missing TextReader.newFromFile")
	}
	f := &RunFrame{Task: task, RP: rp}
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("TextReader")))
	t.Fatal("expected panic")
}

// Compound Find AssertSR panics outside runExecWord; REPL must report and continue.
func TestRunInterpretLoopRecoversRestrictedCompoundFind(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Threads
	rp := NewRunProvider()
	tr, err := NewTextReaderFromString("Pipe.newWithCapacity^\n99\n")
	if err != nil {
		t.Fatal(err)
	}
	rp.Stdin = tr
	var outSB StringBuilder
	rp.Stdout = &outSB
	if err := BootstrapREPL(task, rp); err != nil {
		t.Fatalf("BootstrapREPL: %v", err)
	}
	if !strings.Contains(outSB.ToString(), "operation not allowed") {
		t.Fatalf("stdout = %q, want operation not allowed", outSB.ToString())
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d after continued REPL, want 1", rp.DataStack.Len())
	}
	if got := rp.DataStack.Peek(); got != Int(99) {
		t.Fatalf("DS TOS = %v, want 99 (REPL continued after SR deny)", got)
	}
}

func TestDbProviderNewAssertsSR_Db(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Db
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	mm := GetTypeRegistrar().Get("DbProvider").GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing DbProvider.new")
	}
	rp.DataStack.Push(String("sqlite"))
	rp.DataStack.Push(String(":memory:"))
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("DbProvider")))
	t.Fatal("expected panic")
}

func TestDbProviderNewSqliteAssertsFileSRs(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_FileWrite // Db allowed; sqlite still needs FileWrite
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	mm := GetTypeRegistrar().Get("DbProvider").GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("missing DbProvider.new")
	}
	rp.DataStack.Push(String("sqlite"))
	rp.DataStack.Push(String(":memory:"))
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("DbProvider")))
	t.Fatal("expected panic")
}

func TestInstallImportAxiomsCarryModuleSR(t *testing.T) {
	ResetTestRuntime()
	install := globalDictionaryContext.dict.GetByName("install")
	if install == nil {
		t.Fatal("missing install")
	}
	if install.SRFlags()&SR_ModuleInstall == 0 {
		t.Fatal("install should carry SR_ModuleInstall")
	}
	imp := globalDictionaryContext.dict.GetByName("import")
	if imp == nil {
		t.Fatal("missing import")
	}
	if imp.SRFlags()&SR_ModuleImport == 0 {
		t.Fatal("import should carry SR_ModuleImport")
	}
	lp := globalDictionaryContext.dict.GetByName("loadProject")
	if lp == nil {
		t.Fatal("missing loadProject")
	}
	if lp.SRFlags() != 0 {
		t.Fatalf("loadProject SRFlags = %#x, want 0 (asserts live in installLeafModule/importModule)", lp.SRFlags())
	}
}

func TestSystemFlagOnNeverUserWordsOnly(t *testing.T) {
	ResetTestRuntime()
	keep := []string{"setWordFlags", "_unnest", "rpeek", "cs,", "(cs)", "armCs"}
	for _, name := range keep {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if !w.IsSystem() {
			t.Fatalf("%q should be WordFlagSystem", name)
		}
	}
	strip := []string{"(spawnPip)", "(mathMonadic)", "(dtDyadic)", "(strOp)", "(pipeOp)", "(cast)",
		"(vecUnary)", "(setOp)", "odrop"}
	for _, name := range strip {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if w.IsSystem() {
			t.Fatalf("%q should not be WordFlagSystem (user-runtime worker)", name)
		}
	}
}

func TestLocalTaskDeniesSystemFindAndExec(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	w := globalDictionaryContext.dict.GetByName("setWordFlags")
	if w == nil {
		t.Fatal("missing setWordFlags")
	}

	func() {
		defer func() {
			r := recover()
			if r != "operation not allowed" {
				t.Fatalf("Find setWordFlags panic = %v, want %q", r, "operation not allowed")
			}
		}()
		FindNameInTask(task, rp, "setWordFlags")
		t.Fatal("expected Find panic")
	}()

	func() {
		defer func() {
			r := recover()
			if r != "operation not allowed" {
				t.Fatalf("armExecWord setWordFlags panic = %v, want %q", r, "operation not allowed")
			}
		}()
		var runSubCode int
		armExecWord(task, w, &task.dc, nil, &runSubCode, nil, nil)
		t.Fatal("expected exec panic")
	}()
}

func TestCoreTaskAllowsSystemFind(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	if !task.systemMode {
		t.Fatal("NewTestTask (core) should have systemMode")
	}
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	w, _, ok := FindNameInTask(task, rp, "setWordFlags")
	if !ok || w == nil || !w.IsSystem() {
		t.Fatal("core task should Find setWordFlags")
	}
}

func TestSpawnColonCarriesSR_ThreadsVirally(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("spawn:")
	if w == nil {
		t.Fatal("missing spawn:")
	}
	if w.SRFlags()&SR_Threads == 0 {
		t.Fatalf("spawn: srFlags = %#x, want SR_Threads from viral (spawnPip)", w.srFlags)
	}
}

func TestSetWordFlagsDoesNotTouchSRFlags(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "noSROverride")
	outer := cs.DC.dict.PeekBuilding()
	if outer == nil {
		t.Fatal("missing building word")
	}
	outer.srFlags = uint32(SR_FileRead)
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(WordFlagImmediate))
	if err := runWordInlineExec(task, rp, requireDictWord(t, "setWordFlags")); err != nil {
		t.Fatal(err)
	}
	bootstrapFinalizeColonWord(cs)
	w := globalDictionaryContext.dict.GetByName("noSROverride")
	if w == nil {
		t.Fatal("missing noSROverride")
	}
	if w.SRFlags() != SR_FileRead {
		t.Fatalf("setWordFlags must not clear srFlags: got %#x", w.SRFlags())
	}
	if w.flags&WordFlagImmediate == 0 {
		t.Fatal("expected WordFlagImmediate from setWordFlags")
	}
}
