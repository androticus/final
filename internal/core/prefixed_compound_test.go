package core

import (
	"strings"
	"testing"
)

func TestDispatchSimplePrefixedOpRejectsUnflaggedWord(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	segs := CompoundPathWithSelectorForTest([]string{"find?"}, csTStack)
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected error for t:find?")
	}
	if !strings.Contains(err.Error(), "does not accept a stack selector prefix") {
		t.Fatalf("err = %v", err)
	}
}

func TestDispatchSimplePrefixedOpLoneSelectorErrors(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	cp := CompoundPathWithSelectorForTest(nil, csTStack)
	cp.path = ""
	err := dispatchCompoundToken(p, task, rp, cp)
	if err == nil {
		t.Fatal("expected error for lone t:")
	}
	if !strings.Contains(err.Error(), "requires an operand or compound path") {
		t.Fatalf("err = %v", err)
	}
}

func TestDispatchSimplePrefixedOpTDrop(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.TempStack.Push(Int(42))
	segs := CompoundPathWithSelectorForTest([]string{"drop"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0", rp.TempStack.Len())
	}
}

func TestCompileSimplePrefixedOpNoHandleCompound(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: ctprobe t:drop ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("ctprobe")
	if w == nil {
		t.Fatal("missing ctprobe")
	}
	ct := globalDictionaryContext.ct
	for off := int(w.colonCode); off < ct.Len(); off++ {
		op := OpCode(ct.Read(off))
		if op == OpHandleCompound {
			t.Fatalf("CT[%d] = OpHandleCompound, want simple prefixed compile path", off)
		}
		if op == OpStop {
			break
		}
	}
}
