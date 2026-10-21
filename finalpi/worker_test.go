package finalpi

import (
	"testing"
	"time"
)

func TestWorkerStopAndDoubleStop(t *testing.T) {
	w := NewWorker()
	started := make(chan struct{})
	w.Go(func(cancel <-chan struct{}) {
		close(started)
		<-cancel
	})
	<-started
	if !w.Running() {
		t.Fatal("expected running")
	}
	if !w.Stop(time.Second) {
		t.Fatal("Stop timed out")
	}
	if w.Running() {
		t.Fatal("expected stopped")
	}
	if !w.Stop(time.Millisecond) {
		t.Fatal("second Stop should succeed")
	}
}

func TestWaitCancelTimeoutAndCancel(t *testing.T) {
	if got := WaitCancel(make(chan struct{}), 5*time.Millisecond); got != WaitTimedOut {
		t.Fatalf("want TimedOut got %v", got)
	}
	cancel := make(chan struct{})
	close(cancel)
	if got := WaitCancel(cancel, time.Second); got != WaitCanceled {
		t.Fatalf("want Canceled got %v", got)
	}
}

func TestWorkerCancelUnblocks(t *testing.T) {
	w := NewWorker()
	blocked := make(chan struct{})
	w.Go(func(cancel <-chan struct{}) {
		close(blocked)
		if WaitCancel(cancel, time.Minute) != WaitCanceled {
			t.Error("expected cancel")
		}
	})
	<-blocked
	if !w.Stop(time.Second) {
		t.Fatal("Stop timed out")
	}
}
