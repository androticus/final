package core

import (
	"bufio"
	"strings"
	"testing"
)

func testParserWithInput(t *testing.T, input string) (*Task, *RunProvider, *Parser) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(input)),
		label: "test",
	}})
	rp.SetParser(p)
	return task, rp, p
}

func runTransition(t *testing.T, task *Task, rp *RunProvider, id RuntimeDispatchID) {
	t.Helper()
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, int(id))
	ct.Append(ModuleIDThis, int(OpTransition))
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
}

// runSessionDispatch runs TrSessionDispatch then OpExecWord when dispatch queued ( w true -- ).
func runSessionDispatch(t *testing.T, task *Task, rp *RunProvider) {
	t.Helper()
	runTransition(t, task, rp, TrSessionDispatch)
	if rp.DataStack.Len() < 2 {
		return
	}
	top, ok := rp.DataStack.Peek().(Bool)
	if !ok || !top {
		return
	}
	rp.DataStack.Pop()
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpExecWord))
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
}

func TestDispatchExecWordInRunLoopArmsCS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dropW := globalDictionaryContext.dict.GetByName("drop")
	if dropW == nil {
		t.Fatal("drop missing")
	}
	rp.DataStack.Push(Int(1))
	if err := dispatchExecWord(task, rp, dropW, true); err != nil {
		t.Fatalf("dispatchExecWord(drop, inRunLoop): %v", err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0 after armed drop", rp.DataStack.Len())
	}
}

func TestDispatchExecWordInRunLoopQueuesNonCS(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	nopW := globalDictionaryContext.dict.GetByName("nop")
	if nopW == nil {
		t.Fatal("nop missing")
	}
	if err := dispatchExecWord(task, rp, nopW, true); err != nil {
		t.Fatalf("dispatchExecWord(nop, inRunLoop): %v", err)
	}
	if rp.DataStack.Len() != 2 {
		t.Fatalf("DS len = %d, want 2 (w true)", rp.DataStack.Len())
	}
	top, ok := rp.DataStack.Peek().(Bool)
	if !ok || !top {
		t.Fatalf("TOS = %v, want Bool(true)", rp.DataStack.Peek())
	}
	w, ok := rp.DataStack.At(1).(*Word)
	if !ok || w != nopW {
		t.Fatalf("ds[1] = %v, want nop word", rp.DataStack.At(1))
	}
}

func TestRunProviderSetParserNilPanics(t *testing.T) {
	rp := NewTestRunProvider()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("SetParser(nil) should panic")
		}
	}()
	rp.SetParser(nil)
}

func TestTrCheckResetByeReqFalse(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	SetByeReq(globalDictionaryContext, false)
	runTransition(t, task, rp, TrCheckResetByeReq)
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got {
		t.Fatalf("CheckResetByeReq false = %v, want Bool(false)", got)
	}
}

func TestTrCheckResetByeReqTrueClearsSlot(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	SetByeReq(globalDictionaryContext, true)
	runTransition(t, task, rp, TrCheckResetByeReq)
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || !got {
		t.Fatalf("CheckResetByeReq true = %v, want Bool(true)", got)
	}
	if GetByeReq(globalDictionaryContext) {
		t.Fatal("byeReq slot should be cleared")
	}
}

func TestOpTransitionNullParserEOF(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if rp.Parser != GetNullParser() {
		t.Fatal("NewRunProvider should start with GetNullParser")
	}
	ct := &VirtCTResolver{}
	ip := ct.Append(ModuleIDThis, int(OpDoSubCode))
	ct.Append(ModuleIDThis, int(TrNextToken))
	ct.Append(ModuleIDThis, int(OpTransition))
	ct.Append(ModuleIDThis, int(OpStop))
	task.ip = ip
	Run(task, rp)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want TokenEOF", rp.DataStack.Len())
	}
	if got, ok := rp.DataStack.Pop().(Int); !ok || got != Int(TokenEOF) {
		t.Fatalf("TOS = %v, want TokenEOF", got)
	}
}

func TestTrNextTokenDoesNotWriteREPLOK(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "1 2\n")
	var out StringBuilder
	rp.Stdout = &out
	runTransition(t, task, rp, TrNextToken)
	if out.ToString() != "" {
		t.Fatalf("next-token wrote %q to stdout, want silent core loop", out.ToString())
	}
}

