package finalpi

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadWhenReadyImmediateWhenQueued(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	rx := newUartPeriph("uartReadyImm", 96, "pipe-b", 4, 5, 6, 7)
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	defer rx.release()

	var fired atomic.Bool
	rx.onReadReady = func(_ any, _ ...any) { fired.Store(true) }

	// Push directly into queue (simulates RX already buffered).
	rx.rxQ.push([]byte{0x41})

	if err := rx.readWhenReady(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && !fired.Load() {
		time.Sleep(5 * time.Millisecond)
	}
	if !fired.Load() {
		t.Fatal("expected onReadReady when data already queued")
	}
	rx.rxMu.Lock()
	armed := rx.armed
	rx.rxMu.Unlock()
	if armed {
		t.Fatal("should disarm after fire")
	}
}

func TestReadWhenReadyNopWhenAlreadyArmed(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	rx := newUartPeriph("uartReadyNop", 97, "pipe-b", 4, 5, 6, 7)
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	defer rx.release()

	var n atomic.Int32
	rx.onReadReady = func(_ any, _ ...any) { n.Add(1) }

	if err := rx.readWhenReady(); err != nil {
		t.Fatal(err)
	}
	if err := rx.readWhenReady(); err != nil {
		t.Fatal(err)
	}
	// Deliver one byte after double-arm.
	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = a.Write([]byte{1})
	}()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && n.Load() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	if n.Load() != 1 {
		t.Fatalf("fires=%d want 1 (second arm is nop)", n.Load())
	}
}

func TestReadWhenReadyRequiresHandler(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	rx := newUartPeriph("uartReadyNoH", 98, "pipe-b", 4, 5, 6, 7)
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	defer rx.release()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if !strings.Contains(fmt.Sprint(r), "onReadReady") {
			t.Fatalf("panic=%v", r)
		}
	}()
	_ = rx.readWhenReady()
}

func TestReadWhenReadyClearedOnRelease(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	rx := newUartPeriph("uartReadyRel", 99, "pipe-b", 4, 5, 6, 7)
	if err := rx.activateWithPort(pipePort{b}); err != nil {
		t.Fatal(err)
	}
	var fired atomic.Bool
	rx.onReadReady = func(_ any, _ ...any) { fired.Store(true) }
	if err := rx.readWhenReady(); err != nil {
		t.Fatal(err)
	}
	rx.release()
	_, _ = a.Write([]byte{1})
	time.Sleep(50 * time.Millisecond)
	if fired.Load() {
		t.Fatal("must not mint onReadReady after release")
	}
	rx.rxMu.Lock()
	defer rx.rxMu.Unlock()
	if rx.onReadReady != nil || rx.armed {
		t.Fatal("release should clear handler and armed")
	}
}

func TestReadWhenReadyNotActiveModal(t *testing.T) {
	rx := newUartPeriph("uartReadyOff", 100, "none", 4, 5, 6, 7)
	rx.onReadReady = func(_ any, _ ...any) {}
	err := rx.readWhenReady()
	if err == nil {
		t.Fatal("expected not-active error")
	}
}
