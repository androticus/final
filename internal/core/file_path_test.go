package core

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathTypeRegistered(t *testing.T) {
	typ := GetTypeRegistrar().Get("Path")
	if typ == nil {
		t.Fatal("Path type not registered")
	}
	for _, name := range []string{"join", "split", "clean", "abs", "basename", "dirname", "extname"} {
		if typ.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing Path.%s", name)
		}
	}
}

func TestPathBasenameExtname(t *testing.T) {
	base, ext := pathBaseExt("out/reports/q1.md")
	if base != "q1" || ext != "md" {
		t.Fatalf("got base=%q ext=%q", base, ext)
	}
	base, ext = pathBaseExt(".gitignore")
	if base != ".gitignore" || ext != "" {
		t.Fatalf("dotfile: base=%q ext=%q", base, ext)
	}
	base, ext = pathBaseExt("archive/report.final.md")
	if base != "report.final" || ext != "md" {
		t.Fatalf("multi-dot: base=%q ext=%q", base, ext)
	}
}

func TestPathSplitRelative(t *testing.T) {
	list := pathSplit(filepath.Join("a", "b", "c"))
	if list.Len() != 3 {
		t.Fatalf("len = %d, want 3", list.Len())
	}
	if list.At(0) != "a" || list.At(1) != "b" || list.At(2) != "c" {
		t.Fatalf("parts = %v", []string{list.At(0), list.At(1), list.At(2)})
	}
}

func TestPathSplitAbsolute(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-style absolute split")
	}
	list := pathSplit("/var/log/syslog")
	if list.Len() != 4 {
		t.Fatalf("len = %d want 4, parts via Len", list.Len())
	}
	if list.At(0) != "/" {
		t.Fatalf("root = %q", list.At(0))
	}
	if list.At(1) != "var" || list.At(2) != "log" || list.At(3) != "syslog" {
		t.Fatalf("got %q %q %q", list.At(1), list.At(2), list.At(3))
	}
}

func TestPathSplitEmpty(t *testing.T) {
	if pathSplit("").Len() != 0 {
		t.Fatal("empty should be empty list")
	}
}

func TestPathAbs(t *testing.T) {
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(abs) {
		t.Fatalf("not abs: %q", abs)
	}
}
