package core

import (
	"fmt"
	"testing"
)

func TestNewWordFactory(t *testing.T) {
	w := NewWord("probe", WordFlagColon|WordFlagDefined)
	if w == nil {
		t.Fatal("NewWord returned nil")
	}
	if w.name != "probe" {
		t.Fatalf("name = %q, want probe", w.name)
	}
	if w.ordIdx != -1 {
		t.Fatalf("ordIdx = %d, want -1 (not in dict)", w.ordIdx)
	}
	if !w.IsColon() || !w.IsDefined() {
		t.Fatalf("flags = %#x, want Colon|Defined", w.flags)
	}
}

func TestWordNewClassMethodInterpret(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	const wordName = "factoryWord"
	flags := int(WordFlagColon | WordFlagDefined)
	runInterpretScript(t, task, rp, fmt.Sprintf("%d\n\"%s\"\nWord.new^\n", flags, wordName))
	got, ok := rp.DataStack.Pop().(*Word)
	if !ok || got == nil {
		t.Fatalf("TOS = %v, want *Word", rp.DataStack.Peek())
	}
	if got.name != wordName {
		t.Fatalf("name = %q, want %q", got.name, wordName)
	}
	if got.ordIdx != -1 {
		t.Fatalf("ordIdx = %d, want -1 (not in dict)", got.ordIdx)
	}
	if !got.IsColon() || !got.IsDefined() {
		t.Fatalf("flags = %#x, want Colon|Defined", got.flags)
	}
	if globalDictionaryContext.dict.HasName(wordName) {
		t.Fatalf("%q should not be in dictionary after Word.new", wordName)
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
}
