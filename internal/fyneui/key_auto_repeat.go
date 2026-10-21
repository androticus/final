package fyneui

import (
	"sync"
	"time"

	"github.com/androticus/final/internal/core"
)

// Built-in timings (not author options). 180.1-constructs.md Keys / autoRepeat.
const (
	keyARHoldDelay   = 400 * time.Millisecond
	keyARTurboDelay  = 200 * time.Millisecond
	keyARInterval    = 50 * time.Millisecond
	keyARTurboWindow = 400 * time.Millisecond
)

// keyAutoRepeat is a Go adjunct on Window. Not a Final catalog type.
// Last-key only: one remembered name, hold vs turbo delay, then interval synthetics.
type keyAutoRepeat struct {
	mu         sync.Mutex
	win        *Window
	mintDown   core.HandlerMintEventFn
	remembered string
	lastUpName string
	lastUpAt   time.Time
	timer      *time.Timer
}

func newKeyAutoRepeat(win *Window) *keyAutoRepeat {
	return &keyAutoRepeat{win: win}
}

func (ar *keyAutoRepeat) setMintDown(fn core.HandlerMintEventFn) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.stopTimerLocked()
	ar.remembered = ""
	ar.mintDown = fn
}

func (ar *keyAutoRepeat) close() {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.stopTimerLocked()
	ar.remembered = ""
	ar.mintDown = nil
}

func (ar *keyAutoRepeat) onDown(name string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.stopTimerLocked()
	ar.remembered = name
	ar.armLocked(ar.initialDelayLocked(name, time.Now()))
}

func (ar *keyAutoRepeat) onUp(name string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.lastUpName = name
	ar.lastUpAt = time.Now()
	if name == ar.remembered {
		ar.stopTimerLocked()
		ar.remembered = ""
	}
}

func (ar *keyAutoRepeat) initialDelayLocked(name string, now time.Time) time.Duration {
	if name == ar.lastUpName && !ar.lastUpAt.IsZero() && now.Sub(ar.lastUpAt) < keyARTurboWindow {
		return keyARTurboDelay
	}
	return keyARHoldDelay
}

func (ar *keyAutoRepeat) armLocked(d time.Duration) {
	ar.timer = time.AfterFunc(d, ar.fire)
}

func (ar *keyAutoRepeat) stopTimerLocked() {
	if ar.timer != nil {
		ar.timer.Stop()
		ar.timer = nil
	}
}

// fire mints a synthetic down through mintDown — not the Fyne KeyDown functor.
func (ar *keyAutoRepeat) fire() {
	ar.mu.Lock()
	name := ar.remembered
	mint := ar.mintDown
	win := ar.win
	if name == "" || mint == nil {
		ar.timer = nil
		ar.mu.Unlock()
		return
	}
	ar.armLocked(keyARInterval)
	ar.mu.Unlock()
	mint(win, core.String(name), core.Bool(true))
}
