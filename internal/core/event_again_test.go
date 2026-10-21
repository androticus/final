package core

import "testing"

func TestEventAgainMintsAndClose(t *testing.T) {
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	mm := &MethodMetadata{name: "h"}
	handler := &MethodRef{MM: mm}
	ea := NewEventAgain(handler)

	ea.Again()
	select {
	case v := <-TheEventLoop.Chan():
		ev, ok := v.(*Event)
		if !ok {
			t.Fatalf("got %T, want *Event", v)
		}
		if ev.Source != ea {
			t.Fatalf("source = %v, want EventAgain", ev.Source)
		}
		if ev.Handler == nil || ev.Handler.MM == nil || ev.Handler.MM.name != "_onFire" {
			t.Fatalf("wire handler = %v, want #_onFire", ev.Handler)
		}
		if ev.Params == nil || ev.Params.Len() != 0 {
			t.Fatalf("params len = %v, want empty", ev.Params)
		}
	default:
		t.Fatal("expected Event on loop pipe after Again")
	}

	ea.Close()
	ea.Close() // idempotent
	ea.Again()
	select {
	case <-TheEventLoop.Chan():
		t.Fatal("Again after Close must not mint")
	default:
	}
}

func TestEventAgainRearm(t *testing.T) {
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	mm := &MethodMetadata{name: "h"}
	ea := NewEventAgain(&MethodRef{MM: mm})
	ea.Again()
	ea.Again()
	n := 0
drain:
	for {
		select {
		case <-TheEventLoop.Chan():
			n++
		default:
			break drain
		}
	}
	if n != 2 {
		t.Fatalf("minted %d events, want 2", n)
	}
}
