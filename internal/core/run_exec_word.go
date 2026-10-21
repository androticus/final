package core

import "fmt"

// armExecWord sets runSubCode and dt operands from w, then dispatches axiom reenter or colon Nest.
func armExecWord(t *Task, w *Word, execDC **DictionaryContext, ds *Stack[any], runSubCode *int, rs *ReturnStackType, ip *IP) (op OpCode, nested bool) {
	if w == nil {
		panic("armExecWord: nil word")
	}
	assertWordAllowed(t, w)
	if w.subCode >= 0 {
		*runSubCode = w.subCode
	}
	if w.dtIdx >= 0 {
		pushExecDtOperands(ds, w.dtIdx)
	}
	if w.IsAxiom() {
		return w.opCode, false
	}
	if w.colonCode < 0 {
		panic(fmt.Errorf("exec: colon word %q has no entry", w.name))
	}
	nestIntoColonWord(execDC, rs, ip, w)
	return OpINVALID, true
}

// queueExecWord leaves ( w true -- ) for the in-VM repl loop to run OpExecWord after dispatch.
func queueExecWord(ds *Stack[any], w *Word) {
	ds.Push(w)
	ds.Push(true)
}

// runExecWord runs w through the Run loop (pendingExecWord; entry ip at ReservedCTBootStop).
// RequiresStackContext words default to csDataStack.
func runExecWord(t *Task, rp *RunProvider, w *Word) error {
	if w != nil && w.RequiresStackContext() {
		return runExecWordWithCS(t, rp, w, csDataStack)
	}
	return runPendingExec(t, rp, w, 0)
}

// runExecWordWithCS runs a RequiresStackContext word with an explicit CS selector (no CT Append).
func runExecWordWithCS(t *Task, rp *RunProvider, w *Word, sel ContextStackIndex) error {
	if t == nil || rp == nil || w == nil {
		return fmt.Errorf("runExecWordWithCS: nil argument")
	}
	if !w.RequiresStackContext() {
		return fmt.Errorf("runExecWordWithCS: word %q does not require stack context", w.name)
	}
	if err := checkOStackExecAllowed(sel, w); err != nil {
		return err
	}
	return runPendingExec(t, rp, w, sel)
}

func runPendingExec(t *Task, rp *RunProvider, w *Word, sel ContextStackIndex) error {
	if t == nil || rp == nil || w == nil {
		return fmt.Errorf("runPendingExec: nil argument")
	}
	t.pendingExecWord = w
	t.pendingExecCS = sel
	savedIP := t.ip
	rsDepth := rp.ReturnStack.Len()
	savedFloor := t.rsNestFloor
	t.rsNestFloor = rsDepth
	t.ip = ReservedCTBootStop
	done := make(chan error, 1)
	prevDone := t.doneChan
	t.doneChan = done
	// Nested Run installs its own forceOp and clears t.forceOp on exit. Restore the
	// caller's pointer so Session.Break still arms an outer runLoop (e.g. forever
	// after OpHandleCompound materializes a const via runExecWord).
	savedForceOp := t.forceOp
	Run(t, rp)
	t.forceOp = savedForceOp
	t.doneChan = prevDone
	t.ip = savedIP
	t.rsNestFloor = savedFloor
	unnestReturnStackTo(rp.ReturnStack, rsDepth)
	if err := <-done; err != nil {
		return err
	}
	return nil
}

// runInterruptNest runs OpInterrupt to completion outside an active ForceOp arming
// (interpret-time compound walk / bare RunFrame). Operands: moduleID, entry on DS.
func runInterruptNest(t *Task, rp *RunProvider, moduleID ModuleID, entry int) {
	if t == nil || rp == nil {
		panic("runInterruptNest: nil Task or RunProvider")
	}
	rp.DataStack.Push(Int(moduleID))
	rp.DataStack.Push(Int(entry))
	savedIP := t.ip
	rsDepth := rp.ReturnStack.Len()
	savedFloor := t.rsNestFloor
	t.rsNestFloor = rsDepth
	t.ip = ReservedCTBootStop
	t.pendingForceOp = OpInterrupt
	done := make(chan error, 1)
	prevDone := t.doneChan
	t.doneChan = done
	savedForceOp := t.forceOp
	Run(t, rp)
	t.forceOp = savedForceOp
	t.doneChan = prevDone
	t.ip = savedIP
	t.rsNestFloor = savedFloor
	unnestReturnStackTo(rp.ReturnStack, rsDepth)
	if err := <-done; err != nil {
		panic(err)
	}
}

// unnestReturnStackTo unnests ColonFrameMarks above depth (restoring lastDC into a
// scratch execDC). Panics if the stack is already below the nest floor.
func unnestReturnStackTo(rs *ReturnStackType, depth int) {
	if rs == nil {
		return
	}
	if rs.Len() < depth {
		panic(fmt.Errorf("return stack underflow: len %d < nest floor %d", rs.Len(), depth))
	}
	var scratch *DictionaryContext
	var ip IP
	for rs.Len() > depth {
		unnestFromColon(&scratch, rs, &ip)
	}
}

// dispatchExecWord queues or runs execWord per inRunLoop (in-VM repl vs Go-hosted session).
func dispatchExecWord(t *Task, rp *RunProvider, w *Word, inRunLoop bool) error {
	if err := checkInterpretWordAllowed(t, w); err != nil {
		return err
	}
	if inRunLoop {
		if w.RequiresStackContext() {
			return runExecWordWithCS(t, rp, w, csDataStack)
		}
		queueExecWord(rp.DataStack, w)
		return nil
	}
	return runExecWord(t, rp, w)
}
