package core

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestCompileBootstrapInitScript(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{
		">mark", ">resolve", "<mark", "<resolve", "breakResolve",
		"if", "nif", "else", "then",
		"begin", "break", "loopEndBreakStep", "again", "until", "while", "repeat",
		"forever", "/forever", "continue",
		"foreach", "/foreach", "i", "j", "k",
		"panicRequireNKind", "panicRequireOneOfKind", "setBlockMark", "continueMark",
		"case", "/case", "switch", "default", "/default", "/switch",
		"a", "b", "c",
		"try", "catch", "finally", "/try", "throw",
		"compileLiteral",
		"bye", "checkResetByeReq", "exit",
		"within", "min", "max",
		"sqr", "isClose", "pow",
	} {
		if globalDictionaryContext.dict.GetByName(name) == nil {
			t.Fatalf("missing compiled word %q", name)
		}
	}
	for _, name := range []string{
		"if", "nif", "else", "then",
		"begin", "again", "until", "while", "repeat",
		"forever", "/forever", "break", "continue",
		"foreach", "/foreach", "i", "j", "k",
		"switch", "case", "/case", "default", "/default", "/switch",
		"a", "b", "c",
		"try", "catch", "finally", "/try",
	} {
		w := globalDictionaryContext.dict.GetByName(name)
		if w == nil {
			t.Fatalf("missing %q", name)
		}
		if !w.IsImmediate() {
			t.Fatalf("%q should be Immediate", name)
		}
		if w.flags&WordFlagCompileOnly == 0 {
			t.Fatalf("%q should be CompileOnly", name)
		}
	}
}

func TestMinMaxWithinAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()

	cases := []struct {
		script string
		want   any
	}{
		{"3 5 min\n", Int(3)},
		{"5 3 min\n", Int(3)},
		{"3 5 max\n", Int(5)},
		{"5 3 max\n", Int(5)},
		{"3.5 1.2 max\n", Float(3.5)},
		{"10 0 20 within\n", Bool(true)},
		{"-1 0 20 within\n", Bool(false)},
		{"20 0 20 within\n", Bool(true)},
	}
	for _, tc := range cases {
		rp.DataStack.Clear()
		runInterpretScript(t, tk, rp, tc.script)
		if rp.DataStack.Len() != 1 {
			t.Fatalf("%q: DS len = %d, want 1", tc.script, rp.DataStack.Len())
		}
		got := rp.DataStack.Pop()
		if got != tc.want {
			t.Fatalf("%q = %v (%T), want %v (%T)", tc.script, got, got, tc.want, tc.want)
		}
	}
}

func TestSqrIsClosePowAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()

	cases := []struct {
		script string
		want   any
	}{
		{"4 sqr\n", Int(16)},
		{"-3 sqr\n", Int(9)},
		{"1.5 sqr\n", Float(2.25)},
		{"2 10 pow\n", Float(1024)},
		{"2.0 3.0 pow\n", Float(8)},
		{"1.0 1.0001 0.001 isClose\n", Bool(true)},
		{"1.0 1.1 0.01 isClose\n", Bool(false)},
		{"10 10.0 0 isClose\n", Bool(true)},
	}
	for _, tc := range cases {
		rp.DataStack.Clear()
		runInterpretScript(t, tk, rp, tc.script)
		if rp.DataStack.Len() != 1 {
			t.Fatalf("%q: DS len = %d, want 1", tc.script, rp.DataStack.Len())
		}
		got := rp.DataStack.Pop()
		if got != tc.want {
			t.Fatalf("%q = %v (%T), want %v (%T)", tc.script, got, got, tc.want, tc.want)
		}
	}
}

