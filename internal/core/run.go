package core

import (
	"fmt"
	"math"
	"time"
)

type ContextStackIndex = int

const (
	csDataStack ContextStackIndex = 0
	csTStack    ContextStackIndex = 1
	csXStack    ContextStackIndex = 2
	csYStack    ContextStackIndex = 3
	csZStack    ContextStackIndex = 4
	csOStack    ContextStackIndex = 5
	csCount     ContextStackIndex = 6
)

// Dispatch a task to run in a new goroutine.
func DoRun(t *Task, makeDoneChan bool) (<-chan error, error) {
	if makeDoneChan {
		t.doneChan = make(chan error, 1)
	}
	go Run(t, NewRunProvider())
	return t.doneChan, nil
}

func Run(t *Task, rp *RunProvider) {
	if rp == nil {
		panic(fmt.Errorf("Run: nil RunProvider"))
	}
	runLoop(t, rp)
}

type Stacks = [csCount]*Stack[any]

func runLoop(t *Task, rp *RunProvider) {
	forceOp := OpCode(0)
	t.forceOp = &forceOp
	if rp != nil {
		rp.WithTaskFindScope(t)
	}
	ds := rp.DataStack
	rs := rp.ReturnStack
	ts := rp.TempStack
	xs := rp.XStack
	ys := rp.YStack
	zs := rp.ZStack
	os := rp.OStack
	var stacks Stacks = [csCount]*Stack[any]{ds, ts, xs, ys, zs, os}
	var cs *Stack[any] = nil // nil by default, to detect missing arming                                       // default
	runSubCode := -1
	stdin := rp.Stdin
	_ = stdin // TEMP: unused
	stdout := rp.Stdout
	stderr := rp.Stderr
	_ = stderr // TEMP: unused
	var ip IP
	var op OpCode
	var execDC *DictionaryContext
	runFrame := RunFrame{Task: t, RP: rp, SubCode: &runSubCode, ForceOp: &forceOp}

	defer func() {
		if r := recover(); r != nil {
			rp.LastOp = op
			rp.LastIP = ip
			rp.LastExecDC = execDC
			cs := t.compilerForTarget()
			drainSource := cs.txnActiveForTest()
			_ = cs.tryRollbackTransaction()
			if drainSource && rp.Parser != nil {
				drainParserToEOF(rp.Parser)
			}
			fault := captureRunFault(t, rp, r, op, ip)
			if t.doneChan != nil {
				t.doneChan <- fault
				t.escalateFaultUp(fault, rp)
			} else {
				panic(r)
			}
		} else if t.doneChan != nil {
			t.doneChan <- nil
		}
	}()

	// demuxDC resolves a packed VirtualAddr to (DC, plain index).
	// High half: moduleId 0 = execute DC ("this"); ≥1 = ModuleTable slot. No special "core" case.
	demuxDC := func(virtAddr VirtualAddr) (*DictionaryContext, int) {
		moduleId, packedOff := unpackVirtAddr(virtAddr)
		switch moduleId {
		case ModuleIDThis:
			if execDC != nil {
				return execDC, packedOff
			}
			return globalDictionaryContext, packedOff
		default:
			dc := ModuleTable.DCForModule(t, moduleId)
			return dc, packedOff - int(moduleId)
		}
	}
	execDC, ip = demuxDC(t.ip)

	readCTOperand := func() int {
		v := execDC.ct.Read(int(ip))
		ip++
		return v
	}

	if t.pendingExecWord != nil {
		w := t.pendingExecWord
		t.pendingExecWord = nil
		if w.RequiresStackContext() {
			sel := t.pendingExecCS
			if sel < 0 || sel >= csCount {
				panic(fmt.Errorf("pendingExec: selector %d out of range [0,%d)", sel, csCount))
			}
			cs = stacks[sel]
		}
		var nested bool
		op, nested = armExecWord(t, w, &execDC, ds, &runSubCode, rs, &ip)
		if !nested {
			forceOp = op
		}
	} else if t.pendingForceOp != 0 {
		forceOp = t.pendingForceOp
		t.pendingForceOp = 0
	}

	assertCS := func() {
		if cs == nil {
			panic(fmt.Errorf("CS is nil"))
		}
	}
	resetCS := func() {
		cs = nil
	}

	for {
		if forceOp == 0 {
			//op = OpCode(readCTOperand())
			// unwound.. hot path
			op = execDC.ct.Read(int(ip))
			ip++
		
		} else {
			op = forceOp
			forceOp = 0
		}

	reenter:
		switch op {
		case OpINVALID:
			panic(fmt.Errorf("Atempt to execute OpCode 0 (INVALID)"))
		case OpStop:
			goto runDone

		// stack axioms (CS-bound; pick index still from DS)
		case OpDup:
			assertCS()
			cs.Push(cs.Peek())
			resetCS()
		case OpDDup:
			ds.Push(ds.Peek())
		case OpQDup:
			assertCS()
			if AnyToTruthy(cs.Peek()) {
				cs.Push(cs.Peek())
			}
			resetCS()
		case OpDup2: // c( a b -- a b a b )
			assertCS()
			cs.Push(cs.At(1))
			cs.Push(cs.At(1))
			resetCS()
		case OpDDup2: // c( a b -- a b a b )
			ds.Push(ds.At(1))
			ds.Push(ds.At(1))
		case OpDrop:
			assertCS()
			cs.Pop()
			resetCS()
		case OpDDrop:
			ds.Pop()
		case OpDrop2:
			assertCS()
			cs.Pop()
			cs.Pop()
			resetCS()
		case OpDDrop2:
			ds.Pop()
			ds.Pop()
		case OpSwap: //(a b -- b a)
			assertCS()
			pb := cs.AddrAt(0)
			pa := cs.AddrAt(1)
			*pa, *pb = *pb, *pa
			resetCS()
		case OpDSwap: //(a b -- b a)
			pb := ds.AddrAt(0)
			pa := ds.AddrAt(1)
			*pa, *pb = *pb, *pa
		case OpOver: // ( a b -- a b a )
			assertCS()
			b := cs.Pop()
			a := cs.Pop()
			cs.Push(a)
			cs.Push(b)
			cs.Push(a)
			resetCS()
		case OpNip: //(a b -- b)
			assertCS()
			b := cs.Pop()
			cs.Pop()
			cs.Push(b)
			resetCS()
		case OpTuck: // ( a b -- b a b )
			assertCS()
			b := cs.Pop()
			a := cs.Pop()
			cs.Push(b)
			cs.Push(a)
			cs.Push(b)
			resetCS()
		case OpPick: //(z..a n:iu -- z..a ?) nth element from top
			assertCS()
			var i int
			switch n := ds.Pop().(type) {
			case Int:
				i = int(n)
			case Uint:
				i = int(n)
			default:
				panic(fmt.Errorf("OpPick: invalid index type: %T", n))
			}
			elem := *(cs.AddrAt(i))
			cs.Push(elem)
			resetCS()
		case OpRot:
			assertCS()
			c := cs.Pop()
			b := cs.Pop()
			a := cs.Pop()
			cs.Push(b)
			cs.Push(c)
			cs.Push(a)
			resetCS()
		case OpDStoTS:
			ts.Push(ds.Pop())
		case OpTStoDS:
			ds.Push(ts.Pop())
		case OpTDrop:
			ts.Pop()
		case OpRPeek:
			ds.Push(rs.Peek())
		case OpDStoXS:
			xs.Push(ds.Pop())
		case OpXStoDS:
			ds.Push(xs.Pop())
		case OpXDrop:
			xs.Pop()
		case OpDStoYS:
			ys.Push(ds.Pop())
		case OpYStoDS:
			ds.Push(ys.Pop())
		case OpYDrop:
			ys.Pop()
		case OpDStoZS:
			zs.Push(ds.Pop())
		case OpZStoDS:
			ds.Push(zs.Pop())
		case OpZDrop:
			zs.Pop()

		case OpDepth:
			assertCS()
			ds.Push(Int(cs.Len()))
			resetCS()
		case OpIsTyped:
			if _, ok := ds.Peek().(Typed); ok {
				ds.Push(Bool(true))
			} else {
				ds.Push(Bool(false))
			}

		case OpDSto:
			assertCS()
			cs.Push(ds.Pop())
			resetCS()
		case OpToDS:
			assertCS()
			ds.Push(cs.Pop())
			resetCS()
		case OpPeek:
			assertCS()
			ds.Push(cs.Peek())
			resetCS()
		case OpODrop:
			os.Pop()

		case OpBitNot:
			doIntegralBitMonadic(ds, op)
		case OpBitAnd, OpBitOr, OpBitXor, OpBitSet, OpBitClear, OpBitToggle:
			doIntegralBitDyadic(ds, op)

		case OpBoolAnd, OpBoolOr:
			doDyadicBooleanOp(ds, op)
		case OpBoolNot:
			doMonadicBooleanOp(ds, op)

		case OpEqual, OpNotEqual:
			doDyadicEquateOp(ds, op)
		case OpZeroEqual, OpZeroNotEqual,
			OpZeroLessThan, OpZeroLessThanOrEqual,
			OpZeroGreaterThan, OpZeroGreaterThanOrEqual:
			doMonadicCompare0Op(ds, op)

		case OpLessThan, OpLessThanOrEqual, OpGreaterThan, OpGreaterThanOrEqual:
			doDyadicCompareOp(ds, op)
		case OpCompare:
			doCompareOp(ds)

		case OpAdd, OpSub, OpMul, OpDiv:
			doDyadicArithOp(ds, op)
		case OpAbs, OpNeg:
			doMonadicNumericOp(ds, op)

		case OpShiftLeft:
			integralShiftLeft(ds)
		case OpShiftRight:
			integralShiftRight(ds)
		case OpRotateLeft:
			integralRotateLeft(ds)
		case OpRotateRight:
			integralRotateRight(ds)
		case OpRem:
			integralRem(ds)
		case OpModRem:
			integralModRem(ds)

		case OpCompileLit: // (v:Any -- ) c( -- op v) compile literal to CT; op is the OpCode to execute
			compileEmitLiteralCells(t, t.compilerForTarget(), t.dc, ds.Pop())
		case OpDoLit0:
			ds.Push(Int(0))
		case OpDoLitB:
			ds.Push(Bool(readCTOperand() != 0))
		case OpDoLitI:
			ds.Push(Int(readCTOperand()))
		case OpDoLitU:
			ds.Push(Uint(readCTOperand()))
		case OpDoLitF: // need to binary decode the float from the CT
			ds.Push(math.Float64frombits(uint64(readCTOperand())))
		case OpDoLitNull:
			ds.Push(Any(nil))
		case OpDoLit:
			slotIdx := readCTOperand()
			ds.Push(execDC.dt.Read(slotIdx))

		case OpNest:
			moduleID := ModuleID(readCTOperand())
			entry := readCTOperand()
			nestIntoColon(&execDC, rs, &ip, moduleID, entry)

		case OpUnnest:
			unnestFromColon(&execDC, rs, &ip)

		case OpInterrupt:
			entry := int(ds.Pop().(Int))
			moduleID := ModuleID(ds.Pop().(Int))
			nestIntoColon(&execDC, rs, &ip, moduleID, entry)

		case OpDefer:
			slotIdx := int(ds.Pop().(Int))
			xtVirt := XT(execDC.dt.Read(slotIdx).(Int))
			ds.Push(xtVirt)
			op = OpExec
			goto reenter

		case OpBootstrapDefer:
			slotIdx := int(ds.Pop().(Int))
			coreDC := ModuleTable.DCForModule(t, CoreModuleSlot)
			xtVirt := XT(coreDC.dt.Read(slotIdx).(Int))
			if xtVirt == 0 {
				panic(fmt.Errorf("bootstrapDefer: unset target at core DT[%d]", slotIdx))
			}
			ds.Push(xtVirt)
			op = OpExec
			goto reenter

		case OpExec:
			xtVirt := popExecXT(ds)
			if xtVirt == 0 {
				panic(fmt.Errorf("exec: VirtXT was 0"))
			}
			operandDC, baseAddr := demuxDC(xtVirt)
			w := operandDC.dict.GetByOrdinal(int(baseAddr))
			if w == nil {
				panic(fmt.Errorf("exec: invalid xt %v", xtVirt))
			}
			var nested bool
			op, nested = armExecWord(t, w, &execDC, ds, &runSubCode, rs, &ip)
			if !nested {
				goto reenter
			}

		case OpExecWord:
			w := ds.Pop().(*Word)
			var nested bool
			op, nested = armExecWord(t, w, &execDC, ds, &runSubCode, rs, &ip)
			if !nested {
				goto reenter
			}

		case OpBranch:
			j := readCTOperand()
			ip += IP(j) // branches only ever in a single module
		case OpBranchQ:
			a := ds.Pop()
			b := AnyToTruthy(a)
			j := readCTOperand()
			if !b {
				ip += IP(j) // branches only ever in a single module
			}
		case OpBrSub:
			j := readCTOperand()
			// ip is already past the offset cell — that is the return address.
			rs.Peek().PushSubRet(ip)
			ip += IP(j)
		case OpRetSub:
			ip = rs.Peek().PopSubRet()
		case OpTryGetIter:
			// Prefer Iterator over Iterable when both apply:
			// an existing Iterator on DS is used as-is; GetIter only for pure Iterables.
			provider := ds.Peek()
			switch v := provider.(type) {
			case Iterator:
				ds.Push(true)
			case Iterable:
				ds.Pop()
				ds.Push(v.GetIter())
				ds.Push(true)
			default:
				ds.Push(false)
			}
		case OpForeachLoopTop:
			// RS TOS = ColonFrameMark; iter on Slot.Foreach (helpers panic if missing).
			cm := rs.Peek()
			if cm.PeekIterator().GetNext() {
				ds.Push(true) // branch? falls through into body
			} else {
				// Slot drop is shared exit epilogue after /foreach (break + exhaust).
				ds.Push(false) // branch? skips to epilogue
			}
		case OpI:
			cm := rs.Peek()
			ds.Push(cm.IteratorAt(0).Current())
		case OpJ:
			cm := rs.Peek()
			ds.Push(cm.IteratorAt(1).Current())
		case OpK:
			cm := rs.Peek()
			ds.Push(cm.IteratorAt(2).Current())
		case OpPushIterator:
			cm := rs.Peek()
			cm.PushIterator(ds.Pop().(Iterator))
		case OpDropIterator:
			cm := rs.Peek()
			cm.DropIterator()

		case OpLocals:
			// Same-colon only: RS TOS is this activation's ColonFrameMark (no intervening marks).
			cm := rs.Peek()
			ds.Push(cm.GetLocals())

		case OpPushSwitch:
			cm := rs.Peek()
			cm.PushSwitchTarget(ds.Pop())
		case OpDropSwitch:
			cm := rs.Peek()
			cm.DropSwitchTarget()
		case OpSwitchA:
			cm := rs.Peek()
			ds.Push(cm.SwitchTargetAt(0))
		case OpSwitchB:
			cm := rs.Peek()
			ds.Push(cm.SwitchTargetAt(1))
		case OpSwitchC:
			cm := rs.Peek()
			ds.Push(cm.SwitchTargetAt(2))
		case OpAssertSwitchNest:
			panicIfSwitchNotInArm(rp.TempStack, "'switch' cannot nest directly inside 'switch' (only inside case/default)")

		case OpCTHere:
			ds.Push(Int(t.compileHere(t.compilerForTarget())))
		case OpCTComma:
			n := ds.Pop().(Int)
			_ = t.compileAppend(t.compilerForTarget(), int(n))
		case OpCTStore:
			idx := int(ds.Pop().(Int))
			val := int(ds.Pop().(Int))
			t.compilerForTarget().ct().Write(idx, val)
		case OpCTPatchForward:
			orig := int(ds.Pop().(Int))
			j := t.compileHere(t.compilerForTarget()) - (orig + 1) // resolveIP - markIP; orig is placeholder at markIP-1
			t.compilerForTarget().ct().Write(orig, j)
		case OpCTPatchForwardNext:
			orig := int(ds.Pop().(Int))
			j := t.compileHere(t.compilerForTarget()) + 1 - (orig + 1)
			t.compilerForTarget().ct().Write(orig, j)

		case OpSetWordFlags:
			// WordFlags only — never touches srFlags (SR has no Final write path).
			u, ok := integralToUint32(ds.Pop())
			if !ok {
				panic(fmt.Errorf("setWordFlags: expected Int or Uint"))
			}
			flags := WordFlags(u)
			if w := execDC.dict.PeekBuilding(); w != nil && w.IsColon() {
				w.flags |= flags
				break
			}
			words := execDC.dict.byOrdinal
			flagged := false
			for i := len(words) - 1; i >= 0; i-- {
				w := words[i]
				if !w.IsColon() {
					continue
				}
				w.flags |= flags
				flagged = true
				break
			}
			if !flagged {
				panic(fmt.Errorf("immediate: no compiled word to flag"))
			}

		case OpTransition:
			id := RuntimeDispatchID(runSubCode)
			runRuntimeDispatch(id, &runFrame)

		case OpFetch:
			assertCS()
			runFetch(&runFrame, cs)
			resetCS()
		case OpStore:
			assertCS()
			runStore(&runFrame, cs)
			resetCS()

		case OpDoVar:
			assertCS()
			slotIdx := int(ds.Pop().(Int))
			ref := NewDTSlotRef(execDC.dt, slotIdx, false)
			cs.Push(ref)
			resetCS()

		case OpBootstrapDoVar:
			assertCS()
			slotIdx := int(ds.Pop().(Int))
			coreDC := ModuleTable.DCForModule(t, CoreModuleSlot)
			ref := NewDTSlotRef(coreDC.dt, slotIdx, false)
			cs.Push(ref)
			resetCS()

		case OpCtImmDs:
			slotIdx := readCTOperand()
			ds.Push(Int(slotIdx))

		case OpS:
			runS(ds, stdout)
		case OpSS:
			runS(ds, stderr)
		case OpRtoS:
			runRtoS(ds)
		case OpDsS:
			runDsS(ds, stdout)
		case OpStarS:
			runStarS(rp, stdout, execDC, ip, op)

		case OpHandleCompound:
			handleCompound(&runFrame)

		case OpTypeMemberDispatch:
			assertCS()
			runTypeMemberDispatch(&runFrame, runSubCode, cs)
			resetCS()

		case OpNativeInvoke:
			assertCS()
			kind := RuntimeDispatchKind(runSubCode)
			runNativeInvokeOpcode(&runFrame, cs, kind)
			resetCS()

		case OpAtIndex:
			assertCS()
			runAtIndex(&runFrame, cs)
			resetCS()
		case OpBangIndex:
			assertCS()
			runBangIndex(&runFrame, cs)
			resetCS()
		case OpAtBrace:
			assertCS()
			runAtBrace(&runFrame, cs)
			resetCS()
		case OpBangBrace:
			assertCS()
			runBangBrace(&runFrame, cs)
			resetCS()
		case OpQueryBrace:
			assertCS()
			runQueryBrace(&runFrame, cs)
			resetCS()
		case OpAtMulti:
			assertCS()
			runAtMulti(&runFrame, cs)
			resetCS()
		case OpBangMulti:
			assertCS()
			runBangMulti(&runFrame, cs)
			resetCS()
		case OpMethodInvoke:
			assertCS()
			runMethodInvoke(&runFrame, cs)
			resetCS()

		case OpGetDictionaryContext:
			// Session/compile-target DC (task.dc), not transient execDC (Nest may switch to core).
			if t.dc == nil {
				panic("dc: nil Task.dc")
			}
			ds.Push(t.dc)
		case OpGetSession:
			ds.Push(t.session)
		case OpGetLogger:
			ds.Push(t.session.logger)
		case OpLogAt:
			level := Level(ds.Pop().(Int))
			data := ds.Pop()
			msg := string(ds.Pop().(String))
			t.session.logger.Log(level, msg, data)
		case OpGetSys:
			ds.Push(sys)

		case OpCompileCSSelect:
			if compileEmitMode(t) {
				idx := Int(ds.Pop().(Int))
				_ = t.compileAppend(t.compilerForTarget(), int(OpDoCSSelect))
				_ = t.compileAppend(t.compilerForTarget(), idx)
			}
		case OpDoCSSelect:
			// inlined for hot path, and use Go runtime for bounds check
			/*idx := ContextStackIndex(readCTOperand())
			if idx < 0 || idx >= csCount {
				panic(fmt.Errorf("(cs): selector %d out of range [0,%d)", idx, csCount))
			}*/
			idx := execDC.ct.Read(int(ip))
			ip++
			cs = stacks[ContextStackIndex(idx)]
		case OpCSSelectFromDS:
			idx := ContextStackIndex(ds.Pop().(Int))
			if idx < 0 || idx >= csCount {
				panic(fmt.Errorf("armCs: selector %d out of range [0,%d)", idx, csCount))
			}
			cs = stacks[idx]
		case OpCompileSubCode:
			if compileEmitMode(t) {
				code := Int(ds.Pop().(Int))
				_ = t.compileAppend(t.compilerForTarget(), int(OpDoSubCode))
				_ = t.compileAppend(t.compilerForTarget(), code)
			}
		case OpDoSubCode:
			runSubCode = readCTOperand()
		case OpSubCodeFromDS:
			runSubCode = int(ds.Pop().(Int))
		case OpPushType:
			typeDC, localIdx := demuxDC(runSubCode) // runSubCode is a VirtTypeIdx
			if typeDC == nil || typeDC.tr == nil {
				panic("pushType: nil DictionaryContext or TypeRegistrar")
			}
			typ := typeDC.tr.GetByTypeIdx(localIdx)
			if typ == nil {
				panic(fmt.Errorf("pushType: unknown VirtTypeIdx %d (module idx %d)", runSubCode, localIdx))
			}
			ds.Push(typ)

		case OpError:
			panic(ds.Pop().(String))

		case OpPushCatchMark:
			j := readCTOperand()
			catchIP := int(ip) + j
			pushCatchMark(rp, catchIP)

		case OpDropCatchMark:
			dropCatchMark(rs)

		case OpThrow:
			runOpThrow(t, rp, &execDC, &ip)

		case OpAllotDtSlotToWord:
			runAllotDtSlotToWord(t, rp)

		case OpDTRef:
			runDTRef(execDC, rp)

		case OpDTConst:
			runDTConst(execDC, rp)

		case OpPushBlock:
			kind := BlockKind(int(ds.Pop().(Int)))
			if kind == BlockKindTry {
				if w := t.compilerForTarget().DC.dict.PeekBuilding(); w != nil && w.openSubStart >= 0 {
					panic(fmt.Errorf("try is not allowed inside a subword"))
				}
			}
			ts.Push(NewBlockToken(kind))
		case OpAcceptDropBlock:
			kind := int(ds.Pop().(Int))
			// note: a follow-on required check will panic with msg
			if ts.Len() == 0 {
				break
			}
			// intended to panic if not a BlockToken
			block := ts.Peek().(*BlockToken)
			if block.Kind() == kind {
				ts.Pop()
			}
		case OpPanicDropBlock:
			kind := BlockKind(int(ds.Pop().(Int)))
			msg := string(ds.Pop().(String))
			panicDropBlock(ts, msg, kind)
		case OpPanicRequireBlock:
			kind := int(ds.Pop().(Int))
			msg := String(ds.Pop().(String))
			if ts.Len() == 0 {
				panic(fmt.Errorf("%s", msg))
			}
			// intended to panic if not a BlockToken
			block := ts.Peek().(*BlockToken)
			if block.Kind() != kind {
				panic(fmt.Errorf("%s", msg))
			}
		case OpPanicRequireNKind:
			kind := BlockKind(int(ds.Pop().(Int)))
			msg := string(ds.Pop().(String))
			n := int(ds.Pop().(Int))
			panicRequireNKind(ts, n, msg, kind)
		case OpPanicRequireOneOfKind:
			msg := string(ds.Pop().(String))
			kindsVal := ds.Pop()
			kinds, ok := kindsVal.(hasItemFromAny)
			if !ok {
				panic(fmt.Errorf("panicRequireOneOfKind: expected List, got %T", kindsVal))
			}
			panicRequireOneOfKind(ts, msg, kinds)

		case OpSetBlockMark:
			mark := int(ds.Pop().(Int))
			block := ts.Peek().(*BlockToken)
			block.SetMark(mark)

		case OpContinueMark:
			// Push CT continue dest from innermost foreach or forever block.
			found := false
			for i := 0; i < ts.Len(); i++ {
				block := ts.At(i).(*BlockToken)
				if block.Kind() == BlockKindColon {
					break
				}
				if block.Kind() == BlockKindForeach || block.Kind() == BlockKindForever {
					if block.Mark() < 0 {
						panic(fmt.Errorf("continueMark: loop block has no continue mark"))
					}
					ds.Push(Int(block.Mark()))
					found = true
					break
				}
			}
			if !found {
				panic(fmt.Errorf("'continue' requires 'foreach' or 'forever'"))
			}

		case OpRange:
			r := &Range{
				end: int(ds.Pop().(Int)),
				beg: int(ds.Pop().(Int)),
			}
			ds.Push(r)
		case OpRangeBy:
			ds.Push(NewRangeBy(rp))

		case OpGetCompileStateFlags:
			ds.Push(Int(t.compilerState.compileState))
		case OpSetCompileStateFlags:
			flags := int(ds.Pop().(Int))
			if (flags & CompileStateFlagCompiling) != 0 {
				t.compilerForTarget().DC.panicIfLockedOrSealed("set compile state")
			}
			t.compilerState.compileState = flags

		case OpRecurse:
			if err := bootstrapEmitRecurse(t.compilerForTarget()); err != nil {
				panic(err)
			}

		case OpDoEmit:
			xtVirt := VirtualAddr(readCTOperand())
			operandDC, baseAddr := demuxDC(xtVirt)
			w := operandDC.dict.GetByOrdinal(baseAddr)
			if w == nil {
				panic(fmt.Errorf("doEmit: invalid virtual word xt %x", xtVirt))
			}
			if err := w.CompileWord(&runFrame); err != nil {
				panic(err)
			}
		case OpMarkContext:
			// Push ContextMark on Slot stack (nested paren ops). Parks on Slot, not RS.
			owner := ensureInnermostColonFrameMark(rs)
			owner.PushContextMark(&ContextMark{
				stacks: [csCount]int{
					ds.Len(),
					ts.Len(),
					xs.Len(),
					ys.Len(),
					zs.Len(),
					os.Len(),
				},
				dict: execDC.dict.Len(),
				ct:   execDC.ct.Len(),
				dt:   execDC.dt.Len(),
			})
		case OpParen, OpBrace, OpBracket:
			op = OpMarkContext
			goto reenter
		case OpGetContextMark:
			// Walk ColonFrameMarks (closers nest above the owner). Soft GetClear; panic if none.
			assertCS()
			ctx, ok := takeContextMarkFromRS(rs)
			if !ok {
				panic(fmt.Errorf("getContextMark: no ContextMark found"))
			}
			cs.Push(ctx)
			resetCS()
		case OpMathMonadic:
			var p Float
			switch v := ds.Pop().(type) {
			case Int:
				p = float64(v)
			case Uint:
				p = float64(v)
			case Float:
				p = v
			default:
				panic(fmt.Errorf("OpMathMonadic: invalid type: %T", v))
			}
			ds.Push(Float(mathMonadicFns[runSubCode](p)))

		case OpDtDyadic:
			if runSubCode < 0 || runSubCode >= int(SCDtDyadicCount) {
				panic(fmt.Errorf("OpDtDyadic: invalid subcode %d", runSubCode))
			}
			dtDyadicFns[runSubCode](ds)

		case OpStrOp:
			if runSubCode < 0 || runSubCode >= int(SCStrOpCount) {
				panic(fmt.Errorf("OpStrOp: invalid subcode %d", runSubCode))
			}
			strOpFns[runSubCode](ds)

		case OpPipeOp:
			if runSubCode < 0 || runSubCode >= int(SCPipeOpCount) {
				panic(fmt.Errorf("OpPipeOp: invalid subcode %d", runSubCode))
			}
			pipeOpFns[runSubCode](ds)

		case OpSetOp:
			if runSubCode < 0 || runSubCode >= int(SCSetOpCount) {
				panic(fmt.Errorf("OpSetOp: invalid subcode %d", runSubCode))
			}
			setOpFns[runSubCode](ds)

		case OpCast:
			if runSubCode < 0 || runSubCode >= int(SCCastOpCount) {
				panic(fmt.Errorf("OpCast: invalid subcode %d", runSubCode))
			}
			castOpFns[runSubCode](ds)

		case OpVecUnary:
			if runSubCode < 0 || runSubCode >= int(SCVecUnaryCount) {
				panic(fmt.Errorf("OpVecUnary: invalid subcode %d", runSubCode))
			}
			vecUnaryFns[runSubCode](ds)
		case OpVecVec:
			if runSubCode < 0 || runSubCode >= int(SCVecVecCount) {
				panic(fmt.Errorf("OpVecVec: invalid subcode %d", runSubCode))
			}
			vecVecFns[runSubCode](ds)
		case OpVecScalar:
			if runSubCode < 0 || runSubCode >= int(SCVecScalarCount) {
				panic(fmt.Errorf("OpVecScalar: invalid subcode %d", runSubCode))
			}
			vecScalarFns[runSubCode](ds)
		case OpMatUnary:
			if runSubCode < 0 || runSubCode >= int(SCMatUnaryCount) {
				panic(fmt.Errorf("OpMatUnary: invalid subcode %d", runSubCode))
			}
			matUnaryFns[runSubCode](ds)
		case OpMatMat:
			if runSubCode < 0 || runSubCode >= int(SCMatMatCount) {
				panic(fmt.Errorf("OpMatMat: invalid subcode %d", runSubCode))
			}
			matMatFns[runSubCode](ds)
		case OpMatVec:
			if runSubCode < 0 || runSubCode >= int(SCMatVecCount) {
				panic(fmt.Errorf("OpMatVec: invalid subcode %d", runSubCode))
			}
			matVecFns[runSubCode](ds)
		case OpMatScalar:
			if runSubCode < 0 || runSubCode >= int(SCMatScalarCount) {
				panic(fmt.Errorf("OpMatScalar: invalid subcode %d", runSubCode))
			}
			matScalarFns[runSubCode](ds)
		case OpMatIndex:
			if runSubCode < 0 || runSubCode >= int(SCMatIndexCount) {
				panic(fmt.Errorf("OpMatIndex: invalid subcode %d", runSubCode))
			}
			matIndexFns[runSubCode](ds)
		case OpMatFactory:
			if runSubCode < 0 || runSubCode >= int(SCMatFactoryCount) {
				panic(fmt.Errorf("OpMatFactory: invalid subcode %d", runSubCode))
			}
			matFactoryFns[runSubCode](ds)

		case OpCompileForget:
			runCompileForget(t)

		case OpDoForget:
			// Truncate may wipe this mark's CT body; unnest before the next CT fetch.
			runDoForget(t, rp)
			unnestFromColon(&execDC, rs, &ip)

		case OpDelayMs:
			delayMs := int(ds.Pop().(Int))
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		case OpDelayUs:
			delayUs := int(ds.Pop().(Int))
			time.Sleep(time.Duration(delayUs) * time.Microsecond)

		case OpSpawnPip: // ( pp:AtomicDynamicObject xt:XT -- )
			xt := XT(ds.Pop().(Int))
			pp := ds.Pop().(*AtomicDynamicObject)
			_ = spawnPip(t, xt, pp)

		case OpPls:
			ds.Push(t.GetPls())

		case OpUsing:
			runUsing(t, rp)

		case OpUsingDefault:
			runUsingDefault(t)

		case OpInstall:
			runInstall(t, rp)

		case OpImport:
			runImport(t, rp)

		case OpLoadProject:
			runLoadProject(t, rp)

		case OpNop:
		default:
			panic(fmt.Errorf("Invalid opcode: %s", OpCodeName(op)))
		}
	}
runDone:
	t.forceOp = nil
	t.ip = EncodeExecIP(execDC, ip)
}
