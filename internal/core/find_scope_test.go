package core

import "testing"

func TestFindScopeGlobalBase(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	if rp.FindScope == nil {
		t.Fatal("NewTestRunProvider: nil FindScope")
	}
	w, _, ok := FindNameInRun(rp, "dup")
	if !ok || w == nil || w.name != "dup" {
		t.Fatalf("FindNameInRun(dup) = %v ok=%v", w, ok)
	}
}

func TestFindScopeLocalBeforeGlobal(t *testing.T) {
	ResetTestRuntime()
	localDC := NewDictionaryContext("local", 64, 0, 64, 0, 64, 0, GetTypeRegistrar())
	localW := NewWord("localOnly", WordFlagColon)
	localW.colonCode = CTP(0)
	localDC.dict.Add(localW)

	rp := NewTestRunProvider()
	rp.FindScope.PushScope(localDC)

	w, _, ok := FindNameInRun(rp, "localOnly")
	if !ok || w != localW {
		t.Fatalf("local hit = %v ok=%v", w, ok)
	}
	_, _, ok = FindNameInRun(rp, "dup")
	if !ok {
		t.Fatal("global word should resolve through base layer")
	}
}

func TestWithTaskFindScopeUsesGlobalWhenTaskIsGlobal(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	if rp.FindScope.Scopes.Len() != 1 {
		t.Fatalf("scope len = %d, want 1 (global only)", rp.FindScope.Scopes.Len())
	}
}

func TestWithTaskFindScopePushesLocalDC(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()

	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	if rp.FindScope.Scopes.Len() != 2 {
		t.Fatalf("scope len = %d, want 2 (local + core from CSO)", rp.FindScope.Scopes.Len())
	}
}

func TestBootstrapFindReturnsWord(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push("dup")
	f := &RunFrame{Task: NewTestTask(), RP: rp}
	bootstrapFinalFind(f)
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatal("want true flag for dup")
	}
	w, ok := rp.DataStack.Pop().(*Word)
	if !ok || w == nil || w.name != "dup" {
		t.Fatalf("find dup = %T %v", w, w)
	}
}

func TestTypeCatalogAssignsTypeIdx(t *testing.T) {
	cat := NewTypeCatalog()
	a := &Type{name: "Alpha"}
	b := &Type{name: "Beta"}
	cat.Add("Alpha", a)
	cat.Add("Beta", b)
	if a.typeIdx != 0 || b.typeIdx != 1 {
		t.Fatalf("typeIdx = %d, %d; want 0, 1", a.typeIdx, b.typeIdx)
	}
}
