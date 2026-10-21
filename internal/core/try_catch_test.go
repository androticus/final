package core

import (
	"strings"
	"testing"
)

func TestCompileTryCatchTry(t *testing.T) {
	mustCompileColonDef(t, `: tct try 1 catch 0 /try ;`)
}

func TestCompileTryCatchFinallyTry(t *testing.T) {
	mustCompileColonDef(t, `: tcft try 1 catch 0 finally 2 /try ;`)
}

func TestCompileCatchRequiresTry(t *testing.T) {
	expectCompileColonDefError(t, `: bad catch 0 /try ;`, "'catch' requires a preceding 'try'")
}

func TestCompileFinallyRequiresCatch(t *testing.T) {
	expectCompileColonDefError(t, `: bad try 1 finally 2 /try ;`, "'finally' requires a preceding 'catch'")
}

func TestCompileEndTryRequiresCatch(t *testing.T) {
	expectCompileColonDefError(t, `: bad try 1 /try ;`, "'/try' requires a preceding 'catch'")
}

func TestCompileEndTryRequiresCatchNotBare(t *testing.T) {
	expectCompileColonDefError(t, `: bad /try ;`, "'/try' requires a preceding 'catch'")
}

func TestCompileTryCrossIfRejected(t *testing.T) {
	expectCompileColonDefError(t, `: bad if try 1 catch 0 then /try ;`, "'then' requires a preceding 'if'")
}

func TestTryCatchHappyPathFinallyRuns(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  0 >ts
  try
    ts> 1 + >ts
    10
  catch
    99
  finally
    ts> 1 + >ts
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	if got := rp.DataStack.Pop().(Int); got != 10 {
		t.Fatalf("result = %v, want 10 (catch skipped)", got)
	}
	if got := rp.TempStack.Pop().(Int); got != 2 {
		t.Fatalf("finally counter = %v, want 2", got)
	}
	if rp.ReturnStack.Len() != 0 {
		t.Fatalf("RS len = %d, want 0 (CatchMark on CFM, not RS)", rp.ReturnStack.Len())
	}
}

func TestTryCatchThrowRestoresAndCatchRuns(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: boom "fail" throw ;
: probe
  7
  try
    1 2 3
    boom
    99
  catch
    // restored to [7] then msg; leave 42 only
    drop drop
    42
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	if got := rp.DataStack.Pop().(Int); got != 42 {
		t.Fatalf("result = %v, want 42", got)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS leftover len = %d", rp.DataStack.Len())
	}
}

func TestTryCatchThrowMessageOnCatchEntry(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  try
    "oops" throw
  catch
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "oops" {
		t.Fatalf("msg = %v, want String oops", got)
	}
}

func TestTryCatchNestedColonThrow(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: deep "inner" throw ;
: mid deep ;
: probe
  try
    mid
    false
  catch
    true
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	if got := rp.DataStack.Pop().(Bool); !bool(got) {
		t.Fatal("want true from catch")
	}
	msg, ok := rp.DataStack.Pop().(String)
	if !ok || !strings.Contains(string(msg), "inner") {
		t.Fatalf("msg = %v", msg)
	}
}

func TestThrowJoinOuterFirst(t *testing.T) {
	got := "outer" + throwCausedByPrefix + "inner"
	if !strings.HasPrefix(got, "outer") || !strings.Contains(got, "caused by: inner") {
		t.Fatalf("join format = %q", got)
	}
}

func TestThrowArmsPendingOpThrow(t *testing.T) {
	forceOp := OpCode(0)
	task := NewTestTask()
	task.forceOp = &forceOp
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp, ForceOp: &forceOp}

	Throw("native boom", f)

	if forceOp != OpThrow {
		t.Fatalf("ForceOp = %v, want OpThrow", forceOp)
	}
	if !task.pendingThrow {
		t.Fatal("want pendingThrow")
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "native boom" {
		t.Fatalf("DS msg = %v, want native boom", got)
	}
}

