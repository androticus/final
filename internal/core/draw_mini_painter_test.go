package core

import "testing"

func TestMinipaintPutPixelMat2i1(t *testing.T) {
	mat := NewMat2i1(8, 8)
	pal := NewPalette1(NewColor(0, 0, 0, 255), NewColor(255, 255, 255, 255))
	mp := NewMinipaint(NewBlitSrc1(pal, mat))

	if mp.Width() != 8 || mp.Height() != 8 {
		t.Fatalf("size %d×%d, want 8×8", mp.Width(), mp.Height())
	}
	if mp.Ink() != 1 || mp.Bg() != 0 {
		t.Fatalf("defaults ink=%d bg=%d, want 1/0", mp.Ink(), mp.Bg())
	}

	mp.PutPixel(3, 2, 1)
	if mat.At(2, 3) != 1 { // row=y, col=x
		t.Fatalf("PutPixel(3,2,1): mat[%d,%d]=%d, want 1", 2, 3, mat.At(2, 3))
	}

	mp.PutPixel(3, 2, 0)
	if mat.At(2, 3) != 0 {
		t.Fatalf("clear: mat[2,3]=%d, want 0", mat.At(2, 3))
	}

	mp.SetInk(1)
	mp.Pixel(0, 0)
	if mat.At(0, 0) != 1 {
		t.Fatalf("Pixel with ink: mat[0,0]=%d, want 1", mat.At(0, 0))
	}

	// OOB no-op (and must not panic)
	mp.PutPixel(-1, 0, 1)
	mp.PutPixel(0, -1, 1)
	mp.PutPixel(8, 0, 1)
	mp.PutPixel(0, 8, 1)
	if mat.At(0, 0) != 1 {
		t.Fatal("OOB put altered in-bounds pixel")
	}
}

func TestMinipaintPutPixelMat2i4AndU8(t *testing.T) {
	m4 := NewMat2i4(4, 4)
	p4 := NewEmptyPalette4()
	mp4 := NewMinipaint(NewBlitSrc4(p4, m4))
	mp4.PutPixel(1, 2, 7)
	if m4.At(2, 1) != 7 {
		t.Fatalf("i4: got %d, want 7", m4.At(2, 1))
	}

	m8 := NewMat2u8(4, 4)
	p8 := NewEmptyPalette8()
	mp8 := NewMinipaint(NewBlitSrc8(p8, m8))
	mp8.PutPixel(2, 1, 200)
	if m8.At_(1, 2) != 200 {
		t.Fatalf("u8: got %d, want 200", m8.At_(1, 2))
	}
}

func TestMinipaintRejectsBlitSrcC(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for BlitSrcC")
		}
	}()
	_ = NewMinipaint(NewBlitSrcC(NewMat2C(2, 2)))
}

func newTestMinipaintI1(t *testing.T, rows, cols int) (*Minipaint, *Mat2i1) {
	t.Helper()
	mat := NewMat2i1(rows, cols)
	pal := NewPalette1(NewColor(0, 0, 0, 255), NewColor(255, 255, 255, 255))
	return NewMinipaint(NewBlitSrc1(pal, mat)), mat
}

func mat2i1Count(mat *Mat2i1) int {
	n := 0
	for r := 0; r < mat.Rows(); r++ {
		for c := 0; c < mat.Cols(); c++ {
			if mat.At(r, c) != 0 {
				n++
			}
		}
	}
	return n
}

func TestMinipaintFillScreenAndRect(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 8, 8)
	mp.FillRect(2, 1, 3, 2, 1)
	if mat.At(1, 2) != 1 || mat.At(2, 4) != 1 {
		t.Fatal("fillRect missed expected cells")
	}
	if mat.At(0, 0) != 0 || mat.At(1, 1) != 0 {
		t.Fatal("fillRect spilled")
	}
	if mat2i1Count(mat) != 6 {
		t.Fatalf("fillRect count=%d, want 6", mat2i1Count(mat))
	}

	mp.FillScreen(0)
	if mat2i1Count(mat) != 0 {
		t.Fatal("fillScreen 0 should clear")
	}
	mp.FillScreen(1)
	if mat2i1Count(mat) != 64 {
		t.Fatalf("fillScreen 1 count=%d, want 64", mat2i1Count(mat))
	}
}

