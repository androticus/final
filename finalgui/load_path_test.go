package finalgui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStartDirDefaultsToCwd(t *testing.T) {
	g := &FinalGui{}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got := g.loadStartDir()
	if got != cwd {
		t.Fatalf("loadStartDir = %q, want cwd %q", got, cwd)
	}
}

func TestLoadStartDirUsesLastFileDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.fnl")
	if err := os.WriteFile(file, []byte(": nop ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g := &FinalGui{lastFileLoadPath: file}
	got := g.loadStartDir()
	if got != dir {
		t.Fatalf("loadStartDir = %q, want %q", got, dir)
	}
}

func TestLoadStartDirFallsBackWhenLastDirGone(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	g := &FinalGui{lastFileLoadPath: filepath.Join(t.TempDir(), "missing", "gone.fnl")}
	got := g.loadStartDir()
	if got != cwd {
		t.Fatalf("loadStartDir = %q, want cwd %q", got, cwd)
	}
}
