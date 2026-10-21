package core

import "testing"

func TestAppendOpsWithCSArmingMatchesWordFlags(t *testing.T) {
	ResetTestRuntime()
	ct := &VirtCTResolver{}
	gct := globalDictionaryContext.ct
	for op := OpCode(0); op <= opAxiomLAST; op++ {
		wantArm := axiomWordForOp(op).RequiresStackContext()
		before := gct.Len()
		appendOpsWithCSArming(ct, csDataStack, op)
		wantCells := 1
		if wantArm {
			wantCells = 3
		}
		if got := gct.Len() - before; got != wantCells {
			t.Fatalf("op %s: appended %d CT cells, want %d (RequiresStackContext=%v)", OpCodeName(op), got, wantCells, wantArm)
		}
		gct.Truncate(before)
	}
}

func TestContextStackAwareWordsRequireStackContextFlag(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"@", "!", "@()", "!()", "@{}", "!{}", "^", ">ds", "ds>", "peek", "drop", "dup", "swap", "typeMemberDispatch", "nativeInvoke", "getContextMark"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing dictionary word %q", name)
		}
		if w.flags&WordFlagRequiresStackContext == 0 {
			t.Fatalf("word %q missing WordFlagRequiresStackContext", name)
		}
	}
	for _, name := range []string{"find?", "+", "stop", "(cs)", "cs,", "armCs", "handleCompound", "$"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing dictionary word %q", name)
		}
		if w.flags&WordFlagRequiresStackContext != 0 {
			t.Fatalf("word %q should not have WordFlagRequiresStackContext", name)
		}
	}
}
