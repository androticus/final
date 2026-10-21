package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestBootWordsAfterAxioms(t *testing.T) {
	ResetTestRuntime()
	d := globalDictionaryContext.dict
	cold := d.GetByName(bootWordColdBoot)
	warm := d.GetByName(bootWordWarmBoot)
	if cold == nil {
		t.Fatal("missing ColdBoot word")
	}
	if warm == nil {
		t.Fatal("missing WarmBoot word")
	}
	if cold.ordIdx <= len(axioms)+len(typeWords) {
		t.Fatalf("ColdBoot ordIdx = %d, want after axioms + type words", cold.ordIdx)
	}
	if cold.opCode != OpBootstrapDefer || cold.dtIdx != ReservedDTSlotColdBootTarget {
		t.Fatalf("ColdBoot = opCode %v dtIdx %d, want OpBootstrapDefer slot %d", OpCodeName(cold.opCode), cold.dtIdx, ReservedDTSlotColdBootTarget)
	}
	if warm.opCode != OpBootstrapDefer || warm.dtIdx != ReservedDTSlotWarmBootTarget {
		t.Fatalf("WarmBoot = opCode %v dtIdx %d, want OpBootstrapDefer slot %d", OpCodeName(warm.opCode), warm.dtIdx, ReservedDTSlotWarmBootTarget)
	}
}

func TestBootCTVectorsAndReplLoop(t *testing.T) {
	ResetTestRuntime()
	ct := globalDictionaryContext.ct
	if ct.Read(ReservedCTBootColdEntry) != int(OpCtImmDs) {
		t.Fatalf("CT[0] = %d, want OpCtImmDs", ct.Read(ReservedCTBootColdEntry))
	}
	if ct.Read(ReservedCTBootColdSlot) != ReservedDTSlotColdBootTarget {
		t.Fatalf("CT[1] = %d, want cold DT slot %d", ct.Read(ReservedCTBootColdSlot), ReservedDTSlotColdBootTarget)
	}
	if ct.Read(ReservedCTBootColdDefer) != int(OpBootstrapDefer) {
		t.Fatalf("CT[2] = %d, want OpBootstrapDefer", ct.Read(ReservedCTBootColdDefer))
	}
	if ct.Read(ReservedCTBootWarmCtImmDs) != int(OpCtImmDs) {
		t.Fatalf("CT[3] = %d, want OpCtImmDs", ct.Read(ReservedCTBootWarmCtImmDs))
	}
	if ct.Read(ReservedCTBootWarmSlot) != ReservedDTSlotWarmBootTarget {
		t.Fatalf("CT[4] = %d, want warm DT slot %d", ct.Read(ReservedCTBootWarmSlot), ReservedDTSlotWarmBootTarget)
	}
	if ct.Read(ReservedCTBootWarmDefer) != int(OpBootstrapDefer) {
		t.Fatalf("CT[5] = %d, want OpBootstrapDefer", ct.Read(ReservedCTBootWarmDefer))
	}
	if ct.Read(ReservedCTBootStop) != int(OpStop) {
		t.Fatalf("CT[6] = %d, want OpStop", ct.Read(ReservedCTBootStop))
	}
	if ct.Read(int(BootstrapReplLoopHeadCTP())) != int(OpDoSubCode) {
		t.Fatalf("repl body should start with OpDoSubCode, got %d", ct.Read(int(BootstrapReplLoopHeadCTP())))
	}
	if ct.Read(int(BootstrapReplLoopHeadCTP())+1) != int(TrCheckResetByeReq) {
		t.Fatalf("repl body CT[%d] = %d, want TrCheckResetByeReq", int(BootstrapReplLoopHeadCTP())+1, ct.Read(int(BootstrapReplLoopHeadCTP())+1))
	}
	nextTokenIdx := int(BootstrapReplLoopHeadCTP()) + 5
	if ct.Read(nextTokenIdx) != int(OpDoSubCode) {
		t.Fatalf("repl body CT[%d] = %d, want OpDoSubCode before TrNextToken", nextTokenIdx, ct.Read(nextTokenIdx))
	}
	if ct.Read(nextTokenIdx+1) != int(TrNextToken) {
		t.Fatalf("repl body CT[%d] = %d, want TrNextToken", nextTokenIdx+1, ct.Read(nextTokenIdx+1))
	}
	exitIdx := int(BootstrapWarmBodyCTP()) - 1
	if OpCode(ct.Read(exitIdx)) != OpUnnest {
		t.Fatalf("repl loop exit CT[%d] = %s, want OpUnnest", exitIdx, OpCodeName(OpCode(ct.Read(exitIdx))))
	}
}

