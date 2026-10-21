package fyneui

import (
	"testing"
	"time"

	"github.com/androticus/final/internal/core"
)

func TestKeyAutoRepeatHoldVsTurboDelay(t *testing.T) {
	ar := newKeyAutoRepeat(nil)
	now := time.Now()
	if d := ar.initialDelayLocked("Left", now); d != keyARHoldDelay {
		t.Fatalf("fresh down delay %v, want hold %v", d, keyARHoldDelay)
	}
	ar.lastUpName = "Left"
	ar.lastUpAt = now.Add(-100 * time.Millisecond)
	if d := ar.initialDelayLocked("Left", now); d != keyARTurboDelay {
		t.Fatalf("quick same-key down delay %v, want turbo %v", d, keyARTurboDelay)
	}
	if d := ar.initialDelayLocked("Right", now); d != keyARHoldDelay {
		t.Fatalf("other-key down delay %v, want hold %v", d, keyARHoldDelay)
	}
	ar.lastUpAt = now.Add(-keyARTurboWindow - time.Millisecond)
	if d := ar.initialDelayLocked("Left", now); d != keyARHoldDelay {
		t.Fatalf("stale same-key down delay %v, want hold %v", d, keyARHoldDelay)
	}
}

func TestKeyAutoRepeatOrphanUpDoesNotCancel(t *testing.T) {
	ar := newKeyAutoRepeat(nil)
	ar.onDown("Left")
	if ar.remembered != "Left" {
		t.Fatalf("remembered %q, want Left", ar.remembered)
	}
	ar.onUp("Right")
	if ar.remembered != "Left" {
		t.Fatalf("orphan up cleared remembered %q", ar.remembered)
	}
	if ar.timer == nil {
		t.Fatal("orphan up stopped the timer")
	}
	ar.onUp("Left")
	if ar.remembered != "" {
		t.Fatalf("matching up left remembered %q", ar.remembered)
	}
	if ar.timer != nil {
		t.Fatal("matching up left timer running")
	}
	ar.close()
}

func TestKeyAutoRepeatFireMintsWithoutOnDown(t *testing.T) {
	n := 0
	var last string
	ar := newKeyAutoRepeat(&Window{})
	ar.setMintDown(func(_ any, params ...any) {
		n++
		last = string(params[0].(core.String))
	})
	ar.onDown("Space")
	ar.mu.Lock()
	ar.stopTimerLocked()
	ar.mu.Unlock()
	ar.fire()
	if n != 1 || last != "Space" {
		t.Fatalf("mint n=%d last=%q, want 1 Space", n, last)
	}
	if ar.remembered != "Space" {
		t.Fatalf("fire cleared remembered %q", ar.remembered)
	}
	ar.onUp("Space")
	ar.fire()
	if n != 1 {
		t.Fatalf("fire after up minted extra, n=%d", n)
	}
	ar.close()
}
