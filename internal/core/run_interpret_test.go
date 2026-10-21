package core

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestNestParserPromotesFaultToParentOnError(t *testing.T) {
	outer := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("")),
		label: "outer",
	}})
	inner := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("bad\n")),
		label: "inner",
	}})
	inner.setParserLocation("/tmp/inner.fnl", 7, 4)
	inner.lineText = "  bad"
	inner.col = 3
	inner.line = 7

	rp := NewTestRunProvider()
	rp.SetParser(outer)

	err := nestParser(rp, inner, func() error {
		return fmt.Errorf("compile failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if rp.Parser != outer {
		t.Fatal("outer parser not restored")
	}
	if outer.file != "/tmp/inner.fnl" || outer.line != 7 || outer.col != 3 || outer.lineText != "  bad" {
		t.Fatalf("outer fault = file %q line %d col %d lineText %q", outer.file, outer.line, outer.col, outer.lineText)
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 (nestParser must not park Parser on TS)", rp.TempStack.Len())
	}
}

func tsHasParser(rp *RunProvider) bool {
	n := rp.TempStack.Len()
	for i := 0; i < n; i++ {
		if _, ok := rp.TempStack.Get(i).(*Parser); ok {
			return true
		}
	}
	return false
}

func TestNestParserDoesNotUseTS(t *testing.T) {
	outer := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("")),
		label: "outer",
	}})
	inner := NewParserFromTextRead(&TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(strings.NewReader("")),
		label: "inner",
	}})
	rp := NewTestRunProvider()
	rp.SetParser(outer)
	rp.TempStack.Push(Int(99))

	err := nestParser(rp, inner, func() error {
		if rp.Parser != inner {
			t.Fatal("inner parser not installed")
		}
		if tsHasParser(rp) {
			t.Fatal("Parser must not be parked on TS during nestParser")
		}
		rp.TempStack.Push(Int(1))
		return nil
	})
	if err != nil {
		t.Fatalf("nestParser: %v", err)
	}
	if rp.Parser != outer {
		t.Fatal("outer parser not restored")
	}
	if rp.TempStack.Len() != 2 {
		t.Fatalf("TS len = %d, want 2 (pre-nest 99 plus fn's 1; no hidden trim)", rp.TempStack.Len())
	}
	if got := rp.TempStack.Pop().(Int); got != 1 {
		t.Fatalf("TS TOS = %v, want 1", got)
	}
	if got := rp.TempStack.Pop().(Int); got != 99 {
		t.Fatalf("TS bottom = %v, want 99", got)
	}
}

func TestBootstrapWriteREPLErrorFormatsAndClears(t *testing.T) {
	rp := NewTestRunProvider()
	var sb StringBuilder
	rp.Stderr = &sb
	p := NewParser(nil)
	p.setParserLocation("/tmp/a.fnl", 2, 1)
	p.lineText = "x"
	p.col = 1
	p.line = 2
	rp.SetParser(p)

	bootstrapWriteREPLError(rp, fmt.Errorf("oops"))
	got := strings.TrimSuffix(sb.ToString(), "\n")
	want := "Error compiling file \"/tmp/a.fnl\"\nline 2; column 1 :\n" +
		"> x\n" +
		"> ^\n" +
		"Error: oops"
	if got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
	if p.file != "" || p.line != 1 || p.col != 0 {
		t.Fatalf("after report: file=%q line=%d col=%d", p.file, p.line, p.col)
	}
}

func TestInterpretSourceCapturesStdout(t *testing.T) {
	ResetTestRuntime()
	task := CreateLocalTask()
	rp := NewTestRunProvider()
	rp.WithTaskFindScope(task)
	out, err := InterpretSource(task, rp, "1 2 ds$")
	if err != nil {
		t.Fatalf("InterpretSource: %v", err)
	}
	if out != "1 2 " {
		t.Fatalf("stdout = %q, want %q", out, "1 2 ")
	}

	out, err = InterpretSource(task, rp, "1 2 + $")
	if err != nil {
		t.Fatalf("InterpretSource 1 2 + $: %v", err)
	}
	if out != "3" {
		t.Fatalf("stdout = %q, want %q", out, "3")
	}
	if rp.TempStack.Len() != 0 {
		t.Fatalf("TS len = %d, want 0 after InterpretSource", rp.TempStack.Len())
	}
	if tsHasParser(rp) {
		t.Fatal("InterpretSource must not leave Parser on TS")
	}
}
