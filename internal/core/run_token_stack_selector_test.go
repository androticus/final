package core

import "testing"

func TestWalkCompoundDSelectorDepthFromDS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	segs := CompoundPathWithSelectorForTest([]string{"depth"}, csDataStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 3 {
		t.Fatalf("DS len = %d, want 3", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("d:depth = %v, want 2", got)
	}
}

func TestWalkCompoundTSelectorDropFromTS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.TempStack.Push(Int(99))
	segs := CompoundPathWithSelectorForTest([]string{"drop"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after t:drop", rp.TempStack.Len())
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want unchanged", rp.DataStack.Len())
	}
}

func TestWalkCompoundTSelectorDupOnTS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.TempStack.Push(String("x"))
	segs := CompoundPathWithSelectorForTest([]string{"dup"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if rp.TempStack.Len() != 2 {
		t.Fatalf("TS len = %d, want 2 after t:dup", rp.TempStack.Len())
	}
	top := rp.TempStack.Pop().(String)
	below := rp.TempStack.Pop().(String)
	if top != "x" || below != "x" {
		t.Fatal("t:dup should duplicate TS TOS")
	}
}

func TestWalkCompoundTSelectorDotNameFetchFromTS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("fromTS", 0)
	rp.TempStack.Push(w)
	segs := CompoundPathWithSelectorForTest([]string{".", "name", "@"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "fromTS" {
		t.Fatalf("fetch = %v, want fromTS on DS", got)
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (receiver kept on CS)", rp.TempStack.Len())
	}
}

func TestParseLineTSelectorDispatchedViaParser(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "")
	w := NewWord("magic", 0)
	rp.TempStack.Push(w)
	tr, err := NewTextReaderFromString("t:.name@\n")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "magic" {
		t.Fatalf("fetch = %v, want magic on DS", got)
	}
}
