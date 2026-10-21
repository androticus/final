package core

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestBlitTypeOrdinalsMatchDrawOp(t *testing.T) {
	if BlitBlend != 0 || BlitType(draw.Over) != BlitBlend {
		t.Fatalf("BlitBlend=%d, draw.Over=%d", BlitBlend, draw.Over)
	}
	if BlitStomp != 1 || BlitType(draw.Src) != BlitStomp {
		t.Fatalf("BlitStomp=%d, draw.Src=%d", BlitStomp, draw.Src)
	}
	bt := GetTypeRegistrar().Get("BlitType")
	if bt.GetClassPropertyMetadata("blend").get(nil).(BlitType) != BlitBlend {
		t.Fatal("BlitType.blend")
	}
	if bt.GetClassPropertyMetadata("stomp").get(nil).(BlitType) != BlitStomp {
		t.Fatal("BlitType.stomp")
	}
}

func TestBlitParamsOnPixmap(t *testing.T) {
	pm := NewPixmap(2, 2)
	o := pm.Params()
	if o.Mask() != BlitMaskOff || o.Op() != int(BlitStomp) || o.CellPx() != 1 {
		t.Fatalf("pixmap params base = %+v", *o)
	}
	if NewPainter(NewSize(3, 3)).Pixmap().Params().Op() != int(BlitStomp) {
		t.Fatal("Painter pixmap params missing/default")
	}
	o.SetOp(int(BlitBlend))
	if o.Op() != int(BlitBlend) {
		t.Fatal("SetOp")
	}
}

func TestPixmapClearAndFill(t *testing.T) {
	pm := NewPixmap(2, 2)
	ink := NewColor(10, 20, 30, 255)
	pm.Fill(ink)
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("Fill stored=%v, want %v", got, colorToPremulRGBA(ink))
	}
	pm.Clear()
	if got := pixmapStoredAt(pm, 0, 0); got.A != 0 {
		t.Fatalf("Clear stored=%v, want transparent", got)
	}
}

func TestBlit1StompDoesNotClearOutside(t *testing.T) {
	pm := NewPixmap(2, 2)
	bg := NewColor(255, 0, 0, 255)
	pm.Fill(bg)
	mat := NewMat2i1(1, 1)
	mat.Put(0, 0, 1)
	fg := NewColor(0, 255, 0, 255)
	pal := NewPalette1(NewColor(0, 0, 0, 255), fg)
	pm.Blit1(pal, mat, NewPoint(1, 1))
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(bg) {
		t.Fatalf("outside stamp = %v, want leftover %v", got, colorToPremulRGBA(bg))
	}
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(fg) {
		t.Fatalf("stamp = %v, want %v", got, colorToPremulRGBA(fg))
	}
}

