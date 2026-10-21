package core

import "testing"

func beginCompileEmitTask(t *testing.T) (*Task, *RunProvider, *CompilerState) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "compile-emit-test")
	return task, rp, cs
}

func requireDictWord(t *testing.T, name string) *Word {
	t.Helper()
	w := globalDictionaryContext.dict.GetByName(name)
	if w == nil {
		t.Fatalf("word %q missing from dictionary", name)
	}
	return w
}

func TestOpCTHere(t *testing.T) {
	task, rp, _ := beginCompileEmitTask(t)
	before := globalDictionaryContext.ct.Len()
	if err := runWordInlineExec(task, rp, requireDictWord(t, "here")); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != Int(before) {
		t.Fatalf("here = %v, want %d", got, before)
	}
}

func TestOpCTComma(t *testing.T) {
	task, rp, _ := beginCompileEmitTask(t)
	before := globalDictionaryContext.ct.Len()
	rp.DataStack.Push(Int(12345))
	if err := runWordInlineExec(task, rp, requireDictWord(t, ",")); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.ct.Len() != before+1 {
		t.Fatalf("ct len = %d, want %d", globalDictionaryContext.ct.Len(), before+1)
	}
	if globalDictionaryContext.ct.Read(before) != 12345 {
		t.Fatalf("CT[%d] = %d, want 12345", before, globalDictionaryContext.ct.Read(before))
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack not tidy after comma: len %d", rp.DataStack.Len())
	}
}

func TestOpCTStore_forwardResolve(t *testing.T) {
	task, rp, _ := beginCompileEmitTask(t)
	ct := globalDictionaryContext.ct
	jSlot := ct.Append(int(OpBranchQ))
	branchBase := jSlot + 1
	merge := ct.Append(int(OpNop))
	j := merge - branchBase
	rp.DataStack.Push(Int(j))
	rp.DataStack.Push(Int(jSlot))
	if err := runWordInlineExec(task, rp, requireDictWord(t, "ct!")); err != nil {
		t.Fatal(err)
	}
	if ct.Read(jSlot) != j {
		t.Fatalf("patched j = %d, want %d", ct.Read(jSlot), j)
	}
}

func TestOpCTPatchForward(t *testing.T) {
	task, rp, _ := beginCompileEmitTask(t)
	ct := globalDictionaryContext.ct
	_ = ct.Append(int(OpBranchQ))
	orig := ct.Append(0) // placeholder at markIP-1
	_ = ct.Append(int(OpNop))
	want := ct.Len() - (orig + 1)
	rp.DataStack.Push(Int(orig))
	if err := runWordInlineExec(task, rp, requireDictWord(t, ">resolve")); err != nil {
		t.Fatal(err)
	}
	if ct.Read(orig) != want {
		t.Fatalf("patched j = %d, want %d", ct.Read(orig), want)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("stack not tidy after >resolve: len %d", rp.DataStack.Len())
	}
}

func TestMarkViaInlineExec(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "probe")
	wMark := globalDictionaryContext.dict.GetByName(">mark")
	if wMark == nil {
		t.Fatal(">mark missing")
	}
	if err := runWordInlineExec(task, rp, wMark); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("ds len %d, want 1", rp.DataStack.Len())
	}
}

func TestMarkColonBodyRuns(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "probe")
	wMark := globalDictionaryContext.dict.GetByName(">mark")
	if wMark == nil {
		t.Fatal(">mark missing")
	}
	off := int(wMark.colonCode)
	ct := globalDictionaryContext.ct
	stopIP := ct.Len()
	ct.Append(int(OpStop))
	rp.ReturnStack.PushNew(stopIP)
	task.ip = EncodeVirt(ModuleIDThis, off)
	Run(task, rp)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("after >mark body ds len %d, want 1", rp.DataStack.Len())
	}
}

func TestMarkResolveStackTidy(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "use-mark")
	wMark := globalDictionaryContext.dict.GetByName(">mark")
	if wMark == nil {
		t.Fatal(">mark missing")
	}
	if err := runWordInlineExec(task, rp, wMark); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("after >mark ds len %d, want 1", rp.DataStack.Len())
	}
	if err := runWordInlineExec(task, rp, requireDictWord(t, ">resolve")); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("after >resolve ds len %d, want 0", rp.DataStack.Len())
	}
}

func TestOpRecurse_emitsNestToBuildingWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "selfrec")
	w := cs.DC.dict.PeekBuilding()
	entry := int(w.colonCode)
	if err := runWordInlineExec(task, rp, requireDictWord(t, "recurse")); err != nil {
		t.Fatal(err)
	}
	ct := globalDictionaryContext.ct
	if ct.Read(entry) != int(OpNest) {
		t.Fatalf("body[0] = %s, want Nest", OpCodeName(OpCode(ct.Read(entry))))
	}
	if ct.Read(entry+2) != entry {
		t.Fatalf("nest entry = %d, want building colonCode %d", ct.Read(entry+2), entry)
	}
	cs.DC.dict.ClearBuilding()
	SetCompileState(cs, false)
}

func TestOpSetWordFlags(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	SetCompileState(cs, true)
	w := NewWord("flagme", WordFlagColon)
	w.colonCode = CTP(0)
	globalDictionaryContext.dict.Add(w)
	rp.DataStack.Push(Int(WordFlagImmediate))
	if err := runWordInlineExec(task, rp, requireDictWord(t, "setWordFlags")); err != nil {
		t.Fatal(err)
	}
	if !w.IsImmediate() {
		t.Fatal("expected WordFlagImmediate on flagme")
	}
	SetCompileState(cs, false)
}

