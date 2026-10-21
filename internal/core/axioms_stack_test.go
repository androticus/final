package core

import (
	"strings"
	"testing"
)

func wantDS(t *testing.T, rp *RunProvider, topFirst ...any) {
	t.Helper()
	if got := rp.DataStack.Len(); got != len(topFirst) {
		t.Fatalf("DS depth=%d, want %d", got, len(topFirst))
	}
	for i, want := range topFirst {
		got := *rp.DataStack.AddrAt(i)
		if !stackValsEqual(got, want) {
			t.Fatalf("AddrAt(%d)=%#v, want %#v", i, got, want)
		}
	}
}

func stackValsEqual(got, want any) bool {
	switch w := want.(type) {
	case Int:
		g, ok := got.(Int)
		return ok && g == w
	case Uint:
		g, ok := got.(Uint)
		return ok && g == w
	case String:
		g, ok := got.(String)
		return ok && g == w
	default:
		return got == want
	}
}

func runStackOp(t *testing.T, rp *RunProvider, op OpCode, setup ...any) {
	t.Helper()
	ResetTestRuntime()
	for _, v := range setup {
		rp.DataStack.Push(v)
	}
	RunOps(NewTestTask(), rp, op, OpStop)
}

// --- OpDup ---

func TestOpDup(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDup, Int(7))
		wantDS(t, rp, Int(7), Int(7))
	})

	t.Run("underflow_panics", func(t *testing.T) {
		ResetTestRuntime()
		runOpsExpectPanic(t, NewTestRunProvider(), "underflow", OpDup, OpStop)
	})
}

// --- OpDrop ---

func TestOpDrop(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDrop, Int(9), Int(1))
		wantDS(t, rp, Int(9))
	})

	t.Run("underflow_panics", func(t *testing.T) {
		ResetTestRuntime()
		runOpsExpectPanic(t, NewTestRunProvider(), "underflow", OpDrop, OpStop)
	})
}

func TestOpDropWithTStackSelector(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.TempStack.Push(Int(77))
	rp.DataStack.Push(Int(1))
	RunOpsArmedCS(NewTestTask(), rp, csTStack, OpDrop, OpStop)
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0", rp.TempStack.Len())
	}
	if rp.DataStack.Len() != 1 || rp.DataStack.Pop().(Int) != 1 {
		t.Fatal("DS should be unchanged when selector=TS")
	}
}

// --- OpSwap ---

func TestOpSwap(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpSwap, Int(1), Int(2))
		wantDS(t, rp, Int(1), Int(2))
	})
}

// --- OpOver ---

func TestOpOver(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpOver, Int(10), Int(20))
		wantDS(t, rp, Int(10), Int(20), Int(10))
	})
}

// --- OpNip ---

func TestOpNip(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpNip, Int(10), Int(20))
		wantDS(t, rp, Int(20))
	})
}

// --- OpTuck ---

func TestOpTuck(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpTuck, Int(1), Int(2))
		wantDS(t, rp, Int(2), Int(1), Int(2))
	})

	t.Run("differs_from_over", func(t *testing.T) {
		rpOver := NewTestRunProvider()
		runStackOp(t, rpOver, OpOver, Int(1), Int(2))
		rpTuck := NewTestRunProvider()
		runStackOp(t, rpTuck, OpTuck, Int(1), Int(2))
		if rpOver.DataStack.Len() != rpTuck.DataStack.Len() {
			t.Fatal("depth mismatch")
		}
		for i := 0; i < rpOver.DataStack.Len(); i++ {
			a := *rpOver.DataStack.AddrAt(i)
			b := *rpTuck.DataStack.AddrAt(i)
			if stackValsEqual(a, b) {
				continue
			}
			return
		}
		t.Fatal("over and tuck produced identical stacks")
	})
}

// --- OpPick ---

func TestOpPick(t *testing.T) {
	t.Run("pick_tos", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpPick, Int(1), Int(2), Int(3), Uint(0))
		wantDS(t, rp, Int(3), Int(3), Int(2), Int(1))
	})

	t.Run("pick_second", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpPick, Int(1), Int(2), Int(3), Int(1))
		wantDS(t, rp, Int(2), Int(3), Int(2), Int(1))
	})

	t.Run("invalid_index_type_panics", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Int(1))
		rp.DataStack.Push(String("nope"))
		runOpsExpectPanic(t, rp, "OpPick", OpPick, OpStop)
	})
}

// --- OpRot ---

