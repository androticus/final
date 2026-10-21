package core

import (
	"bytes"
	"strings"
	"sync"
)

// StringLineBuffer is a specialized, efficient component for buffering a
// fixed number of string lines, typically for output to a gui control
// that will only display the last N lines.
type StringLineBuffer struct {
	lines    int
	maxLines int
	ticks    *AtomicInt
	buffer   []byte
}

type StringLineBufferer interface {
	Append(value string)
	AppendLine(line string)
	ToString() string
	GetTick() int
}

func NewStringLineBuffer(maxLines int) *StringLineBuffer {
	if maxLines < 1 {
		panic("maxLines must be at least 1")
	}
	return &StringLineBuffer{maxLines: maxLines, ticks: NewAtomicInt(), buffer: make([]byte, 0)}
}

func NewStringLineBufferSync(maxLines int) *StringLineBufferSync	 {
	if maxLines < 1 {
		panic("maxLines must be at least 1")
	}
	return &StringLineBufferSync{StringLineBuffer: NewStringLineBuffer(maxLines), mu: sync.Mutex{}}
}

func (s *StringLineBuffer) Append(value string) {
	if value == "" {
		return
	}
	s.ticks.Inc()
	k := countNL(value)
	n := s.maxLines
	if k == 0 {
		s.buffer = append(s.buffer, value...)
		return
	}
	if k > n {
		// Input taller than the window: keep the tail after the (k-N)th newline.
		s.buffer = append(s.buffer[:0], value[indexAfterNthNLString(value, k-n):]...)
		s.lines = n
		return
	}
	if d := s.lines + k - n; d > 0 {
		s.buffer = dropLeadingLines(s.buffer, d)
		s.lines -= d
	}
	s.buffer = append(s.buffer, value...)
	s.lines += k
}

func (s *StringLineBuffer) AppendLine(line string) {
	s.Append(line)
	s.Append("\n")
}

func (s *StringLineBuffer) ToString() string {
	return string(s.buffer)
}

func (s *StringLineBuffer) GetTick() int {
	return s.ticks.Value()
}

func countNL(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

// indexAfterNthNLString is the byte index after the nth 0x0A in s (nth >= 1).
func indexAfterNthNLString(s string, nth int) int {
	off := 0
	for i := 0; i < nth; i++ {
		j := strings.IndexByte(s[off:], '\n')
		if j < 0 {
			panic("StringLineBuffer: newline count mismatch")
		}
		off += j + 1
	}
	return off
}

func dropLeadingLines(buf []byte, d int) []byte {
	off := 0
	for i := 0; i < d; i++ {
		j := bytes.IndexByte(buf[off:], '\n')
		if j < 0 {
			panic("StringLineBuffer: newline count mismatch")
		}
		off += j + 1
	}
	n := copy(buf, buf[off:])
	return buf[:n]
}

type StringLineBufferSync struct {
	*StringLineBuffer
	mu sync.Mutex
}

func (s *StringLineBufferSync) Append(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StringLineBuffer.Append(value)
}

func (s *StringLineBufferSync) AppendLine(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StringLineBuffer.AppendLine(line)
}

func (s *StringLineBufferSync) ToString() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.StringLineBuffer.ToString()
}
