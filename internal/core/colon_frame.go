package core

// Adjunct stack caps on ColonFrameMark (same-colon nesting only).
const (
	colonSlotItersMax    = 8  // foreach nesting; i/j/k only expose 3
	colonSlotCatchMax    = 8  // nested try in one colon
	colonSlotCtxMarksMax = 64 // nested ( { [ sandwiches in one colon
	colonSlotSwitchMax   = 3  // a / b / c ([162])
)

// ColonFrameMark is the return-stack frame for OpNest / colon entry (OpExec, OpExecWord).
// lastDC == nil: same execute tranche (moduleId 0). lastDC != nil: restore on OpUnnest.
// Adjunct *Stack / *DynamicObject fields are nil until a feature helper allocates them
// (sticky for the activation). subRets is a plain []int (subword return IPs); nil until first push.
// Toss zeros the whole RS cell.
// Final-method OS cleanup is compiled odrop before OpUnnest (not a dropCS field).
type ColonFrameMark struct {
	lastIP        int
	lastDC        *DictionaryContext
	iters         *Stack[Iterator]
	ctxMarks      *Stack[*ContextMark] // nested paren/brace/bracket; empty stack left sticky
	catchMarks    *Stack[*CatchMark]
	switchTargets *Stack[any] // switch targets ([162]); TOS = innermost / a
	locals        *DynamicObject
	subRets       []int // subword return IPs; append/truncate (187)
}

func (cf *ColonFrameMark) ensureIters() *Stack[Iterator] {
	if cf.iters == nil {
		cf.iters = NewStack[Iterator]("colonIters", colonSlotItersMax)
	}
	return cf.iters
}

func (cf *ColonFrameMark) ensureCatchMarks() *Stack[*CatchMark] {
	if cf.catchMarks == nil {
		cf.catchMarks = NewStack[*CatchMark]("colonCatch", colonSlotCatchMax)
	}
	return cf.catchMarks
}

func (cf *ColonFrameMark) ensureContextMarks() *Stack[*ContextMark] {
	if cf.ctxMarks == nil {
		cf.ctxMarks = NewStack[*ContextMark]("colonCtxMarks", colonSlotCtxMarksMax)
	}
	return cf.ctxMarks
}

func (cf *ColonFrameMark) ensureSwitchTargets() *Stack[any] {
	if cf.switchTargets == nil {
		cf.switchTargets = NewStack[any]("colonSwitchTargets", colonSlotSwitchMax)
	}
	return cf.switchTargets
}

// --- CatchMark (stack; nested try in one colon) ---

// PushCatchMark parks a CatchMark on this activation's catch stack.
func (cf *ColonFrameMark) PushCatchMark(cm *CatchMark) {
	cf.ensureCatchMarks().Push(cm)
}

// GetCatchMark soft-probes catch TOS. Does not allocate; does not pop.
func (cf *ColonFrameMark) GetCatchMark() (cm *CatchMark, ok bool) {
	if cf.catchMarks == nil || cf.catchMarks.Len() == 0 {
		return nil, false
	}
	return cf.catchMarks.Peek(), true
}

// PopCatchMark removes catch TOS. Requires catch stack; panics otherwise.
// Empty catchMarks stack stays allocated (sticky).
func (cf *ColonFrameMark) PopCatchMark() *CatchMark {
	if cf.catchMarks == nil {
		panic("ColonFrameMark.PopCatchMark: catchMarks is nil")
	}
	return cf.catchMarks.Pop()
}

// --- ContextMark (stack; nested paren/brace/bracket in one colon) ---

// PushContextMark parks a ContextMark on this activation's context stack.
func (cf *ColonFrameMark) PushContextMark(cm *ContextMark) {
	cf.ensureContextMarks().Push(cm)
}

// GetContextMark soft-probes context TOS. Does not allocate; does not pop.
func (cf *ColonFrameMark) GetContextMark() (cm *ContextMark, ok bool) {
	if cf.ctxMarks == nil || cf.ctxMarks.Len() == 0 {
		return nil, false
	}
	return cf.ctxMarks.Peek(), true
}

