package core

import "strings"

// TextRead is used to read Utf-8 encodedtext from a stream.
// It always returns valid whole Rune values.
type TextRead interface {
	// GetNextRune reads the next rune from the stream.
	// If the end of the stream is reached, the eof flag is set, and rune is not valid.
	GetNextRune() (r rune, eof bool) // note: go complains about ReadRune name
}

type TextWrite interface {
	// WriteRune writes a single rune to the stream
	WriteRune(rune)

	// WriteLine writes a single line to the stream, terminating with \n.
	// Any trailing whitespace or newlines is included.
	WriteLine(string)

	// WriteString writes a single string to the stream.
	// It does not append a newline.
	WriteString(string)
}

type TextReadWrite interface {
	TextRead
	TextWrite
}

// The Byte interfaces follow the same semantics as the Text ones
type ByteRead interface {
	GetNextByte() (b byte, eof bool)
	ReadBytes(n int) []byte
}

type ByteWrite interface {
	WriteNextByte(byte) // go complains about WriteByte name
	WriteBytes([]byte)
}

type ByteReadWrite interface {
	ByteRead
	ByteWrite
}

type Closeable interface {
	Close()
}

type Flushable interface {
	Flush()
}

/* not in v1
type Seekable interface {
	Seek(offset int64, whence int) (int64, error)
}*/

// Provider interfaces — implemented by sidecar stream_* provider types; use typed factories.

type TextReadProvider interface {
	TextRead
}

type TextWriteProvider interface {
	TextWrite
}

type TextReadWriteProvider interface {
	TextReadWrite
}

type ByteReadProvider interface {
	ByteRead
}

type ByteWriteProvider interface {
	ByteWrite
}

type ByteReadWriteProvider interface {
	ByteReadWrite
}

// The 6 application level objects. All the methods are below them.

// TextReader is a reader for Text.
type TextReader struct {
	p TextReadProvider
}

type TextWriter struct {
	p TextWriteProvider
}

type TextReaderWriter struct {
	p TextReadWriteProvider
}

type ByteReader struct {
	p ByteReadProvider
}

type ByteWriter struct {
	p ByteWriteProvider
}

type ByteReaderWriter struct {
	p ByteReadWriteProvider
}

// The methods are below the objects.

// TextReader methods

func (r *TextReader) GetNextRune() (rv rune, eof bool) {
	return r.p.GetNextRune()
}

// ReadLine reads one line (excluding \n; strips trailing \r). Returns eof when no more data.
func (r *TextReader) ReadLine() (line string, eof bool) {
	var sb strings.Builder
	gotAny := false
	for {
		ch, done := r.GetNextRune()
		if done {
			if !gotAny {
				return "", true
			}
			return sb.String(), false
		}
		gotAny = true
		if ch == '\n' {
			return strings.TrimSuffix(sb.String(), "\r"), false
		}
		sb.WriteRune(ch)
	}
}

func (r *TextReader) Close() {
	if c, ok := r.p.(Closeable); ok {
		c.Close()
	}
}

// TextWriter methods

func (w *TextWriter) WriteRune(r rune) {
	w.p.WriteRune(r)
}

func (w *TextWriter) WriteLine(line string) {
	w.p.WriteLine(line)
}

func (w *TextWriter) WriteString(s string) {
	w.p.WriteString(s)
}

func (w *TextWriter) Close() {
	if c, ok := w.p.(Closeable); ok {
		c.Close()
	}
}

func (w *TextWriter) Flush() {
	if f, ok := w.p.(Flushable); ok {
		f.Flush()
	}
}

// TextReaderWriter methods

func (rw *TextReaderWriter) GetNextRune() (rv rune, eof bool) {
	return rw.p.GetNextRune()
}

func (rw *TextReaderWriter) WriteRune(r rune) {
	rw.p.WriteRune(r)
}

func (rw *TextReaderWriter) WriteLine(line string) {
	rw.p.WriteLine(line)
}

func (rw *TextReaderWriter) WriteString(s string) {
	rw.p.WriteString(s)
}

func (rw *TextReaderWriter) Close() {
	if c, ok := rw.p.(Closeable); ok {
		c.Close()
	}
}

func (rw *TextReaderWriter) Flush() {
	if f, ok := rw.p.(Flushable); ok {
		f.Flush()
	}
}

// ByteReader methods

func (r *ByteReader) GetNextByte() (b byte, eof bool) {
	return r.p.GetNextByte()
}

func (r *ByteReader) ReadBytes(n int) []byte {
	return r.p.ReadBytes(n)
}

func (r *ByteReader) Close() {
	if c, ok := r.p.(Closeable); ok {
		c.Close()
	}
}

// ByteWriter methods

func (w *ByteWriter) WriteNextByte(b byte) {
	w.p.WriteNextByte(b)
}

func (w *ByteWriter) WriteBytes(bytes []byte) {
	w.p.WriteBytes(bytes)
}

func (w *ByteWriter) Close() {
	if c, ok := w.p.(Closeable); ok {
		c.Close()
	}
}

func (w *ByteWriter) Flush() {
	if f, ok := w.p.(Flushable); ok {
		f.Flush()
	}
}

// ByteReaderWriter methods

func (rw *ByteReaderWriter) GetNextByte() (b byte, eof bool) {
	return rw.p.GetNextByte()
}

func (rw *ByteReaderWriter) ReadBytes(n int) []byte {
	return rw.p.ReadBytes(n)
}

func (rw *ByteReaderWriter) WriteNextByte(b byte) {
	rw.p.WriteNextByte(b)
}

func (rw *ByteReaderWriter) WriteBytes(bytes []byte) {
	rw.p.WriteBytes(bytes)
}

func (rw *ByteReaderWriter) Close() {
	if c, ok := rw.p.(Closeable); ok {
		c.Close()
	}
}

func (rw *ByteReaderWriter) Flush() {
	if f, ok := rw.p.(Flushable); ok {
		f.Flush()
	}
}

// Compile-time interface checks.
var (
	_ TextRead      = (*TextReader)(nil)
	_ TextWrite     = (*TextWriter)(nil)
	_ TextReadWrite = (*TextReaderWriter)(nil)
	_ ByteRead      = (*ByteReader)(nil)
	_ ByteWrite     = (*ByteWriter)(nil)
	_ ByteReadWrite = (*ByteReaderWriter)(nil)
)
