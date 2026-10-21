package core

import "testing"

func TestNullTextStreams(t *testing.T) {
	r := GetNullTextReader()
	_, eof := r.GetNextRune()
	if !eof {
		t.Fatal("expected eof")
	}

	w := GetNullTextWriter()
	w.WriteString("discarded")

	rw := GetNullTextReaderWriter()
	if rw == nil {
		t.Fatal("expected non-nil read/write null stream")
	}
}
