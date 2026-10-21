package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompileFromManifestWordRegistered(t *testing.T) {
	ResetTestRuntime()
	if globalDictionaryContext.dict.GetByName("compileFromManifest") == nil {
		t.Fatal("missing compileFromManifest dictionary word")
	}
}

func TestCompileBootstrapManifestFileDefinesWords(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	a := writeTempFinalSource(t, ": manifA ;\n")
	b := writeTempFinalSource(t, ": manifB ;\n")
	manifestPath := filepath.Join(t.TempDir(), "app.fnlm")
	body := "// test manifest\n" + a + "\n// stubbed\n" + b + "\n"
	if err := os.WriteFile(manifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapManifestFile(task, rp, manifestPath); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("manifA") == nil {
		t.Fatal("manifA missing")
	}
	if globalDictionaryContext.dict.GetByName("manifB") == nil {
		t.Fatal("manifB missing")
	}
}

func TestDeferCompileFromManifestDefinesWord(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	src := writeTempFinalSource(t, ": manifDefer ;\n")
	manifestPath := filepath.Join(t.TempDir(), "defer.fnlm")
	if err := os.WriteFile(manifestPath, []byte(src+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rp.DataStack.Push(String(filepath.Clean(manifestPath)))
	runDeferWordByName(t, task, rp, "compileFromManifest")
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS len = %d, want 0", rp.DataStack.Len())
	}
	if globalDictionaryContext.dict.GetByName("manifDefer") == nil {
		t.Fatal("manifDefer missing")
	}
}

func TestCompileBootstrapManifestCommentOnlyOK(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	manifestPath := filepath.Join(t.TempDir(), "empty.fnlm")
	body := "// Darwin platform manifest (tview)\n// nothing yet\n"
	if err := os.WriteFile(manifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapManifestFile(task, rp, manifestPath); err != nil {
		t.Fatal(err)
	}
}

func TestCompileBootstrapManifestRelativeToManifestDir(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dir := t.TempDir()
	sub := filepath.Join(dir, "proj")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(sub, "app.fnl")
	if err := os.WriteFile(src, []byte(": manifRel ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(sub, "app.fnlm")
	// Relative to the .fnlm — not repo/cmd root.
	if err := os.WriteFile(manifestPath, []byte("app.fnl\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapManifestFile(task, rp, manifestPath); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("manifRel") == nil {
		t.Fatal("manifRel missing")
	}
}

// Manifest children must open Join(baseDir, line) directly. Re-applying resolveCompilePath
// would prepend bootstrap/cmd root again and can produce root+root+… paths.
func TestManifestChildNotReResolvedThroughModuleRoot(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dir := t.TempDir()
	pack := filepath.Join(dir, "Users", "baisa", "proj", "final-go", "finaltui", "final")
	if err := os.MkdirAll(pack, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(pack, "000-tui.fnl")
	if err := os.WriteFile(src, []byte(": manifNoDouble ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(pack, "package.fnlm")
	if err := os.WriteFile(manifestPath, []byte("000-tui.fnl\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapManifestFile(task, rp, manifestPath); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("manifNoDouble") == nil {
		t.Fatal("manifNoDouble missing")
	}
}

func TestInstallLayoutPresent(t *testing.T) {
	dir := t.TempDir()
	if installLayoutPresent(dir) {
		t.Fatal("empty dir should not look like an install")
	}
	if err := os.MkdirAll(filepath.Join(dir, "finaltui", "final"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !installLayoutPresent(dir) {
		t.Fatal("finaltui/final should count as install layout")
	}
}

func TestCompileBootstrapManifestRelativeSubdirAndSibling(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib")
	app := filepath.Join(dir, "app")
	if err := os.Mkdir(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "util.fnl"), []byte(": manifUtil ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "main.fnl"), []byte(": manifMain ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(app, "app.fnlm")
	body := "../lib/util.fnl\nmain.fnl\n"
	if err := os.WriteFile(manifestPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := compileBootstrapManifestFile(task, rp, manifestPath); err != nil {
		t.Fatal(err)
	}
	if globalDictionaryContext.dict.GetByName("manifUtil") == nil {
		t.Fatal("manifUtil missing")
	}
	if globalDictionaryContext.dict.GetByName("manifMain") == nil {
		t.Fatal("manifMain missing")
	}
}

func TestCompileBootstrapManifestRollbackOnLaterFile(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	dictBefore := len(globalDictionaryContext.dict.byOrdinal)
	good := writeTempFinalSource(t, ": manifGood ;\n")
	bad := writeTempFinalSource(t, ": manifBad nosuchword ;\n")
	manifestPath := filepath.Join(t.TempDir(), "bad.fnlm")
	if err := os.WriteFile(manifestPath, []byte(good+"\n"+bad+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := compileBootstrapManifestFile(task, rp, manifestPath)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if globalDictionaryContext.dict.GetByName("manifGood") == nil {
		t.Fatal("manifGood should remain")
	}
	if globalDictionaryContext.dict.GetByName("manifBad") != nil {
		t.Fatal("manifBad should be rolled back")
	}
	if len(globalDictionaryContext.dict.byOrdinal) != dictBefore+1 {
		t.Fatalf("dict len = %d, want %d", len(globalDictionaryContext.dict.byOrdinal), dictBefore+1)
	}
}

func TestCompileBootstrapManifestMissingPathErrors(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	err := compileBootstrapManifestFile(task, rp, filepath.Join(t.TempDir(), "nosuch.fnlm"))
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestResolveCompilePathAbsolute(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "x.fnl")
	got, err := resolveCompilePath(abs)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(abs) {
		t.Fatalf("got %q, want %q", got, filepath.Clean(abs))
	}
}

func TestResolveCompilePathRelativeUsesBootstrapRoot(t *testing.T) {
	root := bootstrapRoot()
	if root == "" {
		t.Skip("bootstrapRoot unavailable")
	}
	got, err := resolveCompilePath("internal/final/core")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "internal/final/core")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCmdRootNonEmpty(t *testing.T) {
	root := cmdRoot()
	if root == "" {
		t.Fatal("cmdRoot empty (Executable failed?)")
	}
	if !filepath.IsAbs(root) {
		t.Fatalf("cmdRoot not absolute: %q", root)
	}
}