func TestMinipaintDrawRectHollow(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 8, 8)
	mp.DrawRect(1, 1, 4, 4, 1)
	if mat.At(2, 2) != 0 {
		t.Fatal("drawRect filled interior")
	}
	if mat.At(1, 1) != 1 || mat.At(1, 4) != 1 || mat.At(4, 1) != 1 || mat.At(4, 4) != 1 {
		t.Fatal("drawRect missing corners")
	}
}

func TestMinipaintLines(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 8, 8)
	mp.DrawFastHLine(1, 3, 4, 1)
	if mat2i1Count(mat) != 4 || mat.At(3, 1) != 1 || mat.At(3, 4) != 1 {
		t.Fatalf("HLine: count=%d", mat2i1Count(mat))
	}
	mp.FillScreen(0)
	mp.DrawFastVLine(2, 0, 8, 1)
	if mat2i1Count(mat) != 8 {
		t.Fatalf("VLine: count=%d, want 8", mat2i1Count(mat))
	}
	mp.FillScreen(0)
	mp.DrawLine(0, 0, 7, 7, 1)
	if mat.At(0, 0) != 1 || mat.At(7, 7) != 1 {
		t.Fatal("diagonal missing endpoints")
	}
	if mat2i1Count(mat) != 8 {
		t.Fatalf("diagonal count=%d, want 8", mat2i1Count(mat))
	}
}

func TestMinipaintCircleAndTriangle(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 16, 16)
	mp.FillCircle(8, 8, 3, 1)
	if mat.At(8, 8) != 1 {
		t.Fatal("fillCircle missed center")
	}
	nFill := mat2i1Count(mat)
	if nFill < 20 || nFill > 45 {
		t.Fatalf("fillCircle r=3 count=%d, expected ~29", nFill)
	}

	mp.FillScreen(0)
	mp.DrawCircle(8, 8, 4, 1)
	if mat.At(8, 8) != 0 {
		t.Fatal("drawCircle filled center")
	}
	if mat.At(8, 12) != 1 || mat.At(8, 4) != 1 {
		t.Fatal("drawCircle missing axis points")
	}

	mp.FillScreen(0)
	mp.DrawTriangle(1, 1, 10, 1, 1, 10, 1)
	if mat.At(1, 1) != 1 || mat.At(1, 10) != 1 || mat.At(10, 1) != 1 {
		t.Fatal("drawTriangle missing vertices")
	}

	mp.FillScreen(0)
	mp.FillTriangle(1, 1, 10, 1, 1, 10, 1)
	if mat.At(2, 2) != 1 {
		t.Fatal("fillTriangle missed interior")
	}
	if mat2i1Count(mat) < 10 {
		t.Fatalf("fillTriangle too sparse: %d", mat2i1Count(mat))
	}
}

func TestMinipaintEllipse(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 24, 40)
	mp.FillEllipse(20, 12, 12, 6, 1)
	if mat.At(12, 20) != 1 {
		t.Fatal("fillEllipse missed center")
	}
	if mat.At(12, 20+10) != 1 || mat.At(12+4, 20) != 1 {
		t.Fatal("fillEllipse missed interior along the axes")
	}
	if mat.At(12+10, 20) != 0 {
		t.Fatal("fillEllipse spilled past the vertical radius")
	}
	nFill := mat2i1Count(mat)
	if nFill < 80 || nFill > 280 {
		t.Fatalf("fillEllipse count=%d", nFill)
	}

	mp.FillScreen(0)
	mp.DrawEllipse(20, 12, 12, 6, 1)
	if mat.At(12, 20) != 0 {
		t.Fatal("drawEllipse filled center")
	}
	if mat.At(12, 20+12) != 1 || mat.At(12, 20-12) != 1 ||
		mat.At(12+6, 20) != 1 || mat.At(12-6, 20) != 1 {
		t.Fatal("drawEllipse missing axis endpoints")
	}
}

