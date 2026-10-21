package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const csvModuleID = "github.com/androticus/final/modules/csv"

func runInterpretScriptErr(t *testing.T, task *Task, rp *RunProvider, script string) error {
	t.Helper()
	tr, err := NewTextReaderFromString(script)
	if err != nil {
		t.Fatal(err)
	}
	rp.SetParser(NewParserFromTextRead(tr))
	return interpretSession(task, rp)
}

func TestInstallImportCsvViaReplaces(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
DynamicObject.new^ >ts
"`+csvModuleID+`" t:peek install
t:drop
`)

	mod := ModuleTable.moduleByName(csvModuleID)
	if mod == nil {
		t.Fatal("expected csv module installed")
	}
	if mod.version != "v0.1.0" {
		t.Fatalf("version = %q, want v0.1.0", mod.version)
	}
	if mod.prefix != "csv" {
		t.Fatalf("prefix = %q, want csv", mod.prefix)
	}
	if !mod.dc.sealed {
		t.Fatal("expected installed module DC sealed")
	}
	if mod.dc.dict.GetByName("getDefaultCsvOptions") == nil {
		t.Fatal("expected getDefaultCsvOptions in installed module")
	}

	// Re-install is a nop.
	runInterpretScript(t, task, rp, `
DynamicObject.new^ >ts
"`+csvModuleID+`" t:peek install
t:drop
`)

	runInterpretScript(t, task, rp, `"" "`+csvModuleID+`" import`)
	if task.session.modules.Len() != 3 {
		t.Fatalf("modules len = %d, want 3 (local, core, csv)", task.session.modules.Len())
	}
	runInterpretScript(t, task, rp, `"" "`+csvModuleID+`" import`)
	if task.session.modules.Len() != 3 {
		t.Fatalf("after re-import modules len = %d, want 3", task.session.modules.Len())
	}

	runInterpretScript(t, task, rp, `csv\getDefaultCsvOptions`)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("after csv\\getDefaultCsvOptions DS len = %d, want 1", rp.DataStack.Len())
	}
	_ = rp.DataStack.Pop()

	runInterpretScript(t, task, rp, `getDefaultCsvOptions`)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("bare getDefaultCsvOptions DS len = %d, want 1", rp.DataStack.Len())
	}
}

func TestImportSynonymConflict(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
DynamicObject.new^
"`+csvModuleID+`" swap install
"" "`+csvModuleID+`" import
`)

	err = runInterpretScriptErr(t, task, rp, `"c" "`+csvModuleID+`" import`)
	if err == nil {
		t.Fatal("expected error on conflicting re-import prefix")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already attached") {
		t.Fatalf("err = %v, want already attached", err)
	}
}

func TestImportSealedRefuse(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
DynamicObject.new^
"`+csvModuleID+`" swap install
`)
	task.dc.Seal()
	err = runInterpretScriptErr(t, task, rp, `"" "`+csvModuleID+`" import`)
	if err == nil {
		t.Fatal("expected error importing into sealed local DC")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "sealed") {
		t.Fatalf("err = %v, want sealed", err)
	}
}

func TestInstallLockedRefuse(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp
	task.dc.lockedFn = func() bool { return true }

	err = runInterpretScriptErr(t, task, rp, `
DynamicObject.new^
"`+csvModuleID+`" swap install
`)
	if err == nil {
		t.Fatal("expected error installing while locked")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "locked") {
		t.Fatalf("err = %v, want locked", err)
	}
}

func TestInstallRejectsNonEmptyRequires(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	modJSON := `{
  "name": "test.local/leaf-with-req",
  "description": "x",
  "final": "v0.0.7-rc1",
  "version": "v0.0.1",
  "requires": [ { "name": "other", "version": "v1.0.0" } ],
  "files": [ "a.fnl" ]
}`
	if err := os.WriteFile(filepath.Join(dir, "_MODULE_"), []byte(modJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.fnl"), []byte(": hi 1 ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	script := `
DynamicObject.new^ >ts
"` + dir + `" t:.path!
"test.local/leaf-with-req" t:peek install
`
	err = runInterpretScriptErr(t, task, rp, script)
	if err == nil {
		t.Fatal("expected error for non-empty requires")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "requires") {
		t.Fatalf("err = %v, want requires", err)
	}
}

func TestInstallViaPathOverride(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
DynamicObject.new^ >ts
"modules/csv" t:.path!
"ignored-identity" t:peek install
t:drop
`)
	mod := ModuleTable.moduleByName(csvModuleID)
	if mod == nil {
		t.Fatal("expected csv installed via path override")
	}
}

func TestImportSynonymPrefix(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	task := s.task
	rp := s.rp

	runInterpretScript(t, task, rp, `
DynamicObject.new^
"`+csvModuleID+`" swap install
"c" "`+csvModuleID+`" import
c\getDefaultCsvOptions
`)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
	}
}

func expectOpNotAllowed(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	fn()
	t.Fatal("expected panic")
}

func TestInstallDeniedBySR_ModuleInstall(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.task.restrict = SR_ModuleInstall
	expectOpNotAllowed(t, func() {
		runInterpretScript(t, s.task, s.rp, `
DynamicObject.new^ >ts
"`+csvModuleID+`" t:peek install
`)
	})
}

func TestImportDeniedBySR_ModuleImport(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runInterpretScript(t, s.task, s.rp, `
DynamicObject.new^ >ts
"`+csvModuleID+`" t:peek install
t:drop
`)
	s.task.restrict = SR_ModuleImport
	expectOpNotAllowed(t, func() {
		runInterpretScript(t, s.task, s.rp, `"" "`+csvModuleID+`" import`)
	})
}

func TestInstallLeafModuleGoPathAssertsSR(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.task.restrict = SR_ModuleInstall
	expectOpNotAllowed(t, func() {
		installLeafModule(s.task, csvModuleID, NewDynamicObject())
	})
}

func TestImportModuleGoPathAssertsSR(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	runInterpretScript(t, s.task, s.rp, `
DynamicObject.new^ >ts
"`+csvModuleID+`" t:peek install
t:drop
`)
	s.task.restrict = SR_ModuleImport
	expectOpNotAllowed(t, func() {
		importModule(s.task, "", csvModuleID)
	})
}

func TestInstallCompilesWhenParentFileSRDenied(t *testing.T) {
	ResetTestRuntime()
	dir := t.TempDir()
	identity := "test.local/sr-compile-" + filepath.Base(dir)
	modJSON := `{
  "name": "` + identity + `",
  "description": "x",
  "final": "v0.0.7-rc1",
  "version": "v0.0.1",
  "requires": [],
  "files": [ "a.fnl" ]
}`
	if err := os.WriteFile(filepath.Join(dir, "_MODULE_"), []byte(modJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `
: usesWrite
  "/nope" "hi" File.writeText^
;
`
	if err := os.WriteFile(filepath.Join(dir, "a.fnl"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Persist: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.task.restrict = SR_FileWrite
	runInterpretScript(t, s.task, s.rp, `
DynamicObject.new^ >ts
"`+dir+`" t:.path!
"`+identity+`" t:peek install
t:drop
`)
	mod := ModuleTable.moduleByName(identity)
	if mod == nil {
		t.Fatal("compile session must stay unrestricted so install can emit FileWrite ops")
	}
	if mod.dc.dict.GetByName("usesWrite") == nil {
		t.Fatal("expected usesWrite compiled into installed module")
	}
}
