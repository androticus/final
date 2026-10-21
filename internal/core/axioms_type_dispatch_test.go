package core

import "testing"

func TestTypeMemberDispatchDupNameSmoke(t *testing.T) {
	ResetTestRuntime()
	dupW := globalDictionaryContext.dict.GetByName("dup")
	if dupW == nil {
		t.Fatal("dup missing from dictionary")
	}
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, dupW)
	rp.DataStack.Push(String("name"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceGet), OpTypeMemberDispatch, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
}

func TestTypeMemberInvokeColonWord(t *testing.T) {
	ResetTestRuntime()
	_, xt, ok := FindName("typeMemberInvoke")
	if !ok {
		t.Fatal("typeMemberInvoke missing from bootstrapInitScript")
	}
	dupW := globalDictionaryContext.dict.GetByName("dup")
	if dupW == nil {
		t.Fatal("dup missing from dictionary")
	}
	rp := NewTestRunProvider()
	pushTypedOnContextStack(rp, dupW)
	rp.DataStack.Push(String("name"))
	rp.DataStack.Push(Int(TypeDispatchKindInstanceGet))
	rp.DataStack.Push(xt)
	RunOps(NewTestTask(), rp, OpExec, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
}

// Word.exec is a veneer over OpExec (130-worker-word-pattern.md): park Word on Y, y:.exec^.
func TestWordExecVeneerViaMethodRef(t *testing.T) {
	ResetTestRuntime()
	plusW := globalDictionaryContext.dict.GetByName("+")
	if plusW == nil {
		t.Fatal("+ missing from dictionary")
	}
	mm := plusW.FinalType().MethodByName("exec")
	if mm == nil {
		t.Fatal("Word.exec method missing")
	}
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	ths := rp.Resolve(plusW)
	rp.YStack.Push(plusW)
	rp.YStack.Push(&MethodRef{Parent: ths, MM: mm})
	RunOpsArmedCS(NewTestTask(), rp, csYStack, OpMethodInvoke, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("DS TOS = %v, want Int(3)", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after +", rp.DataStack.Len())
	}
	if rp.YStack.Len() != 1 || rp.YStack.Peek() != plusW {
		t.Fatalf("YS = %v, want parked + Word", rp.YStack.Peek())
	}
}

func TestOpNestUnnestReturns(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	ct := &VirtCTResolver{}
	body := ct.Append(ModuleIDThis, int(OpUnnest))
	ip := ct.Append(ModuleIDThis, int(OpNest))
	ct.Append(ModuleIDThis, 0)
	ct.Append(ModuleIDThis, decodeCTOffset(body))
	ct.Append(ModuleIDThis, int(OpStop))
	task := NewTestTask()
	task.ip = ip
	Run(task, rp)
}
