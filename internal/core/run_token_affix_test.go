package core

import "testing"

func TestAffixAxiomsRegistered(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"@()", "!()", "@{}", "!{}", "@[]", "![]"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing dictionary word %q", name)
		}
	}
}

func TestOpAtIndexOnListViaCompoundWalk(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	l := NewAnyList()
	l.Append(String("a"))
	l.Append(String("b"))
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(l)
	segs := CompoundPathForTest([]string{segAtIndex})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "b" {
		t.Fatalf("At(1) = %v, want b", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 (receiver elided at collection bind)", rp.DataStack.Len())
	}
}

func TestOpAtIndexAxiomDirect(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	l := NewAnyList()
	l.Append(Int(10))
	l.Append(Int(20))
	rp.DataStack.Push(l)
	rp.DataStack.Push(Int(0))
	RunOps(task, rp, OpAtIndex, OpStop)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 10 {
		t.Fatalf("got %v, want 10", got)
	}
}
