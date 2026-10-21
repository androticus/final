package core

import "testing"

func TestTextReaderFromStringRunes(t *testing.T) {
	r, err := NewTextReaderFromString("go")
	if err != nil {
		t.Fatal(err)
	}
	ch, eof := r.GetNextRune()
	if eof || ch != 'g' {
		t.Fatalf("first rune = %q eof=%v", ch, eof)
	}
	ch, eof = r.GetNextRune()
	if eof || ch != 'o' {
		t.Fatalf("second rune = %q eof=%v", ch, eof)
	}
	_, eof = r.GetNextRune()
	if !eof {
		t.Fatal("expected eof")
	}
}

func TestTextReaderFromStringNewlinesAsRunes(t *testing.T) {
	r, err := NewTextReaderFromString("a\nb\r\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []rune{'a', '\n', 'b', '\r', '\n', 'c', '\n'}
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
