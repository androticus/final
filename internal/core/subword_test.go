package core

import (
	"testing"
)

func TestOpBrSubRetSub(t *testing.T) {
	t.Run("call_and_return", func(t *testing.T) {
		ResetTestRuntime()
		rp := NewTestRunProvider()
		ct := &VirtCTResolver{}
		_ = rp.ReturnStack.PushNew(0)

		// brsub → sub (lit 42; retsub) → stop
		brOp := ct.Append(ModuleIDThis, int(OpBrSub))
		jSlot := ct.Append(ModuleIDThis, 0)
		stopAt := ct.Append(ModuleIDThis, int(OpStop))
		subEntry := ct.Append(ModuleIDThis, int(OpDoLitI))
		ct.Append(ModuleIDThis, 42)
		ct.Append(ModuleIDThis, int(OpRetSub))

		j := int(subEntry - (jSlot + 1))
		ct.Write(jSlot, j)

		task := NewTestTask()
		task.ip = brOp
		Run(task, rp)
		got, ok := rp.DataStack.Pop().(Int)
		if !ok || got != 42 {
			t.Fatalf("got %v, want Int(42)", got)
		}
		if len(rp.ReturnStack.Peek().subRets) != 0 {
			t.Fatalf("subRets not empty after retsub: %v", rp.ReturnStack.Peek().subRets)
		}
		_ = stopAt
	})
}

func TestCompileSubwordBasic(t *testing.T) {
	mustCompileColonDef(t, `
: sw_basic
  :s helper
    1 +
  ;s
  10 helper
;
`)
}

func TestCompileSubwordTwoAndCallEarlier(t *testing.T) {
	mustCompileColonDef(t, `
: sw_two
  :s a
    1
  ;s
  :s b
    a 1 +
  ;s
  b
;
`)
}

func TestRunSubwordIntegration(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: sw_run
  :s init
    0
  ;s
  :s bump
    1 +
  ;s
  init
  bump
  bump
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "sw_run")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 2 {
		t.Fatalf("got %v, want Int(2)", got)
	}
}

func TestRunSubwordSharedLocals(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := `
: sw_loc
  :s setx
    7 l.x!
  ;s
  :s getx
    l.x@
  ;s
  setx
  getx
;
`
	if err := compileBootstrapSource(task, rp, src); err != nil {
		t.Fatal(err)
	}
	runCompiledWord(t, task, rp, "sw_loc")
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 7 {
		t.Fatalf("got %v, want Int(7)", got)
	}
}

func TestCompileSubwordPanics(t *testing.T) {
	cases := []struct {
		name, src, wantSubstr string
	}{
		{"semicolon with open subword", `: bad :s foo 1 ;`, semicolonBlockCheckMsg},
		{"nested subword", `: bad :s a :s b 1 ;s ;s ;`, subwordNestMsg},
		{"duplicate name", `: bad :s foo 1 ;s :s foo 2 ;s ;`, "duplicate subword name"},
		{"collides with parent", `: bad :s bad 1 ;s ;`, "collides with the word being defined"},
		{"after main body", `: bad 1 :s foo 2 ;s ;`, subwordLateMsg},
		{"open name use", `: bad :s foo foo ;s ;`, "not yet closed"},
		{"try inside subword", `: bad :s foo try 1 catch 0 /try ;s ;`, "try is not allowed inside a subword"},
		{";s without :s", `: bad ;s ;`, subwordCloseMsg},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectCompileColonDefError(t, tc.src, tc.wantSubstr)
		})
	}
}
