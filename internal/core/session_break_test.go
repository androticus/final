package core

import "testing"

func TestTaskBreakNoOpWhenIdle(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	tk.Break() // forceOp nil outside runLoop
}

func TestSessionBreakNoOpWhenIdle(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.Break()
}

// Nested runExecWord (compound root materialize) must not leave t.forceOp nil while an
// outer runLoop is still live — Break publishes through that pointer.
func TestNestedRunRestoresForceOpPointer(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	outer := OpCode(0)
	task.forceOp = &outer

	drop := globalDictionaryContext.dict.GetByName("drop")
	if drop == nil {
		t.Fatal("missing drop")
	}
	rp.DataStack.Push(Int(1))
	if err := runExecWord(task, rp, drop); err != nil {
		t.Fatal(err)
	}
	if task.forceOp != &outer {
		t.Fatalf("forceOp = %p, want outer %p", task.forceOp, &outer)
	}
	if outer != 0 {
		t.Fatalf("outer forceOp value = %v, want 0", outer)
	}
	task.Break()
	if outer != OpStop {
		t.Fatalf("Break after nest: forceOp = %v, want OpStop", outer)
	}
}