// GetClearContextMark soft-probes and pops context TOS if present. Does not allocate.
// Empty ctxMarks stack stays allocated (sticky). Used by RS walkers (getContextMark closers).
func (cf *ColonFrameMark) GetClearContextMark() (cm *ContextMark, ok bool) {
	if cf.ctxMarks == nil || cf.ctxMarks.Len() == 0 {
		return nil, false
	}
	return cf.ctxMarks.Pop(), true
}

// TakeContextMark pops and returns context TOS. Requires non-empty stack; panics otherwise.
// Empty ctxMarks stack stays allocated (sticky).
func (cf *ColonFrameMark) TakeContextMark() *ContextMark {
	if cf.ctxMarks == nil {
		panic("ColonFrameMark.TakeContextMark: ctxMarks is nil")
	}
	return cf.ctxMarks.Pop()
}

// --- Locals ---

// GetLocals returns the locals DynamicObject, creating it if needed.
func (cf *ColonFrameMark) GetLocals() *DynamicObject {
	if cf.locals == nil {
		cf.locals = NewDynamicObject()
	}
	return cf.locals
}

// --- Foreach iterators ---

// PushIterator parks an iterator on this activation's foreach stack.
func (cf *ColonFrameMark) PushIterator(iter Iterator) {
	cf.ensureIters().Push(iter)
}

// PopIterator removes and returns foreach TOS. Requires iters; panics if missing/empty.
func (cf *ColonFrameMark) PopIterator() Iterator {
	if cf.iters == nil {
		panic("ColonFrameMark.PopIterator: iters is nil")
	}
	return cf.iters.Pop()
}

// DropIterator pops foreach TOS without returning it. Requires iters; panics if missing/empty.
func (cf *ColonFrameMark) DropIterator() {
	_ = cf.PopIterator()
}

// TruncateItersTo pops foreach slots until Len() == n. Nil/empty with n==0 is a no-op.
// Panics if current depth is below n (try-entry invariant violated).
func (cf *ColonFrameMark) TruncateItersTo(n int) {
	truncateStackTo(cf.iters, n, "iters")
}

// PeekIterator returns foreach TOS without popping. Requires iters; panics if missing/empty.
func (cf *ColonFrameMark) PeekIterator() Iterator {
	if cf.iters == nil {
		panic("ColonFrameMark.PeekIterator: iters is nil")
	}
	return cf.iters.Peek()
}

// IteratorAt returns the iterator at TOS-relative index (0 = innermost). Requires iters; panics if OOB.
func (cf *ColonFrameMark) IteratorAt(tosIdx int) Iterator {
	if cf.iters == nil {
		panic("ColonFrameMark.IteratorAt: iters is nil")
	}
	return cf.iters.At(tosIdx)
}

// --- Switch targets ([162]) ---

// PushSwitchTarget parks a switch target.
func (cf *ColonFrameMark) PushSwitchTarget(v any) {
	cf.ensureSwitchTargets().Push(v)
}

// GetSwitchTarget soft-probes switch TOS. Does not allocate; does not pop.
func (cf *ColonFrameMark) GetSwitchTarget() (v any, ok bool) {
	if cf.switchTargets == nil || cf.switchTargets.Len() == 0 {
		return nil, false
	}
	return cf.switchTargets.Peek(), true
}

// SwitchTargetAt returns target at TOS-relative index (0 = innermost / a). Panics if missing/OOB.
func (cf *ColonFrameMark) SwitchTargetAt(tosIdx int) any {
	if cf.switchTargets == nil {
		panic("ColonFrameMark.SwitchTargetAt: switchTargets is nil")
	}
	return cf.switchTargets.At(tosIdx)
}

// PopSwitchTarget removes switch TOS. Requires stack; panics if missing/empty.
func (cf *ColonFrameMark) PopSwitchTarget() any {
	if cf.switchTargets == nil {
		panic("ColonFrameMark.PopSwitchTarget: switchTargets is nil")
	}
	return cf.switchTargets.Pop()
}

// DropSwitchTarget pops switch TOS without returning it.
func (cf *ColonFrameMark) DropSwitchTarget() {
	_ = cf.PopSwitchTarget()
}

