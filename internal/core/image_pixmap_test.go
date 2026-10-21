package core

import (
	"image/color"
	"testing"
)

func pixmapExpectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func pixmapStoredAt(pm *Pixmap, row, col int) color.RGBA {
	x, y := pm.xy(row, col)
	return pm.P.RGBAAt(x, y)
}

func pixmapColorAt(pm *Pixmap, row, col int) Color {
	return colorFromPremulRGBA(pixmapStoredAt(pm, row, col))
}

func TestPixmapRowsCols(t *testing.T) {
	pm := NewPixmap(2, 3)
	if pm.Rows() != 2 || pm.Cols() != 3 {
		t.Fatalf("shape %d×%d, want 2×3", pm.Rows(), pm.Cols())
	}
	p := NewPainter(NewSize(8, 5))
	if p.Pixmap().Rows() != 5 || p.Pixmap().Cols() != 8 {
		t.Fatalf("painter pixmap shape %d×%d, want 5×8", p.Pixmap().Rows(), p.Pixmap().Cols())
	}
}

func TestPixmapWrap(t *testing.T) {
	ResetTestRuntime()
	pm := NewPixmap(2, 3)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pm)
	mm := pixmapType.GetMethodMetadata("wrap")
	if mm == nil {
		t.Fatal("missing Pixmap.wrap")
	}
	rp.DataStack.Push(Int(-1))
	rp.DataStack.Push(Int(3))
	mm.invokeCall(f, ths)
	col := rp.DataStack.Pop().(Int)
	row := rp.DataStack.Pop().(Int)
	if row != 1 || col != 0 {
		t.Fatalf("wrap(-1, 3) on 2×3 = (%d, %d), want (1, 0)", row, col)
	}
}

func TestPixmapPutOpaqueRoundTripBytes(t *testing.T) {
	pm := NewPixmap(2, 3)
	ink := NewColor(255, 128, 0, 255)
	pm.Put(1, 2, ink)
	want := colorToPremulRGBA(ink)
	if got := pixmapStoredAt(pm, 1, 2); got != want {
		t.Fatalf("stored=%v, want %v", got, want)
	}
	pm.PutFlat(0, ink)
	if got := pixmapStoredAt(pm, 0, 0); got != want {
		t.Fatalf("PutFlat stored=%v, want %v", got, want)
	}
}

func TestPixmapPutTranslucentStoredPremul(t *testing.T) {
	pm := NewPixmap(2, 3)
	trans := NewColor(255, 0, 0, 128)
	pm.Put(0, 1, trans)
	stored := pixmapStoredAt(pm, 0, 1)
	want := colorToPremulRGBA(trans)
	if stored != want {
		t.Fatalf("stored=%v, want premul %v", stored, want)
	}
}

func TestPixmapImageAtReturnsRGBA(t *testing.T) {
	pm := NewPixmap(2, 3)
	ink := NewColor(1, 2, 3, 255)
	pm.Put(1, 2, ink)
	x, y := pm.xy(1, 2)
	got := pm.At(x, y)
	want := colorToPremulRGBA(ink)
	if got != want {
		t.Fatalf("image.Image At=%v, want %v", got, want)
	}
}

func TestPixmapPutOutOfRangePanics(t *testing.T) {
	pm := NewPixmap(2, 2)
	pixmapExpectPanic(t, func() { pm.Put(0, 2, NewColor(1, 2, 3, 255)) })
}

func TestPixmapToMat2COpaqueRoundTrip(t *testing.T) {
	pm := NewPixmap(2, 3)
	ink := NewColor(255, 128, 0, 255)
	pm.Put(1, 2, ink)
	pm.PutFlat(0, NewColor(1, 2, 3, 255))
	got := pm.ToMat2C()
	if got.At_(1, 2) != ink {
		t.Fatalf("ToMat2C(1,2)=%v, want %v", got.At_(1, 2), ink)
	}
	if got.At_(0, 0) != NewColor(1, 2, 3, 255) {
		t.Fatalf("ToMat2C(0,0)=%v", got.At_(0, 0))
	}
	if got.Rows() != 2 || got.Cols() != 3 {
		t.Fatalf("shape %d×%d, want 2×3", got.Rows(), got.Cols())
	}
}

func TestPixmapToMat2CClearIsTransparent(t *testing.T) {
	pm := NewPixmap(2, 2)
	pm.Fill(NewColor(9, 8, 7, 255))
	pm.Clear()
	got := pm.ToMat2C()
	z := NewColor(0, 0, 0, 0)
	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			if c := got.At_(row, col); c != z {
				t.Fatalf("Clear ToMat2C(%d,%d)=%v, want zero", row, col, c)
			}
		}
	}
}

