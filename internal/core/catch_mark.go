package core

// CatchMark is the try/catch checkpoint: full value-stack cell snapshot (no DC)
// plus try-entry adjunct depths on the owning ColonFrameMark only.
// Parked on ColonFrameMark.catchMarks (stack), not on RS.
type CatchMark struct {
	snaps   [csCount][]any
	catchIP int // absolute CT index of catch body entry

	// Adjunct depths on the owning ColonFrameMark at try entry.
	// Throw truncates those stacks back to these lengths (drop only what was
	// pushed inside the try). Other RS frames are not involved.
	itersLen    int
	switchLen   int
	ctxMarksLen int
}

func snapshotValueStacks(rp *RunProvider) [csCount][]any {
	return [csCount][]any{
		rp.DataStack.SnapshotItems(),
		rp.TempStack.SnapshotItems(),
		rp.XStack.SnapshotItems(),
		rp.YStack.SnapshotItems(),
		rp.ZStack.SnapshotItems(),
		rp.OStack.SnapshotItems(),
	}
}

func restoreValueStacks(rp *RunProvider, snaps [csCount][]any) {
	rp.DataStack.RestoreItems(snaps[csDataStack])
	rp.TempStack.RestoreItems(snaps[csTStack])
	rp.XStack.RestoreItems(snaps[csXStack])
	rp.YStack.RestoreItems(snaps[csYStack])
	rp.ZStack.RestoreItems(snaps[csZStack])
	rp.OStack.RestoreItems(snaps[csOStack])
}

func adjunctLens(cf *ColonFrameMark) (iters, switches, ctx int) {
	if cf.iters != nil {
		iters = cf.iters.Len()
	}
	if cf.switchTargets != nil {
		switches = cf.switchTargets.Len()
	}
	if cf.ctxMarks != nil {
		ctx = cf.ctxMarks.Len()
	}
	return
}

// pushCatchMark snapshots value stacks and adjunct depths, parks CatchMark on innermost CFM.
func pushCatchMark(rp *RunProvider, catchIP int) {
	cf := ensureInnermostColonFrameMark(rp.ReturnStack)
	iters, switches, ctx := adjunctLens(cf)
	cf.PushCatchMark(&CatchMark{
		snaps:       snapshotValueStacks(rp),
		catchIP:     catchIP,
		itersLen:    iters,
		switchLen:   switches,
		ctxMarksLen: ctx,
	})
}

// dropCatchMark soft-drops catch TOS on the innermost ColonFrameMark.
// No-op if already destacked (throw path popped before catch body).
func dropCatchMark(rs *ReturnStackType) {
	cf := rs.Peek()
	if _, ok := cf.GetCatchMark(); !ok {
		return
	}
	cf.PopCatchMark()
}