func TestThrowNestedJoinsOuterFirst(t *testing.T) {
	forceOp := OpCode(0)
	task := NewTestTask()
	task.forceOp = &forceOp
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp, ForceOp: &forceOp}

	Throw("inner detail", f)
	Throw("outer failure", f)

	if forceOp != OpThrow {
		t.Fatalf("ForceOp = %v, want OpThrow still armed", forceOp)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok {
		t.Fatalf("DS TOS = %T, want String", rp.DataStack.Peek())
	}
	want := "outer failure" + throwCausedByPrefix + "inner detail"
	if string(got) != want {
		t.Fatalf("joined = %q, want %q", got, want)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS leftover len = %d (want single joined msg)", rp.DataStack.Len())
	}
}

func TestNativeThrowProbeBoomCaught(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  try
    ThrowProbe.boom^
    false
  catch
    true
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	if got := rp.DataStack.Pop().(Bool); !bool(got) {
		t.Fatal("want true from catch")
	}
	msg, ok := rp.DataStack.Pop().(String)
	if !ok || msg != "native boom" {
		t.Fatalf("msg = %v, want native boom", msg)
	}
}

func TestTryCatchForeachJInsideTry(t *testing.T) {
	// nesting-snafu try arm: j inside try must see outer foreach (CatchMark not on RS)
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: snafuTry
  0
  0 0 2 range foreach
    0 0 0 range foreach
      try
        j +
      catch
        drop
      /try
    /foreach
  /foreach
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "snafuTry")
	// 0 0 2 range → 0,1; add j each → 0+1 = 2
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("snafuTry = %v, want 2", got)
	}
}

func TestNativeThrowProbeBoomJoinCaught(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  try
    ThrowProbe.boomJoin^
  catch
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	msg, ok := rp.DataStack.Pop().(String)
	if !ok {
		t.Fatalf("msg = %T, want String", rp.DataStack.Peek())
	}
	want := "outer failure" + throwCausedByPrefix + "inner detail"
	if string(msg) != want {
		t.Fatalf("msg = %q, want %q", msg, want)
	}
}

// Throw inside foreach must drop the orphan iter so a later foreach sees clean nesting.
func TestThrowCleansForeachInsideTry(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  0
  try
    0 9 range foreach
      "boom" throw
    /foreach
  catch
    drop
  /try
  0 2 range foreach
    i +
  /foreach
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 { // 0+1+2
		t.Fatalf("sum = %v, want 3", got)
	}
}

// Try inside foreach: outer iter must survive throw (itersLen >= 1 at try entry).
func TestThrowPreservesOuterForeach(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  0
  0 2 range foreach
    try
      i 1 == if "skip" throw then
      i +
    catch
      drop
    /try
  /foreach
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 { // 0+2; skip 1 via throw
		t.Fatalf("sum = %v, want 2", got)
	}
}

func TestThrowCleansSwitchInsideTry(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  try
    99 switch
      a 99 == case
        "boom" throw
      /case
    /switch
  catch
    drop
  /try
  7 switch
    a
  /switch
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("a = %v, want 7", got)
	}
}

func TestThrowCleansParenInsideTry(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  try
    ( 1 2
      "boom" throw
    )
  catch
    drop
  /try
  ( 10 20 )
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	lst, ok := rp.DataStack.Pop().(*List[any])
	if !ok || lst.Len() != 2 {
		t.Fatalf("got %v, want List len 2", rp.DataStack.Peek())
	}
	if lst.At(0) != Int(10) || lst.At(1) != Int(20) {
		t.Fatalf("list = %v, want [10 20]", lst)
	}
}

// Nested try: inner throw restores only to inner try-entry adjunct depths.
func TestThrowNestedTryInnerForeach(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: probe
  0
  try
    0 0 range foreach
      try
        0 9 range foreach
          "inner" throw
        /foreach
      catch
        drop
      /try
      // outer foreach still live; i is 0
      i +
    /foreach
  catch
    drop 99
  /try
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "probe")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 0 {
		t.Fatalf("got %v, want 0 (outer foreach completed after inner throw)", got)
	}
}
