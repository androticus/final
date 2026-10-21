package core

import "fmt"

var bootstrapReplBodyEndCTP CTP

// InitializeBootstrapCT reserves CT[0..6]: cold/warm OpCtImmDs·dtIdx·OpBootstrapDefer chains, then OpStop.
// Defer targets live in DT[ReservedDTSlotColdBootTarget] and DT[ReservedDTSlotWarmBootTarget].
func InitializeBootstrapCT(dc *DictionaryContext) {
	if dc == nil || dc.ct == nil {
		panic("InitializeBootstrapCT: nil DictionaryContext or CT")
	}
	ct := dc.ct
	if ct.Len() != 0 {
		panic(fmt.Errorf("InitializeBootstrapCT: CT len %d, want 0 before init", ct.Len()))
	}
	ct.Append(int(OpCtImmDs))
	ct.Append(ReservedDTSlotColdBootTarget)
	ct.Append(int(OpBootstrapDefer))
	ct.Append(int(OpCtImmDs))
	ct.Append(ReservedDTSlotWarmBootTarget)
	ct.Append(int(OpBootstrapDefer))
	ct.Append(int(OpStop))
}

// SetColdBootTarget stores implXT in the cold-boot defer DT slot (rebind without IS).
func (t *Task) SetColdBootTarget(implXT VirtualAddr) {
	if t == nil || t.dc == nil {
		panic("SetColdBootTarget: nil Task or DictionaryContext")
	}
	t.dc.dt.Write(ReservedDTSlotColdBootTarget, Int(implXT))
}

// SetWarmBootTarget stores implXT in the warm-boot defer DT slot.
func (t *Task) SetWarmBootTarget(implXT VirtualAddr) {
	if t == nil || t.dc == nil {
		panic("SetWarmBootTarget: nil Task or DictionaryContext")
	}
	t.dc.dt.Write(ReservedDTSlotWarmBootTarget, Int(implXT))
}

// BootstrapGlobalColdEntry is the VirtualAddr for the global boot vector (CT[0] OpCtImmDs).
func BootstrapGlobalColdEntry() VirtualAddr {
	return EncodeVirt(ModuleIDThis, ReservedCTBootColdEntry)
}

// bootstrapRunColdBoot runs the in-VM repl driver until OpStop (stream EOF on parser).
func bootstrapRunColdBoot(t *Task, rp *RunProvider, p *Parser) {
	if t == nil || rp == nil || p == nil {
		panic("bootstrapRunColdBoot: nil Task, RunProvider, or Parser")
	}
	rp.SetParser(p)
	t.ip = BootstrapGlobalColdEntry()
	RunREPLSession(t, rp)
	cs := t.compilerForTarget()
	if GetCompileState(cs) {
		SetCompileState(cs, false)
	}
}

// BootstrapReplLoopHeadCTP is the loop head inside the warm repl body (fault re-entry target).
func BootstrapReplLoopHeadCTP() CTP {
	return ReservedCTBootBodiesStart
}

// BootstrapWarmBodyCTP returns the CT index just past the warm repl body (for fault labeling).
func BootstrapWarmBodyCTP() CTP {
	return bootstrapReplBodyEndCTP
}

func appendBootTransition(ct *CodeTable, id RuntimeDispatchID) {
	ct.Append(int(OpDoSubCode))
	ct.Append(int(id))
	ct.Append(int(OpTransition))
}

func patchBootBranchOffset(ct *CodeTable, branchOpcodeIdx int, targetIdx int) {
	jSlot := branchOpcodeIdx + 1
	branchBase := jSlot + 1
	ct.Write(jSlot, targetIdx-branchBase)
}

// buildWarmBootReplBody emits the in-VM repl loop at the current CT end.
//
//	loopHead:
//	  TrCheckResetByeReq
//	  branch?  loopExit
//	  TrNextToken
//	  branch?  loopExit
//	  TrSessionDispatch
//	  branch?  skipExec
//	  OpExecWord
//	  branch   loopHead
//	loopExit:
//	  OpUnnest   (return to CT[8] OpStop via warm defer nest)
func buildWarmBootReplBody(ct *CodeTable) CTP {
	loopHead := ct.Len()

	appendBootTransition(ct, TrCheckResetByeReq)
	branchQBye := ct.Len()
	ct.Append(int(OpBranchQ))
	ct.Append(0)

	appendBootTransition(ct, TrNextToken)
	branchQEof := ct.Len()
	ct.Append(int(OpBranchQ))
	ct.Append(0)

	appendBootTransition(ct, TrSessionDispatch)
	branchQExec := ct.Len()
	ct.Append(int(OpBranchQ))
	ct.Append(0)

	ct.Append(int(OpExecWord))

	branchBack := ct.Len()
	ct.Append(int(OpBranch))
	ct.Append(0)

	loopExit := ct.Len()
	ct.Append(int(OpUnnest))

	patchBootBranchOffset(ct, branchQBye, loopExit)
	patchBootBranchOffset(ct, branchQEof, loopExit)
	patchBootBranchOffset(ct, branchQExec, branchBack)
	patchBootBranchOffset(ct, branchBack, loopHead)

	return CTP(loopHead)
}

// wireBootstrapBootTargets builds the warm repl body and stores defer target XTs in reserved DT slots.
// Cold falls through to warm: CT[0..2] cold defer chain → CT[3..5] warm defer chain → CT[6] OpStop.
func wireBootstrapBootTargets(t *Task, dc *DictionaryContext) {
	if t == nil || dc == nil {
		panic("wireBootstrapBootTargets: nil Task or DictionaryContext")
	}
	nopW := dc.dict.GetByName("nop")
	if nopW == nil {
		panic("wireBootstrapBootTargets: missing nop axiom")
	}
	nopXT, ok := dc.dict.XTOf(nopW)
	if !ok {
		panic("wireBootstrapBootTargets: XTOf nop")
	}
	replHead := buildWarmBootReplBody(dc.ct)
	bootstrapReplBodyEndCTP = CTP(dc.ct.Len())
	warmReplW := NewWord(bootWordWarmReplBody, WordFlagHidden|WordFlagColon)
	warmReplW.colonCode = replHead
	dc.dict.Add(warmReplW)
	warmReplXT, ok := dc.dict.XTOf(warmReplW)
	if !ok {
		panic("wireBootstrapBootTargets: XTOf warm repl body")
	}
	dc.dt.Write(ReservedDTSlotColdBootTarget, Int(nopXT))
	dc.dt.Write(ReservedDTSlotWarmBootTarget, Int(warmReplXT))
}
