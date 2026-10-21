package core

import "testing"

type slotTestIter struct {
	cur int
}

func (it *slotTestIter) GetNext() bool {
	it.cur++
	return it.cur <= 1
}

func (it *slotTestIter) Current() any {
	return Int(it.cur)
}

func TestColonFrameMarkSoftProbesNoAlloc(t *testing.T) {
	cf := &ColonFrameMark{}
	if _, ok := cf.GetCatchMark(); ok {
		t.Fatal("GetCatchMark ok on empty mark")
	}
	if _, ok := cf.GetContextMark(); ok {
		t.Fatal("GetContextMark ok on empty mark")
	}
	if _, ok := cf.GetClearContextMark(); ok {
		t.Fatal("GetClearContextMark ok on empty mark")
	}
	if _, ok := cf.GetSwitchTarget(); ok {
		t.Fatal("GetSwitchTarget ok on empty mark")
	}
	if cf.iters != nil || cf.ctxMarks != nil || cf.catchMarks != nil || cf.switchTargets != nil || cf.locals != nil || cf.subRets != nil {
		t.Fatal("soft probes must not allocate adjunct stacks")
	}
}

func TestColonFrameMarkIterators(t *testing.T) {
	cf := &ColonFrameMark{}
	outer, inner := &slotTestIter{}, &slotTestIter{}
	cf.PushIterator(outer)
	cf.PushIterator(inner)
	if cf.IteratorAt(0) != inner || cf.IteratorAt(1) != outer {
		t.Fatalf("IteratorAt: got %v %v", cf.IteratorAt(0), cf.IteratorAt(1))
	}
	if cf.PeekIterator() != inner {
		t.Fatal("PeekIterator")
	}
	if got := cf.PopIterator(); got != inner {
		t.Fatalf("PopIterator = %v, want inner", got)
	}
	cf.DropIterator()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("PopIterator on empty should panic")
			}
		}()
		cf.PopIterator()
	}()
	if cf.iters == nil {
		t.Fatal("iters stack should remain sticky after pops")
	}
}

func TestColonFrameMarkCatchStack(t *testing.T) {
	cf := &ColonFrameMark{}
	a, b := &CatchMark{catchIP: 1}, &CatchMark{catchIP: 2}
	cf.PushCatchMark(a)
	cf.PushCatchMark(b)
	if cm, ok := cf.GetCatchMark(); !ok || cm != b {
		t.Fatalf("GetCatchMark = %v %v, want b", cm, ok)
	}
	if got := cf.PopCatchMark(); got != b {
		t.Fatalf("PopCatchMark = %v, want b", got)
	}
	if got := cf.PopCatchMark(); got != a {
		t.Fatalf("PopCatchMark = %v, want a", got)
	}
	if cf.catchMarks == nil {
		t.Fatal("catchMarks stack sticky after catch pops")
	}
}

func TestColonFrameMarkContextStack(t *testing.T) {
	cf := &ColonFrameMark{}
	a, b := &ContextMark{}, &ContextMark{}
	cf.PushContextMark(a)
	cf.PushContextMark(b)
	if cm, ok := cf.GetContextMark(); !ok || cm != b {
		t.Fatalf("GetContextMark = %v %v, want b", cm, ok)
	}
	if cm, ok := cf.GetClearContextMark(); !ok || cm != b {
		t.Fatalf("GetClearContextMark = %v %v, want b", cm, ok)
	}
	if cm, ok := cf.GetContextMark(); !ok || cm != a {
		t.Fatalf("GetContextMark after clear = %v %v, want a", cm, ok)
	}
	if cf.TakeContextMark() != a {
		t.Fatal("TakeContextMark")
	}
	if _, ok := cf.GetContextMark(); ok {
		t.Fatal("context stack should be empty")
	}
	if cf.ctxMarks == nil {
		t.Fatal("empty ctxMarks stack must remain sticky (not niled)")
	}
	if cf.ctxMarks.Len() != 0 {
		t.Fatalf("ctxMarks.Len() = %d, want 0", cf.ctxMarks.Len())
	}
}

func TestColonFrameMarkLocals(t *testing.T) {
	cf := &ColonFrameMark{}
	loc := cf.GetLocals()
	if loc == nil {
		t.Fatal("GetLocals nil")
	}
	if cf.GetLocals() != loc {
		t.Fatal("GetLocals should return same DO")
	}
}