func TestFloatNaNIsNaNIsInfViaScript(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()

	rp.DataStack.Clear()
	runInterpretScript(t, tk, rp, "Float.NaN Float.isNaN^\n")
	if rp.DataStack.Len() != 1 {
		t.Fatalf("Float.NaN Float.isNaN^: DS len = %d, want 1", rp.DataStack.Len())
	}
	if got := rp.DataStack.Pop(); got != Bool(true) {
		t.Fatalf("Float.NaN Float.isNaN^ = %v, want true", got)
	}

	rp.DataStack.Clear()
	runInterpretScript(t, tk, rp, "1.0 Float.isNaN^\n")
	if got := rp.DataStack.Pop(); got != Bool(false) {
		t.Fatalf("1.0 Float.isNaN^ = %v, want false", got)
	}

	rp.DataStack.Clear()
	runInterpretScript(t, tk, rp, "1.0 0.0 / Float.isInf^\n")
	if got := rp.DataStack.Pop(); got != Bool(true) {
		t.Fatalf("1.0 0.0 / Float.isInf^ = %v, want true", got)
	}

	rp.DataStack.Clear()
	runInterpretScript(t, tk, rp, "1.0 Float.isInf^\n")
	if got := rp.DataStack.Pop(); got != Bool(false) {
		t.Fatalf("1.0 Float.isInf^ = %v, want false", got)
	}
}

