package core

import (
	"strings"
	"testing"
)

func TestExpandCompoundDotNameRecursiveTokens(t *testing.T) {
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name"})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	assertDSNameValue(t, rp, "hello")
}

func TestExpandCompoundDotNameFetch(t *testing.T) {
	task, rp, p := testParserWithInput(t, "")
	w := NewWord("hello", 0)
	pushWordOnDS(t, rp, w)
	segs := CompoundPathForTest([]string{".", "name", "@"})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	assertDSNameValue(t, rp, "hello")
}

func TestParseCompoundFusedAffixes(t *testing.T) {
	tokens := parseLineTokens(t, "myObj.field@{}")
	expectCompoundPath(t, tokens[0].Value, "myObj", ".", "field", "@{}")
}

func TestExpandCompoundToOS(t *testing.T) {
	t.Skip(">os removed; use t: prefix for temp-stack compound paths")
}

func TestParseCompoundGtOsPath(t *testing.T) {
	_, err := splitCompoundSegments("find>os.name@")
	if err == nil {
		t.Fatal("want error for find>os compound (use t: prefix)")
	}
}

func TestExpandCompoundFindGtOsNameFetch(t *testing.T) {
	t.Skip(">os removed; use t: prefix for temp-stack compound paths")
}

func TestExpandCompoundDictionaryContextFindWordByNameCaret(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	dc := globalDictionaryContext
	rp.DataStack.Push(String("dup"))
	segs := CompoundPathForTest([]string{"dc", ".", "findWordByName", "^"})
	if err := dispatchCompoundToken(p, task, rp, segs); err != nil {
		t.Fatal(err)
	}
	if !rp.DataStack.Pop().(Bool) {
		t.Fatal("FindWordByName should return true")
	}
	if rp.DataStack.Pop().(*Word) != dc.dict.GetByName("dup") {
		t.Fatal("FindWordByName should return dup word")
	}
}

func TestExpandCompoundDictionaryContextFindWordByNameBareInvoke(t *testing.T) {
	ResetTestRuntime()
	task, rp, p := testParserWithInput(t, "")
	dc := globalDictionaryContext
	rp.DataStack.Push(String("dup"))
	rp.DataStack.Push(dc)
	segs := CompoundPathForTest([]string{".", "findWordByName"})
	err := dispatchCompoundToken(p, task, rp, segs)
	if err == nil {
		t.Fatal("expected error for bare .method without ^")
	}
	if !strings.Contains(err.Error(), `"^"`) {
		t.Fatalf("err = %v", err)
	}
}
