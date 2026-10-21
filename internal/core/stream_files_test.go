package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileTextWriterAppendOrCreate(t *testing.T) {
	dir := t.TempDir()

	appendPath := filepath.Join(dir, "append.txt")
	if err := os.WriteFile(appendPath, []byte("start"), 0o666); err != nil {
		t.Fatal(err)
	}
	w, err := NewFileTextWriterAppendOrCreate(appendPath)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString("end")
	w.Close()
	data, err := os.ReadFile(appendPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "startend" {
		t.Fatalf("append content = %q", string(data))
	}

	missingPath := filepath.Join(dir, "newlog.txt")
	w, err = NewFileTextWriterAppendOrCreate(missingPath)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString("first")
	w.Close()
	data, err = os.ReadFile(missingPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first" {
		t.Fatalf("create+append content = %q", string(data))
	}
}

func TestFileTextWriterTruncateOrCreate(t *testing.T) {
	dir := t.TempDir()

	truncPath := filepath.Join(dir, "trunc.txt")
	if err := os.WriteFile(truncPath, []byte("old"), 0o666); err != nil {
		t.Fatal(err)
	}
	w, err := NewFileTextWriterTruncateOrCreate(truncPath)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString("fresh")
	w.Close()
	data, err := os.ReadFile(truncPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fresh" {
		t.Fatalf("truncate content = %q", string(data))
	}

	createPath := filepath.Join(dir, "new.txt")
	w, err = NewFileTextWriterTruncateOrCreate(createPath)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteString("new")
	w.Close()
	data, err = os.ReadFile(createPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("create content = %q", string(data))
	}
}

func TestFileTextReaderMissing(t *testing.T) {
	_, err := NewFileTextReader(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFileTextReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read.txt")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileTextReader(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []rune("line1\nline2\n")
	for i, w := range want {
		ch, eof := r.GetNextRune()
		if eof || ch != w {
			t.Fatalf("rune[%d] = %q eof=%v, want %q", i, ch, eof, w)
		}
	}
	_, eof := r.GetNextRune()
	if !eof {
		t.Fatal("expected eof")
	}
	r.Close()
}

func TestFileTextReaderCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crlf.txt")
	if err := os.WriteFile(path, []byte("a\r\nb"), 0o666); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileTextReader(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []rune("a\r\nb")
	for i, w := range want {
		ch, eof := r.GetNextRune()
		if eof || ch != w {
			t.Fatalf("rune[%d] = %q eof=%v, want %q", i, ch, eof, w)
		}
	}
	_, eof := r.GetNextRune()
	if !eof {
		t.Fatal("expected eof")
	}
}

func TestFileTextReaderFinalLineWithoutNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.txt")
	if err := os.WriteFile(path, []byte("only"), 0o666); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileTextReader(path)
	if err != nil {
		t.Fatal(err)
	}
	ch, eof := r.GetNextRune()
	if eof || ch != 'o' {
		t.Fatalf("first rune = %q eof=%v", ch, eof)
	}
	for _, w := range []rune("nly") {
		ch, eof = r.GetNextRune()
		if eof || ch != w {
			t.Fatalf("rune = %q eof=%v, want %q", ch, eof, w)
		}
	}
	_, eof = r.GetNextRune()
	if !eof {
		t.Fatal("expected eof")
	}
}
