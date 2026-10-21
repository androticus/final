package core

import (
	"math"
	"testing"
)

func TestOpDoLit(t *testing.T) {
	t.Run("loads_dt_literal", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		dt := &VirtDTResolver{}
		dtIdx := dt.Append(ModuleIDThis, Int(42))
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpDoLit))
		ct.Append(ModuleIDThis, decodeCTOffset(dtIdx))
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 42 {
			t.Fatalf("got %v, want Int(42)", got)
		}
	})
}

func TestOpDoLit0(t *testing.T) {
	t.Run("loads_zero", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		RunOps(NewTestTask(), rp, OpDoLit0, OpStop)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 0 {
			t.Fatalf("got %v, want Int(0)", got)
		}
	})
}

func TestOpDoLitNull(t *testing.T) {
	t.Run("loads_untyped_nil", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		RunOps(NewTestTask(), rp, OpDoLitNull, OpStop)
		got := rp.DataStack.Pop()
		if got != nil {
			t.Fatalf("got %v (%T), want untyped nil", got, got)
		}
	})
}

func TestOpDoLitI(t *testing.T) {
	t.Run("loads_int_operand", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 99)
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 99 {
			t.Fatalf("got %v, want Int(99)", got)
		}
	})
}

func TestOpDoLitB(t *testing.T) {
	t.Run("loads_bool_operand", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpDoLitB))
		ct.Append(ModuleIDThis, 1)
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Bool)
		if !ok || !got {
			t.Fatalf("got %v, want true", got)
		}
	})
}

func TestOpDoLitU(t *testing.T) {
	t.Run("loads_uint_operand", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpDoLitU))
		ct.Append(ModuleIDThis, 7)
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Uint)
		if !ok || got != 7 {
			t.Fatalf("got %v, want Uint(7)", got)
		}
	})
}

func TestOpDoLitF(t *testing.T) {
	t.Run("loads_float_operand", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpDoLitF))
		ct.Append(ModuleIDThis, int(math.Float64bits(3.5)))
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Float)
		if !ok || got != 3.5 {
			t.Fatalf("got %v, want Float(3.5)", got)
		}
	})
}

func TestOpNestUnnest(t *testing.T) {
	t.Run("subroutine_call_returns", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Int(3))
		rp.DataStack.Push(Int(1))
		ct := &VirtCTResolver{}
		body := ct.Append(ModuleIDThis, int(OpAdd))
		ct.Append(ModuleIDThis, int(OpUnnest))
		ip := ct.Append(ModuleIDThis, int(OpNest))
		ct.Append(ModuleIDThis, 0)
		ct.Append(ModuleIDThis, decodeCTOffset(body))
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 4 {
			t.Fatalf("got %v, want Int(4)", got)
		}
	})
}

func TestOpExec(t *testing.T) {
	t.Run("axiom_via_xt", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		_, xt, ok := FindName("+")
		if !ok {
			t.Fatal("FindName +")
		}
		rp.DataStack.Push(Int(1))
		rp.DataStack.Push(Int(2))
		rp.DataStack.Push(xt)
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpExec))
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 3 {
			t.Fatalf("got %v, want Int(3)", got)
		}
	})
}

func TestOpBranch(t *testing.T) {
	t.Run("j_zero_no_op", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		ip := ct.Append(ModuleIDThis, int(OpBranch))
		ct.Append(ModuleIDThis, 0)
		ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 42)
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 42 {
			t.Fatalf("got %v, want Int(42)", got)
		}
	})

	t.Run("forward_skip", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		branchOp := ct.Append(ModuleIDThis, int(OpBranch))
		jSlot := ct.Append(ModuleIDThis, 0)
		ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 77)
		merge := ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 99)
		ct.Append(ModuleIDThis, int(OpStop))
		branchBase := jSlot + 1
		j := int(merge - branchBase)
		ct.Write(jSlot, j)
		task := NewTestTask()
		task.ip = branchOp
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 99 {
			t.Fatalf("got %v, want Int(99)", got)
		}
	})

}

func TestOpBranchQ(t *testing.T) {
	t.Run("truthy_fall_through", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Bool(true))
		ct := &VirtCTResolver{}
		branchOp := ct.Append(ModuleIDThis, int(OpBranchQ))
		jSlot := ct.Append(ModuleIDThis, 0)
		ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 77)
		stop := ct.Append(ModuleIDThis, int(OpStop))
		branchBase := jSlot + 1
		j := int(stop - branchBase)
		ct.Write(jSlot, j)
		task := NewTestTask()
		task.ip = branchOp
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 77 {
			t.Fatalf("got %v, want Int(77)", got)
		}
	})

	t.Run("falsy_forward_branch", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		rp.DataStack.Push(Bool(false))
		ct := &VirtCTResolver{}
		branchOp := ct.Append(ModuleIDThis, int(OpBranchQ))
		jSlot := ct.Append(ModuleIDThis, 0)
		ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 77)
		merge := ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 88)
		ct.Append(ModuleIDThis, int(OpStop))
		branchBase := jSlot + 1
		j := int(merge - branchBase)
		ct.Write(jSlot, j)
		task := NewTestTask()
		task.ip = branchOp
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 88 {
			t.Fatalf("got %v, want Int(88)", got)
		}
	})
}

func TestOpExec_colonWord(t *testing.T) {
	t.Run("colon_word_via_xt", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		body := ct.Append(ModuleIDThis, int(OpAdd))
		ct.Append(ModuleIDThis, int(OpUnnest))
		w := NewWord("inc2", WordFlagColon)
		w.colonCode = CTP(decodeCTOffset(body))
		globalDictionaryContext.dict.Add(w)
		xt, ok := globalDictionaryContext.dict.XTOf(w)
		if !ok {
			t.Fatal("XTOf inc2")
		}
		rp.DataStack.Push(Int(1))
		rp.DataStack.Push(Int(2))
		rp.DataStack.Push(xt)
		ip := ct.Append(ModuleIDThis, int(OpExec))
		ct.Append(ModuleIDThis, int(OpStop))
		task := NewTestTask()
		task.ip = ip
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 3 {
			t.Fatalf("got %v, want Int(3)", got)
		}
	})
}
