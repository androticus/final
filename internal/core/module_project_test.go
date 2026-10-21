package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProjectCsvRequireAndLocalFile(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	projPath := filepath.Join(dir, "app.fnlp")
	projJSON := `{
  "name": "test.local/proj-csv",
  "version": "v0.0.1",
  "requires": [
    { "name": "github.com/androticus/final/modules/csv", "version": "v0.1.0" }
  ],
  "files": [ "main.fnl" ]
}`
	if err := os.WriteFile(projPath, []byte(projJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	main := `
: pingCsv
  getDefaultCsvOptions drop
  99
;
`
	if err := os.WriteFile(filepath.Join(dir, "main.fnl"), []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `"`+projPath+`" loadProject`)

	if ModuleTable.moduleByName(csvModuleID) == nil {
		t.Fatal("expected csv installed via project requires")
	}
	if task.session.modules.Len() != 3 {
		t.Fatalf("modules len = %d, want 3 (local, core, csv)", task.session.modules.Len())
	}
	if task.dc.dict.GetByName("pingCsv") == nil {
		t.Fatal("expected pingCsv compiled into local")
	}

	runInterpretScript(t, task, rp, `pingCsv`)
	if got := rp.DataStack.Pop(); got != Int(99) {
		t.Fatalf("pingCsv = %v, want 99", got)
	}
}

func TestLoadProjectRelaxedFilePaths(t *testing.T) {
	ResetTestRuntime()
	base := t.TempDir()
	projDir := filepath.Join(base, "proj")
	sharedDir := filepath.Join(base, "shared")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sharedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sharedFile := filepath.Join(sharedDir, "shared.fnl")
	if err := os.WriteFile(sharedFile, []byte(": fromShared  7 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	otherShared := filepath.Join(sharedDir, "other.fnl")
	if err := os.WriteFile(otherShared, []byte(": fromOther  8 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	relOther := filepath.Join("..", "shared", "other.fnl")

	projPath := filepath.Join(projDir, "paths.fnlp")
	projJSON := `{
  "name": "test.local/proj-paths",
  "version": "v0.0.1",
  "files": [ "` + filepath.ToSlash(sharedFile) + `", "` + filepath.ToSlash(relOther) + `" ]
}`
	if err := os.WriteFile(projPath, []byte(projJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `"`+projPath+`" loadProject`)
	runInterpretScript(t, task, rp, `fromShared fromOther`)
	a := rp.DataStack.Pop()
	b := rp.DataStack.Pop()
	if a != Int(8) || b != Int(7) {
		t.Fatalf("got %v %v, want 7 then 8 (stack pop order)", b, a)
	}
}

func TestLoadProjectRequirePrefix(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	projPath := filepath.Join(dir, "pref.fnlp")
	projJSON := `{
  "name": "test.local/proj-prefix",
  "version": "v0.0.1",
  "requires": [
    { "name": "github.com/androticus/final/modules/csv", "version": "v0.1.0", "prefix": "c" }
  ],
  "files": []
}`
	if err := os.WriteFile(projPath, []byte(projJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `"`+projPath+`" loadProject`)
	runInterpretScript(t, task, rp, `c\getDefaultCsvOptions`)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
}

func TestLoadProjectSealedRefuse(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	projPath := filepath.Join(dir, "x.fnlp")
	if err := os.WriteFile(projPath, []byte(`{
  "name": "test.local/x",
  "version": "v0.0.1",
  "files": []
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp
	task.dc.Seal()

	err = runInterpretScriptErr(t, task, rp, `"`+projPath+`" loadProject`)
	if err == nil {
		t.Fatal("expected sealed error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "sealed") {
		t.Fatalf("err = %v, want sealed", err)
	}
}

func TestLoadProjectRejectsNonFnlp(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	err = runInterpretScriptErr(t, s.task, s.rp, `"foo.json" loadProject`)
	if err == nil {
		t.Fatal("expected .fnlp suffix error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), ".fnlp") {
		t.Fatalf("err = %v, want .fnlp", err)
	}
}

func TestResolveProjectFilePath(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "proj")
	abs := filepath.Join(string(filepath.Separator), "abs", "a.fnl")
	if got := resolveProjectFilePath(root, abs); got != filepath.Clean(abs) {
		t.Fatalf("abs = %q, want %q", got, filepath.Clean(abs))
	}
	rel := resolveProjectFilePath(root, filepath.Join("..", "shared", "b.fnl"))
	want := filepath.Clean(filepath.Join(root, "..", "shared", "b.fnl"))
	if rel != want {
		t.Fatalf("rel = %q, want %q", rel, want)
	}
}

func TestCsvDemoProjectFixture(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp
	outB := NewStringBuilder()
	rp.Stdout = outB
	runInterpretScript(t, task, rp, `"projects/csv-demo/csv-demo.fnlp" loadProject`)
	out := outB.ToString()
	if !strings.Contains(out, "Try: csv-demo") {
		t.Fatalf("missing blurb invite, stdout:\n%s", out)
	}
	runInterpretScript(t, task, rp, `csv-demo`)
	out = outB.ToString()
	if !strings.Contains(out, "Creature,Legs,Eyes") {
		t.Fatalf("missing CSV header, stdout:\n%s", out)
	}
	if !strings.Contains(out, "octopus,8,2") {
		t.Fatalf("missing CSV row, stdout:\n%s", out)
	}
}

func TestLoadProjectRequiresDeniedByModuleInstallSR(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	projPath := filepath.Join(dir, "app.fnlp")
	projJSON := `{
  "name": "test.local/proj-sr-install",
  "version": "v0.0.1",
  "requires": [
    { "name": "github.com/androticus/final/modules/csv", "version": "v0.1.0" }
  ],
  "files": [ "main.fnl" ]
}`
	if err := os.WriteFile(projPath, []byte(projJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.fnl"), []byte(": ping 1 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.task.restrict = SR_ModuleInstall
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	loadProject(s.task, s.rp, projPath)
	t.Fatal("expected panic")
}

func TestLoadProjectFilesOnlyAllowedUnderModuleSR(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	projPath := filepath.Join(dir, "app.fnlp")
	projJSON := `{
  "name": "test.local/proj-sr-files",
  "version": "v0.0.1",
  "files": [ "main.fnl" ]
}`
	if err := os.WriteFile(projPath, []byte(projJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.fnl"), []byte(": pingSr 3 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.task.restrict = SR_ModuleInstall | SR_ModuleImport
	runInterpretScript(t, s.task, s.rp, `"`+projPath+`" loadProject`)
	if s.task.dc.dict.GetByName("pingSr") == nil {
		t.Fatal("loadProject with no requires should compile files under module SRs")
	}
}
