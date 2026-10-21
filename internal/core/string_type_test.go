package core

import (
	"testing"
)

func TestByteListAsString_StringAsBytes_roundtrip(t *testing.T) {
	want := "café {\"a\": 1}"
	bytes := NewByteListFromString(want)
	if bytes.Len() != len(want) {
		t.Fatalf("asBytes len = %d, want %d", bytes.Len(), len(want))
	}
	got := String(string(bytes.items))
	if got != String(want) {
		t.Fatalf("asString = %q, want %q", got, want)
	}
	// Mutual independence: mutating list must not change original string bytes path
	bytes.items[0] = 'X'
	again := NewByteListFromString(want)
	if again.items[0] != want[0] {
		t.Fatal("NewByteListFromString must copy; shared backing with string conversion path")
	}
}


func TestStringStartsEndsWith(t *testing.T) {
	s := String("internal/final/core/0040-json/0046-parse.fnl")
	if !stringEndsWith(s, ".fnl") {
		t.Fatal("expected endsWith .fnl")
	}
	if stringEndsWith(s, ".json") {
		t.Fatal("expected not endsWith .json")
	}
	if !stringStartsWith(s, "internal/") {
		t.Fatal("expected startsWith internal/")
	}
}

func TestStringSplitLines(t *testing.T) {
	list := stringSplitLines("a\nb\n")
	if list.Len() != 2 || list.At(0) != "a" || list.At(1) != "b" {
		t.Fatalf("splitLines = %#v", list.items)
	}
}

func TestTextReaderReadLine(t *testing.T) {
	r, err := NewTextReaderFromString("line1\nline2\n")
	if err != nil {
		t.Fatal(err)
	}
	line, eof := r.ReadLine()
	if eof || line != "line1" {
		t.Fatalf("line1 = %q eof=%v", line, eof)
	}
	line, eof = r.ReadLine()
	if eof || line != "line2" {
		t.Fatalf("line2 = %q eof=%v", line, eof)
	}
	_, eof = r.ReadLine()
	if !eof {
		t.Fatal("expected eof")
	}
}

func TestInterpretStringEndsWithTrimSpace(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t, `"  x.fnl  " str::trim ".fnl" str::endsWith`+"\n")
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
	if rp.DataStack.Peek() != Bool(true) {
		t.Fatalf("endsWith = %v, want true", rp.DataStack.Peek())
	}
}

func TestCompileBootstrapFileStringHelpersFinalSource(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	path := writeTempFinalSource(t, ": strtest \"  x  \" str::trim drop ;\n")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("strtest") == nil {
		t.Fatal("strtest should be defined")
	}
}