func TestMinipaintRoundRect(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 16, 16)
	mp.FillRoundRect(2, 2, 10, 8, 2, 1)
	if mat.At(6, 7) != 1 {
		t.Fatal("fillRoundRect missed interior")
	}
	n := mat2i1Count(mat)
	if n < 50 || n > 80 {
		t.Fatalf("fillRoundRect count=%d", n)
	}
	mp.FillScreen(0)
	mp.DrawRoundRect(2, 2, 10, 8, 2, 1)
	if mat.At(6, 7) != 0 {
		t.Fatal("drawRoundRect filled interior")
	}
	if mat2i1Count(mat) < 20 {
		t.Fatalf("drawRoundRect too sparse: %d", mat2i1Count(mat))
	}
}

func TestMinipaintDrawCharAndPrint(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 32, 64)
	if mp.TextSize() != 1 || !mp.Wrap() {
		t.Fatalf("defaults textSize=%d wrap=%v", mp.TextSize(), mp.Wrap())
	}
	mp.DrawChar(0, 0, 'A')
	n := mat2i1Count(mat)
	if n < 8 {
		t.Fatalf("drawChar 'A' too sparse: %d", n)
	}
	// Transparent bg (ink==bg): no fill of spacing column-only noise beyond glyph
	mp.FillScreen(0)
	mp.SetInk(1)
	mp.SetBg(1)
	mp.DrawChar(0, 0, ' ')
	if mat2i1Count(mat) != 0 {
		t.Fatalf("space with ink==bg should be empty, got %d", mat2i1Count(mat))
	}

	mp.FillScreen(0)
	mp.SetBg(0)
	mp.SetCursor(0, 0)
	mp.Print("Hi")
	if mp.CursorX() != 12 { // 2 chars × 6
		t.Fatalf("cursorX after Hi = %d, want 12", mp.CursorX())
	}
	if mat2i1Count(mat) < 10 {
		t.Fatalf("Print Hi sparse: %d", mat2i1Count(mat))
	}

	mp.FillScreen(0)
	mp.SetCursor(0, 0)
	mp.WriteChar('\n')
	if mp.CursorX() != 0 || mp.CursorY() != 8 {
		t.Fatalf("newline cursor = (%d,%d), want (0,8)", mp.CursorX(), mp.CursorY())
	}

	// wrap at right edge
	mp.SetCursor(60, 0) // width 64; next char needs 6 → wrap
	mp.WriteChar('X')
	if mp.CursorY() != 8 || mp.CursorX() != 6 {
		t.Fatalf("wrap cursor = (%d,%d), want (6,8)", mp.CursorX(), mp.CursorY())
	}
}

func TestMinipaintTextSizeAndOpaqueBg(t *testing.T) {
	mp, mat := newTestMinipaintI1(t, 24, 24)
	mp.SetTextSize(2)
	mp.SetInk(1)
	mp.SetBg(0)
	mp.DrawChar(0, 0, 'I')
	n1 := mat2i1Count(mat)
	if n1 < 8 {
		t.Fatalf("size2 glyph sparse: %d", n1)
	}
	// Opaque bg (ink != bg): space clears a 12×16 cell on a filled screen.
	mp.FillScreen(1)
	mp.SetInk(1)
	mp.SetBg(0)
	mp.DrawChar(0, 0, ' ')
	cell := minipaintCharAdvance * 2 * minipaintFontHeight * 2 // 192
	ones := mat2i1Count(mat)
	wantOnes := 24*24 - cell
	if ones != wantOnes {
		t.Fatalf("opaque space size2 ones=%d, want %d (cleared %d)", ones, wantOnes, cell)
	}
}

func TestMinipaintFontTableLen(t *testing.T) {
	if len(minipaintFont) != 256*minipaintFontWidth {
		t.Fatalf("font len=%d, want %d", len(minipaintFont), 256*minipaintFontWidth)
	}
}
