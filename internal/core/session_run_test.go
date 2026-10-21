package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewSessionLocalUsesLocalModule(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if s.task.compileModID <= CoreModuleSlot {
		t.Fatalf("compileModID = %d, want > %d", s.task.compileModID, CoreModuleSlot)
	}
	if s.task.dc == globalDictionaryContext {
		t.Fatal("local session must not use core dc")
	}
}

func TestSessionRunScriptAdd(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	res := s.RunScript("1 2 +")
	if res.Status != RunCompleted {
		t.Fatalf("status = %v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
}

func TestSessionRunScriptWithOutputDsDollar(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	res := s.RunScriptWithOutput("1 2 ds$")
	if res.Status != RunCompleted {
		t.Fatalf("status = %v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
	if res.Output != "1 2 " {
		t.Fatalf("output = %q, want %q", res.Output, "1 2 ")
	}
}

func TestSessionRunNilInputNoOp(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	h, err := s.Run(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	res := h.Wait()
	if res.Status != RunCompleted {
		t.Fatalf("status = %v", res.Status)
	}
}

func TestSessionRunEmptyScriptError(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Run(RunInputString{Script: ""}, nil)
	if err != ErrEmptyScript {
		t.Fatalf("err = %v, want ErrEmptyScript", err)
	}
}

func TestSessionRunNilReaderError(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.Run(RunInputReader{Source: nil}, nil)
	if err != ErrNilReader {
		t.Fatalf("err = %v, want ErrNilReader", err)
	}
}

func TestSessionRunInterpretError(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	res := s.RunScript("nosuchword")
	if res.Status != RunInterpretError {
		t.Fatalf("status = %v, want RunInterpretError", res.Status)
	}
}

func TestSessionRunFault(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	res := s.RunScript("1 0 /")
	if res.Status != RunFaulted {
		t.Fatalf("status = %v, want RunFaulted", res.Status)
	}
	if res.Fault == nil {
		t.Fatal("expected fault")
	}
}

func TestSessionMultiStepCompileAndRun(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	h, err := s.Run(RunInputString{Script: ": double 2 * ;"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h.Wait().Status != RunCompleted {
		t.Fatal("define failed")
	}

	res := s.RunScript("4 double")
	if res.Status != RunCompleted {
		t.Fatalf("run double: status=%v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
}

func TestSessionCompileIntoGlobalDictionary(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	h, err := s.Run(RunInputString{Script: ": sessionword 99 ;"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h.Wait().Status != RunCompleted {
		t.Fatal("define failed")
	}
	if globalDictionaryContext.dict.GetByName("sessionword") == nil {
		t.Fatal("colon def not in global dictionary")
	}
	_, _, ok := FindNameInRun(NewTestRunProvider(), "sessionword")
	if !ok {
		t.Fatal("word not found via task dictionary context")
	}
}

func TestSessionOutputWriter(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	var buf bytes.Buffer
	h, err := s.Run(RunInputString{Script: "1 2 ds$"}, RunOutputWriter{Sink: &buf})
	if err != nil {
		t.Fatal(err)
	}
	res := h.Wait()
	if res.Status != RunCompleted {
		t.Fatalf("status = %v", res.Status)
	}
	if buf.String() != "1 2 " {
		t.Fatalf("buf = %q", buf.String())
	}
}

func TestSessionInputPipe(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	inParam, clientW := NewRunInputPipe()
	h, err := s.Run(inParam, RunOutputCapture{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := clientW.Write([]byte("1 2 +\n")); err != nil {
		t.Fatal(err)
	}
	clientW.Close()
	res := h.Wait()
	if res.Status != RunCompleted {
		t.Fatalf("status = %v fault=%v err=%v", res.Status, res.Fault, res.Err)
	}
}

func TestSessionClose(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	id := s.StatsID()
	if SessionStatsSnapshot(id) == nil {
		t.Fatal("expected stats")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if SessionStatsSnapshot(id) != nil {
		t.Fatal("stats should be unregistered")
	}
	_, err = s.Run(RunInputString{Script: "1"}, nil)
	if err != ErrSessionClosed {
		t.Fatalf("err = %v", err)
	}
}

func TestSessionRunSourceReader(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	src := strings.NewReader("3 4 +")
	h, err := s.Run(RunInputReader{Source: src}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h.Wait().Status != RunCompleted {
		t.Fatal("expected completed")
	}
}

func TestNewSessionSetsTaskRestrict(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindLocal, Restrict: SR_Threads})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.restrict != SR_Threads {
		t.Fatalf("session.restrict = %#x, want SR_Threads", s.restrict)
	}
	if s.task.restrict != SR_Threads {
		t.Fatalf("task.restrict = %#x, want SR_Threads %#x", s.task.restrict, SR_Threads)
	}
}
