package core

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustWrite := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("a.txt", "a")
	mustWrite("b.txt", "b")
	mustWrite("sub/c.txt", "c")
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func collectFileSystemIter(it *FileSystemIter) []*FileInfo {
	var out []*FileInfo
	for it.GetNext() {
		out = append(out, it.Current().(*FileInfo))
	}
	return out
}

func namesOf(fis []*FileInfo) []string {
	out := make([]string, len(fis))
	for i, fi := range fis {
		out[i] = fi.Name
	}
	return out
}

func TestFileSystemIterShallowFiles(t *testing.T) {
	root := writeTempTree(t)
	it := NewFileSystemIter(NewDir(root), DirIterFiles)
	got := collectFileSystemIter(it)
	names := namesOf(got)
	// ReadDir order is host-defined; assert set membership and path/name split.
	want := map[string]bool{"a.txt": true, "b.txt": true}
	if len(names) != 2 {
		t.Fatalf("names = %v, want 2 files", names)
	}
	for _, n := range names {
		if !want[n] {
			t.Fatalf("unexpected name %q in %v", n, names)
		}
	}
	for _, fi := range got {
		if fi.Path != root {
			t.Fatalf("Path = %q, want root %q", fi.Path, root)
		}
		if fi.IsDir {
			t.Fatalf("%s: IsDir true, want false", fi.Name)
		}
		full := filepath.Join(fi.Path, fi.Name)
		if _, err := os.Stat(full); err != nil {
			t.Fatalf("join path+name: %v", err)
		}
	}
}

func TestFileSystemIterShallowDirs(t *testing.T) {
	root := writeTempTree(t)
	it := NewFileSystemIter(NewDir(root), DirIterDirs)
	got := collectFileSystemIter(it)
	names := namesOf(got)
	want := map[string]bool{"sub": true, "empty": true}
	if len(names) != 2 {
		t.Fatalf("names = %v, want sub and empty", names)
	}
	for _, n := range names {
		if !want[n] {
			t.Fatalf("unexpected name %q in %v", n, names)
		}
	}
	for _, fi := range got {
		if !fi.IsDir {
			t.Fatalf("%s: IsDir false", fi.Name)
		}
	}
}

func TestFileSystemIterDeepFiles(t *testing.T) {
	root := writeTempTree(t)
	it := NewFileSystemIter(NewDir(root), DirIterFiles|DirIterDeep)
	got := collectFileSystemIter(it)
	names := namesOf(got)
	want := map[string]bool{"a.txt": true, "b.txt": true, "c.txt": true}
	if len(names) != 3 {
		t.Fatalf("names = %v, want 3 files", names)
	}
	for _, n := range names {
		if !want[n] {
			t.Fatalf("unexpected name %q in %v", n, names)
		}
	}
	var foundC bool
	for _, fi := range got {
		if fi.Name == "c.txt" {
			foundC = true
			wantParent := filepath.Join(root, "sub")
			if fi.Path != wantParent {
				t.Fatalf("c.txt Path = %q, want %q", fi.Path, wantParent)
			}
		}
	}
	if !foundC {
		t.Fatal("missing c.txt")
	}
}

func TestFileSystemIterNotDirPanics(t *testing.T) {
	root := writeTempTree(t)
	file := filepath.Join(root, "a.txt")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for non-directory root")
		}
	}()
	_ = NewFileSystemIter(NewDir(file), DirIterFiles)
}

func TestDirTypeRegistered(t *testing.T) {
	if GetTypeRegistrar().Get("Dir") == nil {
		t.Fatal("Dir type not registered")
	}
	if GetTypeRegistrar().Get("FileInfo") == nil {
		t.Fatal("FileInfo type not registered")
	}
	d := NewDir("/tmp")
	if d.FinalType().Name() != "Dir" {
		t.Fatalf("FinalType = %q", d.FinalType().Name())
	}
	typ := GetTypeRegistrar().Get("Dir")
	for _, name := range []string{"cwd", "setCwd", "exists?", "info", "mkdir", "new"} {
		if typ.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing Dir class method %s", name)
		}
	}
	if typ.GetMethodMetadata("iter") == nil {
		t.Fatal("missing Dir.iter")
	}
}

func TestDirExists(t *testing.T) {
	root := writeTempTree(t)
	if !FilePathExists(root) {
		t.Fatal("root should exist")
	}
	if FilePathExists(filepath.Join(root, "nope")) {
		t.Fatal("missing path should be false")
	}
}

func TestDirMkdirAll(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := dirMkdirAll(nested); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(nested)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected dir, err=%v", err)
	}
	// idempotent
	if err := dirMkdirAll(nested); err != nil {
		t.Fatalf("idempotent mkdir: %v", err)
	}
	// file in the way
	file := filepath.Join(root, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := dirMkdirAll(file); err == nil {
		t.Fatal("expected error when path is a file")
	}
}

func TestFileInfoFromPath(t *testing.T) {
	root := writeTempTree(t)
	file := filepath.Join(root, "a.txt")
	fi, err := fileInfoFromPath(file)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Name != "a.txt" || fi.Path != root {
		t.Fatalf("got path=%q name=%q", fi.Path, fi.Name)
	}
	if fi.IsDir {
		t.Fatal("file should not be dir")
	}
	sub := filepath.Join(root, "sub")
	di, err := fileInfoFromPath(sub)
	if err != nil {
		t.Fatal(err)
	}
	if !di.IsDir || di.Name != "sub" {
		t.Fatalf("got %#v", di)
	}
}

func TestDirCwdSetCwd(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Resolve symlinks (macOS /var -> /private/var)
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(wd)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("cwd = %q, want %q", got, want)
	}
	d := NewDir(wd)
	if d.Path() != wd {
		t.Fatalf("Dir path = %q", d.Path())
	}
}

