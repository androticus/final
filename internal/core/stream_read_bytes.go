package core

import (
	"bufio"
	"fmt"
	"io"
)

func readNextByteFromBufio(r *bufio.Reader, label string) (b byte, eof bool) {
	ch, err := r.ReadByte()
	if err == io.EOF {
		return 0, true
	}
	if err != nil {
		panic(fmt.Errorf("%s: read byte: %w", label, err))
	}
	return ch, false
}

func readBytesFromBufio(r *bufio.Reader, label string, n int) []byte {
	if n <= 0 {
		return nil
	}
	buf := make([]byte, n)
	nr, err := r.Read(buf)
	if nr == 0 {
		return nil
	}
	if err != nil && err != io.EOF {
		panic(fmt.Errorf("%s: read bytes: %w", label, err))
	}
	return buf[:nr]
}
