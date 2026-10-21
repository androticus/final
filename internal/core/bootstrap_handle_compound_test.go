package core

import (
	"strings"
	"testing"
)

func TestHandleCompoundAxiomRegistered(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("handleCompound")
	if w == nil {
		t.Fatal("handleCompound missing from dictionary")
	}
	if w.opCode != OpHandleCompound {
		t.Fatalf("handleCompound opCode = %v, want OpHandleCompound", OpCodeName(w.opCode))
	}
}

func TestHandleCompoundTransitionDoesNothing(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	dsLen := rp.DataStack.Len()
	runTransition(t, task, rp, FinalHandleCompound)
	if rp.DataStack.Len() != dsLen {
		t.Fatalf("DS len = %d, want %d (no-op)", rp.DataStack.Len(), dsLen)
	}
	if rp.DataStack.Pop().(Int) != 2 || rp.DataStack.Pop().(Int) != 1 {
		t.Fatal("DS values changed")
	}
}

func pushCompoundPath(ds *Stack[any], segs []string) {
	ds.Push(String(strings.Join(segs, " ")))
}

func pushWordOnTS(t *testing.T, rp *RunProvider, w *Word) {
	t.Helper()
	rp.TempStack.Push(w)
}

func pushWordOnDS(t *testing.T, rp *RunProvider, w *Word) {
	t.Helper()
	rp.DataStack.Push(w)
}

func assertDSPropertyRefName(t *testing.T, rp *RunProvider, w *Word) {
	t.Helper()
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + PropertyRef)", rp.DataStack.Len())
	}
	ref, ok := rp.DataStack.Peek().(*PropertyRef)
	if !ok || ref.PM.PropertyName() != "name" {
		t.Fatalf("DS TOS = %T, want PropertyRef name", rp.DataStack.Peek())
	}
	if ref.Parent.Obj != w {
		t.Fatalf("PropertyRef parent = %v, want %v", ref.Parent.Obj, w)
	}
}

func assertDSNameValue(t *testing.T, rp *RunProvider, want String) {
	t.Helper()
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + implied read value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != want {
		t.Fatalf("TOS = %v, want %q", got, want)
	}
}

func runHandleCompoundWithSegs(task *Task, rp *RunProvider, segs []string) {
	pushCompoundPath(rp.DataStack, segs)
	RunOps(task, rp, OpHandleCompound, OpStop)
}

func TestHandleCompoundDotNameAtPathExplicitFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name", "@"})
	assertDSNameValue(t, rp, "hello")
}

func TestHandleCompoundLeadingAtFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	ths := NewTypeObjPair(w)
	pm := GetTypeRegistrar().Get("Word").PropertyByName("name")
	rp.DataStack.Push(&PropertyRef{Parent: ths, PM: pm})
	runHandleCompoundWithSegs(task, rp, []string{"@"})
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (PropertyRef elided, value only)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("TOS = %v, want hello", got)
	}
}

func TestHandleCompoundPropertyRefRMW(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("old", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name", "~"})
	RunOps(task, rp, OpDup, OpFetch, OpStop)
	rp.DataStack.Pop() // drop fetched value before store (R-M-W would combine here)
	rp.DataStack.Push(String("fresh"))
	RunOps(task, rp, OpSwap, OpStore, OpStop)
	if w.name != "fresh" {
		t.Fatalf("w.name = %q, want fresh", w.name)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want receiver only", rp.DataStack.Len())
	}
	if _, ok := rp.DataStack.Peek().(*Word); !ok {
		t.Fatalf("TOS = %T, want *Word", rp.DataStack.Peek())
	}
}

func TestHandleCompoundDotNameImpliedRead(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name"})
	assertDSNameValue(t, rp, "hello")
}

func TestHandleCompoundConstClassPropertyPushesValue(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	tt := GetTypeRegistrar().Get("TokenType")
	if tt == nil {
		t.Fatal("missing TokenType")
	}
	rp.DataStack.Push(tt)
	runHandleCompoundWithSegs(task, rp, []string{".", "TokenEOF"})
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + terminal value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int(TokenEOF) {
		t.Fatalf("TOS = %v, want Int(TokenEOF)", got)
	}
}

func TestHandleCompoundTokenTypeTokenEOFEphemeralElided(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runHandleCompoundWithSegs(task, rp, []string{"TokenType", ".", "TokenEOF"})
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (terminal only)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int(TokenEOF) {
		t.Fatalf("TOS = %v, want Int(TokenEOF)", got)
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after ephemeral elision", rp.TempStack.Len())
	}
}

func TestHandleCompoundDCDictMakeNewWordEphemeralChain(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	const wordName = "create4testword"
	rp.DataStack.Push(Int(WordFlagColon))
	rp.DataStack.Push(String(wordName))
	runHandleCompoundWithSegs(task, rp, []string{"dc", ".", "dict", ".", "makeNewWord", "^"})
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after ephemeral close", rp.TempStack.Len())
	}
	w, ok := rp.DataStack.Pop().(*Word)
	if !ok || w == nil || w.name != wordName || !w.IsColon() {
		t.Fatalf("TOS = %v, want colon *Word %q", w, wordName)
	}
}

