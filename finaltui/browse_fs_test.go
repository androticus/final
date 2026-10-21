package finaltui

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestListDirSplitsAndHidesDots(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".dotdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := ListDir(dir)
	if got.Err != nil {
		t.Fatal(got.Err)
	}
	if len(got.Folders) != 1 || got.Folders[0] != "sub" {
		t.Fatalf("folders = %v", got.Folders)
	}
	if len(got.Files) != 1 || got.Files[0] != "a.txt" {
		t.Fatalf("files = %v", got.Files)
	}
}

func TestParentDirAtRoot(t *testing.T) {
	root := "/"
	if runtime.GOOS == "windows" {
		roots := WindowsDriveRoots()
		if len(roots) == 0 {
			t.Skip("no windows drives")
		}
		root = roots[0]
	}
	if ParentDir(root) != filepath.Clean(root) {
		t.Fatalf("ParentDir(%q) = %q", root, ParentDir(root))
	}
	if !IsFSRoot(root) {
		t.Fatalf("expected root %q", root)
	}
}

func TestResolveBrowsePathFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "x.fnl")
	if err := os.WriteFile(f, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	d, base, err := ResolveBrowsePath(f)
	if err != nil {
		t.Fatal(err)
	}
	if d != filepath.Clean(dir) || base != "x.fnl" {
		t.Fatalf("dir=%q base=%q", d, base)
	}
}

func TestShowBrowseFileFilter(t *testing.T) {
	if !showBrowseFile("a.fnl", true) || !showBrowseFile("A.FNL", true) {
		t.Fatal("expected .fnl match")
	}
	if !showBrowseFile("m.fnlm", true) || !showBrowseFile("p.fnlp", true) {
		t.Fatal("expected .fnlm / .fnlp match")
	}
	if showBrowseFile("a.txt", true) {
		t.Fatal("txt should be hidden when finalOnly")
	}
	if !showBrowseFile("a.txt", false) {
		t.Fatal("all files should show txt")
	}
}