func TestSwitchCaseBranchesPatched(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	if err := compileBootstrapSource(tk, rp, `: casetest null switch 1 case 111 /case 222 /switch ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("casetest")
	if w == nil {
		t.Fatal("missing casetest")
	}
	ct := globalDictionaryContext.ct
	off := int(w.colonCode)
	var branchOps []int
	for off < ct.Len() {
		op := ct.Read(off)
		if OpCode(op) == OpBranch {
			j := ct.Read(off + 1)
			branchOps = append(branchOps, j)
		}
		off++
		if OpCode(op) == OpStop {
			break
		}
	}
	if len(branchOps) == 0 {
		t.Fatal("no OpBranch in casetest body")
	}
	for i, j := range branchOps {
		if j == 0 {
			t.Fatalf("/case branch[%d] still 0 (unpatched)", i)
		}
	}
}

func TestCompileTargetSwitchAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	src := `: pick2 switch a 1 == case 1 /case a 2 == case 2 /case default 0 /default /switch ;`
	if err := compileBootstrapSource(tk, rp, src); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("pick2") == nil {
		t.Fatal("missing pick2")
	}
}

func TestRunTargetSwitchFirstMatch(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": first switch a 1 == case 111 /case a 1 == case 222 /case /switch ; 1 first\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 111 {
		t.Fatalf("first = %v, want 111", got)
	}
}

func TestRunTargetSwitchDefault(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: pick2 switch a 1 == case 1 /case a 2 == case 2 /case default 0 /default /switch ;
3 pick2
5 pick2
1 pick2
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	want := []Int{0, 0, 1} // stack pushes: 3→0, 5→0, 1→1 (TOS first on pop)
	for i := len(want) - 1; i >= 0; i-- {
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != want[i] {
			t.Fatalf("result[%d] = %v, want %v", i, got, want[i])
		}
	}
}

func TestCompileLiteralEmitsInt(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cl := globalDictionaryContext.dict.GetByName("compileLiteral")
	if cl == nil {
		t.Fatal("missing compileLiteral")
	}
	litWord := globalDictionaryContext.dict.GetByName("literal,")
	if litWord == nil {
		t.Fatal("missing literal,")
	}
	xt, ok := globalDictionaryContext.dict.XTOf(litWord)
	if !ok {
		t.Fatal("XTOf literal,")
	}
	off := int(cl.colonCode)
	ct := globalDictionaryContext.ct
	if ct.Read(off) != int(OpDoCSSelect) {
		t.Fatalf("compileLiteral body[0] = %d, want OpDoCSSelect", ct.Read(off))
	}
	if ct.Read(off+1) != int(csDataStack) {
		t.Fatalf("compileLiteral body[1] = %d, want csDataStack", ct.Read(off+1))
	}
	if ct.Read(off+2) != int(OpDrop) {
		t.Fatalf("compileLiteral body[2] = %d, want OpDrop", ct.Read(off+2))
	}
	if ct.Read(off+3) != int(OpDoEmit) {
		t.Fatalf("compileLiteral body[3] = %d, want OpDoEmit", ct.Read(off+3))
	}
	if ct.Read(off+4) != int(xt) {
		t.Fatalf("compileLiteral body[4] = %d, want xt %d", ct.Read(off+4), int(xt))
	}

	cs := task.compilerForTarget()
	bootstrapCreate(cs, "litprobe")
	emitAt := cs.ct().Len()
	rp.DataStack.Push(Int(42))
	rp.DataStack.Push(TokenInt)
	if err := runWordInlineExec(task, rp, cl); err != nil {
		t.Fatal(err)
	}
	if ct.Read(emitAt) != int(OpCompileLit) {
		t.Fatalf("CT[%d] = %d, want OpCompileLit", emitAt, ct.Read(emitAt))
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("ds len %d after compileLiteral, want 1 (value pending for literal,)", rp.DataStack.Len())
	}
	if err := runWordInlineExec(task, rp, litWord); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("ds len %d after literal,, want 0", rp.DataStack.Len())
	}
	if ct.Read(emitAt+1) != int(OpDoLitI) {
		t.Fatalf("CT[%d] = %d, want OpDoLitI", emitAt+1, ct.Read(emitAt+1))
	}
	if ct.Read(emitAt+2) != 42 {
		t.Fatalf("CT[%d] = %d, want 42", emitAt+2, ct.Read(emitAt+2))
	}
}

func TestCompileLiteralEmitsNull(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cl := globalDictionaryContext.dict.GetByName("compileLiteral")
	if cl == nil {
		t.Fatal("missing compileLiteral")
	}
	litWord := globalDictionaryContext.dict.GetByName("literal,")
	if litWord == nil {
		t.Fatal("missing literal,")
	}
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "nullprobe")
	emitAt := cs.ct().Len()
	rp.DataStack.Push(Any(nil))
	rp.DataStack.Push(LiteralTokenType(TokenNull))
	if err := runWordInlineExec(task, rp, cl); err != nil {
		t.Fatal(err)
	}
	if ct := globalDictionaryContext.ct; ct.Read(emitAt) != int(OpCompileLit) {
		t.Fatalf("CT[%d] = %d, want OpCompileLit", emitAt, ct.Read(emitAt))
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("ds len %d after compileLiteral, want 1 (nil pending for literal,)", rp.DataStack.Len())
	}
	if err := runWordInlineExec(task, rp, litWord); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("ds len %d after literal,, want 0", rp.DataStack.Len())
	}
	ct := globalDictionaryContext.ct
	if ct.Read(emitAt+1) != int(OpDoLitNull) {
		t.Fatalf("CT[%d] = %d, want OpDoLitNull", emitAt+1, ct.Read(emitAt+1))
	}
}

func TestRunBreakUntilAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: early 0 begin 1 + dup 2 == if break then again ;
early
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("early = %v, want 2 (break when counter hits 2)", got)
	}
}

func TestRunBreakImmediateExitAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(": brk 0 begin break 999 until ; brk\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 0 {
		t.Fatalf("brk = %v, want 0", got)
	}
}

func TestRunBreakRepeatAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: brkrep 0 begin dup 3 < while 1 + dup 2 == if break then repeat ;
brkrep
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("brkrep = %v, want 2", got)
	}
}

func TestRunForeachBreakAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: fb 0 0 4 range foreach i 2 == if break then i + /foreach ;
fb
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 1 {
		t.Fatalf("fb = %v, want 1 (0+1, break at i==2)", got)
	}
}

func TestRunForeachContinueAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: fc 0 0 3 range foreach i 1 == if continue then i + /foreach ;
fc
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 5 {
		t.Fatalf("fc = %v, want 5 (0+2+3, skip 1)", got)
	}
}

func TestRunForeachNestedJAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: nestj 0 0 2 range foreach 0 2 range foreach j + /foreach /foreach ;
nestj
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	// 0 0 2 range → 0,1,2 (inclusive end, same as sum10's 0 0 4 → 0..4).
	// For each outer j, inner runs 3 times adding j: 0*3 + 1*3 + 2*3 = 9
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 9 {
		t.Fatalf("nestj = %v, want 9", got)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("RS len = %d, want 0", rp.ReturnStack.Len())
	}
}

func TestRunForeverBreakContinueAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: fe 0 forever 1 + dup 2 == if continue then dup 4 == if break then /forever ;
fe
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 4 {
		t.Fatalf("fe = %v, want 4", got)
	}
}

// Nested if around break must not steal the outer if's >mark (else a falsy outer
// if jumps to the forever exit). Regression for hello-led Esc check pattern.
func TestRunForeverNestedIfBreakAfterBootstrap(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: nestbrk 0 forever
    1 +
    dup 5 > if break then
    false if
      0 ==0 if
        break
      then
    then
  /forever ;
nestbrk
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 6 {
		t.Fatalf("nestbrk = %v, want 6 (must not exit on falsy outer if)", got)
	}
}

func TestCompileBreakBranchPatched(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewRunProvider()
	if err := compileBootstrapSource(tk, rp, `: brktest 0 begin break 999 until ;`); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("brktest")
	if w == nil {
		t.Fatal("missing brktest")
	}
	if err := runExecWord(tk, rp, w); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 0 {
		t.Fatalf("brktest = %v, want 0 (break skips 999 until body)", got)
	}
}

func TestRunTargetSwitchNoMatch(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": noopsw null switch 0 case 9 /case /switch ; 5 noopsw\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 5 {
		t.Fatalf("noopsw = %v, want 5 (unchanged, no default)", got)
	}
}

func TestRunTargetSwitchDefaultOnlyAlwaysRuns(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: onlydef switch default 42 /default /switch ;
1 onlydef
0 onlydef
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 42 {
			t.Fatalf("onlydef[%d] = %v, want 42 (default-only switch always runs)", i, got)
		}
	}
}

func TestRunTargetSwitchEmptyDropsTarget(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader(": emptysw null switch /switch ; 5 emptysw\n")),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 5 {
		t.Fatalf("emptysw = %v, want 5 (nil target dropped; 5 preserved)", got)
	}
}

func TestRunSwitchJsonParseValueShapeLeavesTrue(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: jpt true ;
: sim 116 switch a false == case "eof" /case a 116 == case jpt /case /switch ;
sim
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got != Bool(true) {
		t.Fatalf("sim = %v, want true", got)
	}
}

func TestDisasmPick2Switch(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	src := `: pick2 switch a 1 == case 1 /case /switch ;`
	if err := compileBootstrapSource(tk, rp, src); err != nil {
		t.Fatal(err)
	}
	w := globalDictionaryContext.dict.GetByName("pick2")
	ct := globalDictionaryContext.ct
	off := int(w.colonCode)
	var lines []string
	for off < ct.Len() {
		op := OpCode(ct.Read(off))
		if op == OpStop {
			break
		}
		if op == OpDoLitI || op == OpDoLitB || op == OpDoLit0 {
			lines = append(lines, fmt.Sprintf("%d: %s %d", off, OpCodeName(op), ct.Read(off+1)))
			off += 2
			continue
		}
		if op == OpBranch || op == OpBranchQ {
			lines = append(lines, fmt.Sprintf("%d: %s %d", off, OpCodeName(op), ct.Read(off+1)))
			off += 2
			continue
		}
		lines = append(lines, fmt.Sprintf("%d: %s", off, OpCodeName(op)))
		off++
	}
	t.Log(strings.Join(lines, "\n"))
}

func TestJsonParseTrueLeavesValueOnStack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	// 0042-lex.fnl and 0046-parse.fnl are part of the production bootstrap
	// (see bootstrapTranche2Manifest), so json> is already defined.
	rp.DataStack.Push(String("true"))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("json>")); err != nil {
		t.Fatalf("json>: %v", err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1; stack = %#v", rp.DataStack.Len(), rp.DataStack)
	}
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got != true {
		t.Fatalf("json> true = %v (%T), want true", got, got)
	}
}

func TestJsonParseStringLeavesValueOnStack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	// 0042-lex.fnl and 0046-parse.fnl are part of the production bootstrap
	// (see bootstrapTranche2Manifest), so json> is already defined.
	rp.DataStack.Push(String(`"49"`))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("json>")); err != nil {
		t.Fatalf("json>: %v", err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1; stack = %#v", rp.DataStack.Len(), rp.DataStack)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "49" {
		t.Fatalf("json> string = %q (%T), want 49", got, got)
	}
}

func TestJsonParseObjectOneMember(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	// 0042-lex.fnl and 0046-parse.fnl are part of the production bootstrap
	// (see bootstrapTranche2Manifest), so json> is already defined.
	rp.DataStack.Push(String(`{"fortynine":49}`))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("json>")); err != nil {
		t.Fatalf("json>: %v", err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	obj, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("json> = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	if got := obj.Get("fortynine"); got != Int(49) {
		t.Fatalf("fortynine = %v (%T), want Int(49)", got, got)
	}
}

func requireJsonParseFromBootstrap(t *testing.T) {
	t.Helper()
	if globalDictionaryContext.dict.GetByName("json>") == nil {
		t.Fatal("missing json> (bootstrap core2 parse.fnl?)")
	}
}

func TestJsonParsePrettyMultiMemberObject(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	requireJsonParseFromBootstrap(t)
	const pretty = `{
  "Accept-Encoding": [
    "gzip"
  ],
  "Host": [
    "httpbingo.org"
  ]
}`
	rp.DataStack.Push(String(pretty))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("json>")); err != nil {
		t.Fatalf("json>: %v", err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	obj, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("json> = %T, want *DynamicObject", obj)
	}
	ae, ok := obj.Get("Accept-Encoding").(*List[any])
	if !ok || ae.Len() != 1 || ae.AtAsAny(0) != String("gzip") {
		t.Fatalf("Accept-Encoding = %#v, want [gzip]", obj.Get("Accept-Encoding"))
	}
	host, ok := obj.Get("Host").(*List[any])
	if !ok || host.Len() != 1 || host.AtAsAny(0) != String("httpbingo.org") {
		t.Fatalf("Host = %#v, want [httpbingo.org]", obj.Get("Host"))
	}
}

func TestJsonParseHttpBingoHeadersShape(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	requireJsonParseFromBootstrap(t)
	const headers = `{
  "Accept-Encoding": [
    "gzip"
  ],
  "Host": [
    "httpbingo.org"
  ],
  "User-Agent": [
    "Go-http-client/2.0"
  ],
  "Via": [
    "2 fly.io"
  ]
}`
	rp.DataStack.Push(String(headers))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("json>")); err != nil {
		t.Fatalf("json>: %v", err)
	}
	obj, ok := rp.DataStack.Pop().(*DynamicObject)
	if !ok {
		t.Fatalf("json> = %T, want *DynamicObject", rp.DataStack.Peek())
	}
	if obj.Len() != 4 {
		t.Fatalf("headers len = %d, want 4", obj.Len())
	}
	if obj.Get("Via") == nil {
		t.Fatal("missing Via member")
	}
}

func TestRunSwitchJsonParseValueShapeLeavesTrueInline(t *testing.T) {
	ResetTestRuntime()
	tk := NewTestTask()
	rp := NewTestRunProvider()
	p := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r: bufio.NewReader(strings.NewReader(`
: sim 116 switch a false == case "eof" /case a 116 == case true /case /switch ;
sim
`)),
		label: "test",
	}})
	rp.SetParser(p)
	if err := interpretSession(tk, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("ds len = %d, want 1", rp.DataStack.Len())
	}
	got, ok := rp.DataStack.Pop().(Bool)
	if !ok || got != true {
		t.Fatalf("sim = %v, want true", got)
	}
}