func TestHandleCompoundCaretMustBeTerminal(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name", "^", ".", "other"})
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected error when ^ is not terminal")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Fatalf("err = %v", err)
	}
}

func TestHandleCompoundTildeMustBeTerminal(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name", "~", "@"})
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected error when ~ is not terminal")
	}
	if !strings.Contains(err.Error(), "terminal") {
		t.Fatalf("err = %v", err)
	}
}

func TestHandleCompoundTSelectorDotNameTildeRefOnDS(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("onTS", 0)
	rp.TempStack.Push(w)
	segs := CompoundPathWithSelectorForTest([]string{".", "name", "~"}, csTStack)
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (receiver only)", rp.TempStack.Len())
	}
	if rp.TempStack.Peek() != w {
		t.Fatalf("TS TOS = %v, want receiver %v", rp.TempStack.Peek(), w)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (PropertyRef on DS)", rp.DataStack.Len())
	}
	ref, ok := rp.DataStack.Peek().(*PropertyRef)
	if !ok || ref.PM.PropertyName() != "name" || ref.Parent.Obj != w {
		t.Fatalf("DS TOS = %T, want PropertyRef name for %v", rp.DataStack.Peek(), w)
	}
}

func TestHandleCompoundDotNameThenFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name", "~"})
	RunOps(task, rp, OpFetch, OpStop)
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "hello" {
		t.Fatalf("OpFetch name = %v, want hello", got)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want receiver only after @ consumes ref", rp.DataStack.Len())
	}
	if _, ok := rp.DataStack.Peek().(*Word); !ok {
		t.Fatalf("TOS = %T, want *Word", rp.DataStack.Peek())
	}
}

func TestHandleCompoundDotNameTildePushesPropertyRef(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	runHandleCompoundWithSegs(task, rp, []string{".", "name", "~"})
	assertDSPropertyRefName(t, rp, w)
}

func TestHandleCompoundUnknownNamePanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("probe", 0)
	pushWordOnDS(t, rp, w)
	pushCompoundPath(rp.DataStack, []string{".", "nosuch"})
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown accessor name")
		}
	}()
	RunOps(task, rp, OpHandleCompound, OpStop)
}

func TestDispatchInterpretCompoundDotName(t *testing.T) {
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name"})
	err := dispatchInterpretParsed(p, task, rp, segs, TokenCompoundIdent)
	if err != nil {
		t.Fatal(err)
	}
	assertDSNameValue(t, rp, "hello")
}

func TestTrSessionDispatchCompoundDotName(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	rp.DataStack.Push(CompoundPathForTest([]string{".", "name"}))
	rp.DataStack.Push(Int(TokenCompoundIdent))
	runTransition(t, task, rp, TrSessionDispatch)
	assertDSNameValue(t, rp, "hello")
}

func TestInterpretSessionFindDsOsDotNameFetch(t *testing.T) {
	t.Skip(">os .name @ compound path pending DS-centric fetch spike")
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	tr, err := NewTextReaderFromString("\"dup\" find? if >os .name @ then\n")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
}

