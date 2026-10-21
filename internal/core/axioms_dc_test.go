package core

import "testing"

func TestOpGetDictionaryContextPushesTaskDC(t *testing.T) {
	InitTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	// CT cells are in core (ModuleIDThis), so execDC would be core — dc must still push task.dc.
	RunOps(task, rp, OpGetDictionaryContext, OpStop)
	got, ok := rp.DataStack.Pop().(*DictionaryContext)
	if !ok || got != task.dc {
		t.Fatalf("dc = %v, want task.dc %v", got, task.dc)
	}
	if got == globalDictionaryContext {
		t.Fatal("local task dc must not be core")
	}
}

func TestOpGetDictionaryContextCoreTask(t *testing.T) {
	InitTestRuntime()
	task := CreateCoreTask()
	rp := NewTestRunProvider()
	RunOps(task, rp, OpGetDictionaryContext, OpStop)
	got, ok := rp.DataStack.Pop().(*DictionaryContext)
	if !ok || got != globalDictionaryContext {
		t.Fatalf("dc = %v, want core", got)
	}
}
