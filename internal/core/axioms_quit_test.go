package core

import (
	"strings"
	"testing"
)

func TestDeferGetNextTokenReadsLine(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "42\n")
	runDeferWordByName(t, task, rp, "getNextToken")
	wantDS(t, rp, Int(LiteralTokenType(TokenInt)), Int(42))
}

func TestDeferGetNextTokenStreamEOF(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	runDeferWordByName(t, task, rp, "getNextToken")
	wantDS(t, rp, Int(TokenEOF))
}

func TestDeferGetNextTokenNullParserEOF(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := globalDictionaryContext.dict.GetByName("getNextToken")
	if w == nil {
		t.Fatal("missing getNextToken")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("runWordInlineExec: %v", err)
	}
	if rp.DataStack.Len() != 1 || rp.DataStack.Pop().(Int) != Int(TokenEOF) {
		t.Fatalf("want TokenEOF on DS from null parser, got %v", rp.DataStack)
	}
}

func TestDeferGetNextTokenParserErrorPanics(t *testing.T) {
	task, rp, _ := testParserWithInput(t, ".\n")
	done := make(chan error, 1)
	w := globalDictionaryContext.dict.GetByName("getNextToken")
	if w == nil {
		t.Fatal("missing getNextToken")
	}
	ct := &VirtCTResolver{}
	task.ip = ct.Append(ModuleIDThis, int(OpCtImmDs))
	ct.Append(ModuleIDThis, w.dtIdx)
	ct.Append(ModuleIDThis, int(OpBootstrapDefer))
	ct.Append(ModuleIDThis, int(OpStop))
	task.doneChan = done
	Run(task, rp)
	err := <-done
	if err == nil {
		t.Fatal("expected panic on parser error")
	}
	if !strings.Contains(err.Error(), "getNextToken") {
		t.Fatalf("panic error = %q, want getNextToken", err.Error())
	}
}

func TestDeferReplError(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "skipped rest\n")
	sb := NewStringBuilder()
	rp.Stderr = sb
	cs := task.compilerForTarget()
	SetCompileState(cs, true)
	defer SetCompileState(cs, false)

	rp.DataStack.Push(String("test failure"))
	runDeferWordByName(t, task, rp, "replError")

	if got := strings.TrimSuffix(sb.ToString(), "\n"); got != "test failure" {
		t.Fatalf("stderr = %q, want %q", got, "test failure")
	}
	if GetCompileState(cs) {
		t.Fatal("compile state still true after replError")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack depth = %d, want 0", rp.DataStack.Len())
	}
}

func TestDeferReplErrorBadMessageType(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	sb := NewStringBuilder()
	rp.Stderr = sb
	rp.DataStack.Push(Int(1))
	runDeferWordByName(t, task, rp, "replError")
	if !strings.Contains(sb.ToString(), "replError: expected String message") {
		t.Fatalf("stderr = %q", sb.ToString())
	}
}

func TestWarmReplLoopByeReqExitsViaUnnest(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	SetByeReq(globalDictionaryContext, true)
	rp.ReturnStack.PushNew(int(ReservedCTBootStop))
	task.ip = EncodeVirt(ModuleIDThis, int(BootstrapReplLoopHeadCTP()))
	Run(task, rp)
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq should be cleared")
	}
}

func TestByeSetsByeReq(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	w := globalDictionaryContext.dict.GetByName("bye")
	if w == nil {
		t.Fatal("missing bye")
	}
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("runWordInlineExec(bye): %v", err)
	}
	if !GetByeReq(globalDictionaryContext) {
		t.Fatal("bye should set byeReq true")
	}
}

func TestCheckResetByeReqFinalWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()

	SetByeReq(globalDictionaryContext, false)
	runDeferWordByName(t, task, rp, "checkResetByeReq")
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got {
		t.Fatalf("checkResetByeReq false = %v, want Bool(false)", got)
	}

	SetByeReq(globalDictionaryContext, true)
	runDeferWordByName(t, task, rp, "checkResetByeReq")
	got, ok = rp.DataStack.Pop().(Bool)
	if !ok || !got {
		t.Fatalf("checkResetByeReq true = %v, want Bool(true)", got)
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq should be cleared")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack depth after true path = %d, want 0 (no ref leak)", rp.DataStack.Len())
	}

	SetByeReq(globalDictionaryContext, true)
	runDeferWordByName(t, task, rp, "checkResetByeReq")
	got, ok = rp.DataStack.Pop().(Bool)
	if !ok || !got {
		t.Fatalf("second checkResetByeReq true = %v, want Bool(true)", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack depth after second true path = %d, want 0", rp.DataStack.Len())
	}
}
