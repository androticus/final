package core

import "testing"

func collectUserWordNames(it Iterator) []string {
	var names []string
	for it.GetNext() {
		names = append(names, it.Current().(*Word).name)
	}
	return names
}

func TestWordIsUserVisible(t *testing.T) {
	plain := NewWord("plain", WordFlagColon)
	hidden := NewWord("hidden", WordFlagHidden|WordFlagColon)
	system := NewWord("system", WordFlagSystem|WordFlagColon)
	crypto := NewWord("crypto", WordFlagColon)
	crypto.srFlags = uint32(SR_Crypto)

	if !plain.IsUserVisible(nil) {
		t.Fatal("plain word should be user-visible with nil Task")
	}
	if hidden.IsUserVisible(nil) {
		t.Fatal("Hidden word should not be user-visible")
	}
	if system.IsUserVisible(nil) {
		t.Fatal("System word should not be user-visible")
	}
	if !crypto.IsUserVisible(nil) {
		t.Fatal("SR-flagged word should be visible when Task is unrestricted")
	}

	task := NewTestTask()
	if !crypto.IsUserVisible(task) {
		t.Fatal("SR-flagged word should be visible when Task.restrict is 0")
	}
	task.restrict = SR_Crypto
	if crypto.IsUserVisible(task) {
		t.Fatal("SR-flagged word should not be user-visible when Task restricts that SR")
	}
	if !plain.IsUserVisible(task) {
		t.Fatal("plain word should stay visible under unrelated SR restrict")
	}
}

func TestDictionaryGetUserWords(t *testing.T) {
	d := NewDictionary("userWords", 16, 8)
	plain := NewWord("plain", WordFlagColon)
	hidden := NewWord("hidden", WordFlagHidden|WordFlagColon)
	system := NewWord("system", WordFlagSystem|WordFlagColon)
	crypto := NewWord("crypto", WordFlagColon)
	crypto.srFlags = uint32(SR_Crypto)
	d.Add(plain)
	d.Add(hidden)
	d.Add(system)
	d.Add(crypto)

	task := NewTestTask()
	got := collectUserWordNames(d.GetUserWords(task))
	if len(got) != 2 || got[0] != "plain" || got[1] != "crypto" {
		t.Fatalf("unrestricted GetUserWords = %v, want [plain crypto]", got)
	}

	task.restrict = SR_Crypto
	got = collectUserWordNames(d.GetUserWords(task))
	if len(got) != 1 || got[0] != "plain" {
		t.Fatalf("SR_Crypto GetUserWords = %v, want [plain]", got)
	}

	it := d.GetUserWords(task)
	if !it.GetNext() {
		t.Fatal("expected GetNext true")
	}
	cur, ok := it.Current().(*Word)
	if !ok || cur != plain {
		t.Fatalf("Current = %v, want *Word plain", it.Current())
	}
}

func TestDictionaryGetUserWordsSnapshotIgnoresLaterAdd(t *testing.T) {
	d := NewDictionary("userWordsSnap", 16, 8)
	d.Add(NewWord("a", WordFlagColon))
	it := d.GetUserWords(NewTestTask())
	d.Add(NewWord("b", WordFlagColon))
	got := collectUserWordNames(it)
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("snapshot iter = %v, want [a]", got)
	}
}

func TestDictionaryBuildingStackColonCompile(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	cs := task.compilerForTarget()
	bootstrapCreate(cs, "buildme")
	if !globalDictionaryContext.dict.IsBuilding() {
		t.Fatal("expected building stack active")
	}
	if globalDictionaryContext.dict.GetByName("buildme") != nil {
		t.Fatal("building word should not be in dictionary until ;")
	}
	if globalDictionaryContext.dict.PeekBuilding().ordIdx >= 0 {
		t.Fatal("building word ordIdx should be unset until ;")
	}
	bootstrapFinalizeColonWord(cs)
	if globalDictionaryContext.dict.IsBuilding() {
		t.Fatal("building stack should be empty after ;")
	}
	if globalDictionaryContext.dict.GetByName("buildme") == nil {
		t.Fatal("word should be in dict after ;")
	}
}

func TestDictionaryPeekBuildingEmptyReturnsNil(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.PeekBuilding() != nil {
		t.Fatal("PeekBuilding on empty stack should return nil")
	}
}