func TestOpRot(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpRot, Int(1), Int(2), Int(3))
		wantDS(t, rp, Int(1), Int(3), Int(2))
	})
}

// --- OpDStoTS ---

func TestOpDStoTS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDStoTS, Int(42))
		wantDS(t, rp)
		if got := rp.TempStack.Pop().(Int); got != 42 {
			t.Fatalf("TS got %v, want 42", got)
		}
	})
}

// --- OpTStoDS ---

func TestOpTStoDS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.TempStack.Push(Int(99))
		runStackOp(t, rp, OpTStoDS)
		wantDS(t, rp, Int(99))
	})
}

// --- OpTDrop ---

func TestOpTDrop(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.TempStack.Push(Int(42))
		runStackOp(t, rp, OpTDrop)
		if rp.TempStack.Len() != 0 {
			t.Fatalf("TS len = %d, want 0 after tdrop", rp.TempStack.Len())
		}
	})
}

// --- OpDStoXS ---

func TestOpDStoXS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDStoXS, Int(42))
		wantDS(t, rp)
		if got := rp.XStack.Pop().(Int); got != 42 {
			t.Fatalf("XS got %v, want 42", got)
		}
	})
}

// --- OpXStoDS ---

func TestOpXStoDS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.XStack.Push(Int(99))
		runStackOp(t, rp, OpXStoDS)
		wantDS(t, rp, Int(99))
	})
}

// --- OpXDrop ---

func TestOpXDrop(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.XStack.Push(Int(42))
		runStackOp(t, rp, OpXDrop)
		if rp.XStack.Len() != 0 {
			t.Fatalf("XS len = %d, want 0 after xdrop", rp.XStack.Len())
		}
	})
}

// --- OpDStoYS ---

func TestOpDStoYS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDStoYS, Int(42))
		wantDS(t, rp)
		if got := rp.YStack.Pop().(Int); got != 42 {
			t.Fatalf("YS got %v, want 42", got)
		}
	})
}

// --- OpYStoDS ---

func TestOpYStoDS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.YStack.Push(Int(99))
		runStackOp(t, rp, OpYStoDS)
		wantDS(t, rp, Int(99))
	})
}

// --- OpYDrop ---

func TestOpYDrop(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.YStack.Push(Int(42))
		runStackOp(t, rp, OpYDrop)
		if rp.YStack.Len() != 0 {
			t.Fatalf("YS len = %d, want 0 after ydrop", rp.YStack.Len())
		}
	})
}

// --- OpDStoZS ---

func TestOpDStoZS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		runStackOp(t, rp, OpDStoZS, Int(42))
		wantDS(t, rp)
		if got := rp.ZStack.Pop().(Int); got != 42 {
			t.Fatalf("ZS got %v, want 42", got)
		}
	})
}

// --- OpZStoDS ---

func TestOpZStoDS(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.ZStack.Push(Int(99))
		runStackOp(t, rp, OpZStoDS)
		wantDS(t, rp, Int(99))
	})
}

// --- OpZDrop ---

func TestOpZDrop(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		rp.ZStack.Push(Int(42))
		runStackOp(t, rp, OpZDrop)
		if rp.ZStack.Len() != 0 {
			t.Fatalf("ZS len = %d, want 0 after zdrop", rp.ZStack.Len())
		}
	})
}

// --- OpRPeek ---

func TestOpRPeek(t *testing.T) {
	t.Run("valid_inputs_work", func(t *testing.T) {
		rp := NewTestRunProvider()
		cm := rp.ReturnStack.PushNew(11)
		runStackOp(t, rp, OpRPeek)
		wantDS(t, rp, cm)
		if got := rp.ReturnStack.Peek(); got != cm {
			t.Fatalf("RS peek %v, want same *ColonFrameMark", got)
		}
	})

	t.Run("rs_underflow_panics", func(t *testing.T) {
		ResetTestRuntime()
		runOpsExpectPanic(t, NewTestRunProvider(), "underflow", OpRPeek, OpStop)
	})
}

func contextStackFor(rp *RunProvider, sel ContextStackIndex) *Stack[any] {
	switch sel {
	case csTStack:
		return rp.TempStack
	case csXStack:
		return rp.XStack
	case csYStack:
		return rp.YStack
	case csZStack:
		return rp.ZStack
	case csOStack:
		return rp.OStack
	default:
		return rp.DataStack
	}
}