func TestColonFrameMarkSwitchTargets(t *testing.T) {
	cf := &ColonFrameMark{}
	cf.PushSwitchTarget(Int(10))
	cf.PushSwitchTarget(Int(20))
	if v, ok := cf.GetSwitchTarget(); !ok || v != Int(20) {
		t.Fatalf("GetSwitchTarget = %v %v", v, ok)
	}
	if cf.SwitchTargetAt(1) != Int(10) {
		t.Fatalf("SwitchTargetAt(1) = %v", cf.SwitchTargetAt(1))
	}
	if cf.PopSwitchTarget() != Int(20) {
		t.Fatal("PopSwitchTarget")
	}
	cf.DropSwitchTarget()
}

func TestColonFrameMarkSubRets(t *testing.T) {
	cf := &ColonFrameMark{}
	if cf.subRets != nil {
		t.Fatal("subRets should start nil")
	}
	cf.PushSubRet(10)
	cf.PushSubRet(20)
	if len(cf.subRets) != 2 {
		t.Fatalf("len(subRets) = %d, want 2", len(cf.subRets))
	}
	if got := cf.PopSubRet(); got != 20 {
		t.Fatalf("PopSubRet = %d, want 20", got)
	}
	if got := cf.PopSubRet(); got != 10 {
		t.Fatalf("PopSubRet = %d, want 10", got)
	}
	if len(cf.subRets) != 0 {
		t.Fatalf("len after pops = %d, want 0", len(cf.subRets))
	}
	// truncate keeps backing array; slice is non-nil empty
	if cf.subRets == nil {
		t.Fatal("empty subRets after pops should remain non-nil (truncated)")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("PopSubRet on empty should panic")
			}
		}()
		cf.PopSubRet()
	}()
	cfNil := &ColonFrameMark{}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("PopSubRet on nil subRets should panic")
			}
		}()
		cfNil.PopSubRet()
	}()
}

func TestUnnestPopsFrame(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	execDC := task.dc
	rs := rp.ReturnStack
	var ip IP
	nestIntoColon(&execDC, rs, &ip, ModuleIDThis, 100)
	cf := rs.Peek()
	cf.PushIterator(&slotTestIter{})
	cf.GetLocals()
	cf.PushSubRet(42)
	if cf.iters == nil || cf.locals == nil || cf.subRets == nil {
		t.Fatal("expected adjuncts after writes")
	}
	unnestFromColon(&execDC, rs, &ip)
	if rs.Len() != 0 {
		t.Fatalf("RS len = %d, want 0 after unnest", rs.Len())
	}
	if ip != 0 {
		t.Fatalf("ip = %d, want restored lastIP 0", ip)
	}
}

func TestTruncateAdjunctsToTryEntry(t *testing.T) {
	cf := &ColonFrameMark{}
	cf.PushIterator(&slotTestIter{})
	cf.PushIterator(&slotTestIter{})
	cf.PushSwitchTarget(Int(1))
	cf.PushSwitchTarget(Int(2))
	cf.PushContextMark(&ContextMark{})
	cf.PushContextMark(&ContextMark{})

	cf.TruncateItersTo(1)
	cf.TruncateSwitchTargetsTo(1)
	cf.TruncateCtxMarksTo(1)
	if cf.iters.Len() != 1 || cf.switchTargets.Len() != 1 || cf.ctxMarks.Len() != 1 {
		t.Fatalf("depths = iters %d switch %d ctx %d, want 1 each",
			cf.iters.Len(), cf.switchTargets.Len(), cf.ctxMarks.Len())
	}

	cf.TruncateItersTo(0)
	cf.TruncateSwitchTargetsTo(0)
	cf.TruncateCtxMarksTo(0)
	if cf.iters.Len() != 0 || cf.switchTargets.Len() != 0 || cf.ctxMarks.Len() != 0 {
		t.Fatal("want empty adjuncts after truncate to 0")
	}

	// nil stacks + depth 0: no-op
	cf2 := &ColonFrameMark{}
	cf2.TruncateItersTo(0)
	cf2.TruncateSwitchTargetsTo(0)
	cf2.TruncateCtxMarksTo(0)
}