func TestDispatchInterpretCompoundEmptyOSError(t *testing.T) {
	task, rp, p := testParserWithInput(t, "")
	segs := CompoundPathForTest([]string{".", "name"})
	err := dispatchInterpretParsed(p, task, rp, segs, TokenCompoundIdent)
	if err == nil {
		t.Fatal("expected error when context stack has no receiver")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after error", rp.DataStack.Len())
	}
}

func TestDispatchInterpretCompoundUnknownPrefixErrors(t *testing.T) {
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("probe", 0)
	pushWordOnTS(t, rp, w)
	segs := CompoundPathForTest([]string{"w", ".", "name"})
	err := dispatchInterpretParsed(p, task, rp, segs, TokenCompoundIdent)
	if err == nil {
		t.Fatal("expected error for unknown prefix segment")
	}
	if !strings.Contains(err.Error(), `unknown word "w"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestCompileCompoundTokenTypeTokenEOF(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: tok TokenType.TokenEOF ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("tok")
	if w == nil {
		t.Fatal("tok missing")
	}
	ct := globalDictionaryContext.ct
	body := int(w.colonCode)
	var hasDoLit, hasHandleCompound bool
	for i := body; i < ct.Len(); i++ {
		switch OpCode(ct.Read(i)) {
		case OpDoLit:
			hasDoLit = true
		case OpHandleCompound:
			hasHandleCompound = true
		case OpUnnest:
			if !hasDoLit || !hasHandleCompound {
				t.Fatalf("tok body missing OpDoLit=%v OpHandleCompound=%v", hasDoLit, hasHandleCompound)
			}
			runCompiledWord(t, task, rp, "tok")
			if rp.DataStack.Len() != 1 {
				t.Fatalf("DS len = %d, want 1 after tok", rp.DataStack.Len())
			}
			got, ok := rp.DataStack.Pop().(Int)
			if !ok || got != Int(TokenEOF) {
				t.Fatalf("TOS = %v, want Int(TokenEOF)", got)
			}
			return
		}
	}
	t.Fatal("tok body ended without OpUnnest")
}

func TestCompileCompoundDotName(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: probe "dup" find? if .name then ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("probe")
	if w == nil {
		t.Fatal("probe missing")
	}
	ct := globalDictionaryContext.ct
	body := int(w.colonCode)
	var hasDoLit, hasHandleCompound bool
	for i := body; i < ct.Len(); i++ {
		switch OpCode(ct.Read(i)) {
		case OpDoLit:
			hasDoLit = true
		case OpHandleCompound:
			hasHandleCompound = true
		case OpUnnest:
			if !hasDoLit || !hasHandleCompound {
				t.Fatalf("probe body missing OpDoLit=%v OpHandleCompound=%v", hasDoLit, hasHandleCompound)
			}
			runCompiledWord(t, task, rp, "probe")
			got, ok := rp.DataStack.Pop().(String)
			if !ok || got != "dup" {
				t.Fatalf("TOS = %v, want String(dup)", got)
			}
			return
		}
	}
	t.Fatal("probe body ended without OpUnnest")
}

func TestCompileCreate4MakeNewWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	const wordName = "create4testword"
	if err := compileBootstrapSource(task, rp, `: mk dc.dict.makeNewWord^ ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("mk")
	if w == nil {
		t.Fatal("mk missing")
	}
	rp.DataStack.Push(Int(WordFlagColon))
	rp.DataStack.Push(String(wordName))
	runCompiledWord(t, task, rp, "mk")
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0", rp.TempStack.Len())
	}
	got, ok := rp.DataStack.Pop().(*Word)
	if !ok || got == nil || got.name != wordName || !got.IsColon() {
		t.Fatalf("TOS = %v, want colon *Word %q", got, wordName)
	}
}

func runCompiledWord(t *testing.T, task *Task, rp *RunProvider, name string) {
	t.Helper()
	tr, err := NewTextReaderFromString(name + "\n")
	if err != nil {
		t.Fatal(err)
	}
	p := NewParserFromTextRead(tr)
	rp.SetParser(p)
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
}

func TestColonFindDsOsPrefix(t *testing.T) {
	t.Skip(">os removed; use t: prefix for temp-stack compound paths")
}
