package core

import (
	"bufio"
	"fmt"
	"io"
)

type ioWriterProvider struct {
	w io.Writer
}

func (p *ioWriterProvider) WriteRune(r rune) {
	if _, err := io.WriteString(p.w, string(r)); err != nil {
		panic(fmt.Errorf("io writer: write rune: %w", err))
	}
}

func (p *ioWriterProvider) WriteLine(line string) {
	if _, err := fmt.Fprintln(p.w, line); err != nil {
		panic(fmt.Errorf("io writer: write line: %w", err))
	}
}

func (p *ioWriterProvider) WriteString(s string) {
	if _, err := io.WriteString(p.w, s); err != nil {
		panic(fmt.Errorf("io writer: write string: %w", err))
	}
}

func (p *ioWriterProvider) WriteNextByte(b byte) {
	if _, err := p.w.Write([]byte{b}); err != nil {
		panic(fmt.Errorf("io writer: write byte: %w", err))
	}
}

func (p *ioWriterProvider) WriteBytes(bs []byte) {
	if _, err := p.w.Write(bs); err != nil {
		panic(fmt.Errorf("io writer: write bytes: %w", err))
	}
}

func (p *ioWriterProvider) Close() {
	if c, ok := p.w.(io.Closer); ok {
		if err := c.Close(); err != nil {
			panic(fmt.Errorf("io writer: close: %w", err))
		}
	}
}

// NewTextReaderFromReader wraps r as a TextReader (e.g. *os.File, pipe, bytes.Buffer).
func NewTextReaderFromReader(r io.Reader) (*TextReader, error) {
	if r == nil {
		return nil, fmt.Errorf("NewTextReaderFromReader: nil reader")
	}
	return &TextReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(r),
		label: "reader",
	}}, nil
}

// NewTextWriterFromWriter wraps w as a TextWriter (e.g. *os.File, pipe, bytes.Buffer).
func NewTextWriterFromWriter(w io.Writer) (*TextWriter, error) {
	if w == nil {
		return nil, fmt.Errorf("NewTextWriterFromWriter: nil writer")
	}
	return &TextWriter{p: &ioWriterProvider{w: w}}, nil
}

// NewByteReaderFromReader wraps r as a ByteReader (e.g. *os.File, pipe, bytes.Buffer).
func NewByteReaderFromReader(r io.Reader) (*ByteReader, error) {
	if r == nil {
		return nil, fmt.Errorf("NewByteReaderFromReader: nil reader")
	}
	return &ByteReader{p: &bufioReaderProvider{
		r:     bufio.NewReader(r),
		label: "reader",
	}}, nil
}

// NewByteWriterFromWriter wraps w as a ByteWriter (e.g. *os.File, pipe, bytes.Buffer).
func NewByteWriterFromWriter(w io.Writer) (*ByteWriter, error) {
	if w == nil {
		return nil, fmt.Errorf("NewByteWriterFromWriter: nil writer")
	}
	return &ByteWriter{p: &ioWriterProvider{w: w}}, nil
}

var (
	_ TextWriteProvider = (*ioWriterProvider)(nil)
	_ ByteWriteProvider = (*ioWriterProvider)(nil)
	_ Closeable         = (*ioWriterProvider)(nil)
)
