package core

import (
	"testing"
	"time"
)

func TestStopwatchNewIdle(t *testing.T) {
	sw := NewStopwatch()
	if sw.running {
		t.Fatal("new Stopwatch should not be running")
	}
	if sw.Elapsed() != 0 {
		t.Fatalf("new Stopwatch elapsed = %v, want 0", sw.Elapsed())
	}
}

func TestStopwatchLiveElapsedThenStop(t *testing.T) {
	sw := NewStopwatch()
	sw.Start()
	time.Sleep(20 * time.Millisecond)
	live := sw.Elapsed()
	if live < 15*time.Millisecond {
		t.Fatalf("live elapsed = %v, want >= 15ms", live)
	}
	sw.Stop()
	frozen := sw.Elapsed()
	time.Sleep(20 * time.Millisecond)
	if sw.Elapsed() != frozen {
		t.Fatalf("after stop, elapsed changed: got %v want %v", sw.Elapsed(), frozen)
	}
	if frozen < 15*time.Millisecond {
		t.Fatalf("frozen elapsed = %v, want >= 15ms", frozen)
	}
}

func TestStopwatchStartZerosElapsed(t *testing.T) {
	sw := NewStopwatch()
	sw.Start()
	time.Sleep(20 * time.Millisecond)
	sw.Stop()
	if sw.Elapsed() == 0 {
		t.Fatal("expected non-zero elapsed after first run")
	}
	sw.Start()
	if sw.Elapsed() > 5*time.Millisecond {
		t.Fatalf("start should zero elapsed; got %v", sw.Elapsed())
	}
	if !sw.running {
		t.Fatal("expected running after start")
	}
}

func TestStopwatchStopNoopWhenIdle(t *testing.T) {
	sw := NewStopwatch()
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	sw.Stop()
	frozen := sw.Elapsed()
	sw.Stop() // no-op
	if sw.Elapsed() != frozen {
		t.Fatalf("idle stop changed elapsed: got %v want %v", sw.Elapsed(), frozen)
	}
	if sw.running {
		t.Fatal("should remain not running")
	}
}