func runContextStackAxiom(t *testing.T, rp *RunProvider, sel ContextStackIndex, op OpCode, csSetup []any, dsSetup []any) {
	t.Helper()
	ResetTestRuntime()
	csStack := contextStackFor(rp, sel)
	for _, v := range csSetup {
		csStack.Push(v)
	}
	for _, v := range dsSetup {
		rp.DataStack.Push(v)
	}
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoCSSelect))
	ct.Append(ModuleIDThis, int(sel))
	ct.Append(ModuleIDThis, int(op))
	ct.Append(ModuleIDThis, int(OpStop))
	task := NewTestTask()
	task.ip = ip
	Run(task, rp)
}

func runContextStackAxiomExpectPanic(t *testing.T, rp *RunProvider, sel ContextStackIndex, wantSubstr string, op OpCode, csSetup []any, dsSetup []any) {
	t.Helper()
	ResetTestRuntime()
	csStack := contextStackFor(rp, sel)
	for _, v := range csSetup {
		csStack.Push(v)
	}
	for _, v := range dsSetup {
		rp.DataStack.Push(v)
	}
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoCSSelect))
	ct.Append(ModuleIDThis, int(sel))
	ct.Append(ModuleIDThis, int(op))
	ct.Append(ModuleIDThis, int(OpStop))
	task := NewTestTask()
	task.ip = ip
	done := make(chan error, 1)
	task.doneChan = done
	Run(task, rp)
	err := <-done
	if err == nil {
		t.Fatal("expected panic")
	}
	if wantSubstr != "" && !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("panic error = %q, want substring %q", err.Error(), wantSubstr)
	}
}

// --- OpDSto (>ds) ---

func TestOpDSto(t *testing.T) {
	t.Run("moves_ds_tos_to_cs", func(t *testing.T) {
		rp := NewTestRunProvider()
		runContextStackAxiom(t, rp, csTStack, OpDSto, nil, []any{Int(99)})
		wantDS(t, rp)
		if rp.TempStack.Len() != 1 || rp.TempStack.Peek().(Int) != 99 {
			t.Fatalf("TS = %#v, want [99]", rp.TempStack)
		}
	})

	t.Run("ds_underflow_panics", func(t *testing.T) {
		runContextStackAxiomExpectPanic(t, NewTestRunProvider(), csTStack, "underflow", OpDSto, nil, nil)
	})
}

// --- OpToDS (ds>) ---

func TestOpToDS(t *testing.T) {
	t.Run("moves_cs_tos_to_ds", func(t *testing.T) {
		rp := NewTestRunProvider()
		runContextStackAxiom(t, rp, csTStack, OpToDS, []any{Int(42)}, nil)
		if rp.TempStack.Len() != 0 {
			t.Fatalf("TS len = %d, want 0", rp.TempStack.Len())
		}
		wantDS(t, rp, Int(42))
	})

	t.Run("cs_underflow_panics", func(t *testing.T) {
		runContextStackAxiomExpectPanic(t, NewTestRunProvider(), csTStack, "underflow", OpToDS, nil, nil)
	})
}

// --- OpPeek (c:peek) ---

func TestOpPeek(t *testing.T) {
	t.Run("copies_cs_tos_to_ds_without_popping_cs", func(t *testing.T) {
		rp := NewTestRunProvider()
		runContextStackAxiom(t, rp, csTStack, OpPeek, []any{String("item")}, []any{Int(1)})
		if rp.TempStack.Len() != 1 || rp.TempStack.Peek().(String) != "item" {
			t.Fatalf("TS = %#v, want [item]", rp.TempStack)
		}
		wantDS(t, rp, String("item"), Int(1))
	})

	t.Run("cs_underflow_panics", func(t *testing.T) {
		runContextStackAxiomExpectPanic(t, NewTestRunProvider(), csTStack, "underflow", OpPeek, nil, nil)
	})
}

func TestStackOps_underflow_pair(t *testing.T) {
	underflow := []OpCode{OpSwap, OpOver, OpNip, OpTuck, OpRot}
	for _, op := range underflow {
		t.Run(OpCodeName(op), func(t *testing.T) {
			ResetTestRuntime()
			runOpsExpectPanic(t, NewTestRunProvider(), "underflow", op, OpStop)
		})
	}
}

func TestOpPick_underflow(t *testing.T) {
	ResetTestRuntime()
	runOpsExpectPanic(t, NewTestRunProvider(), "underflow", OpPick, OpStop)
}

func TestOpPick_bad_index(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(9))
	runOpsExpectPanic(t, rp, "underflow", OpPick, OpStop)
}