func TestTrNextTokenReadsLine(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "42\n")
	runTransition(t, task, rp, TrNextToken)
	if rp.DataStack.Len() != 2 {
		t.Fatalf("stack len = %d, want value+type", rp.DataStack.Len())
	}
	if TokenType(rp.DataStack.Pop().(Int)) != LiteralTokenType(TokenInt) {
		t.Fatal("want TokenInt")
	}
	if rp.DataStack.Pop().(Int) != 42 {
		t.Fatal("want 42")
	}
}

func TestTrNextTokenStreamEOF(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	runTransition(t, task, rp, TrNextToken)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("stack len = %d, want TokenEOF only", rp.DataStack.Len())
	}
	if TokenType(rp.DataStack.Pop().(Int)) != TokenEOF {
		t.Fatal("want TokenEOF on stream EOF")
	}
}

func TestTrFind(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	wantW, _, ok := FindName("+")
	if !ok {
		t.Fatal("FindName +")
	}
	rp.DataStack.Push(String("+"))
	runTransition(t, task, rp, TrFind)
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatal("want true flag for known name")
	}
	got, ok := rp.DataStack.Pop().(*Word)
	if !ok || got == nil || got.name != wantW.name {
		t.Fatalf("got %v, want Word %q", got, wantW.name)
	}
	rp.DataStack.Push(String("nosuchword"))
	runTransition(t, task, rp, TrFind)
	if rp.DataStack.Pop().(Bool) != false {
		t.Fatal("want false for unknown name")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack depth = %d, want 0", rp.DataStack.Len())
	}
}

func TestTrSessionDispatchLiteral(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(Int(7))
	rp.DataStack.Push(Int(LiteralTokenType(TokenInt)))
	runTransition(t, task, rp, TrSessionDispatch)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("stack len = %d, want 7", rp.DataStack.Len())
	}
	if rp.DataStack.Pop().(Int) != 7 {
		t.Fatal("want 7")
	}
}

func TestTrSessionDispatchExec(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "")
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	rp.DataStack.Push(String("+"))
	rp.DataStack.Push(Int(TokenIdent))
	runSessionDispatch(t, task, rp)
	gotSum, ok := rp.DataStack.Pop().(Int)
	if !ok || gotSum != 3 {
		t.Fatalf("got %v, want 3", gotSum)
	}
}

func TestTrSessionDispatchUnknownWord(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "rest\n")
	rp.DataStack.Push(String("nosuchword"))
	rp.DataStack.Push(Int(TokenIdent))
	runTransition(t, task, rp, TrSessionDispatch)
	if rp.DataStack.Len() != 0 {
		t.Fatalf("after error want empty DS, len=%d", rp.DataStack.Len())
	}
}

func TestTrSessionDispatchColonDefine(t *testing.T) {
	task, rp, _ := testParserWithInput(t, ": foo ;\n")
	runTransition(t, task, rp, TrNextToken) // :
	if rp.DataStack.Len() != 2 {
		t.Fatalf("after next token len=%d", rp.DataStack.Len())
	}
	if TokenType(rp.DataStack.Pop().(Int)) != TokenIdent {
		t.Fatal("want TokenIdent")
	}
	if rp.DataStack.Pop().(String) != ":" {
		t.Fatal("want :")
	}
	rp.DataStack.Push(String(":"))
	rp.DataStack.Push(Int(TokenIdent))
	runSessionDispatch(t, task, rp)
	cs := task.compilerState
	if cs == nil || !GetCompileState(cs) {
		t.Fatalf("after : expected compile state")
	}
	if globalDictionaryContext.dict.GetByName("foo") != nil {
		t.Fatal("foo should not be in dictionary until ;")
	}
	if w := globalDictionaryContext.dict.PeekBuilding(); w == nil || w.name != "foo" {
		t.Fatal("foo should be on building stack after :")
	}
	runTransition(t, task, rp, TrNextToken) // ;
	if TokenType(rp.DataStack.Pop().(Int)) != TokenIdent {
		t.Fatal("want TokenIdent for ;")
	}
	if rp.DataStack.Pop().(String) != ";" {
		t.Fatal("want ;")
	}
	rp.DataStack.Push(String(";"))
	rp.DataStack.Push(Int(TokenIdent))
	runSessionDispatch(t, task, rp)
	if globalDictionaryContext.dict.GetByName("foo") == nil {
		t.Fatal("foo not in dict after ;")
	}
	if GetCompileState(cs) {
		t.Fatal("still compiling after ;")
	}
}

func TestTrReplError(t *testing.T) {
	task, rp, _ := testParserWithInput(t, "x\n")
	rp.DataStack.Push(String("test failure"))
	runTransition(t, task, rp, TrReplError)
}
