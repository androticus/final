package core

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type stdinProvider struct {
	r *bufio.Reader
}

func (p *stdinProvider) GetNextRune() (rune, bool) {
	return readNextRuneFromBufio(p.r, "stdin")
}

type stdioWriterProvider struct {
	mu    sync.Mutex
	w     *bufio.Writer
	label string
}

func (p *stdioWriterProvider) WriteRune(r rune) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.w.WriteRune(r); err != nil {
		panic(fmt.Errorf("%s: write rune: %w", p.label, err))
	}
}

func (p *stdioWriterProvider) WriteLine(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.w.WriteString(line); err != nil {
		panic(fmt.Errorf("%s: write line: %w", p.label, err))
	}
	if _, err := p.w.WriteRune('\n'); err != nil {
		panic(fmt.Errorf("%s: write line newline: %w", p.label, err))
	}
}

func (p *stdioWriterProvider) WriteString(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.w.WriteString(s); err != nil {
		panic(fmt.Errorf("%s: write string: %w", p.label, err))
	}
}

func (p *stdioWriterProvider) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.w.Flush(); err != nil {
		panic(fmt.Errorf("%s: flush: %w", p.label, err))
	}
}

func NewStdinTextReader() (*TextReader, error) {
	return &TextReader{p: &stdinProvider{r: bufio.NewReader(os.Stdin)}}, nil
}

func NewStdoutTextWriter() (*TextWriter, error) {
	return &TextWriter{p: &stdioWriterProvider{
		w:     bufio.NewWriter(os.Stdout),
		label: "stdout",
	}}, nil
}

func NewStderrTextWriter() (*TextWriter, error) {
	return &TextWriter{p: &stdioWriterProvider{
		w:     bufio.NewWriter(os.Stderr),
		label: "stderr",
	}}, nil
}

// NewRunProviderWithProcessStdio returns a RunProvider whose stdin/stdout/stderr
// are wired to os.Stdin, os.Stdout, and os.Stderr.
func NewRunProviderWithProcessStdio() (*RunProvider, error) {
	rp := NewRunProvider()
	var err error
	rp.Stdin, err = NewStdinTextReader()
	if err != nil {
		return nil, err
	}
	rp.Stdout, err = NewStdoutTextWriter()
	if err != nil {
		return nil, err
	}
	rp.Stderr, err = NewStderrTextWriter()
	if err != nil {
		return nil, err
	}
	return rp, nil
}

var (
	_ TextReadProvider  = (*stdinProvider)(nil)
	_ TextWriteProvider = (*stdioWriterProvider)(nil)
	_ Flushable         = (*stdioWriterProvider)(nil)
)
