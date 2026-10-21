package core

import (
	"strings"
	"testing"
)

func TestOStackAllowsPrefixedWord(t *testing.T) {
	ResetTestRuntime()
	peek := globalDictionaryContext.dict.GetByName("peek")
	drop := globalDictionaryContext.dict.GetByName("drop")
	depth := globalDictionaryContext.dict.GetByName("depth")
	caret := globalDictionaryContext.dict.GetByName("^")
	if !oStackAllowsPrefixedWord(peek) {
		t.Fatal("peek should be allowed on o:")
	}
	if oStackAllowsPrefixedWord(drop) {
		t.Fatal("drop should not be allowed on o:")
	}
	if !oStackAllowsPrefixedWord(depth) {
		t.Fatal("depth should be allowed on o:")
	}
	if !oStackAllowsPrefixedWord(caret) {
		t.Fatal("^ should be allowed on o:")
	}
}

func TestOStackRestrictWordsFlagged(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"drop", "dup", "swap", ">ds", "ds>", "getContextMark"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if !w.OSContextRestrict() {
			t.Fatalf("word %q missing WordFlagOSContextRestrict", name)
		}
	}
	for _, name := range []string{"peek", "depth", "^", "@", "@()", "!", "typeMemberDispatch"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if w.OSContextRestrict() {
			t.Fatalf("word %q should not have WordFlagOSContextRestrict", name)
		}
	}
}

func TestWalkCompoundODotNameFetchAllowed(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("fromO", 0)
	rp.OStack.Push(w)
	segs := CompoundPathWithSelectorForTest([]string{".", "name", "@"}, csOStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "fromO" {
		t.Fatalf("fetch = %v, want fromO on DS", got)
	}
	if rp.OStack.Len() != 1 {
		t.Fatalf("OStack len = %d, want 1 (receiver kept on o:)", rp.OStack.Len())
	}
}

func TestWalkCompoundOListIndexFetchAllowed(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	list := NewAnyList()
	list.Append(Int(42))
	rp.OStack.Push(list)
	rp.DataStack.Push(Int(0))
	segs := CompoundPathWithSelectorForTest([]string{"@()"}, csOStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 42 {
		t.Fatalf("@() = %v, want 42", got)
	}
	if rp.OStack.Len() != 1 {
		t.Fatalf("OStack len = %d, want 1 (list receiver kept on o:)", rp.OStack.Len())
	}
}

func TestCSSelectWordsAreSystem(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"armCs", "(cs)", "cs,"} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if !w.IsSystem() {
			t.Fatalf("word %q should be WordFlagSystem", name)
		}
	}
}

func TestWalkCompoundODropRejected(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.OStack.Push(Int(99))
	segs := CompoundPathWithSelectorForTest([]string{"drop"}, csOStack)
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected o:drop to be rejected")
	}
	if !strings.Contains(err.Error(), "o:") {
		t.Fatalf("error = %v, want o: restriction", err)
	}
	if rp.OStack.Len() != 1 {
		t.Fatalf("OStack len = %d, want 1 unchanged", rp.OStack.Len())
	}
}

func TestWalkCompoundOPeekAllowed(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	list := NewAnyList()
	rp.OStack.Push(list)
	segs := CompoundPathWithSelectorForTest([]string{"peek"}, csOStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(*List[any])
	if !ok || got != list {
		t.Fatalf("o:peek = %v, want list on DS", got)
	}
	if rp.OStack.Len() != 1 {
		t.Fatalf("OStack len = %d, want 1 (peek non-consuming on o:)", rp.OStack.Len())
	}
}

func TestWalkCompoundODepthAllowed(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	rp.OStack.Push(Int(1))
	rp.OStack.Push(Int(2))
	segs := CompoundPathWithSelectorForTest([]string{"depth"}, csOStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("o:depth = %v, want 2", got)
	}
}

func TestCompileColonOStackDropRejected(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad o:drop ;`)
	if err == nil {
		t.Fatal("expected compile error for o:drop")
	}
	if !strings.Contains(err.Error(), "o:") {
		t.Fatalf("error = %v, want o: restriction", err)
	}
}

func TestCompileColonOStackPeekAllowed(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: ok o:peek drop ;`); err != nil {
		t.Fatal(err)
	}
}

func TestPrefixedDollarRejected(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	segs := CompoundPathWithSelectorForTest([]string{"$"}, csTStack)
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected t:$ to be rejected")
	}
	if !strings.Contains(err.Error(), "stack selector prefix") {
		t.Fatalf("error = %v", err)
	}
}
