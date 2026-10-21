package core

import (
	"io"
	"testing"
)

func TestReaderFromTextReadRoundtrip(t *testing.T) {
	tr, err := NewTextReaderFromString("café")
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewReaderFromTextRead(tr)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "café" {
		t.Fatalf("got %q, want café", got)
	}
}

func TestReaderFromByteReadRoundtrip(t *testing.T) {
	list := NewByteList()
	list.items = append(list.items, 0, 1, 255, 2)
	br, err := NewByteReaderFromBytes(list)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewReaderFromByteRead(br)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 1, 255, 2}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte[%d]=%d want %d", i, got[i], want[i])
		}
	}
}

func TestNewReaderFromTextReadNil(t *testing.T) {
	_, err := NewReaderFromTextRead(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewReaderFromByteReadNil(t *testing.T) {
	_, err := NewReaderFromByteRead(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
