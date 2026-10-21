package core

import "testing"

func TestCSSelectFromDSAxiomBindsRunLoopCS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.TempStack.Push(Int(99))
	rp.DataStack.Push(Int(1))
	RunOpsArmedCS(task, rp, csTStack, OpDrop, OpStop)
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after cs! drop", rp.TempStack.Len())
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (unchanged marker)", rp.DataStack.Len())
	}
}

func TestOpExecStackContextWrapsFetchAxiom(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("probe", 0)
	ths := NewTypeObjPair(w)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(w)
	rp.DataStack.Push(&PropertyRef{Parent: ths, PM: pm})
	_, xt, ok := FindName("@")
	if !ok {
		t.Fatal("FindName @")
	}
	rp.DataStack.Push(xt)
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoCSSelect))
	ct.Append(ModuleIDThis, int(csDataStack))
	ct.Append(ModuleIDThis, int(OpExec))
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "probe" {
		t.Fatalf("fetch = %v, want probe", got)
	}
}

func TestExecuteInterpretWordStackContextWrapsDotNameFetch(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name", "@"})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("fetch = %v, want hello", got)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (receiver kept on CS)", rp.DataStack.Len())
	}
}
