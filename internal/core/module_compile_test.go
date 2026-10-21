package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileModuleCsvIntoLocalSession(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.GetByName("compileModule") == nil {
		t.Fatal("missing compileModule (0502-compile-module.fnl in tranche2?)")
	}

	jsonBytes, err := os.ReadFile(filepath.Join("modules", "csv", "_MODULE_"))
	if err != nil {
		// Tests run with cwd = module under test (internal/core); repo root is ../..
		jsonBytes, err = os.ReadFile(filepath.Join("..", "..", "modules", "csv", "_MODULE_"))
		if err != nil {
			t.Fatalf("read _MODULE_: %v", err)
		}
	}
	root := filepath.Join("modules", "csv")
	if _, err := os.Stat(filepath.Join(root, "10-csv-export.fnl")); err != nil {
		root = filepath.Join("..", "..", "modules", "csv")
	}

	task := CreateLocalTask()
	rp := NewTestRunProvider()
	outB := NewStringBuilder()
	rp.Stdout = outB
	rp.WithTaskFindScope(task)

	rp.DataStack.Push(String(root))
	rp.DataStack.Push(String(jsonBytes))
	w := globalDictionaryContext.dict.GetByName("compileModule")
	if err := runWordInlineExec(task, rp, w); err != nil {
		t.Fatalf("compileModule: %v\nstdout:\n%s", err, outB.ToString())
	}

	out := outB.ToString()
	exportPath := filepath.Join(root, "10-csv-export.fnl")
	importPath := filepath.Join(root, "20-csv-import.fnl")
	wantExport := "Compiled " + exportPath + ".\n"
	wantImport := "Compiled " + importPath + ".\n"
	if !strings.Contains(out, wantExport) {
		t.Fatalf("stdout missing %q\ngot:\n%s", wantExport, out)
	}
	if !strings.Contains(out, wantImport) {
		t.Fatalf("stdout missing %q\ngot:\n%s", wantImport, out)
	}
	if !strings.Contains(out, "zzCMOD_DONE_zz\n") {
		t.Fatalf("stdout missing sentinel zzCMOD_DONE_zz\ngot:\n%s", out)
	}

	if task.dc.dict.GetByName("getDefaultCsvOptions") == nil {
		t.Fatal("expected getDefaultCsvOptions defined in local DC after compileModule")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d after compileModule, want 0; stack=%#v", rp.DataStack.Len(), rp.DataStack)
	}
}
