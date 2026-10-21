package core

import "fmt"

// batchReq and batchRes are internal worker pump types (opaque to embedders).
type batchReq struct {
	script   string
	timeoutS int
	restrict SecurityRestrictions // sender extra; ORed with Session.restrict in the worker
}

type batchRes struct {
	response string
	err      error
}

// Send runs one Final script batch synchronously (web-request semantics).
// timeoutS is the advisory run budget in whole seconds (0 = none); propagated to the
// session task for inward I/O (e.g. HttpReq when timeoutS unset). Does not preempt
// blocking native code by itself — see docs/003-design-notes/134-embedding-final.md.
// Blocks until the VM worker completes; concurrent Send calls serialize on sendMu.
func (s *Session) Send(script string, timeoutS int) (response string, err error) {
	return s.send(script, timeoutS, 0)
}

func (s *Session) send(script string, timeoutS int, extra SecurityRestrictions) (response string, err error) {
	if s == nil {
		return "", fmt.Errorf("session: nil Session")
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return "", ErrSessionClosed
	}
	s.mu.Unlock()

	if script == "" {
		return "", ErrEmptyScript
	}

	s.workerReq <- batchReq{script: script, timeoutS: timeoutS, restrict: extra}
	br := <-s.workerRes
	return br.response, br.err
}

func (s *Session) workerLoop() {
	defer close(s.workerDone)
	for req := range s.workerReq {
		s.task.advisoryTimeoutS = req.timeoutS
		s.mu.Lock()
		floor := s.restrict
		s.mu.Unlock()
		s.task.restrict = floor | req.restrict
		rr := s.runOnce(RunInputString{Script: req.script}, RunOutputCapture{})
		s.task.advisoryTimeoutS = 0
		s.workerRes <- batchRes{response: rr.Output, err: runResultError(rr)}
	}
}

func runResultError(r RunResult) error {
	switch r.Status {
	case RunCompleted:
		return nil
	case RunFaulted:
		if r.Fault != nil {
			return r.Fault
		}
		return fmt.Errorf("session: run faulted")
	case RunInterpretError:
		if r.Err != nil {
			return r.Err
		}
		return fmt.Errorf("session: interpret error")
	default:
		return fmt.Errorf("session: run failed")
	}
}

func startSessionWorker(s *Session) {
	s.workerReq = make(chan batchReq, 1)
	s.workerRes = make(chan batchRes, 1)
	s.workerDone = make(chan struct{})
	go s.workerLoop()
}
