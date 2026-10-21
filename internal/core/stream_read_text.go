package core

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

func readNextRuneFromString(s string, off *int, label string) (r rune, eof bool) {
	if *off >= len(s) {
		return 0, true
	}
	r, size := utf8.DecodeRuneInString(s[*off:])
	if r == utf8.RuneError && size == 1 {
		panic(fmt.Errorf("%s: invalid UTF-8 at offset %d", label, *off))
	}
	*off += size
	return r, false
}

func readNextRuneFromBufio(r *bufio.Reader, label string) (rv rune, eof bool) {
	ch, _, err := r.ReadRune()
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		panic(fmt.Errorf("%s: read rune: %w", label, err))
	}
	return ch, false
}