func TestPixmapToMat2CTransparentBufferCell(t *testing.T) {
	pm := NewPixmap(1, 1)
	pm.Put(0, 0, NewColor(200, 100, 50, 0))
	got := pm.ToMat2C().At_(0, 0)
	want := NewColor(0, 0, 0, 0)
	if got != want {
		t.Fatalf("A=0 buffer cell = %v, want %v", got, want)
	}
}

func TestPixmapToMat2CNativeMethod(t *testing.T) {
	ResetTestRuntime()
	pm := NewPixmap(1, 1)
	ink := NewColor(4, 5, 6, 255)
	pm.Put(0, 0, ink)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pm)
	pixmapType.GetMethodMetadata("toMat2C").invokeCall(f, ths)
	mat, ok := rp.DataStack.Pop().(*Mat2C)
	if !ok || mat.At_(0, 0) != ink {
		t.Fatalf("native toMat2C = %v, want %v", mat, ink)
	}
}

func TestPixmapPutFlatOutOfRangePanics(t *testing.T) {
	pm := NewPixmap(2, 2)
	pixmapExpectPanic(t, func() { pm.PutFlat(4, NewColor(1, 2, 3, 255)) })
}

func TestPixmapPutRowColCheckedEvenWhenFlatInRange(t *testing.T) {
	pm := NewPixmap(2, 3)
	ink := NewColor(1, 2, 3, 255)
	pm.PutFlat(5, ink)
	pixmapExpectPanic(t, func() { pm.Put(0, 5, ink) })
	if got := pixmapStoredAt(pm, 1, 2); got != colorToPremulRGBA(ink) {
		t.Fatalf("PutFlat stored=%v", got)
	}
}

func TestPixmapNoNamedAtPutMethods(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"at", "atFlat", "put", "putFlat"} {
		if pixmapType.GetMethodMetadata(name) != nil {
			t.Fatalf("Pixmap must not expose named method %q (items only, like mats)", name)
		}
	}
}

func TestPixmapItemsMetadataWriteOnly(t *testing.T) {
	ResetTestRuntime()
	im := pixmapType.GetItemsMetadata()
	if im == nil || im.getByOrdinal != nil || im.getByMultiIndex != nil {
		t.Fatal("Pixmap must not register @() / @[] getters")
	}
	if im.setByOrdinal == nil || im.setByMultiIndex == nil || im.multiIndexArity != 2 {
		t.Fatal("Pixmap must register !() / ![] with arity 2")
	}
	if bool(im.FinalType().GetPropertyMetadata("hasGetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasGetByIndices want false for write-only Pixmap")
	}
	if !bool(im.FinalType().GetPropertyMetadata("hasSetByIndices").GetValue(im).(Bool)) {
		t.Fatal("ItemsMetadata.hasSetByIndices want true for Pixmap")
	}
	pm := NewPixmap(2, 2)
	ink := NewColor(9, 8, 7, 255)
	im.SetByOrdinal(pm, Int(3), ink)
	if got := pixmapStoredAt(pm, 1, 1); got != colorToPremulRGBA(ink) {
		t.Fatalf("flat !() stored=%v", got)
	}
	ink2 := NewColor(1, 2, 3, 255)
	im.SetByMultiIndex(pm, ink2, Int(0), Int(1))
	if got := pixmapStoredAt(pm, 0, 1); got != colorToPremulRGBA(ink2) {
		t.Fatalf("![] stored=%v", got)
	}
}

func TestOpBangIndexPixmapFlat(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	pm := NewPixmap(2, 3)
	ink := NewColor(10, 20, 30, 255)
	rp.DataStack.Push(pm)
	rp.DataStack.Push(ink)
	rp.DataStack.Push(Int(5)) // (1,2)
	RunOps(task, rp, OpBangIndex, OpStop)
	if got := pixmapStoredAt(pm, 1, 2); got != colorToPremulRGBA(ink) {
		t.Fatalf("after !() stored=%v", got)
	}
}

func TestOpBangMultiPixmap(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	pm := NewPixmap(2, 3)
	ink := NewColor(4, 5, 6, 255)
	rp.DataStack.Push(pm)
	rp.DataStack.Push(ink)
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	RunOps(task, rp, OpBangMulti, OpStop)
	if got := pixmapStoredAt(pm, 1, 2); got != colorToPremulRGBA(ink) {
		t.Fatalf("after ![] stored=%v", got)
	}
}
