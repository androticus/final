package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestRectHitArea_NewAndReadonly(t *testing.T) {
	r := core.NewRect(core.Point{X: 1, Y: 2}, core.Size{W: 3, H: 4})
	h := NewRectHitArea(3, r)
	if h.Shape() != HitAreaShapeRect {
		t.Fatalf("shape=%v", h.Shape())
	}
	if h.Kind() != 3 {
		t.Fatalf("kind=%v", h.Kind())
	}
	if h.Rect() != r {
		t.Fatalf("rect=%v want %v", h.Rect(), r)
	}
	if h.FinalType() != rectHitAreaType || rectHitAreaType == nil {
		t.Fatal("RectHitArea type not registered")
	}
	pm := rectHitAreaType.GetPropertyMetadata("rect")
	if pm == nil || pm.HasSet() {
		t.Fatal("rect property should be readonly")
	}
	if km := rectHitAreaType.GetPropertyMetadata("kind"); km == nil || km.HasSet() {
		t.Fatal("kind property should be readonly")
	}
	if sm := rectHitAreaType.GetPropertyMetadata("shape"); sm == nil || sm.HasSet() {
		t.Fatal("shape property should be readonly")
	}
}

func TestCircleHitArea_NewAndReadonly(t *testing.T) {
	c := core.Point{X: 10, Y: 20}
	h := NewCircleHitArea(HitKindBorder, c, 5.5)
	if h.Shape() != HitAreaShapeCircle {
		t.Fatalf("shape=%v", h.Shape())
	}
	if h.Kind() != HitKindBorder {
		t.Fatalf("kind=%v", h.Kind())
	}
	if h.Center() != c || h.Radius() != 5.5 {
		t.Fatalf("center=%v radius=%v", h.Center(), h.Radius())
	}
	if h.FinalType() != circleHitAreaType || circleHitAreaType == nil {
		t.Fatal("CircleHitArea type not registered")
	}
	pm := circleHitAreaType.GetPropertyMetadata("radius")
	if pm == nil || pm.HasSet() {
		t.Fatal("radius property should be readonly")
	}
}
