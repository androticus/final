package finalpi

import (
	"sync"
	"time"

	"github.com/androticus/final/internal/core"
)

// Worker is a cancelable long-lived goroutine lifecycle (Phase 5 spine).
// Peripheral/poller code starts work with Go; Stop cancels and waits for exit.
type Worker struct {
	mu      sync.Mutex
	cancel  chan struct{}
	done    chan struct{}
	running bool
}

// NewWorker returns an idle worker (no goroutine yet).
func NewWorker() *Worker {
	return &Worker{}
}

// Go starts fn in a new goroutine. fn must return when cancel is closed.
// Panics if already running.
func (w *Worker) Go(fn func(cancel <-chan struct{})) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		panic("Worker.Go: already running")
	}
	w.cancel = make(chan struct{})
	w.done = make(chan struct{})
	w.running = true
	cancel := w.cancel
	done := w.done
	go func() {
		defer close(done)
		fn(cancel)
	}()
}

// Running reports whether a goroutine is active.
func (w *Worker) Running() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// Stop closes cancel and waits up to timeout for the goroutine to finish.
// Safe to call when not running or repeatedly. Returns false if join timed out.
func (w *Worker) Stop(timeout time.Duration) bool {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return true
	}
	cancel := w.cancel
	done := w.done
	w.mu.Unlock()

	select {
	case <-cancel:
	default:
		close(cancel)
	}

	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		w.mu.Lock()
		w.running = false
		w.cancel = nil
		w.done = nil
		w.mu.Unlock()
		return true
	case <-timer.C:
		return false
	}
}

// WaitOutcome is the result of WaitCancel.
type WaitOutcome int

const (
	WaitOK WaitOutcome = iota
	WaitTimedOut
	WaitCanceled
)

// WaitCancel waits for d or cancel. Used by spine clients instead of naked sleep/read.
func WaitCancel(cancel <-chan struct{}, d time.Duration) WaitOutcome {
	if d <= 0 {
		select {
		case <-cancel:
			return WaitCanceled
		default:
			return WaitOK
		}
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-cancel:
		return WaitCanceled
	case <-t.C:
		return WaitTimedOut
	}
}

// MintEvent wraps EventLoop.CreateEventFn for Pi workers (never Session.Send).
func MintEvent(handler *core.MethodRef, index int) core.HandlerMintEventFn {
	if handler == nil {
		return nil
	}
	return core.GetEventLoop().CreateEventFn(handler, index)
}
