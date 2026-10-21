package core

import "testing"

// runInterpretScriptWithExtraScope runs one script with task FindScope plus an extra inner layer.
func runInterpretScriptWithExtraScope(t *testing.T, task *Task, rp *RunProvider, extra Scope, script string) {
	t.Helper()
	tr, err := NewTextReaderFromString(script)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	rp.WithTaskFindScope(task)
	if extra != nil {
		rp.FindScope.PushScope(extra)
	}
	for {
		if err := pushNextToken(rp.Parser, rp.DataStack); err != nil {
			t.Fatal(err)
		}
		if PopStreamEOF(rp.DataStack) {
			break
		}
		if err := dispatchSessionToken(rp.Parser, task, rp); err != nil {
			t.Fatal(err)
		}
	}
}

// runInterpretScript wires a Final source string into rp.Parser and runs interpretSession.
func runInterpretScript(t *testing.T, task *Task, rp *RunProvider, script string) {
	t.Helper()
	tr, err := NewTextReaderFromString(script)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
}

// pushFindQ runs bootstrap find? and leaves the resolved *Word on DS (val/ok peeled).
func pushFindQ(t *testing.T, task *Task, rp *RunProvider, name string) {
	t.Helper()
	rp.DataStack.Push(String(name))
	runDeferWordByName(t, task, rp, "find?")
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatalf("find? %q: want true flag", name)
	}
}

func TestInterpretSessionTokenTypeTokenEOF(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp, "TokenType.TokenEOF\n")
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1 (ephemeral elided)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int(TokenEOF) {
		t.Fatalf("TOS = %v, want Int(TokenEOF)", got)
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0", rp.TempStack.Len())
	}
}

func TestInterpretSessionDupFindDotNameFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	pushFindQ(t, task, rp, "dup")
	runInterpretScript(t, task, rp, ".name@\n")
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + fetched value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
	w, ok := rp.DataStack.Pop().(*Word)
	if !ok || w == nil || w.name != "dup" {
		t.Fatalf("receiver = %v, want *Word dup", w)
	}
}

func TestInterpretSessionDotFetchUsesTOSReceiver(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: brad: nop ;`); err != nil {
		t.Fatal(err)
	}
	pushFindQ(t, task, rp, "brad:")
	pushFindQ(t, task, rp, "dup")
	runInterpretScript(t, task, rp, ".name@\n")
	if rp.DataStack.Len() != 3 {
		t.Fatalf("DS len = %d, want 3 (brad:, dup, name)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
	dupW, ok := rp.DataStack.Pop().(*Word)
	if !ok || dupW == nil || dupW.name != "dup" {
		t.Fatalf("TOS receiver = %v, want *Word dup", dupW)
	}
	bradW, ok := rp.DataStack.Pop().(*Word)
	if !ok || bradW == nil || bradW.name != "brad:" {
		t.Fatalf("below receiver = %v, want *Word brad:", bradW)
	}
}

func TestInterpretSessionDotFetchStackShape(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: brad: nop ;`); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapSource(task, rp, `: brad1 nop ;`); err != nil {
		t.Fatal(err)
	}
	pushFindQ(t, task, rp, "brad:")
	runInterpretScript(t, task, rp, ".colonCode@\n")
	pushFindQ(t, task, rp, "brad1")
	runInterpretScript(t, task, rp, ".colonCode@\n")
	if rp.DataStack.Len() != 4 {
		t.Fatalf("DS len = %d, want 4 (w int w int)", rp.DataStack.Len())
	}
	brad1Code, ok := rp.DataStack.Pop().(Int)
	if !ok {
		t.Fatalf("TOS = %T, want Int colonCode", rp.DataStack.Peek())
	}
	brad1, ok := rp.DataStack.Pop().(*Word)
	if !ok || brad1 == nil || brad1.name != "brad1" {
		t.Fatalf("receiver = %v, want *Word brad1", brad1)
	}
	bradCode, ok := rp.DataStack.Pop().(Int)
	if !ok {
		t.Fatalf("TOS = %T, want Int colonCode", rp.DataStack.Peek())
	}
	brad, ok := rp.DataStack.Pop().(*Word)
	if !ok || brad == nil || brad.name != "brad:" {
		t.Fatalf("receiver = %v, want *Word brad:", brad)
	}
	if bradCode < 0 {
		t.Fatalf("brad.colonCode = %d, want >= 0", bradCode)
	}
	if brad1Code < 0 {
		t.Fatalf("brad1.colonCode = %d, want >= 0", brad1Code)
	}
}

func TestInterpretSessionFindDotNameImpliedRead(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	pushFindQ(t, task, rp, "dup")
	runInterpretScript(t, task, rp, ".name\n")
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (receiver + implied read value)", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("TOS = %v, want String(dup)", got)
	}
}

func TestInterpretSessionPropertyRefRMW(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := NewWord("oldname", 0)
	rp.DataStack.Push(w)
	runInterpretScript(t, task, rp, ".name~\ndup\n@\ndrop\n\"fresh\"\nswap\n!\n")
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

func TestInterpretSessionDCDictMakeNewWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	const wordName = "compoundMkWord"
	runInterpretScript(t, task, rp, "2\n\""+wordName+"\"\ndc.dict.makeNewWord^\n")
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after ephemeral close", rp.TempStack.Len())
	}
	got, ok := rp.DataStack.Pop().(*Word)
	if !ok || got == nil || got.name != wordName || !got.IsColon() {
		t.Fatalf("TOS = %v, want colon *Word %q", got, wordName)
	}
}

func TestInterpretSessionFindOnTSDotNameFetch(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	pushFindQ(t, task, rp, "dup")
	runInterpretScript(t, task, rp, ">ts\nt:.name@\n")
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "dup" {
		t.Fatalf("fetched name = %v, want String(dup)", got)
	}
	if rp.TempStack.Len() != 1 {
		t.Fatalf("TS len = %d, want 1 (receiver kept)", rp.TempStack.Len())
	}
}
