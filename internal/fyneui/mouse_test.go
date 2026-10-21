package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/androticus/final/internal/core"
)

func TestMouseEventProperties(t *testing.T) {
	if mouseEventType == nil {
		t.Fatal("MouseEvent type not registered")
	}
	ev := &desktop.MouseEvent{
		PointEvent: fyne.PointEvent{
			Position:         fyne.NewPos(10.8, 20.2),
			AbsolutePosition: fyne.NewPos(100, 200),
		},
		Button:   desktop.MouseButtonSecondary,
		Modifier: fyne.KeyModifierShift | fyne.KeyModifierControl,
	}
	pos := mouseEventType.GetPropertyMetadata("position").GetValue(ev).(core.Point)
	wantPos := eventPoint(fyne.NewPos(10.8, 20.2))
	if pos != wantPos {
		t.Fatalf("position = %v, want %v", pos, wantPos)
	}
	abs := mouseEventType.GetPropertyMetadata("absolutePosition").GetValue(ev).(core.Point)
	if abs != core.NewPoint(100, 200) {
		t.Fatalf("absolutePosition = %v, want (100, 200)", abs)
	}
	btn := mouseEventType.GetPropertyMetadata("button").GetValue(ev).(core.Int)
	if btn != core.Int(desktop.MouseButtonSecondary) {
		t.Fatalf("button = %v", btn)
	}
	mod := mouseEventType.GetPropertyMetadata("modifier").GetValue(ev).(core.Int)
	if mod != core.Int(fyne.KeyModifierShift|fyne.KeyModifierControl) {
		t.Fatalf("modifier = %v", mod)
	}
	for _, name := range []string{"position", "absolutePosition", "button", "modifier"} {
		pm := mouseEventType.GetPropertyMetadata(name)
		if pm == nil || pm.HasSet() {
			t.Fatalf("%s should be readonly", name)
		}
	}
	if mouseEventType.GetClassPropertyMetadata("btnPrimary").GetValue(nil).(core.Int) != core.Int(desktop.MouseButtonPrimary) {
		t.Fatal("MouseEvent.btnPrimary")
	}
	if mouseEventType.GetClassPropertyMetadata("keyShift").GetValue(nil).(core.Int) != core.Int(fyne.KeyModifierShift) {
		t.Fatal("MouseEvent.keyShift")
	}
}

func TestScrollEventProperties(t *testing.T) {
	if scrollEventType == nil {
		t.Fatal("ScrollEvent type not registered")
	}
	ev := &fyne.ScrollEvent{
		PointEvent: fyne.PointEvent{
			Position:         fyne.NewPos(3, 4),
			AbsolutePosition: fyne.NewPos(30, 40),
		},
		Scrolled: fyne.NewDelta(-12, 8),
	}
	pos := scrollEventType.GetPropertyMetadata("position").GetValue(ev).(core.Point)
	if pos != core.NewPoint(3, 4) {
		t.Fatalf("position = %v, want (3, 4)", pos)
	}
	abs := scrollEventType.GetPropertyMetadata("absolutePosition").GetValue(ev).(core.Point)
	if abs != core.NewPoint(30, 40) {
		t.Fatalf("absolutePosition = %v, want (30, 40)", abs)
	}
	scrolled := scrollEventType.GetPropertyMetadata("scrolled").GetValue(ev).(core.Point)
	if scrolled != core.NewPoint(-12, 8) {
		t.Fatalf("scrolled = %v, want (-12, 8)", scrolled)
	}
	for _, name := range []string{"position", "absolutePosition", "scrolled"} {
		pm := scrollEventType.GetPropertyMetadata(name)
		if pm == nil || pm.HasSet() {
			t.Fatalf("%s should be readonly", name)
		}
	}
}
