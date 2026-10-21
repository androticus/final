package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileGetBaseExtMakeName(t *testing.T) {
	base, ext := pathBaseExt("archive/report.final.md")
	if base != "report.final" || ext != "md" {
		t.Fatalf("got %q %q", base, ext)
	}
	if fileMakeName("report", "md") != "report.md" {
		t.Fatal(fileMakeName("report", "md"))
	}
	if fileMakeName("out/report", ".md") != "out/report.md" {
		t.Fatal(fileMakeName("out/report", ".md"))
	}
	if fileMakeName("readme", "") != "readme" {
		t.Fatal(fileMakeName("readme", ""))
	}
}

func TestFileDeleteMoveCopy(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(root, "archive")
	if err := os.Mkdir(archive, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := fileCopyToDir(src, archive); err != nil {
		t.Fatal(err)
	}
	copied := filepath.Join(archive, "notes.txt")
	data, err := os.ReadFile(copied)
	if err != nil || string(data) != "hi" {
		t.Fatalf("copy: %v %q", err, data)
	}
	// no clobber
	if err := fileCopyToDir(src, archive); err == nil {
		t.Fatal("expected copy clobber error")
	}

	dst := filepath.Join(archive, "notes-2026.txt")
	if err := fileMove(copied, dst); err != nil {
		t.Fatal(err)
	}
	if FilePathExists(copied) {
		t.Fatal("move left source")
	}
	if err := fileMove(src, dst); err == nil {
		t.Fatal("expected move clobber error")
	}

	if err := fileDelete(dst); err != nil {
		t.Fatal(err)
	}
	if FilePathExists(dst) {
		t.Fatal("delete failed")
	}

	empty := filepath.Join(root, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := fileDelete(empty); err != nil {
		t.Fatal(err)
	}
}

func TestFileMembersRegistered(t *testing.T) {
	typ := GetTypeRegistrar().Get("File")
	for _, name := range []string{
		"exists?", "fileExists?", "info", "delete", "move", "copy",
		"getBaseExt", "makeName",
	} {
		if typ.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing File.%s", name)
		}
	}
}
