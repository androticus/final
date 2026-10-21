package core

import "testing"

func TestEventEscFnStartStopClose(t *testing.T) {
	ClearEscFn()
	got := make(chan int, 4)
	mm := &MethodMetadata{name: "h"} // not forgotten
	ev := &EventEscFn{
		userHandler: &MethodRef{MM: mm},
		mintFn:      func(_ any, params ...any) { got <- int(params[0].(Int)) },
	}
	linkEventEscFn(ev)
	defer ev.Close()

	LatchEscFn(3)
	select {
	case <-got:
		t.Fatal("should not deliver before Start")
	default:
	}
	// Latch still holds for pollers.
	if n, ok := TakeEscFn(); !ok || n != 3 {
		t.Fatalf("TakeEscFn after ignored latch = %d %v", n, ok)
	}

	ev.Start()
	LatchEscFn(7)
	select {
	case n := <-got:
		if n != 7 {
			t.Fatalf("got %d want 7", n)
		}
	default:
		t.Fatal("expected deliver after Start")
	}
	TakeEscFn() // clear poll latch

	ev.Stop()
	LatchEscFn(1)
	select {
	case <-got:
		t.Fatal("should not deliver after Stop")
	default:
	}
	TakeEscFn()

	ev.Start()
	LatchEscFn(EscFnEsc)
	select {
	case n := <-got:
		if n != EscFnEsc {
			t.Fatalf("got %d want Esc", n)
		}
	default:
		t.Fatal("expected deliver after restart")
	}
	TakeEscFn()

	ev.Close()
	LatchEscFn(2)
	select {
	case <-got:
		t.Fatal("should not deliver after Close")
	default:
	}
}

func TestEventEscFnCloseUnlinks(t *testing.T) {
	ClearEscFn()
	got := make(chan int, 1)
	mm := &MethodMetadata{name: "h"}
	ev := &EventEscFn{
		userHandler: &MethodRef{MM: mm},
		mintFn:      func(_ any, params ...any) { got <- int(params[0].(Int)) },
	}
	linkEventEscFn(ev)
	ev.Start()
	ev.Close()
	ev.Close() // idempotent
	LatchEscFn(4)
	select {
	case <-got:
		t.Fatal("closed sub must not receive")
	default:
	}
	TakeEscFn()
}

func TestEventEscFnOutOfRangeNoNotify(t *testing.T) {
	ClearEscFn()
	got := make(chan int, 1)
	mm := &MethodMetadata{name: "h"}
	ev := &EventEscFn{
		userHandler: &MethodRef{MM: mm},
		mintFn:      func(_ any, params ...any) { got <- int(params[0].(Int)) },
	}
	linkEventEscFn(ev)
	defer ev.Close()
	ev.Start()
	LatchEscFn(99)
	select {
	case <-got:
		t.Fatal("out-of-range must not notify")
	default:
	}
}

func TestEventEscFnDropsForgottenHandler(t *testing.T) {
	ClearEscFn()
	got := make(chan int, 1)
	mm := &MethodMetadata{name: "handleEscFn", forgotten: true}
	ev := &EventEscFn{
		userHandler: &MethodRef{MM: mm},
		mintFn:      func(_ any, params ...any) { got <- int(params[0].(Int)) },
	}
	linkEventEscFn(ev)
	ev.Start()
	LatchEscFn(5)
	select {
	case <-got:
		t.Fatal("forgotten MethodRef must not mint")
	default:
	}
	if !ev.closed {
		t.Fatal("expected auto-close after forgotten deliver")
	}
	TakeEscFn()
}