func TestBootDTDeferTargets(t *testing.T) {
	ResetTestRuntime()
	dt := globalDictionaryContext.dt
	nopW := globalDictionaryContext.dict.GetByName("nop")
	nopXT, ok := globalDictionaryContext.dict.XTOf(nopW)
	if !ok {
		t.Fatal("XTOf nop")
	}
	if dt.Read(ReservedDTSlotColdBootTarget) != Int(nopXT) {
		t.Fatalf("coldBootTarget = %v, want nop XT %v", dt.Read(ReservedDTSlotColdBootTarget), nopXT)
	}
	warmRaw, ok := dt.Read(ReservedDTSlotWarmBootTarget).(Int)
	if !ok {
		t.Fatalf("warmBootTarget = %T, want Int XT", dt.Read(ReservedDTSlotWarmBootTarget))
	}
	warmReplW := globalDictionaryContext.dict.GetByName(bootWordWarmReplBody)
	if warmReplW == nil {
		t.Fatal("missing _bootWarmRepl colon word")
	}
	warmReplXT, ok := globalDictionaryContext.dict.XTOf(warmReplW)
	if !ok {
		t.Fatal("XTOf _bootWarmRepl")
	}
	if int(warmRaw) != int(warmReplXT) {
		t.Fatalf("warmBootTarget XT = %d, want _bootWarmRepl XT %d", warmRaw, warmReplXT)
	}
	if int(warmReplW.colonCode) != int(BootstrapReplLoopHeadCTP()) {
		t.Fatalf("_bootWarmRepl colonCode = %d, want repl loop head %d", warmReplW.colonCode, BootstrapReplLoopHeadCTP())
	}
}

func TestColdBootEntryRunsToStop(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	task := NewTestTask()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("")),
		label: "test",
	}})
	bootstrapRunColdBoot(task, rp, p)
}

func TestTaskSetColdBootTargetWritesDT(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	ct := globalDictionaryContext.ct
	body := ct.Len()
	ct.Append(int(OpNop))
	ct.Append(int(OpStop))
	task.SetColdBootTarget(EncodeVirt(ModuleIDThis, body))
	got, ok := globalDictionaryContext.dt.Read(ReservedDTSlotColdBootTarget).(Int)
	if !ok || int(got) != body {
		t.Fatalf("coldBootTarget = %v, want XT %d", got, body)
	}
}

func TestTaskSetWarmBootTargetWritesDT(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	xt := EncodeVirt(ModuleIDThis, 99)
	task.SetWarmBootTarget(xt)
	got, ok := globalDictionaryContext.dt.Read(ReservedDTSlotWarmBootTarget).(Int)
	if !ok || VirtualAddr(got) != xt {
		t.Fatalf("warmBootTarget = %v, want %v", got, xt)
	}
}

func TestResetTestRuntimeRestoresBootCT(t *testing.T) {
	ResetTestRuntime()
	globalDictionaryContext.ct.Append(int(OpNop))
	ResetTestRuntime()
	if globalDictionaryContext.ct.Len() <= int(BootstrapWarmBodyCTP()) {
		t.Fatalf("ct len = %d, want repl body end index %d", globalDictionaryContext.ct.Len(), BootstrapWarmBodyCTP())
	}
	TestBootCTVectorsAndReplLoop(t)
}
