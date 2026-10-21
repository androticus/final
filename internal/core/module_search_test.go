package core

import (
	"testing"
)

func TestModuleSearchOrderUsingNopAlias(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	task := s.task
	rp := s.rp
	if task.session != s {
		t.Fatal("task.session not wired")
	}
	if task.currentSearchOrder == nil || s.modules == nil || task.currentSearchOrder != s.modules {
		t.Fatal("expected shared Session.modules / currentSearchOrder")
	}
	if task.currentSearchOrder.Len() != 2 {
		t.Fatalf("search order len = %d, want 2", task.currentSearchOrder.Len())
	}

	// Local nop that pushes 1; core nop is a no-op.
	runInterpretScript(t, task, rp, `
: nop  1 ;
`)

	runInterpretScript(t, task, rp, `nop`)
	if got := rp.DataStack.Pop(); got != Int(1) {
		t.Fatalf("default order nop = %v, want 1 (local)", got)
	}

	runInterpretScript(t, task, rp, `( "core" "local" s) using`)
	if task.currentSearchOrder == s.modules {
		t.Fatal("using should fork currentSearchOrder")
	}
	runInterpretScript(t, task, rp, `nop`)
	if rp.DataStack.Len() != 0 {
		t.Fatalf("after core-first nop DS len = %d, want 0 (core nop)", rp.DataStack.Len())
	}

	runInterpretScript(t, task, rp, `using-default`)
	if task.currentSearchOrder != s.modules {
		t.Fatal("using-default should restore Session.modules")
	}
	runInterpretScript(t, task, rp, `nop`)
	if got := rp.DataStack.Pop(); got != Int(1) {
		t.Fatalf("after using-default nop = %v, want 1 (local)", got)
	}
}

func TestModuleSearchQualifiedLocalCore(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
: markLocal  42 ;
`)

	runInterpretScript(t, task, rp, `local\markLocal`)
	if got := rp.DataStack.Pop(); got != Int(42) {
		t.Fatalf("local\\markLocal = %v, want 42", got)
	}

	// core\dup exists; exercise qualified find into core.
	rp.DataStack.Push(Int(7))
	runInterpretScript(t, task, rp, `core\dup`)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("after core\\dup DS len = %d, want 2", rp.DataStack.Len())
	}
	a := rp.DataStack.Pop()
	b := rp.DataStack.Pop()
	if a != Int(7) || b != Int(7) {
		t.Fatalf("core\\dup stacked %v %v, want 7 7", b, a)
	}
}

func TestDefiningWordNameRejectsBackslash(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for name with backslash")
		}
	}()
	_ = NewWord(`foo\bar`, WordFlagColon)
}

func TestPipInheritsSearchOrderAndSession(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	parent := s.task
	runInterpretScript(t, parent, NewTestRunProvider(), `( "core" "local" s) using`)
	child, _ := CreateSubTask(parent)
	if child.session != s {
		t.Fatal("pip session not copied")
	}
	if child.currentSearchOrder != parent.currentSearchOrder {
		t.Fatal("pip currentSearchOrder should share parent ref")
	}
}
