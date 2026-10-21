package core

import (
	"fmt"
	"io"
	"os"
)

// Break requests cooperative stop of an in-flight VM run on this session's task.
// See Task.Break; no-op when the task is not inside runLoop.
func (s *Session) Break() {
	if s != nil && s.task != nil {
		s.task.Break()
	}
}

// RunREPL runs the interactive terminal REPL on this session's task (same module
// dictionary as Send) until stdin EOF or bye. Uses process stdin/stdout/stderr.
func (s *Session) RunREPL() error {
	if s == nil {
		return fmt.Errorf("RunREPL: nil Session")
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	s.mu.Unlock()

	t := s.task
	t.name = "repl"
	rp, err := NewRunProviderWithProcessStdio()
	if err != nil {
		return err
	}
	bootstrapWriteREPLBanner(rp)
	return BootstrapREPL(t, rp)
}

// RunEventLoop runs EventLoop#run as this session's outer loop (GUI/TUI worker).
// Holds sendMu for the duration so Send cannot interleave. Persistent sessions only.
// rootWord is the trampoline colon (e.g. "guiRoot", "tuiRoot"); empty uses "guiRoot".
func (s *Session) RunEventLoop(rootWord string) error {
	if s == nil {
		return fmt.Errorf("RunEventLoop: nil Session")
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	s.mu.Unlock()
	if !s.persistent || s.rp == nil {
		return fmt.Errorf("RunEventLoop: persistent Session required")
	}
	t := s.task
	t.name = "eventLoop"
	rp := s.rp
	// Keep a host SetStderr (e.g. final-gui standing log). Only fall back to
	// process stderr when still on the null writer.
	if rp.Stderr == nil || rp.Stderr == GetNullTextWriter() {
		if stderr, err := NewTextWriterFromWriter(os.Stderr); err == nil {
			rp.Stderr = stderr
		}
	}
	RunEventLoopSession(t, rp, rootWord)
	return nil
}

// InterpretReader interprets Final source from r on process stdout until EOF.
// No REPL banner. Faults and interpret errors are returned (caller prints / exits).
func (s *Session) InterpretReader(r io.Reader) error {
	return s.InterpretReaderPath(r, "")
}

// InterpretReaderPath is InterpretReader with an optional source path for fault reports.
func (s *Session) InterpretReaderPath(r io.Reader, path string) error {
	if s == nil {
		return fmt.Errorf("InterpretReader: nil Session")
	}
	if r == nil {
		return ErrNilReader
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	h, err := s.Run(RunInputReader{Source: r, Path: path}, RunOutputWriter{Sink: os.Stdout})
	if err != nil {
		return err
	}
	return runResultError(h.Wait())
}
