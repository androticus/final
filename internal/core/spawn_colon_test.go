package core

import (
	"strings"
	"testing"
	"time"
)

func TestSpawnColonHelloStdout(t *testing.T) {
	ResetTestRuntime()
	parent := CreateLocalTask()
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.Stdout = sb
	src := `
: helloPip drop "Hello from spawn:!\n" $ ;
: kick Pip.newPipParams^ spawn: helloPip ;
`
	if err := compileBootstrapSource(parent, rp, src); err != nil {
		t.Fatal(err)
	}
	if n := rp.DataStack.Len(); n != 0 {
		t.Fatalf("DS len = %d after compile with spawn:, want 0 (assertSpawnSuitable must not leak Words)", n)
	}
	runInterpretScript(t, parent, rp, "kick\n")
	deadline := time.Now().Add(2 * time.Second)
	for parent.descendantCount.Load() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for spawned pip")
		}
		time.Sleep(time.Millisecond)
	}
	if got := sb.ToString(); got != "Hello from spawn:!\n" {
		t.Fatalf("stdout = %q, want %q", got, "Hello from spawn:!\n")
	}
}

func TestSpawnColonOnLocalSessionWithoutSystemMode(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.task.systemMode {
		t.Fatal("local Session must not have systemMode")
	}
	sb := NewStringBuilder()
	s.rp.Stdout = sb
	runInterpretScript(t, s.task, s.rp, `
: helloPip drop "Hello from spawn:!\n" $ ;
: kick Pip.newPipParams^ spawn: helloPip ;
kick
`)
	deadline := time.Now().Add(2 * time.Second)
	for s.task.descendantCount.Load() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for spawned pip")
		}
		time.Sleep(time.Millisecond)
	}
	if got := sb.ToString(); got != "Hello from spawn:!\n" {
		t.Fatalf("stdout = %q, want %q", got, "Hello from spawn:!\n")
	}
}

func TestSpawnColonRejectsAxiom(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad Pip.newPipParams^ spawn: + ;`)
	if err == nil {
		t.Fatal("expected compile error spawning axiom")
	}
	if !strings.Contains(err.Error(), "not a colon word") {
		t.Fatalf("err = %v, want not a colon word", err)
	}
}

func TestSpawnColonRejectsImmediate(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad Pip.newPipParams^ spawn: if ;`)
	if err == nil {
		t.Fatal("expected compile error spawning immediate")
	}
	if !strings.Contains(err.Error(), "immediate") {
		t.Fatalf("err = %v, want immediate", err)
	}
}

func TestSpawnColonMissingName(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad Pip.newPipParams^ spawn: nosuchSpawnColonWord ;`)
	if err == nil {
		t.Fatal("expected compile error for missing name")
	}
	if !strings.Contains(err.Error(), "word not found") {
		t.Fatalf("err = %v, want word not found", err)
	}
}

func TestSpawnColonIsCompileOnlyImmediate(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.dict.GetByName("spawn:")
	if w == nil {
		t.Fatal("missing spawn:")
	}
	if !w.IsImmediate() {
		t.Fatal("spawn: should be Immediate")
	}
	if w.flags&WordFlagCompileOnly == 0 {
		t.Fatal("spawn: should be CompileOnly")
	}
}

func TestAssertSpawnSuitableRejectsHidden(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapSource(task, rp, `: bad "_bootWarmRepl" find assertSpawnSuitable drop ;`)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := NewTextReaderFromString("bad\n")
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	err = interpretSession(task, rp)
	if err == nil {
		t.Fatal("expected error for hidden colon word")
	}
	if !strings.Contains(err.Error(), "hidden") {
		t.Fatalf("err = %v, want hidden", err)
	}
}
