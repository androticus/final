package core

import (
	"bytes"
	"fmt"
)

type ByteListWriter struct {
	ByteWriter
	list *ByteList
}

func NewByteListWriter() *ByteListWriter {
	list := NewByteList()
	return &ByteListWriter{ByteWriter: ByteWriter{
		p: &byteListWriterProvider{list: list}}, list: list}
}

func (w *ByteListWriter) GetBytes() []byte {
	return w.list.items
}

func (w *ByteListWriter) GetByteList() *ByteList {
	return w.list
}

type ByteListReader struct {
	ByteReader
	list *ByteList
}

func NewByteListReader() *ByteListReader {
	list := NewByteList()
	return &ByteListReader{ByteReader: ByteReader{
		p: &byteListReaderProvider{list: list}}, list: list}
}

func (r *ByteListReader) GetBytes() []byte {
	return r.list.items
}

func (r *ByteListReader) GetByteList() *ByteList {
	return r.list
}

type byteListReaderProvider struct {
	list *ByteList
	off  int
}

func (p *byteListReaderProvider) GetNextByte() (byte, bool) {
	if p.off >= len(p.list.items) {
		return 0, true
	}
	b := p.list.items[p.off]
	p.off++
	return b, false
}

func (p *byteListReaderProvider) ReadBytes(n int) []byte {
	if n <= 0 || p.off >= len(p.list.items) {
		return nil
	}
	end := p.off + n
	if end > len(p.list.items) {
		end = len(p.list.items)
	}
	out := make([]byte, end-p.off)
	copy(out, p.list.items[p.off:end])
	p.off = end
	return out
}

type byteListWriterProvider struct {
	list *ByteList
}

func (p *byteListWriterProvider) WriteNextByte(b byte) {
	p.list.Append(b)
}

func (p *byteListWriterProvider) WriteBytes(bs []byte) {
	p.list.items = append(p.list.items, bs...)
}

// NewByteReaderFromBytes returns a read-only byte stream over list (Final Bytes carrier).
func NewByteReaderFromBytes(list *ByteList) (*ByteReader, error) {
	if list == nil {
		return nil, fmt.Errorf("NewByteReaderFromBytes: nil ByteList")
	}
	return &ByteReader{p: &byteListReaderProvider{list: list}}, nil
}

// NewByteWriterFromBytes returns a write-only byte stream that appends to list (Final Bytes carrier).
func NewByteWriterFromBytes(list *ByteList) (*ByteWriter, error) {
	if list == nil {
		return nil, fmt.Errorf("NewByteWriterFromBytes: nil ByteList")
	}
	return &ByteWriter{p: &byteListWriterProvider{list: list}}, nil
}

// NewByteReaderFromSlice returns a read-only byte stream over a Go []byte.
func NewByteReaderFromSlice(b []byte) (*ByteReader, error) {
	return NewByteReaderFromReader(bytes.NewReader(b))
}

func bytesToByteList(b []byte) *ByteList {
	bl := NewByteList()
	bl.items = append(bl.items, b...)
	return bl
}

var (
	_ ByteReadProvider  = (*byteListReaderProvider)(nil)
	_ ByteWriteProvider = (*byteListWriterProvider)(nil)
)
