package core

import "testing"

type finalTestHost struct{}

func (finalTestHost) FinalType() *Type { return finalTestHostType }

var finalTestHostType *Type

func installFinalTestAnswerWord(t *testing.T) *Word {
	t.Helper()
	ct := globalDictionaryContext.ct
	entry := ct.Len()
	ct.Append(int(OpDoLitI))
	ct.Append(42)
	ct.Append(int(OpODrop)) // Final method: pop receiver before Unnest
	ct.Append(int(OpUnnest))
	w := NewWord("FinalTestHost#_answer", WordFlagColon)
	w.colonCode = CTP(entry)
	w.moduleId = globalDictionaryContext.moduleId
	globalDictionaryContext.dict.Add(w)
	return w
}

func installFinalTestHostType(t *testing.T, w *Word) *MethodMetadata {
	t.Helper()
	finalTestHostType = &Type{
		name: "FinalTestHost",
		getZero: func() any {
			return finalTestHost{}
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	return finalTestHostType.RegisterFinalMethodFromWord(w)
}

func TestExposedMethodNameFromWord(t *testing.T) {
	t.Parallel()
	w := NewWord("FinalTestHost#_answer", WordFlagColon)
	if got := exposedMethodNameFromWord(w); got != "_answer" {
		t.Fatalf("exposedMethodNameFromWord = %q, want _answer", got)
	}
}

func TestRegisterFinalMethodFromWord(t *testing.T) {
	t.Parallel()
	typ := &Type{name: "FinalTestHost"}
	w := NewWord("FinalTestHost#_answer", WordFlagColon)
	w.colonCode = 100
	w.moduleId = ModuleIDThis
	mm := typ.RegisterFinalMethodFromWord(w)
	if !mm.isFinalMethod {
		t.Fatal("isFinalMethod = false, want true")
	}
	if mm.finalColonCode != 100 {
		t.Fatalf("finalColonCode = %d, want 100", mm.finalColonCode)
	}
	if typ.MethodByName("_answer") != mm {
		t.Fatal("MethodByName(_answer) mismatch")
	}
	if mm.call == nil {
		t.Fatal("call is nil")
	}
}

func TestProtoTypeHasNoRegisterFinalMethod(t *testing.T) {
	ResetTestRuntime()
	if protoTypeType == nil {
		t.Fatal("protoTypeType not registered")
	}
	if protoTypeType.HasClassMethod("registerFinalMethod") {
		t.Fatal("ProtoType.registerFinalMethod should be removed (tear-off `;` registers)")
	}
}

func TestFinalMethodInvokeViaMethodRef(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := installFinalTestAnswerWord(t)
	mm := installFinalTestHostType(t, w)

	ths := rp.Resolve(finalTestHost{})
	rp.DataStack.Push(&MethodRef{Parent: ths, MM: mm})
	RunOps(task, rp, OpMethodInvoke, OpStop)

	for i := 0; i < rp.DataStack.Len(); i++ {
		if _, ok := (*rp.DataStack.AddrAt(i)).(*MethodRef); ok {
			t.Fatal("MethodRef should be consumed by ^")
		}
	}

	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0 after odrop", rp.OStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
}

func TestFinalMethodInvokeViaCompoundPath(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := installFinalTestAnswerWord(t)
	installFinalTestHostType(t, w)

	rp.DataStack.Push(finalTestHost{})
	runHandleCompoundWithSegs(task, rp, []string{".", "_answer", "^"})

	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0 after odrop", rp.OStack.Len())
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + method result)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
	if _, ok := rp.DataStack.Pop().(finalTestHost); !ok {
		t.Fatalf("below TOS = %T, want finalTestHost receiver", rp.DataStack.Peek())
	}
}

// Interpret-time compounds (compileFromFile / session dispatch) must not allocate a bare
// RunFrame — Final methods need ForceOp for SetInterrupt (regression: typedObject_example.fnl).
func TestFinalMethodInvokeViaInterpretCompoundToken(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := installFinalTestAnswerWord(t)
	installFinalTestHostType(t, w)

	rp.DataStack.Push(finalTestHost{})
	segs := CompoundPathForTest([]string{".", "_answer", "^"})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}

	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0 after odrop", rp.OStack.Len())
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + method result)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
	if _, ok := rp.DataStack.Pop().(finalTestHost); !ok {
		t.Fatalf("below TOS = %T, want finalTestHost receiver", rp.DataStack.Peek())
	}
}

func TestFinalMethodInvokeViaInterpretYStackCompound(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := installFinalTestAnswerWord(t)
	installFinalTestHostType(t, w)

	rp.YStack.Push(finalTestHost{})
	segs := CompoundPathWithSelectorForTest([]string{".", "_answer", "^"}, csYStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}

	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0 after odrop", rp.OStack.Len())
	}
	if rp.YStack.Len() != 1 {
		t.Fatalf("YS len = %d, want 1 (receiver stays)", rp.YStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("TOS = %v, want Int(42)", got)
	}
}

func TestOpODropPopsOS(t *testing.T) {
	rp := NewTestRunProvider()
	rp.OStack.Push(Int(99))
	RunOps(NewTestTask(), rp, OpODrop, OpStop)
	if rp.OStack.Len() != 0 {
		t.Fatalf("OS len = %d, want 0", rp.OStack.Len())
	}
}

func TestNestUnnestNoLongerTouchesOS(t *testing.T) {
	task := NewTestTask()
	rp := NewTestRunProvider()
	rs := rp.ReturnStack
	osStack := rp.OStack
	execDC := task.dc
	osStack.Push(Int(99))
	var ip IP = 100
	nestIntoColon(&execDC, rs, &ip, ModuleIDThis, 200)
	unnestFromColon(&execDC, rs, &ip)
	if osStack.Len() != 1 {
		t.Fatalf("OS len after unnest = %d, want 1 (odrop is separate)", osStack.Len())
	}
	if int(ip) != 100 {
		t.Fatalf("ip = %d, want 100", ip)
	}
}
