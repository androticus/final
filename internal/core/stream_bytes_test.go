package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestByteListByteStreamRoundtrip(t *testing.T) {
	list := NewByteList()
	w, err := NewByteWriterFromBytes(list)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteBytes([]byte{1, 2, 3})
	w.WriteNextByte(4)
	w.Flush()

	r, err := NewByteReaderFromBytes(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []byte{1, 2, 3, 4} {
		b, eof := r.GetNextByte()
		if eof || b != want {
			t.Fatalf("byte = %d eof=%v, want %d", b, eof, want)
		}
	}
	_, eof := r.GetNextByte()
	if !eof {
		t.Fatal("expected eof")
	}
}

func TestByteReaderReadBytes(t *testing.T) {
	r, err := NewByteReaderFromSlice([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(r.ReadBytes(3)); got != "hel" {
		t.Fatalf("ReadBytes(3) = %q", got)
	}
	if got := string(r.ReadBytes(10)); got != "lo" {
		t.Fatalf("ReadBytes(10) = %q", got)
	}
	if got := r.ReadBytes(5); got != nil {
		t.Fatalf("ReadBytes at eof = %v", got)
	}
}

func TestByteReaderFromReader(t *testing.T) {
	r, err := NewByteReaderFromReader(bytes.NewReader([]byte{9, 8}))
	if err != nil {
		t.Fatal(err)
	}
	b, eof := r.GetNextByte()
	if eof || b != 9 {
		t.Fatalf("first = %d eof=%v", b, eof)
	}
}

func TestFileByteWriterModes(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "bin.dat")
	w, err := NewFileByteWriterTruncateOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteBytes([]byte{0x01, 0x02})
	w.Close()

	w, err = NewFileByteWriterAppendOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteBytes([]byte{0x03})
	w.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "\x01\x02\x03" {
		t.Fatalf("content = % x", data)
	}

	missing := filepath.Join(dir, "new.bin")
	w, err = NewFileByteWriterAppendOrCreate(missing)
	if err != nil {
		t.Fatal(err)
	}
	w.WriteNextByte(0xff)
	w.Close()
	data, err = os.ReadFile(missing)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 || data[0] != 0xff {
		t.Fatalf("new file content = % x", data)
	}
}

func TestFileByteReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read.bin")
	if err := os.WriteFile(path, []byte{10, 20, 30}, 0o666); err != nil {
		t.Fatal(err)
	}
	r, err := NewFileByteReader(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []byte{10, 20, 30} {
		b, eof := r.GetNextByte()
		if eof || b != want {
			t.Fatalf("byte = %d eof=%v, want %d", b, eof, want)
		}
	}
	r.Close()
}

func TestFilePathExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "present.txt")
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	if !FilePathExists(path) {
		t.Fatal("expected path to exist")
	}
	if FilePathExists(filepath.Join(dir, "absent.txt")) {
		t.Fatal("expected path to not exist")
	}
}

func TestNewByteReaderFromBytesNil(t *testing.T) {
	_, err := NewByteReaderFromBytes(nil)
	if err == nil {
		t.Fatal("expected error for nil list")
	}
}
