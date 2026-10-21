package core

import (
	"math"
	"testing"
)

func TestPixmapViewOriginDefaultIdentity(t *testing.T) {
	pm := NewPixmap(8, 8)
	if got := pm.ViewOrigin(); got != (Point{}) {
		t.Fatalf("default ViewOrigin = %v, want zero", got)
	}
	if pm.viewOrigin != nil {
		t.Fatal("default viewOrigin pointer should be nil")
	}
	at := NewPoint(3, 4)
	if mapped := pm.MapWorldToView(at); mapped != at {
		t.Fatalf("identity MapWorldToView = %v, want %v", mapped, at)
	}
}

func TestPixmapSetViewOriginZeroClears(t *testing.T) {
	pm := NewPixmap(8, 8)
	pm.SetViewOrigin(NewPoint(10, 20))
	if pm.viewOrigin == nil {
		t.Fatal("expected non-nil after set")
	}
	pm.SetViewOrigin(Point{})
	if pm.viewOrigin != nil {
		t.Fatal("exact (0,0) should clear to nil")
	}
	// in Go, need to use a var to get -0.0
	negZero := math.Copysign(0.0, -1.0)
	pm.SetViewOrigin(NewPointF(negZero, 0.0))
	if pm.viewOrigin != nil {
		t.Fatal("signed zero should clear to nil")
	}
}

func TestPixmapMapWorldViewRoundTrip(t *testing.T) {
	pm := NewPixmap(8, 8)
	pm.SetViewOrigin(NewPoint(100, -50))
	world := NewPoint(130, -20)
	view := pm.MapWorldToView(world)
	if view != (Point{X: 30, Y: 30}) {
		t.Fatalf("MapWorldToView = %v, want (30, 30)", view)
	}
	back := pm.MapViewToWorld(view)
	if back != world {
		t.Fatalf("MapViewToWorld = %v, want %v", back, world)
	}
}

func TestBlitRespectsViewOrigin(t *testing.T) {
	pm := NewPixmap(4, 4)
	pm.Fill(NewColor(0, 0, 0, 255))
	mat := NewMat2i1(1, 1)
	mat.Put(0, 0, 1)
	fg := NewColor(0, 255, 0, 255)
	pal := NewPalette1(NewColor(0, 0, 0, 255), fg)

	// World (5, 5) with origin (4, 4) → pixmap (1, 1)
	pm.SetViewOrigin(NewPoint(4, 4))
	pm.Blit1(pal, mat, NewPoint(5, 5))
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(fg) {
		t.Fatalf("translated stamp at (1,1) = %v, want %v", got, colorToPremulRGBA(fg))
	}
	if got := pixmapStoredAt(pm, 0, 0); got.G == 255 {
		t.Fatal("stamp should not land at (0,0)")
	}
}

func TestPainterPixelRespectsViewOrigin(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	ink := NewColor(255, 0, 0, 255)
	p.Pixmap().SetViewOrigin(NewPoint(2, 3))
	// World (2, 3) → pixmap (0, 0)
	p.PixelC(ink, NewPoint(2, 3))
	if got := pixmapStoredAt(p.Pixmap(), 0, 0); got != colorToPremulRGBA(ink) {
		t.Fatalf("pixel at view (0,0) = %v, want %v", got, colorToPremulRGBA(ink))
	}
}

func TestPainterCircleRespectsViewOrigin(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	p.Stroke().SetLineStyle(LineStyleNone)
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(0, 255, 0, 255))
	p.Pixmap().SetViewOrigin(NewPoint(10, 10))
	// Filled circle centered at world (26, 26) → view (16, 16).
	p.Circle(NewPoint(26, 26), 3)
	if got := pixmapStoredAt(p.Pixmap(), 16, 16); got.A == 0 {
		t.Fatal("expected filled pixels at translated circle center (16,16)")
	}
	if got := pixmapStoredAt(p.Pixmap(), 26, 26); got.A != 0 {
		t.Fatal("world (26,26) must not paint at pixmap (26,26) when origin is (10,10)")
	}
}

func TestPathBuilderRespectsViewOrigin(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(1)
	p.Pixmap().SetViewOrigin(NewPoint(5, 5))
	b := p.PathBuilder()
	b.StartPath()
	b.MoveTo(5, 5)   // → (0, 0)
	b.LineTo(15, 5)  // → (10, 0)
	b.End()
	// Horizontal stroke along top row in view space
	found := false
	for x := 0; x < 10; x++ {
		if pixmapStoredAt(p.Pixmap(), 0, x).A != 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected path stroke along translated y=0")
	}
}

func TestPixmapSetViewOriginNativeFlex(t *testing.T) {
	ResetTestRuntime()
	pm := NewPixmap(4, 4)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pm)
	mm := pixmapType.GetMethodMetadata("setViewOrigin")
	if mm == nil {
		t.Fatal("missing Pixmap.setViewOrigin")
	}
	rp.DataStack.Push(Int(7))
	rp.DataStack.Push(Int(9))
	mm.invokeCall(f, ths)
	if got := pm.ViewOrigin(); got != NewPoint(7, 9) {
		t.Fatalf("setViewOrigin ints = %v, want (7, 9)", got)
	}
	rp.DataStack.Push(NewPoint(1, 2))
	mm.invokeCall(f, ths)
	if got := pm.ViewOrigin(); got != NewPoint(1, 2) {
		t.Fatalf("setViewOrigin Point = %v, want (1, 2)", got)
	}
}

func TestPixmapViewOriginProperty(t *testing.T) {
	pm := NewPixmap(4, 4)
	meta := pixmapType.GetPropertyMetadata("viewOrigin")
	if meta == nil {
		t.Fatal("missing Pixmap.viewOrigin")
	}
	meta.set(pm, NewPoint(3, 4))
	if got := meta.get(pm).(Point); got != NewPoint(3, 4) {
		t.Fatalf("property get = %v, want (3, 4)", got)
	}
	meta.set(pm, Point{})
	if pm.viewOrigin != nil {
		t.Fatal("property set zero should clear")
	}
	if got := meta.get(pm).(Point); got != (Point{}) {
		t.Fatalf("property get after clear = %v, want zero", got)
	}
}
