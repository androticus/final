package core

import (
	"fmt"
)

const throwCausedByPrefix = "\n  caused by: "

// Throw arms OpThrow for the next run-loop iteration. Call sites must return after Throw.
// If OpThrow is already pending, pops the prior message and joins outer-first with "caused by:".
func Throw(msg string, f *RunFrame) {
	if f == nil || f.RP == nil {
		panic("Throw: nil RunFrame")
	}
	rp := f.RP
	ds := rp.DataStack
	if throwPending(f) {
		prior, ok := ds.Pop().(String)
		if !ok {
			panic(fmt.Errorf("Throw: pending throw but DS TOS is %T, want String", ds.Peek()))
		}
		ds.Push(String(msg + throwCausedByPrefix + string(prior)))
		return
	}
	ds.Push(String(msg))
	armThrow(f)
}

func throwPending(f *RunFrame) bool {
	if f.Task != nil && f.Task.pendingThrow {
		return true
	}
	if f.ForceOp != nil && *f.ForceOp == OpThrow {
		return true
	}
	if f.Task != nil && f.Task.forceOp != nil && *f.Task.forceOp == OpThrow {
		return true
	}
	return false
}

func armThrow(f *RunFrame) {
	if f.ForceOp != nil {
		*f.ForceOp = OpThrow
	} else if f.Task != nil {
		f.Task.pendingForceOp = OpThrow
	} else {
		panic("Throw: no ForceOp or Task to arm OpThrow")
	}
	if f.Task != nil {
		f.Task.pendingThrow = true
	}
}

func clearPendingThrow(t *Task) {
	if t == nil {
		return
	}
	t.pendingThrow = false
	if t.pendingForceOp == OpThrow {
		t.pendingForceOp = 0
	}
	if t.forceOp != nil && *t.forceOp == OpThrow {
		*t.forceOp = 0
	}
}

// throwUnnestFloor is the RS depth throw must not pop through: nested Run's saved
// depth, or 1 (session BootStop floor) when no nested Run is active.
func throwUnnestFloor(t *Task) int {
	floor := 1
	if t != nil && t.rsNestFloor > floor {
		floor = t.rsNestFloor
	}
	return floor
}

// runOpThrow unwinds ColonFrameMarks until a CatchMark is found on a catch stack,
// restores value stacks, leaves msg on DS. Does not unnest the nest/session floor.
func runOpThrow(t *Task, rp *RunProvider, execDC **DictionaryContext, ip *IP) {
	ds := rp.DataStack
	rs := rp.ReturnStack
	if ds.Len() == 0 {
		panic(fmt.Errorf("throw: empty data stack"))
	}
	msg, ok := ds.Pop().(String)
	if !ok {
		panic(fmt.Errorf("throw: TOS is %T, want String", ds.Peek()))
	}

	floor := throwUnnestFloor(t)
	var catch *CatchMark
	for rs.Len() > 0 {
		cf := rs.Peek()
		if _, ok := cf.GetCatchMark(); ok {
			catch = cf.PopCatchMark()
			goto found
		}
		if rs.Len() <= floor {
			// Nest/session floor: leave it; uncaught.
			clearPendingThrow(t)
			panic(msg)
		}
		unnestFromColon(execDC, rs, ip)
	}
	clearPendingThrow(t)
	panic(msg)

found:
	cf := rs.Peek()
	restoreValueStacks(rp, catch.snaps)
	cf.TruncateItersTo(catch.itersLen)
	cf.TruncateSwitchTargetsTo(catch.switchLen)
	cf.TruncateCtxMarksTo(catch.ctxMarksLen)
	cf.subRets = nil // 187: abandon subword returns on this frame
	clearPendingThrow(t)
	*ip = IP(catch.catchIP)
	ds.Push(msg)
}
