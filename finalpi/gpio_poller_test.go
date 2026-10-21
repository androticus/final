package finalpi

import (
	"testing"
	"time"
)

func TestGpioPollerStartsAndStops(t *testing.T) {
	p := newGpioPoller()
	pin := newGpioPin(99)
	pin.active = true
	pin.Direction = GpioIn
	pin.DebounceMs = DebounceSensor
	p.syncWatch(pin)
	deadline := time.Now().Add(200 * time.Millisecond)
	for !p.worker.Running() {
		if time.Now().After(deadline) {
			t.Fatal("expected poller running")
		}
		time.Sleep(time.Millisecond)
	}
	p.remove(pin)
	deadline = time.Now().Add(time.Second)
	for p.worker.Running() {
		if time.Now().After(deadline) {
			t.Fatal("expected poller stopped")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestGpioPinWantsWatch(t *testing.T) {
	pin := newGpioPin(4)
	if pin.wantsWatch() {
		t.Fatal("inactive should not watch")
	}
	pin.active = true
	pin.Direction = GpioIn
	if pin.wantsWatch() {
		t.Fatal("active input alone should not watch")
	}
	pin.DebounceMs = DebounceButton
	if !pin.wantsWatch() {
		t.Fatal("debounce should watch")
	}
	pin.DebounceMs = DebounceOff
	pin.onHigh = func(source any, params ...any) {}
	if !pin.wantsWatch() {
		t.Fatal("handler should watch")
	}
}

func TestGpioCtrlResetStopsPoller(t *testing.T) {
	c := GpioCtrlInst()
	c.Reset()
	pin := c.gpio4
	pin.active = true
	pin.Direction = GpioIn
	pin.DebounceMs = DebounceSensor
	c.poller.syncWatch(pin)
	deadline := time.Now().Add(200 * time.Millisecond)
	for !c.poller.worker.Running() {
		if time.Now().After(deadline) {
			t.Fatal("expected poller running")
		}
		time.Sleep(time.Millisecond)
	}
	c.Reset()
	deadline = time.Now().Add(time.Second)
	for c.poller.worker.Running() {
		if time.Now().After(deadline) {
			t.Fatal("Reset should stop poller")
		}
		time.Sleep(time.Millisecond)
	}
}
