package core

import (
	"image"
	"image/color"
	"testing"
)

func fillOnly(p *Painter) {
	p.Stroke().SetLineStyle(LineStyleNone)
}

func TestFillCheckerboardTiles(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleCheckerboard)
	f.SetColor(NewColor(255, 0, 0, 255))
	f.SetColor2(NewColor(0, 0, 255, 255))
	f.SetCheckerSize(8)
	p.Rect(NewPoint(16, 16), NewSize(32, 32))
	a := painterRGBA(p).RGBAAt(2, 2)
	b := painterRGBA(p).RGBAAt(10, 2)
	if a.R < 200 || a.B > 80 {
		t.Fatalf("tile (0,0) = %v, want red-dominant", a)
	}
	if b.B < 200 || b.R > 80 {
		t.Fatalf("tile (1,0) = %v, want blue-dominant", b)
	}
}

func TestFillGradientFromToLeftRight(t *testing.T) {
	p := NewPainter(NewSize(64, 16))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleGradient)
	f.SetColor(NewColor(255, 0, 0, 255))
	f.SetColor2(NewColor(0, 0, 255, 255))
	f.SetGradientFromTo(NewPoint(0, 8), NewPoint(64, 8))
	p.Rect(NewPoint(32, 8), NewSize(64, 16))
	left := painterRGBA(p).RGBAAt(4, 8)
	right := painterRGBA(p).RGBAAt(60, 8)
	if left.R <= right.R {
		t.Fatalf("left %v should be redder than right %v", left, right)
	}
	if right.B <= left.B {
		t.Fatalf("right %v should be bluer than left %v", right, left)
	}
}

func TestFillGradientDegreesZeroIsRightward(t *testing.T) {
	p := NewPainter(NewSize(64, 16))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleGradient)
	f.SetColor(NewColor(255, 0, 0, 255))
	f.SetColor2(NewColor(0, 0, 255, 255))
	f.SetGradientDegrees(0)
	p.Rect(NewPoint(32, 8), NewSize(64, 16))
	left := painterRGBA(p).RGBAAt(4, 8)
	right := painterRGBA(p).RGBAAt(60, 8)
	if left.R <= right.R {
		t.Fatalf("0°: left %v should be redder than right %v", left, right)
	}
}

func TestFillGradientClipsToPath(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleGradient)
	f.SetColor(NewColor(255, 0, 0, 255))
	f.SetColor2(NewColor(0, 0, 255, 255))
	f.SetGradientFromTo(NewPoint(-100, -100), NewPoint(200, 200))
	p.Circle(NewPoint(32, 32), 10)
	if painterRGBA(p).RGBAAt(0, 0).A != 0 {
		t.Fatal("pixel outside the circle must stay empty")
	}
	if painterRGBA(p).RGBAAt(32, 32).A == 0 {
		t.Fatal("circle interior should be filled")
	}
}

func TestFillTilePixmap(t *testing.T) {
	tile := NewPixmap(2, 2)
	tile.Put(0, 0, NewColor(255, 0, 0, 255))
	tile.Put(0, 1, NewColor(0, 255, 0, 255))
	tile.Put(1, 0, NewColor(0, 0, 255, 255))
	tile.Put(1, 1, NewColor(255, 255, 0, 255))

	p := NewPainter(NewSize(16, 16))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleTile)
	f.SetTileImage(tile)
	f.SetTileZoom(4)
	p.Rect(NewPoint(8, 8), NewSize(16, 16))

	c00 := painterRGBA(p).RGBAAt(1, 1)
	c10 := painterRGBA(p).RGBAAt(5, 1)
	if c00.R < 200 {
		t.Fatalf("zoomed tile (0,0) = %v, want red", c00)
	}
	if c10.G < 200 {
		t.Fatalf("zoomed tile (1,0) = %v, want green", c10)
	}
}

func TestFillTileAcceptsImage(t *testing.T) {
	nrgba := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	nrgba.SetNRGBA(0, 0, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
	img := &Image{Image: nrgba}

	p := NewPainter(NewSize(8, 8))
	fillOnly(p)
	f := p.Fill()
	f.SetStyle(FillStyleTile)
	f.SetTileImage(img)
	p.Rect(NewPoint(4, 4), NewSize(8, 8))
	got := painterRGBA(p).RGBAAt(3, 3)
	if got.G < 200 {
		t.Fatalf("tiled Image = %v, want green", got)
	}
}

func TestFillGradientMissingSpecPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	fillOnly(p)
	p.Fill().SetStyle(FillStyleGradient)
	defer func() {
		if recover() == nil {
			t.Fatal("gradient without setter should panic")
		}
	}()
	p.Rect(NewPoint(4, 4), NewSize(8, 8))
}

func TestFillStylePatternConsts(t *testing.T) {
	ResetTestRuntime()
	fs := GetTypeRegistrar().Get("FillStyle")
	if fs.GetClassPropertyMetadata("checkerboard").get(nil).(FillStyle) != FillStyleCheckerboard {
		t.Fatal("FillStyle.checkerboard")
	}
	if fs.GetClassPropertyMetadata("gradient").get(nil).(FillStyle) != FillStyleGradient {
		t.Fatal("FillStyle.gradient")
	}
	if fs.GetClassPropertyMetadata("tile").get(nil).(FillStyle) != FillStyleTile {
		t.Fatal("FillStyle.tile")
	}
}

func TestPixmapAtImageXY(t *testing.T) {
	pm := NewPixmap(2, 3)
	ink := NewColor(1, 2, 3, 255)
	pm.Put(1, 2, ink) // row 1, col 2 → x=2, y=1
	if got := pm.At(2, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("image.Image At(x=2,y=1)=%v, want %v", got, colorToPremulRGBA(ink))
	}
}