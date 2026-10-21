package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempFinalSource(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "final-*.fina")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestCompileFromFileWordRegistered(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.GetByName("compileFromFile") == nil {
		t.Fatal("missing compileFromFile dictionary word")
	}
}

func TestCompileBootstrapFileDefinesWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	path := writeTempFinalSource(t, ": fromfile ;\n")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("fromfile") == nil {
		t.Fatal("fromfile should be defined after compileBootstrapFile")
	}
}

func TestDeferCompileFromFileDefinesWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	path := writeTempFinalSource(t, ": deferfile ;\n")
	rp.DataStack.Push(String(filepath.Clean(path)))
	runDeferWordByName(t, task, rp, "compileFromFile")
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
	if globalDictionaryContext.dict.GetByName("deferfile") == nil {
		t.Fatal("deferfile should be defined after compileFromFile")
	}
}

func TestDeferCompileFromFileBadSourceFaults(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	path := writeTempFinalSource(t, ": badfile nosuchword ;\n")
	rp.DataStack.Push(String(filepath.Clean(path)))
	w := globalDictionaryContext.dict.GetByName("compileFromFile")
	if w == nil {
		t.Fatal("missing compileFromFile")
	}
	err := runWordInlineExec(task, rp, w)
	if _, ok := err.(*RunFault); !ok {
		t.Fatalf("err = %v (%T), want *RunFault", err, err)
	}
}

func TestCompileBootstrapFileErrorRollsBack(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dictBefore := len(globalDictionaryContext.dict.byOrdinal)
	path := writeTempFinalSource(t, ": badfile nosuchword ;\n")
	err := compileBootstrapFile(task, rp, path)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if globalDictionaryContext.dict.GetByName("badfile") != nil {
		t.Fatal("partial word should be rolled back")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictBefore {
		t.Fatalf("dict len = %d, want %d", len(globalDictionaryContext.dict.byOrdinal), dictBefore)
	}
}

func TestCompileBootstrapFileMissingPathErrors(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapFile(task, rp, filepath.Join(t.TempDir(), "nosuch.fina"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCompileFromFileAutoRunsLeadingForgetMark(t *testing.T) {
	_, task, rp := localForgetSession(t)
	path := writeTempFinalSource(t, "forget: reload\n: scratch 2 ;\n")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	if task.dc.dict.GetByName("reload") == nil || task.dc.dict.GetByName("scratch") == nil {
		t.Fatal("reload/scratch missing after first load")
	}
	if err := os.WriteFile(path, []byte("forget: reload\n: scratch 3 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	interpretLocal(t, task, rp, `scratch`)
	got, ok := rp.DataStack.Pop().(Int)
	if !ok || got != 3 {
		t.Fatalf("scratch = %v, want 3", got)
	}
}

func TestCompileFromFileLeadingForgetFailedUnitStaysWiped(t *testing.T) {
	_, task, rp := localForgetSession(t)
	path := writeTempFinalSource(t, "forget: reload\n: scratch 1 ;\n")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("forget: reload\n: scratch nosuchword ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapFile(task, rp, path); err == nil {
		t.Fatal("expected compile error")
	}
	if task.dc.dict.GetByName("scratch") != nil {
		t.Fatal("scratch should stay wiped after failed reload")
	}
	if task.dc.dict.GetByName("reload") != nil {
		t.Fatal("new mark should roll back with the failed unit")
	}
}

func TestCompileFromFileSkipsAutoForgetWhenLine1IsNotSentinel(t *testing.T) {
	_, task, rp := localForgetSession(t)
	path := writeTempFinalSource(t, "// header\nforget: reload\n: scratch 1 ;\n")
	if err := compileBootstrapFile(task, rp, path); err != nil {
		t.Fatal(err)
	}
	err := compileBootstrapFile(task, rp, path)
	if err == nil {
		t.Fatal("expected already-exists on second load without line-1 forget:")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v, want already exists", err)
	}
}

func TestCompileBootstrapFileErrorIncludesAbsolutePath(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	var stderr StringBuilder
	rp.Stderr = &stderr
	path := writeTempFinalSource(t, ": badpath nosuchword ;\n")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	err = compileBootstrapFile(task, rp, path)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if rp.Parser == nil || rp.Parser.file != abs {
		t.Fatalf("outer parser file = %q, want %q", rp.Parser.file, abs)
	}
	msg := err.Error()
	if !IsLocatedFaultMessage(msg) {
		t.Fatalf("err = %q, want located fault", msg)
	}
	if !strings.Contains(msg, abs) {
		t.Fatalf("err = %q, want path %q", msg, abs)
	}
	if !strings.Contains(msg, `unknown word "nosuchword"`) {
		t.Fatalf("err = %q, want unknown word", msg)
	}
	if !strings.Contains(msg, "> ") || !strings.Contains(msg, "^") {
		t.Fatalf("err = %q, want caret lines", msg)
	}
	bootstrapWriteREPLError(rp, err)
	out := stderr.ToString()
	// Already-located error prints once (not re-wrapped).
	if strings.Count(out, "Error compiling") != 1 {
		t.Fatalf("stderr = %q, want one Error compiling", out)
	}
}

func TestCompileBootstrapFileCaretOnFaultLine(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	path := writeTempFinalSource(t, ": myWord\n  nop\n  foo\n;\n")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	err = compileBootstrapFile(task, rp, path)
	if err == nil {
		t.Fatal("expected compile error")
	}
	msg := err.Error()
	if !strings.Contains(msg, abs) {
		t.Fatalf("msg missing abs path: %q", msg)
	}
	if !strings.Contains(msg, "line 3;") {
		t.Fatalf("want line 3 in %q", msg)
	}
	if !strings.Contains(msg, ">   foo\n") {
		t.Fatalf("want source line in %q", msg)
	}
	if !strings.Contains(msg, `Error: unknown word "foo"`) {
		t.Fatalf("want unknown word in %q", msg)
	}
}

// A multi-line raw string must not rewind the live line counter when yielded
// (Token.Line is the opener). Faults after the raw string use the true line.
func TestCompileBootstrapFileLineAfterMultilineRawString(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := "forget: t\n\n`\nlineA\nlineB\nlineC\n` const: blurb\n\n: badLit\n  12bad\n;\n"
	path := writeTempFinalSource(t, src)
	err := compileBootstrapFile(task, rp, path)
	if err == nil {
		t.Fatal("expected compile error")
	}
	msg := err.Error()
	// Physical line of 12bad: 1 forget, 2 blank, 3-7 raw (5 lines), 8 blank, 9 : badLit, 10 12bad
	if !strings.Contains(msg, "line 10;") {
		t.Fatalf("want line 10 after raw string, got %q", msg)
	}
	if !strings.Contains(msg, ">   12bad\n") {
		t.Fatalf("want 12bad source line in %q", msg)
	}
	if !strings.Contains(msg, `invalid int "12bad"`) {
		t.Fatalf("want invalid int in %q", msg)
	}
}