func TestBlitSrcAndWithOverrides(t *testing.T) {
	pm := NewPixmap(2, 2)
	bg := NewColor(10, 10, 10, 255)
	pm.Fill(bg)
	mat := NewMat2i1(2, 2)
	mat.Put(0, 0, 0)
	mat.Put(0, 1, 1)
	fg := NewColor(0, 255, 0, 255)
	key := NewColor(255, 0, 255, 255)
	src := NewBlitSrc1(NewPalette1(key, fg), mat)

	pm.Blit(src, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(key) {
		t.Fatalf("no mask: key cell = %v, want stamped key %v", got, colorToPremulRGBA(key))
	}

	pm.Fill(bg)
	over := NewBlitParams()
	over.SetMask(BlitMaskOn)
	pm.BlitWithOverrides(src, over, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(bg) {
		t.Fatalf("mask on: key cell = %v, want leftover bg %v", got, colorToPremulRGBA(bg))
	}
	if got := pixmapStoredAt(pm, 0, 1); got != colorToPremulRGBA(fg) {
		t.Fatalf("mask on: ink = %v, want %v", got, colorToPremulRGBA(fg))
	}
}

func TestBlitWithOverridesCellPx(t *testing.T) {
	pm := NewPixmap(4, 4)
	mat := NewMat2i1(1, 1)
	mat.Put(0, 0, 1)
	fg := NewColor(0, 0, 255, 255)
	src := NewBlitSrc1(NewPalette1(Color{}, fg), mat)
	over := NewBlitParams()
	over.SetCellPx(2)
	pm.BlitWithOverrides(src, over, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(fg) {
		t.Fatalf("override cellPx: stored(1,1)=%v, want %v", got, colorToPremulRGBA(fg))
	}
	if pm.Params().CellPx() != 1 {
		t.Fatal("override must not mutate pixmap params base")
	}
}


func TestBlit1CellPxN(t *testing.T) {
	pm := NewPixmap(4, 4)
	mat := NewMat2i1(2, 2)
	mat.Put(0, 1, 1)
	fg := NewColor(0, 0, 255, 255)
	back := NewColor(1, 2, 3, 255)
	pal := NewPalette1(back, fg)
	pm.Params().SetCellPx(2)
	pm.Blit1(pal, mat, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 2); got != colorToPremulRGBA(fg) {
		t.Fatalf("expanded live stored(0,2)=%v, want %v", got, colorToPremulRGBA(fg))
	}
	if got := pixmapStoredAt(pm, 1, 3); got != colorToPremulRGBA(fg) {
		t.Fatalf("expanded live stored(1,3)=%v, want %v", got, colorToPremulRGBA(fg))
	}
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(back) {
		t.Fatalf("dead cell stored(0,0)=%v, want pal[0] %v", got, colorToPremulRGBA(back))
	}
}

func TestBlit1BlendMatchesDrawOver(t *testing.T) {
	dstC := NewColor(255, 0, 0, 255)
	srcC := NewColor(0, 0, 255, 128)
	pm := NewPixmap(1, 1)
	pm.Fill(dstC)
	mat := NewMat2i1(1, 1)
	mat.Put(0, 0, 1)
	pal := NewPalette1(Color{}, srcC)
	pm.Params().SetOp(int(BlitBlend))
	pm.Blit1(pal, mat, NewPoint(0, 0))

	ref := image.NewRGBA(image.Rect(0, 0, 1, 1))
	draw.Draw(ref, ref.Bounds(), &image.Uniform{dstC}, image.Point{}, draw.Src)
	draw.Draw(ref, ref.Bounds(), &image.Uniform{srcC}, image.Point{}, draw.Over)
	want := colorFromPremulRGBA(ref.RGBAAt(0, 0))
	if got := pixmapColorAt(pm, 0, 0); got != want {
		t.Fatalf("blend readback=%v, want draw.Over %v", got, want)
	}
}

func TestBlitCAndBlit4AndBlit8(t *testing.T) {
	ink := NewColor(9, 8, 7, 255)
	pm := NewPixmap(1, 2)

	mc := NewMat2C(1, 2)
	mc.Put(0, 1, ink)
	pm.Clear()
	pm.BlitC(mc, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("blitC stored(0,1)=%v, want %v", got, colorToPremulRGBA(ink))
	}

	m4 := NewMat2i4(1, 2)
	m4.Put(0, 1, 1)
	pal4 := NewPalette4FromSlice([]Color{{}, ink}, 255)
	pm.Clear()
	pm.Blit4(pal4, m4, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("blit4 stored(0,1)=%v, want %v", got, colorToPremulRGBA(ink))
	}

	m8 := NewMat2u8(1, 2)
	m8.Put(0, 1, 1)
	pal8 := NewPalette8FromSlice([]Color{{}, ink})
	pm.Clear()
	pm.Blit8(pal8, m8, NewPoint(0, 0))
	if got := pixmapStoredAt(pm, 0, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("blit8 stored(0,1)=%v, want %v", got, colorToPremulRGBA(ink))
	}
}

func TestBlit1PainterRGBA(t *testing.T) {
	p := NewPainter(NewSize(2, 2))
	pm := p.Pixmap()
	fg := NewColor(255, 0, 0, 128)
	mat := NewMat2i1(2, 2)
	mat.Put(1, 1, 1)
	pal := NewPalette1(Color{}, fg)
	pm.Blit1(pal, mat, NewPoint(0, 0))
	got := pixmapColorAt(pm, 1, 1)
	stored := pm.P.RGBAAt(1, 1)
	want := colorFromPremulRGBA(stored)
	if got != want {
		t.Fatalf("readback=%v, want %v stored %v", got, want, stored)
	}
	if stored.A == 0 {
		t.Fatal("expected painted RGBA pixel")
	}
}

func TestBlitClipsToDest(t *testing.T) {
	pm := NewPixmap(2, 2)
	bg := NewColor(1, 2, 3, 255)
	pm.Fill(bg)
	mat := NewMat2i1(2, 2)
	mat.Put(0, 0, 1)
	mat.Put(0, 1, 1)
	mat.Put(1, 0, 1)
	mat.Put(1, 1, 1)
	fg := NewColor(0, 255, 0, 255)
	pal := NewPalette1(Color{}, fg)
	// Stamp overhangs right/bottom by 1; only the overlapping cell is written.
	pm.Blit1(pal, mat, NewPoint(1, 1))
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(fg) {
		t.Fatalf("clipped stamp = %v, want %v", got, colorToPremulRGBA(fg))
	}
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(bg) {
		t.Fatalf("outside = %v, want bg %v", got, colorToPremulRGBA(bg))
	}
}

func TestBlitFullyOutsideIsNoop(t *testing.T) {
	pm := NewPixmap(2, 2)
	bg := NewColor(9, 9, 9, 255)
	pm.Fill(bg)
	mat := NewMat2i1(2, 2)
	mat.Put(0, 0, 1)
	pal := NewPalette1(Color{}, NewColor(255, 0, 0, 255))
	pm.Blit1(pal, mat, NewPoint(10, 10))
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(bg) {
		t.Fatalf("fully outside should leave dest, got %v", got)
	}
}

func TestBlitMethodsRegistered(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"blit1", "blit4", "blit8", "blitC", "clear", "fill"} {
		if pixmapType.GetMethodMetadata(name) == nil {
			t.Fatalf("missing Pixmap.%s", name)
		}
	}
	rp := NewTestRunProvider()
	pm := NewPixmap(1, 1)
	mat := NewMat2i1(1, 1)
	mat.Put(0, 0, 1)
	fg := NewColor(4, 5, 6, 255)
	pal := NewPalette1(Color{}, fg)
	f, ths := memberRunFrame(rp, pm)
	rp.DataStack.Push(pal)
	rp.DataStack.Push(mat)
	rp.DataStack.Push(Int(0))
	rp.DataStack.Push(Int(0))
	pixmapType.GetMethodMetadata("blit1").invokeCall(f, ths)
	if got := pixmapStoredAt(pm, 0, 0); got != colorToPremulRGBA(fg) {
		t.Fatalf("native blit1 stored=%v, want %v", got, colorToPremulRGBA(fg))
	}
}

func TestRgbaOverMatchesDrawOver(t *testing.T) {
	dstC := NewColor(255, 0, 0, 255)
	srcC := NewColor(0, 0, 255, 128)
	dst := colorToPremulRGBA(dstC)
	src := colorToPremulRGBA(srcC)
	got := rgbaOver(dst, src)
	ref := image.NewRGBA(image.Rect(0, 0, 1, 1))
	ref.SetRGBA(0, 0, dst)
	draw.Draw(ref, ref.Bounds(), &image.Uniform{srcC}, image.Point{}, draw.Over)
	want := ref.RGBAAt(0, 0)
	if got != want {
		t.Fatalf("rgbaOver=%v, want draw.Over %v", got, want)
	}
}

var _ color.Color = Color{}
