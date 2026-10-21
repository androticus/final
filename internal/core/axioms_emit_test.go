package core

import "testing"

func beginEmitCompileTask(t *testing.T) (*Task, *RunProvider, *CompilerState) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "emit-test")
	SetEmitting(cs, true)
	return task, rp, cs
}

func TestEmitDoEmitWord_bakesOpDoEmit(t *testing.T) {
	_, _, cs := beginEmitCompileTask(t)
	w := globalDictionaryContext.dict.GetByName("branch?")
	if w == nil {
		t.Fatal("missing branch?")
	}
	offBefore := globalDictionaryContext.ct.Len()
	if err := emitDoEmitWord(cs, w); err != nil {
		t.Fatal(err)
	}
	ct := globalDictionaryContext.ct
	if ct.Read(offBefore) != int(OpDoEmit) {
		t.Fatalf("CT[%d] = %d, want OpDoEmit", offBefore, int(OpDoEmit))
	}
	xt, ok := globalDictionaryContext.dict.XTOf(w)
	if !ok {
		t.Fatal("XTOf branch?")
	}
	if ct.Read(offBefore+1) != int(xt) {
		t.Fatalf("CT[%d] xt = %d, want %d", offBefore+1, ct.Read(offBefore+1), int(xt))
	}
}

func TestOpDoEmit_compilesBranchQIntoOpenColon(t *testing.T) {
	task, rp, cs := beginEmitCompileTask(t)
	wBranchQ := globalDictionaryContext.dict.GetByName("branch?")
	if wBranchQ == nil {
		t.Fatal("missing branch?")
	}
	w := cs.DC.dict.PeekBuilding()
	if w == nil {
		t.Fatal("no open colon")
	}
	bodyStart := int(w.colonCode)
	if err := emitDoEmitWord(cs, wBranchQ); err != nil {
		t.Fatal(err)
	}
	stopIP := globalDictionaryContext.ct.Len()
	globalDictionaryContext.ct.Append(int(OpStop))
	rp.ReturnStack.PushNew(stopIP)
	task.ip = EncodeVirt(ModuleIDThis, bodyStart)
	Run(task, rp)
	ct := globalDictionaryContext.ct
	got := ct.Read(stopIP + 1)
	if got != int(OpBranchQ) {
		t.Fatalf("after doEmit CT[%d] = %d, want OpBranchQ", stopIP+1, got)
	}
}

func TestCompileEmittingMode_emitsBracketLiteral(t *testing.T) {
	task, rp, cs := beginEmitCompileTask(t)
	offBefore := cs.ct().Len()
	if err := handleParsedTokenCompile(nil, task, rp, cs, Int(42), LiteralTokenType(TokenInt), false, nil); err != nil {
		t.Fatal(err)
	}
	ct := globalDictionaryContext.ct
	if OpCode(ct.Read(offBefore)) != OpDoLit {
		t.Fatalf("CT[%d] = %s, want OpDoLit", offBefore, OpCodeName(OpCode(ct.Read(offBefore))))
	}
	found := false
	for i := offBefore; i < ct.Len(); i++ {
		if OpCode(ct.Read(i)) == OpCompileLit {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected OpCompileLit after bracket literal emit")
	}
}

func TestCompileEmittingMode_emitsDoEmitForIdent(t *testing.T) {
	task, rp, cs := beginEmitCompileTask(t)
	offBefore := globalDictionaryContext.ct.Len()
	if err := handleParsedTokenCompile(nil, task, rp, cs, String("branch?"), TokenIdent, false, nil); err != nil {
		t.Fatal(err)
	}
	ct := globalDictionaryContext.ct
	if ct.Read(offBefore) != int(OpDoEmit) {
		t.Fatalf("CT[%d] = %d, want OpDoEmit", offBefore, ct.Read(offBefore))
	}
}
