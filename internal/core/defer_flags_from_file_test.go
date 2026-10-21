package core

import "testing"

func TestDeferColonInstanceFlags(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	if globalDictionaryContext.dict.GetByName("defer:") == nil {
		t.Fatal("defer: missing from bootstrap core1")
	}
	runInterpretScript(t, task, rp, "defer: brad\n")
	inst := globalDictionaryContext.dict.GetByName("brad")
	if inst == nil {
		t.Fatal("brad missing")
	}
	want := WordFlagDefinedWord | WordFlagDefer
	if inst.flags != want {
		t.Fatalf("brad.flags = %#x, want %#x", inst.flags, want)
	}
	if inst.dtIdx < 0 {
		t.Fatalf("brad.dtIdx = %d, want allotted slot", inst.dtIdx)
	}
	if inst.colonCode < 0 {
		t.Fatal("brad.colonCode unset")
	}
}