func TestDictionaryTypeHasName(t *testing.T) {
	ResetTestRuntime()
	d := globalDictionaryContext.dict
	rp := NewTestRunProvider()
	rp.DataStack.Push(d)
	rp.DataStack.Push(String("dup"))
	mm := GetTypeRegistrar().Get("Dictionary").MethodByName("hasName")
	f, ths := memberRunFrame(rp, d)
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) != true {
		t.Fatal("dup should exist")
	}
	rp.DataStack.Push(String("nosuch"))
	mm.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) != false {
		t.Fatal("nosuch should not exist")
	}
}

func TestDictionaryTypeMakeNewWordAndPushBuilding(t *testing.T) {
	ResetTestRuntime()
	d := globalDictionaryContext.dict
	rp := NewTestRunProvider()
	rp.DataStack.Push(d)
	dictT := GetTypeRegistrar().Get("Dictionary")
	makeW := dictT.MethodByName("makeNewWord")
	pushB := dictT.MethodByName("pushBuilding")
	f, ths := memberRunFrame(rp, d)
	rp.DataStack.Push(Int(WordFlagColon))
	rp.DataStack.Push(String("fromfinal"))
	makeW.invokeCall(f, ths)
	w := rp.DataStack.Pop().(*Word)
	if w == nil || w.name != "fromfinal" || !w.IsColon() {
		t.Fatalf("MakeNewWord = %+v", w)
	}
	rp.DataStack.Push(w)
	pushB.invokeCall(f, ths)
	if d.PeekBuilding() != w {
		t.Fatal("PushBuilding should place word on building stack")
	}
	d.ClearBuilding()
}

func TestDictionaryTypeItemsGetByKey(t *testing.T) {
	ResetTestRuntime()
	d := globalDictionaryContext.dict
	dupW := d.GetByName("dup")
	if dupW == nil {
		t.Fatal("missing dup")
	}
	rp := NewTestRunProvider()
	rp.DataStack.Push(d)
	rp.DataStack.Push(String("dup"))
	RunOpsWithSubCode(NewTestTask(), rp, int(TypeDispatchKindInstanceItemGetByKey), OpTypeMemberDispatch, OpStop)
	got := rp.DataStack.Pop().(*Word)
	if got != dupW {
		t.Fatalf("itemsGetByKey = %v, want dup", got)
	}
}

func TestOpSetWordFlagsFlagsBuildingWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	cs := task.compilerForTarget()
	w := bootstrapCreate(cs, "immprobe")
	rp.DataStack.Push(Int(WordFlagImmediate))
	if err := runWordInlineExec(task, rp, globalDictionaryContext.dict.GetByName("setWordFlags")); err != nil {
		t.Fatal(err)
	}
	if !w.IsImmediate() {
		t.Fatal("expected WordFlagImmediate on building word")
	}
	if globalDictionaryContext.dict.GetByName("immprobe") != nil {
		t.Fatal("building word should not be in dictionary until ;")
	}
	if globalDictionaryContext.dict.PeekBuilding().ordIdx >= 0 {
		t.Fatal("building word ordIdx should be unset until ;")
	}
	cs.DC.dict.ClearBuilding()
	SetCompileState(cs, false)
}

func TestDictionaryContextTypePropertiesAndFindWordByName(t *testing.T) {
	ResetTestRuntime()
	dc := globalDictionaryContext
	dcT := GetTypeRegistrar().Get("DictionaryContext")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dc)

	dictProp := dcT.PropertyByName("dict")
	dictProp.invokeGet(f, ths)
	if rp.DataStack.Pop().(*Dictionary) != dc.dict {
		t.Fatal("dict property should return bundled dictionary")
	}

	find := dcT.MethodByName("findWordByName")
	rp.DataStack.Push(String("dup"))
	find.invokeCall(f, ths)
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("dup should be found")
	}
	if rp.DataStack.Pop().(*Word) != dc.dict.GetByName("dup") {
		t.Fatal("FindWordByName should return dup word")
	}

	rp.DataStack.Push(String("nosuchsymbol"))
	find.invokeCall(f, ths)
	if rp.DataStack.Pop().(Bool) {
		t.Fatal("missing name should return false only")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatal("not-found should not push an item")
	}
}