// TruncateSwitchTargetsTo pops switch targets until Len() == n. Nil/empty with n==0 is a no-op.
func (cf *ColonFrameMark) TruncateSwitchTargetsTo(n int) {
	truncateStackTo(cf.switchTargets, n, "switchTargets")
}

// TruncateCtxMarksTo pops context marks until Len() == n. Nil/empty with n==0 is a no-op.
func (cf *ColonFrameMark) TruncateCtxMarksTo(n int) {
	truncateStackTo(cf.ctxMarks, n, "ctxMarks")
}

// truncateStackTo pops until Len() == n. Nil stack with n==0 is a no-op; n<0 or Len()<n panics.
func truncateStackTo[T any](s *Stack[T], n int, name string) {
	if n < 0 {
		panic("ColonFrameMark.Truncate: negative depth for " + name)
	}
	if s == nil {
		if n != 0 {
			panic("ColonFrameMark.Truncate: nil " + name + " below try-entry depth")
		}
		return
	}
	if s.Len() < n {
		panic("ColonFrameMark.Truncate: " + name + " shallower than try-entry depth")
	}
	for s.Len() > n {
		s.Pop()
	}
}

// --- Subword return IPs (187) ---

// PushSubRet appends a return IP for brsub.
func (cf *ColonFrameMark) PushSubRet(ip IP) {
	cf.subRets = append(cf.subRets, int(ip))
}

// PopSubRet for retsub removes and returns the top subword IP.
// Empty or nil subRets panics via ordinary slice indexing.
func (cf *ColonFrameMark) PopSubRet() IP {
	n := len(cf.subRets)
	ip := cf.subRets[n-1]
	cf.subRets = cf.subRets[:n-1]
	return IP(ip)
}

// ensureInnermostColonFrameMark returns RS TOS, pushing a BootStop floor frame
// if RS is empty (interpret / bare paren at session top).
func ensureInnermostColonFrameMark(rs *ReturnStackType) *ColonFrameMark {
	if rs.Len() > 0 {
		return rs.Peek()
	}
	return rs.PushNew(int(ReservedCTBootStop))
}

// takeContextMarkFromRS walks frames from TOS and GetClearContextMark on the first hit.
func takeContextMarkFromRS(rs *ReturnStackType) (*ContextMark, bool) {
	for i := 0; i < rs.Len(); i++ {
		if ctx, ok := rs.At(i).GetClearContextMark(); ok {
			return ctx, true
		}
	}
	return nil, false
}

// nestIntoColon enters a colon body. moduleID 0 keeps execDC; ≥1 switches via ModuleTable.
func nestIntoColon(execDC **DictionaryContext, rs *ReturnStackType, ip *IP, moduleID ModuleID, entry int) {
	if execDC == nil || *execDC == nil {
		panic("nestIntoColon: nil execDC")
	}
	mark := rs.PushNew(*ip)
	if moduleID != ModuleIDThis {
		mark.lastDC = *execDC
		*execDC = ModuleTable.DCForModule(nil, moduleID)
	}
	*ip = IP(entry)
}

// nestIntoColonWord enters w's colon body using same-module vs external moduleID rule.
func nestIntoColonWord(execDC **DictionaryContext, rs *ReturnStackType, ip *IP, w *Word) {
	moduleID := ModuleIDThis
	if w.moduleId != (*execDC).moduleId {
		moduleID = w.moduleId
	}
	nestIntoColon(execDC, rs, ip, moduleID, int(w.colonCode))
}

// unnestFromColon restores ip and optionally execDC from a ColonFrameMark.
func unnestFromColon(execDC **DictionaryContext, rs *ReturnStackType, ip *IP) {
	mark := rs.Peek()
	lastIP := mark.lastIP
	lastDC := mark.lastDC
	rs.Toss()
	*ip = IP(lastIP)
	if lastDC != nil {
		*execDC = lastDC
	}
}

// pushExecDtOperands leaves ( dtIdx:i -- ) on DS for OpDefer / OpDoVar after OpCtImmDs prefix.
func pushExecDtOperands(ds *Stack[any], slotIdx int) {
	ds.Push(Int(slotIdx))
}
