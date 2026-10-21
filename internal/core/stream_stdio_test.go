package core

import (
	"bufio"
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestStdioWriterProviderConcurrentWriteString(t *testing.T) {
	var buf bytes.Buffer
	p := &stdioWriterProvider{
		w:     bufio.NewWriter(&buf),
		label: "test-stdout",
	}
	const (
		n = 500
		a = "AAAAAAAA"
		b = "BBBBBBBB"
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			p.WriteString(a)
			if i%50 == 0 {
				p.Flush()
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			p.WriteString(b)
			if i%50 == 0 {
				p.Flush()
			}
		}
	}()
	wg.Wait()
	p.Flush()

	s := buf.String()
	wantLen := n * (len(a) + len(b))
	if len(s) != wantLen {
		t.Fatalf("len = %d, want %d", len(s), wantLen)
	}
	// Each locked WriteString is atomic: output must be a concatenation of intact markers only.
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], a):
			i += len(a)
		case strings.HasPrefix(s[i:], b):
			i += len(b)
		default:
			end := i + 16
			if end > len(s) {
				end = len(s)
			}
			t.Fatalf("torn write at offset %d: %q", i, s[i:end])
		}
	}
}

func TestStdioWriterProviderConcurrentWriteLine(t *testing.T) {
	var buf bytes.Buffer
	p := &stdioWriterProvider{
		w:     bufio.NewWriter(&buf),
		label: "test-stdout",
	}
	const (
		n     = 300
		lineA = "line-AAAA"
		lineB = "line-BBBB"
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			p.WriteLine(lineA)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			p.WriteLine(lineB)
		}
	}()
	wg.Wait()
	p.Flush()

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 2*n {
		t.Fatalf("line count = %d, want %d", len(lines), 2*n)
	}
	for _, line := range lines {
		if line != lineA && line != lineB {
			t.Fatalf("torn or unexpected line %q", line)
		}
	}
}

func TestNewStdoutTextWriterUsesStdioProvider(t *testing.T) {
	w, err := NewStdoutTextWriter()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := w.p.(*stdioWriterProvider); !ok {
		t.Fatalf("Stdout provider = %T, want *stdioWriterProvider", w.p)
	}
	w2, err := NewStderrTextWriter()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := w2.p.(*stdioWriterProvider); !ok {
		t.Fatalf("Stderr provider = %T, want *stdioWriterProvider", w2.p)
	}
}
