package finaltui

import (
	"testing"

	"github.com/androticus/final/internal/core"
	"github.com/gdamore/tcell/v2"
)

func TestEscFnCodeFromTcell(t *testing.T) {
	cases := []struct {
		key  tcell.Key
		want int
	}{
		{tcell.KeyEscape, 0},
		{tcell.KeyF1, 1},
		{tcell.KeyF5, 5},
		{tcell.KeyF12, 12},
	}
	for _, tc := range cases {
		ev := tcell.NewEventKey(tc.key, 0, tcell.ModNone)
		n, ok := escFnCodeFromTcell(ev)
		if !ok || n != tc.want {
			t.Fatalf("%v → %d %v, want %d true", tc.key, n, ok, tc.want)
		}
	}
	ev := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	if _, ok := escFnCodeFromTcell(ev); ok {
		t.Fatal("rune should not map")
	}
}

func TestGlobalCaptureLatchesFn(t *testing.T) {
	core.ClearEscFn()
	ui := &appUI{}
	ev := tcell.NewEventKey(tcell.KeyF3, 0, tcell.ModNone)
	if got := ui.globalCapture(ev); got != nil {
		t.Fatal("F3 should be consumed")
	}
	n, ok := core.TakeEscFn()
	if !ok || n != 3 {
		t.Fatalf("latch = %d %v, want 3 true", n, ok)
	}
}

func TestGlobalCaptureLatchesEscPassThrough(t *testing.T) {
	core.ClearEscFn()
	ui := &appUI{}
	ev := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	if got := ui.globalCapture(ev); got != ev {
		t.Fatal("Esc should pass through for dialogs/menus")
	}
	n, ok := core.TakeEscFn()
	if !ok || n != 0 {
		t.Fatalf("latch = %d %v, want 0 true", n, ok)
	}
}
