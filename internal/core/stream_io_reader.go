package core

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// textReadIOReader adapts TextRead to io.Reader (UTF-8).
type textReadIOReader struct {
	src TextRead
	buf []byte // leftover UTF-8 from the last rune
}

func (r *textReadIOReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := 0
	for n < len(p) {
		if len(r.buf) > 0 {
			c := copy(p[n:], r.buf)
			r.buf = r.buf[c:]
			n += c
			continue
		}
		ch, eof := r.src.GetNextRune()
		if eof {
			if n == 0 {
				return 0, io.EOF
			}
			return n, nil
		}
		var enc [utf8.UTFMax]byte
		sz := utf8.EncodeRune(enc[:], ch)
		c := copy(p[n:], enc[:sz])
		n += c
		if c < sz {
			r.buf = append(r.buf[:0], enc[c:sz]...)
			return n, nil
		}
	}
	return n, nil
}

// byteReadIOReader adapts ByteRead to io.Reader.
type byteReadIOReader struct {
	src ByteRead
}

func (r *byteReadIOReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	bs := r.src.ReadBytes(len(p))
	if len(bs) == 0 {
		// Distinguish empty ReadBytes from EOF: probe one byte.
		b, eof := r.src.GetNextByte()
		if eof {
			return 0, io.EOF
		}
		p[0] = b
		return 1, nil
	}
	copy(p, bs)
	return len(bs), nil
}

// NewReaderFromTextRead wraps src as an io.Reader that emits UTF-8.
func NewReaderFromTextRead(src TextRead) (io.Reader, error) {
	if src == nil {
		return nil, fmt.Errorf("NewReaderFromTextRead: nil TextRead")
	}
	return &textReadIOReader{src: src}, nil
}

// NewReaderFromByteRead wraps src as an io.Reader.
func NewReaderFromByteRead(src ByteRead) (io.Reader, error) {
	if src == nil {
		return nil, fmt.Errorf("NewReaderFromByteRead: nil ByteRead")
	}
	return &byteReadIOReader{src: src}, nil
}

var (
	_ io.Reader = (*textReadIOReader)(nil)
	_ io.Reader = (*byteReadIOReader)(nil)
)
