package core

import (
	"bufio"
	"fmt"
	"os"
)

const (
	fileWriteAppendOrCreate   = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	fileWriteTruncateOrCreate = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

type fileReaderProvider struct {
	path string
	r    *bufio.Reader
	f    *os.File
}

func (p *fileReaderProvider) GetNextRune() (rune, bool) {
	return readNextRuneFromBufio(p.r, p.path)
}

func (p *fileReaderProvider) GetNextByte() (byte, bool) {
	return readNextByteFromBufio(p.r, p.path)
}

func (p *fileReaderProvider) ReadBytes(n int) []byte {
	return readBytesFromBufio(p.r, p.path, n)
}

func (p *fileReaderProvider) Close() {
	if err := p.f.Close(); err != nil {
		panic(fmt.Errorf("%s: close: %w", p.path, err))
	}
}

type fileWriterProvider struct {
	path string
	w    *bufio.Writer
	f    *os.File
}

func (p *fileWriterProvider) WriteRune(r rune) {
	if _, err := p.w.WriteRune(r); err != nil {
		panic(fmt.Errorf("%s: write rune: %w", p.path, err))
	}
}

func (p *fileWriterProvider) WriteLine(line string) {
	if _, err := p.w.WriteString(line); err != nil {
		panic(fmt.Errorf("%s: write line: %w", p.path, err))
	}
	if _, err := p.w.WriteRune('\n'); err != nil {
		panic(fmt.Errorf("%s: write line newline: %w", p.path, err))
	}
}

func (p *fileWriterProvider) WriteString(s string) {
	if _, err := p.w.WriteString(s); err != nil {
		panic(fmt.Errorf("%s: write string: %w", p.path, err))
	}
}

func (p *fileWriterProvider) WriteNextByte(b byte) {
	if err := p.w.WriteByte(b); err != nil {
		panic(fmt.Errorf("%s: write byte: %w", p.path, err))
	}
}

func (p *fileWriterProvider) WriteBytes(bs []byte) {
	if _, err := p.w.Write(bs); err != nil {
		panic(fmt.Errorf("%s: write bytes: %w", p.path, err))
	}
}

func (p *fileWriterProvider) Flush() {
	if err := p.w.Flush(); err != nil {
		panic(fmt.Errorf("%s: flush: %w", p.path, err))
	}
}

func (p *fileWriterProvider) Close() {
	if err := p.w.Flush(); err != nil {
		panic(fmt.Errorf("%s: close flush: %w", p.path, err))
	}
	if err := p.f.Close(); err != nil {
		panic(fmt.Errorf("%s: close: %w", p.path, err))
	}
}

func openFileReaderProvider(path string) (*fileReaderProvider, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return &fileReaderProvider{
		path: path,
		r:    bufio.NewReader(f),
		f:    f,
	}, nil
}

func openFileWriterProvider(path string, flags int) (*fileWriterProvider, error) {
	f, err := os.OpenFile(path, flags, 0o666)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return &fileWriterProvider{
		path: path,
		w:    bufio.NewWriter(f),
		f:    f,
	}, nil
}

// NewFileTextReader opens path for read-only sequential text access.
func NewFileTextReader(path string) (*TextReader, error) {
	p, err := openFileReaderProvider(path)
	if err != nil {
		return nil, fmt.Errorf("NewFileTextReader(%q): %w", path, err)
	}
	return &TextReader{p: p}, nil
}

// NewFileByteReader opens path for read-only sequential byte access.
func NewFileByteReader(path string) (*ByteReader, error) {
	p, err := openFileReaderProvider(path)
	if err != nil {
		return nil, fmt.Errorf("NewFileByteReader(%q): %w", path, err)
	}
	return &ByteReader{p: p}, nil
}

// NewFileTextWriterAppendOrCreate opens path for append, creating the file if absent.
func NewFileTextWriterAppendOrCreate(path string) (*TextWriter, error) {
	p, err := openFileWriterProvider(path, fileWriteAppendOrCreate)
	if err != nil {
		return nil, fmt.Errorf("NewFileTextWriterAppendOrCreate(%q): %w", path, err)
	}
	return &TextWriter{p: p}, nil
}

// NewFileTextWriterTruncateOrCreate opens path for write, truncating or creating the file.
func NewFileTextWriterTruncateOrCreate(path string) (*TextWriter, error) {
	p, err := openFileWriterProvider(path, fileWriteTruncateOrCreate)
	if err != nil {
		return nil, fmt.Errorf("NewFileTextWriterTruncateOrCreate(%q): %w", path, err)
	}
	return &TextWriter{p: p}, nil
}

// NewFileByteWriterAppendOrCreate opens path for append, creating the file if absent.
func NewFileByteWriterAppendOrCreate(path string) (*ByteWriter, error) {
	p, err := openFileWriterProvider(path, fileWriteAppendOrCreate)
	if err != nil {
		return nil, fmt.Errorf("NewFileByteWriterAppendOrCreate(%q): %w", path, err)
	}
	return &ByteWriter{p: p}, nil
}

// NewFileByteWriterTruncateOrCreate opens path for write, truncating or creating the file.
func NewFileByteWriterTruncateOrCreate(path string) (*ByteWriter, error) {
	p, err := openFileWriterProvider(path, fileWriteTruncateOrCreate)
	if err != nil {
		return nil, fmt.Errorf("NewFileByteWriterTruncateOrCreate(%q): %w", path, err)
	}
	return &ByteWriter{p: p}, nil
}

var (
	_ TextReadProvider  = (*fileReaderProvider)(nil)
	_ ByteReadProvider  = (*fileReaderProvider)(nil)
	_ Closeable         = (*fileReaderProvider)(nil)
	_ TextWriteProvider = (*fileWriterProvider)(nil)
	_ ByteWriteProvider = (*fileWriterProvider)(nil)
	_ Closeable         = (*fileWriterProvider)(nil)
	_ Flushable         = (*fileWriterProvider)(nil)
)
