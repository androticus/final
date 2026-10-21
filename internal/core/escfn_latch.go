package core

import "sync/atomic"

// Esc / Fn key press latch for headed hosts (final-tui, final-gui).
// Last press wins. TakeEscFn is one-shot (clears). Not a held-key model —
// terminals do not reliably report key-up.
//
// Encoding: 0 = Esc, 1..12 = F1..F12. Empty latch is -1.
//
// Hosts only map keys → LatchEscFn. Polling uses TakeEscFn / clearEscFn;
// EventEscFn subscribers are notified from LatchEscFn when started.

const (
	escFnLatchEmpty /*int32*/ = -1
	EscFnEsc              = 0
	EscFnF1               = 1
	EscFnF12              = 12
)

var escFnLatch atomic.Int32

func init() {
	escFnLatch.Store(escFnLatchEmpty)
}

// ClearEscFn discards any latched Esc/Fn press. Does not mint Events.
func ClearEscFn() {
	escFnLatch.Store(escFnLatchEmpty)
}

// LatchEscFn records a press and notifies started EventEscFn subscribers.
// n must be 0 (Esc) or 1..12 (F1..F12); other values ignored.
func LatchEscFn(n int) {
	if n < EscFnEsc || n > EscFnF12 {
		return
	}
	escFnLatch.Store(int32(n))
	notifyEventEscFnSubs(n)
}

// TakeEscFn returns the latched code and true, or false if empty. Always clears.
func TakeEscFn() (int, bool) {
	n := escFnLatch.Swap(escFnLatchEmpty)
	if n < EscFnEsc {
		return 0, false
	}
	return int(n), true
}
