package core

import "testing"

func TestCompileStateOnCompilerState(t *testing.T) {
	cs := NewCompilerState(globalDictionaryContext)
	SetCompileState(cs, true)
	if !GetCompileState(cs) {
		t.Fatal("STATE should be true")
	}
	SetCompileState(cs, false)
	if GetCompileState(cs) {
		t.Fatal("STATE should be false")
	}
}

func TestResetTestRuntimeClearsGlobalDictNotCompilerStateField(t *testing.T) {
	cs := NewCompilerState(globalDictionaryContext)
	SetCompileState(cs, true)
	ResetTestRuntime()
	if globalDictionaryContext.dt.Len() < ReservedDTSlots {
		t.Fatalf("dt len = %d, want >= %d reserved slots", globalDictionaryContext.dt.Len(), ReservedDTSlots)
	}
	// CompilerState is ephemeral; a fresh cs after reset starts false.
	fresh := NewCompilerState(globalDictionaryContext)
	if GetCompileState(fresh) {
		t.Fatal("fresh CompilerState should have STATE false")
	}
}

func TestCompilerStateColonUsesCompileState(t *testing.T) {
	ResetTestRuntime()
	cs := NewCompilerState(globalDictionaryContext)
	if GetCompileState(cs) {
		t.Fatal("fresh runtime")
	}
	bootstrapCreate(cs, "foo")
	if !GetCompileState(cs) {
		t.Fatal(": should set STATE true")
	}
	if globalDictionaryContext.dict.GetByName("foo") != nil {
		t.Fatal("word should not be in dictionary until ;")
	}
	if w := globalDictionaryContext.dict.PeekBuilding(); w == nil || w.name != "foo" {
		t.Fatal("building word should be on building stack at :")
	}
	bootstrapFinalizeColonWord(cs)
	if globalDictionaryContext.dict.GetByName("foo") == nil {
		t.Fatal("word should be in dictionary after ;")
	}
	if GetCompileState(cs) {
		t.Fatal("; should set STATE false")
	}
}

func TestCompilerStateTransactionRollback(t *testing.T) {
	ResetTestRuntime()
	cs := NewCompilerState(globalDictionaryContext)
	cs.beginTransaction()
	bootstrapCreate(cs, "foo")
	cs.rollbackTransaction()
	if GetCompileState(cs) {
		t.Fatal("rollback should clear STATE")
	}
	if globalDictionaryContext.dict.GetByName("foo") != nil {
		t.Fatal("rollback should remove optimistic word")
	}
}
