package core

import "testing"

func TestRunFrameAdvisoryTimeoutS(t *testing.T) {
	task := NewTestTask()
	task.advisoryTimeoutS = 42
	f := &RunFrame{Task: task, RP: NewTestRunProvider()}
	if got := f.AdvisoryTimeoutS(); got != 42 {
		t.Fatalf("AdvisoryTimeoutS = %d, want 42", got)
	}
	if (&RunFrame{}).AdvisoryTimeoutS() != 0 {
		t.Fatal("nil task should return 0")
	}
}

func TestRunFrameBreakRequested(t *testing.T) {
	forceOp := OpCode(0)
	task := NewTestTask()
	task.forceOp = &forceOp
	f := &RunFrame{Task: task, RP: NewTestRunProvider(), ForceOp: &forceOp}

	if f.BreakRequested() {
		t.Fatal("want false when forceOp is 0")
	}
	*task.forceOp = OpStop
	if !f.BreakRequested() {
		t.Fatal("want true when forceOp is OpStop")
	}
	task.forceOp = nil
	if f.BreakRequested() {
		t.Fatal("want false when forceOp nil")
	}
}

func TestSessionWorkerClearsAdvisoryTimeoutAfterSend(t *testing.T) {
	ResetTestRuntime()
	s, err := NewSession(&SessionOptions{Kind: SessionKindCore})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.Send("1 2 +", 30); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if s.task.advisoryTimeoutS != 0 {
		t.Fatalf("advisoryTimeoutS = %d, want 0 after Send", s.task.advisoryTimeoutS)
	}
}
