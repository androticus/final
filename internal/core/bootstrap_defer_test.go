package core

import (
	"bufio"
	"strings"
	"testing"
)

func TestRegisterDeferWordInitTarget(t *testing.T) {
	ResetTestRuntime()
	handler := globalDictionaryContext.dict.GetByName("_boot_colon")
	if handler == nil {
		t.Fatal("missing _boot_colon transition worker")
	}
	deferW := registerDeferWord("_test_defer", WordFlagHidden, handler)
	if deferW.opCode != OpBootstrapDefer {
		t.Fatalf("defer opCode = %v, want OpBootstrapDefer", deferW.opCode)
	}
	if deferW.dtIdx < ReservedDTSlots {
		t.Fatalf("defer dtIdx = %d, want DT slot >= %d", deferW.dtIdx, ReservedDTSlots)
	}
	gotXT, ok := globalDictionaryContext.dict.XTOf(handler)
	if !ok {
		t.Fatal("XTOf handler")
	}
	slotVal := globalDictionaryContext.dt.Read(deferW.dtIdx)
	if slotVal.(Int) != gotXT {
		t.Fatalf("DT[slot] = %v, want handler XT %v", slotVal, gotXT)
	}
}

func TestDeferInterpretOpensColon(t *testing.T) {
	ResetTestRuntime()
	task, _ := NewTask(false)
	rp := NewRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("foo ")),
		label: "test",
	}})
	rp.SetParser(p)
	colonW := globalDictionaryContext.dict.GetByName(":")
	if colonW == nil {
		t.Fatal("missing :")
	}
	if err := runWordInlineExec(task, rp, colonW); err != nil {
		t.Fatal(err)
	}
	if !GetCompileState(task.compilerForTarget()) {
		t.Fatal("defer : should open compile state")
	}
	if task.dc.dict.GetByName("foo") != nil {
		t.Fatal("expected foo not in dictionary until ;")
	}
	if w := task.dc.dict.PeekBuilding(); w == nil || w.name != "foo" {
		t.Fatal("expected foo on building stack")
	}
}

func TestDeferCompileViaPrefixExec(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	semiW := globalDictionaryContext.dict.GetByName(";")
	if semiW == nil {
		t.Fatal("missing ;")
	}
	if semiW.dtIdx < 0 {
		t.Fatal("; defer word missing DT slot")
	}
	ct := globalDictionaryContext.ct
	start := ct.Len()
	ct.Append(int(OpCtImmDs))
	ct.Append(semiW.dtIdx)
	ct.Append(int(OpBootstrapDefer))
	ct.Append(int(OpStop))
	task := NewTestTask()
	task.compilerForTarget()
	bootstrapCreate(task.compilerState, "deferredsemi")
	rp.TempStack.Push(NewBlockToken(BlockKindColon))
	task.ip = EncodeVirt(ModuleIDThis, start)
	Run(task, rp)
	if GetCompileState(task.compilerForTarget()) {
		t.Fatal("compiled defer ; should finalize colon (compile state false)")
	}
	if task.dc.dict.GetByName("deferredsemi") == nil {
		t.Fatal("deferredsemi not in dictionary")
	}
}

func TestDeferFindInterpret(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(String("+"))
	runDeferWordByName(t, task, rp, "find?")
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatal("want true flag for known name")
	}
	got, ok := rp.DataStack.Pop().(*Word)
	if !ok || got == nil || got.name != "+" {
		t.Fatalf("got %v, want Word +", got)
	}
}

func TestBootstrapDeferFindCompiledInLocalModule(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: findQ: readIdentifier find? ;`); err != nil {
		t.Fatal(err)
	}
	findColon := task.dc.dict.GetByName("findQ:")
	if findColon == nil {
		t.Fatal("findQ: missing in local dictionary")
	}
	coreFindQ := globalDictionaryContext.dict.GetByName("find?")
	if coreFindQ == nil {
		t.Fatal("missing core find?")
	}
	ct := task.dc.ct
	body := int(findColon.colonCode)
	var sawBootstrapDefer bool
	for i := body; i < ct.Len(); i++ {
		switch OpCode(ct.Read(i)) {
		case OpBootstrapDefer:
			sawBootstrapDefer = true
			if ct.Read(i-1) != coreFindQ.dtIdx {
				t.Fatalf("OpBootstrapDefer operand = %d, want core find? dtIdx %d", ct.Read(i-1), coreFindQ.dtIdx)
			}
		case OpDefer:
			t.Fatalf("local findQ: body must not emit OpDefer at CT[%d]", i)
		case OpUnnest:
			if !sawBootstrapDefer {
				t.Fatal("findQ: body missing OpBootstrapDefer")
			}
			runInterpretScript(t, task, rp, "findQ: +\n")
			if rp.DataStack.Pop().(Bool) != true {
				t.Fatal("want true flag for +")
			}
			got, ok := rp.DataStack.Pop().(*Word)
			if !ok || got == nil || got.name != "+" {
				t.Fatalf("got %v, want Word +", got)
			}
			return
		}
	}
	t.Fatal("findQ: body ended without OpUnnest")
}

func TestDeferFindUnknown(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(String("nosuchword"))
	runDeferWordByName(t, task, rp, "find?")
	if rp.DataStack.Pop().(Bool) != false {
		t.Fatal("want false for unknown name")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack depth = %d, want 0", rp.DataStack.Len())
	}
}

func TestDeferCurrentDictHasWord(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(String("find?"))
	runDeferWordByName(t, task, rp, "currentDictHasWord")
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatal("find? should exist in dict")
	}
	rp.DataStack.Push(String("nosuchword"))
	runDeferWordByName(t, task, rp, "currentDictHasWord")
	if rp.DataStack.Pop().(Bool) != false {
		t.Fatal("nosuchword should not exist")
	}
}

func TestDeferReadIdentifier(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "foo ")
	rp.SetParser(p)
	runDeferWordByName(t, task, rp, "readIdentifier")
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "foo" {
		t.Fatalf("got %q, want foo", got)
	}
}

func TestDeferValidatePlainIdent(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(String("counter"))
	runDeferWordByName(t, task, rp, "validatePlainIdent")
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "counter" {
		t.Fatalf("got %q, want counter", got)
	}
}

func TestDeferValidatePlainIdentCompoundPanics(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.DataStack.Push(String(".field"))
	w := globalDictionaryContext.dict.GetByName("validatePlainIdent")
	if w == nil {
		t.Fatal("missing validatePlainIdent")
	}
	if err := runWordInlineExec(task, rp, w); err == nil {
		t.Fatal("expected error for compound name")
	}
}

func runDeferWordByName(t *testing.T, task *Task, rp *RunProvider, name string) {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName(name)
	if w == nil {
		t.Fatalf("missing defer word %q", name)
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("runWordInlineExec(%q): %v", name, err)
	}
}