func TestEmitBracketStringLiteral(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: emstr [[ "hello" ]] ;`); err != nil {
		t.Fatal(err)
	}
	emstr := globalDictionaryContext.dict.GetByName("emstr")
	if emstr == nil {
		t.Fatal("missing emstr")
	}
	off := int(emstr.colonCode)
	ct := globalDictionaryContext.ct
	if OpCode(ct.Read(off)) != OpDoLit {
		t.Fatalf("emstr body[0] = %s, want OpDoLit", OpCodeName(OpCode(ct.Read(off))))
	}
	foundCompileLit := false
	for i := off; i < ct.Len(); i++ {
		op := OpCode(ct.Read(i))
		if op == OpUnnest {
			break
		}
		if op == OpCompileLit {
			foundCompileLit = true
			break
		}
	}
	if !foundCompileLit {
		t.Fatal("emstr missing OpCompileLit after OpDoLit (emit bracket literal program)")
	}

	cs := task.compilerForTarget()
	bootstrapCreate(cs, "strclient")
	emitAt := cs.ct().Len()
	if err := runWordInlineExec(task, rp, emstr); err != nil {
		t.Fatal(err)
	}
	gotLit := false
	for i := emitAt; i < cs.ct().Len(); i++ {
		if OpCode(cs.ct().Read(i)) == OpDoLit {
			gotLit = true
			break
		}
	}
	if !gotLit {
		t.Fatalf("client CT [%d:] missing OpDoLit after emstr run", emitAt)
	}
	client := cs.DC.dict.PeekBuilding()
	if client == nil {
		t.Fatal("missing open client word")
	}
	stopIP := cs.ct().Len()
	cs.ct().Append(int(OpStop))
	rp.ReturnStack.PushNew(stopIP)
	task.ip = EncodeVirt(ModuleIDThis, emitAt)
	Run(task, rp)
	s, ok := rp.DataStack.Pop().(String)
	if !ok || s != "hello" {
		t.Fatalf("client literal = %v, want %q", s, "hello")
	}
}

func TestEmitBracketIntLiteral(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(task, rp, `: emint [[ 42 ]] ;`); err != nil {
		t.Fatal(err)
	}
	emint := globalDictionaryContext.dict.GetByName("emint")
	if emint == nil {
		t.Fatal("missing emint")
	}
	off := int(emint.colonCode)
	ct := globalDictionaryContext.ct
	if OpCode(ct.Read(off)) != OpDoLit {
		t.Fatalf("emint body[0] = %s, want OpDoLit", OpCodeName(OpCode(ct.Read(off))))
	}
	foundCompileLit := false
	for i := off; i < ct.Len(); i++ {
		op := OpCode(ct.Read(i))
		if op == OpUnnest {
			break
		}
		if op == OpCompileLit {
			foundCompileLit = true
			break
		}
	}
	if !foundCompileLit {
		t.Fatal("emint missing OpCompileLit after OpDoLit (emit bracket literal program)")
	}

	cs := task.compilerForTarget()
	bootstrapCreate(cs, "intclient")
	emitAt := cs.ct().Len()
	if err := runWordInlineExec(task, rp, emint); err != nil {
		t.Fatal(err)
	}
	if OpCode(cs.ct().Read(emitAt)) != OpDoLitI {
		t.Fatalf("CT[%d] = %s, want OpDoLitI", emitAt, OpCodeName(OpCode(cs.ct().Read(emitAt))))
	}
	if cs.ct().Read(emitAt+1) != 42 {
		t.Fatalf("CT[%d] = %d, want 42", emitAt+1, cs.ct().Read(emitAt+1))
	}
	stopIP := cs.ct().Len()
	cs.ct().Append(int(OpStop))
	rp.ReturnStack.PushNew(stopIP)
	task.ip = EncodeVirt(ModuleIDThis, emitAt)
	Run(task, rp)
	n, ok := rp.DataStack.Pop().(Int)
	if !ok || n != 42 {
		t.Fatalf("client literal = %v, want 42", n)
	}
}

func TestBranchCompileEmit(t *testing.T) {
	task, rp, cs := beginCompileEmitTask(t)
	SetEmitting(cs, true)
	before := globalDictionaryContext.ct.Len()
	if err := handleParsedTokenCompile(nil, task, rp, cs, String("branch?"), TokenIdent, false, nil); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.ct.Len() != before+2 {
		t.Fatalf("ct len = %d, want %d", globalDictionaryContext.ct.Len(), before+2)
	}
	if globalDictionaryContext.ct.Read(before) != int(OpDoEmit) {
		t.Fatalf("emitted %d, want OpDoEmit", globalDictionaryContext.ct.Read(before))
	}
	wBranchQ := requireDictWord(t, "branch?")
	xt, ok := globalDictionaryContext.dict.XTOf(wBranchQ)
	if !ok {
		t.Fatal("XTOf branch?")
	}
	if globalDictionaryContext.ct.Read(before+1) != int(xt) {
		t.Fatalf("CT[%d] xt = %d, want %d", before+1, globalDictionaryContext.ct.Read(before+1), int(xt))
	}
}
