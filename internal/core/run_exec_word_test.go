package core

import (
	"fmt"
	"strings"
	"testing"
)

// TestRunPendingExecFaultRestoresOuterForeachFrame: nested runExecWord fault must not
// leave leftover RS cells; outer foreach-armed CFM + PeekIterator stay intact.
func TestRunPendingExecFaultRestoresOuterForeachFrame(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()

	src := `: boomNest "nested boom" error ;`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := task.dc.dict.GetByName("boomNest")
	if w == nil {
		t.Fatal("missing boomNest")
	}

	rp.DataStack.Clear()
	cf := rp.ReturnStack.PushNew(int(ReservedCTBootStop))
	cf.PushIterator(newRangeFromDS(0, 3).GetIter())
	outerDepth := rp.ReturnStack.Len()

	runErr := runExecWord(task, rp, w)
	if _, ok := runErr.(*RunFault); !ok {
		t.Fatalf("runExecWord err = %v, want *RunFault", runErr)
	}
	if rp.ReturnStack.Len() != outerDepth {
		t.Fatalf("RS len = %d, want outer depth %d", rp.ReturnStack.Len(), outerDepth)
	}
	if got := rp.ReturnStack.Peek().PeekIterator(); got == nil {
		t.Fatal("outer PeekIterator nil after nested fault")
	}
}

// TestInterpretDropDoesNotAppendCT proves RequiresStackContext interpret exec
// uses pendingExecWord + BootStop (no disposable CT Append).
func TestInterpretDropDoesNotAppendCT(t *testing.T) {
	ResetTestRuntime()
	before := globalDictionaryContext.ct.Len()
	drop := globalDictionaryContext.dict.GetByName("drop")
	if drop == nil {
		t.Fatal("missing drop")
	}
	for i := 0; i < 32; i++ {
		task := NewTestTask()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Int(i))
		if err := runExecWord(task, rp, drop); err != nil {
			t.Fatalf("iter %d: runExecWord(drop): %v", i, err)
		}
		if rp.DataStack.Len() != 0 {
			t.Fatalf("iter %d: DS len = %d, want 0", i, rp.DataStack.Len())
		}
	}
	if got := globalDictionaryContext.ct.Len(); got != before {
		t.Fatalf("ct.Len() = %d, want %d (no scratch Append)", got, before)
	}
}

func TestInterpretTDropDoesNotAppendCT(t *testing.T) {
	ResetTestRuntime()
	before := globalDictionaryContext.ct.Len()
	drop := globalDictionaryContext.dict.GetByName("drop")
	if drop == nil {
		t.Fatal("missing drop")
	}
	for i := 0; i < 32; i++ {
		task := NewTestTask()
		rp := NewTestRunProvider()
		rp.TempStack.Push(Int(i))
		if err := runExecWordWithCS(task, rp, drop, csTStack); err != nil {
			t.Fatalf("iter %d: runExecWordWithCS(t:drop): %v", i, err)
		}
		if rp.TempStack.Len() != 0 {
			t.Fatalf("iter %d: TS len = %d, want 0", i, rp.TempStack.Len())
		}
	}
	if got := globalDictionaryContext.ct.Len(); got != before {
		t.Fatalf("ct.Len() = %d, want %d (no scratch Append)", got, before)
	}
}

func TestRunPendingExecThrowDoesNotStealOuterFrames(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()

	src := `: boomThrow "nested throw" throw ;`
	tr, err := NewTextReaderFromString(src)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if err := interpretSession(task, rp); err != nil {
		t.Fatalf("compile: %v", err)
	}
	w := task.dc.dict.GetByName("boomThrow")
	if w == nil {
		t.Fatal("missing boomThrow")
	}

	rp.DataStack.Clear()
	cf := rp.ReturnStack.PushNew(int(ReservedCTBootStop))
	cf.PushIterator(newRangeFromDS(0, 3).GetIter())
	outerDepth := rp.ReturnStack.Len()

	runErr := runExecWord(task, rp, w)
	if _, ok := runErr.(*RunFault); !ok {
		t.Fatalf("runExecWord err = %v, want *RunFault", runErr)
	}
	if rp.ReturnStack.Len() != outerDepth {
		t.Fatalf("RS len = %d, want outer depth %d", rp.ReturnStack.Len(), outerDepth)
	}
	if got := rp.ReturnStack.Peek().PeekIterator(); got == nil {
		t.Fatal("outer PeekIterator nil after nested throw")
	}
}

func TestUnnestReturnStackToUnderflowPanics(t *testing.T) {
	rs := NewReturnStack("rs", 16)
	rs.PushNew(1)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("want panic on RS below nest floor")
		}
		if got := fmt.Sprint(r); !strings.Contains(got, "nest floor") {
			t.Fatalf("panic = %q, want nest floor", got)
		}
	}()
	unnestReturnStackTo(rs, 5)
}

func TestRunDonePacksExecIP(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	RunOps(task, rp, OpNop, OpStop)
	mod, packed := DecodeVirt(task.ip)
	if mod != CoreModuleSlot {
		t.Fatalf("t.ip moduleId = %d, want CoreModuleSlot %d", mod, CoreModuleSlot)
	}
	off := packed - int(mod)
	if off < 0 {
		t.Fatalf("packed offset %d - module %d is negative", packed, mod)
	}
}
