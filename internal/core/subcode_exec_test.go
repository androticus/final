package core

import "testing"

func TestSubCodeFromDSAxiom(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(42))
	RunOps(task, rp, OpSubCodeFromDS, OpStop)
}

func TestOpDoSubCode(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, 7)
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
}

func TestCompileSubCodeEmits(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "subprobe")
	w := cs.DC.dict.PeekBuilding()
	SetCompileState(cs, true)
	rp.DataStack.Push(Int(3))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("subCode,")); err != nil {
		t.Fatal(err)
	}
	body := int(w.colonCode)
	if globalDictionaryContext.ct.Read(body) != int(OpDoSubCode) {
		t.Fatalf("CT[%d] = %d, want OpDoSubCode", body, globalDictionaryContext.ct.Read(body))
	}
	if globalDictionaryContext.ct.Read(body+1) != 3 {
		t.Fatalf("CT[%d] = %d, want operand 3", body+1, globalDictionaryContext.ct.Read(body+1))
	}
}

func TestTransitionUsesRunSubCode(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, int(TrCheckResetByeReq))
	ct.Append(ModuleIDThis, int(OpTransition))
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
}
