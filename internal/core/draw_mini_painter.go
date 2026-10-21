package core

import "fmt"

// Primitive algorithms ported from Adafruit GFX (BSD):
// https://github.com/adafruit/Adafruit-GFX-Library

// Minipaint draws palette-index pixels into a BlitSrc mat (kinds 1/4/8).
// Adafruit-GFX-shaped geometry (top-left, Y-down); not the gg Painter path.
type Minipaint struct {
	blitSrc  *BlitSrc
	ink      int // palette index (text / stroke)
	bg       int // palette index (text background; same as ink ⇒ transparent bg)
	cursorX  int
	cursorY  int
	textSize int  // integer magnification (≥1); classic 6×8 cell × size
	wrap     bool // wrap text at right edge
}

func (mp *Minipaint) FinalType() *Type { return minipaintType }

func (mp *Minipaint) BlitSrc() *BlitSrc { return mp.blitSrc }

func (mp *Minipaint) Ink() int { return mp.ink }

func (mp *Minipaint) SetInk(v int) { mp.ink = v }

func (mp *Minipaint) Bg() int { return mp.bg }

func (mp *Minipaint) SetBg(v int) { mp.bg = v }

func (mp *Minipaint) CursorX() int { return mp.cursorX }

func (mp *Minipaint) CursorY() int { return mp.cursorY }

func (mp *Minipaint) SetCursor(x, y int) {
	mp.cursorX = x
	mp.cursorY = y
}

func (mp *Minipaint) TextSize() int { return mp.textSize }

func (mp *Minipaint) SetTextSize(s int) {
	if s < 1 {
		s = 1
	}
	mp.textSize = s
}

func (mp *Minipaint) Wrap() bool { return mp.wrap }

func (mp *Minipaint) SetWrap(w bool) { mp.wrap = w }

// Width and Height are mat cols × rows (GFX screen size).
func (mp *Minipaint) Width() int {
	_, cols := mp.blitSrc.rowsCols()
	return cols
}

func (mp *Minipaint) Height() int {
	rows, _ := mp.blitSrc.rowsCols()
	return rows
}

// NewMinipaint binds a BlitSrc. Kind must be 1, 4, or 8 (not full-color Mat2c).
func NewMinipaint(src *BlitSrc) *Minipaint {
	if src == nil {
		panic(fmt.Errorf("Minipaint.new: blitSrc required"))
	}
	switch src.kind {
	case blitSrc1, blitSrc4, blitSrc8:
		// ok
	case blitSrcC:
		panic(fmt.Errorf("Minipaint.new: BlitSrcC / full color not supported"))
	default:
		panic(fmt.Errorf("Minipaint.new: unknown BlitSrc kind %d", src.kind))
	}
	return &Minipaint{blitSrc: src, ink: 1, bg: 0, textSize: 1, wrap: true}
}

// PutPixel writes palette index c at GFX (x,y). Out of bounds is a no-op.
func (mp *Minipaint) PutPixel(x, y, c int) {
	rows, cols := mp.blitSrc.rowsCols()
	if x < 0 || y < 0 || x >= cols || y >= rows {
		return
	}
	row, col := y, x // mat is row/col; GFX is x/y Y-down
	switch mp.blitSrc.kind {
	case blitSrc1:
		mp.blitSrc.mat.(*Mat2i1).Put(row, col, c)
	case blitSrc4:
		if c < 0 || c > 15 {
			panic(fmt.Errorf("Minipaint.putPixel: index %d out of range 0..15 for Mat2i4", c))
		}
		mp.blitSrc.mat.(*Mat2i4).Put(row, col, uint8(c))
	case blitSrc8:
		if c < 0 || c > 255 {
			panic(fmt.Errorf("Minipaint.putPixel: index %d out of range 0..255 for Mat2u8", c))
		}
		mp.blitSrc.mat.(*Mat2u8).Put(row, col, uint8(c))
	default:
		panic(fmt.Errorf("Minipaint.putPixel: unsupported kind %d", mp.blitSrc.kind))
	}
}

// Pixel writes .ink at (x,y).
func (mp *Minipaint) Pixel(x, y int) {
	mp.PutPixel(x, y, mp.ink)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func swapInt(a, b *int) { *a, *b = *b, *a }

// writeLine is Bresenham (Adafruit_GFX::writeLine).
func (mp *Minipaint) writeLine(x0, y0, x1, y1, c int) {
	steep := absInt(y1-y0) > absInt(x1-x0)
	if steep {
		swapInt(&x0, &y0)
		swapInt(&x1, &y1)
	}
	if x0 > x1 {
		swapInt(&x0, &x1)
		swapInt(&y0, &y1)
	}
	dx := x1 - x0
	dy := absInt(y1 - y0)
	err := dx / 2
	ystep := 1
	if y0 >= y1 {
		ystep = -1
	}
	for ; x0 <= x1; x0++ {
		if steep {
			mp.PutPixel(y0, x0, c)
		} else {
			mp.PutPixel(x0, y0, c)
		}
		err -= dy
		if err < 0 {
			y0 += ystep
			err += dx
		}
	}
}

func (mp *Minipaint) writeFastVLine(x, y, h, c int) {
	if h <= 0 {
		return
	}
	for i := 0; i < h; i++ {
		mp.PutPixel(x, y+i, c)
	}
}

func (mp *Minipaint) writeFastHLine(x, y, w, c int) {
	if w <= 0 {
		return
	}
	for i := 0; i < w; i++ {
		mp.PutPixel(x+i, y, c)
	}
}

func (mp *Minipaint) writeFillRect(x, y, w, h, c int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < w; i++ {
		mp.writeFastVLine(x+i, y, h, c)
	}
}

func (mp *Minipaint) DrawFastVLine(x, y, h, c int) {
	mp.writeFastVLine(x, y, h, c)
}

func (mp *Minipaint) DrawFastHLine(x, y, w, c int) {
	mp.writeFastHLine(x, y, w, c)
}

func (mp *Minipaint) FillRect(x, y, w, h, c int) {
	mp.writeFillRect(x, y, w, h, c)
}

func (mp *Minipaint) FillScreen(c int) {
	mp.writeFillRect(0, 0, mp.Width(), mp.Height(), c)
}

func (mp *Minipaint) DrawLine(x0, y0, x1, y1, c int) {
	if x0 == x1 {
		if y0 > y1 {
			swapInt(&y0, &y1)
		}
		mp.writeFastVLine(x0, y0, y1-y0+1, c)
		return
	}
	if y0 == y1 {
		if x0 > x1 {
			swapInt(&x0, &x1)
		}
		mp.writeFastHLine(x0, y0, x1-x0+1, c)
		return
	}
	mp.writeLine(x0, y0, x1, y1, c)
}

func (mp *Minipaint) DrawRect(x, y, w, h, c int) {
	if w <= 0 || h <= 0 {
		return
	}
	mp.writeFastHLine(x, y, w, c)
	mp.writeFastHLine(x, y+h-1, w, c)
	mp.writeFastVLine(x, y, h, c)
	mp.writeFastVLine(x+w-1, y, h, c)
}

func (mp *Minipaint) DrawCircle(x0, y0, r, c int) {
	if r < 0 {
		return
	}
	f := 1 - r
	ddF_x := 1
	ddF_y := -2 * r
	x := 0
	y := r
	mp.PutPixel(x0, y0+r, c)
	mp.PutPixel(x0, y0-r, c)
	mp.PutPixel(x0+r, y0, c)
	mp.PutPixel(x0-r, y0, c)
	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x
		mp.PutPixel(x0+x, y0+y, c)
		mp.PutPixel(x0-x, y0+y, c)
		mp.PutPixel(x0+x, y0-y, c)
		mp.PutPixel(x0-x, y0-y, c)
		mp.PutPixel(x0+y, y0+x, c)
		mp.PutPixel(x0-y, y0+x, c)
		mp.PutPixel(x0+y, y0-x, c)
		mp.PutPixel(x0-y, y0-x, c)
	}
}

func (mp *Minipaint) drawCircleHelper(x0, y0, r int, cornername byte, c int) {
	f := 1 - r
	ddF_x := 1
	ddF_y := -2 * r
	x := 0
	y := r
	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x
		if cornername&0x4 != 0 {
			mp.PutPixel(x0+x, y0+y, c)
			mp.PutPixel(x0+y, y0+x, c)
		}
		if cornername&0x2 != 0 {
			mp.PutPixel(x0+x, y0-y, c)
			mp.PutPixel(x0+y, y0-x, c)
		}
		if cornername&0x8 != 0 {
			mp.PutPixel(x0-y, y0+x, c)
			mp.PutPixel(x0-x, y0+y, c)
		}
		if cornername&0x1 != 0 {
			mp.PutPixel(x0-y, y0-x, c)
			mp.PutPixel(x0-x, y0-y, c)
		}
	}
}

func (mp *Minipaint) FillCircle(x0, y0, r, c int) {
	if r < 0 {
		return
	}
	mp.writeFastVLine(x0, y0-r, 2*r+1, c)
	mp.fillCircleHelper(x0, y0, r, 3, 0, c)
}

// DrawEllipse outlines a Bresenham ellipse centered at (x0,y0)
// with horizontal radius rw and vertical radius rh (Adafruit GFX).
func (mp *Minipaint) DrawEllipse(x0, y0, rw, rh, c int) {
	if rw < 0 || rh < 0 {
		return
	}
	x, y := 0, rh
	rw2, rh2 := rw*rw, rh*rh
	twoRw2, twoRh2 := 2*rw2, 2*rh2
	decision := rh2 - (rw2 * rh) + (rw2 / 4)

	for (twoRh2 * x) < (twoRw2 * y) {
		mp.PutPixel(x0+x, y0+y, c)
		mp.PutPixel(x0-x, y0+y, c)
		mp.PutPixel(x0+x, y0-y, c)
		mp.PutPixel(x0-x, y0-y, c)
		x++
		if decision < 0 {
			decision += rh2 + (twoRh2 * x)
		} else {
			decision += rh2 + (twoRh2 * x) - (twoRw2 * y)
			y--
		}
	}

	decision = ((rh2 * (2*x + 1) * (2*x + 1)) >> 2) + (rw2 * (y - 1) * (y - 1)) - (rw2 * rh2)
	for y >= 0 {
		mp.PutPixel(x0+x, y0+y, c)
		mp.PutPixel(x0-x, y0+y, c)
		mp.PutPixel(x0+x, y0-y, c)
		mp.PutPixel(x0-x, y0-y, c)
		y--
		if decision > 0 {
			decision += rw2 - (twoRw2 * y)
		} else {
			decision += rw2 + (twoRh2 * x) - (twoRw2 * y)
			x++
		}
	}
}

// FillEllipse fills a Bresenham ellipse centered at (x0,y0)
// with horizontal radius rw and vertical radius rh (Adafruit GFX).
func (mp *Minipaint) FillEllipse(x0, y0, rw, rh, c int) {
	if rw < 0 || rh < 0 {
		return
	}
	x, y := 0, rh
	rw2, rh2 := rw*rw, rh*rh
	twoRw2, twoRh2 := 2*rw2, 2*rh2
	decision := rh2 - (rw2 * rh) + (rw2 / 4)

	for (twoRh2 * x) < (twoRw2 * y) {
		x++
		if decision < 0 {
			decision += rh2 + (twoRh2 * x)
		} else {
			decision += rh2 + (twoRh2 * x) - (twoRw2 * y)
			mp.writeFastHLine(x0-(x-1), y0+y, 2*(x-1)+1, c)
			mp.writeFastHLine(x0-(x-1), y0-y, 2*(x-1)+1, c)
			y--
		}
	}

	decision = ((rh2 * (2*x + 1) * (2*x + 1)) >> 2) + (rw2 * (y - 1) * (y - 1)) - (rw2 * rh2)
	for y >= 0 {
		mp.writeFastHLine(x0-x, y0+y, 2*x+1, c)
		mp.writeFastHLine(x0-x, y0-y, 2*x+1, c)
		y--
		if decision > 0 {
			decision += rw2 - (twoRw2 * y)
		} else {
			decision += rw2 + (twoRh2 * x) - (twoRw2 * y)
			x++
		}
	}
}

func (mp *Minipaint) fillCircleHelper(x0, y0, r int, corners byte, delta, c int) {
	f := 1 - r
	ddF_x := 1
	ddF_y := -2 * r
	x := 0
	y := r
	px := x
	py := y
	delta++
	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x
		if x < (y + 1) {
			if corners&1 != 0 {
				mp.writeFastVLine(x0+x, y0-y, 2*y+delta, c)
			}
			if corners&2 != 0 {
				mp.writeFastVLine(x0-x, y0-y, 2*y+delta, c)
			}
		}
		if y != py {
			if corners&1 != 0 {
				mp.writeFastVLine(x0+py, y0-px, 2*px+delta, c)
			}
			if corners&2 != 0 {
				mp.writeFastVLine(x0-py, y0-px, 2*px+delta, c)
			}
			py = y
		}
		px = x
	}
}

func (mp *Minipaint) DrawRoundRect(x, y, w, h, r, c int) {
	if w <= 0 || h <= 0 {
		return
	}
	maxR := w
	if h < w {
		maxR = h
	}
	maxR /= 2
	if r > maxR {
		r = maxR
	}
	mp.writeFastHLine(x+r, y, w-2*r, c)
	mp.writeFastHLine(x+r, y+h-1, w-2*r, c)
	mp.writeFastVLine(x, y+r, h-2*r, c)
	mp.writeFastVLine(x+w-1, y+r, h-2*r, c)
	mp.drawCircleHelper(x+r, y+r, r, 1, c)
	mp.drawCircleHelper(x+w-r-1, y+r, r, 2, c)
	mp.drawCircleHelper(x+w-r-1, y+h-r-1, r, 4, c)
	mp.drawCircleHelper(x+r, y+h-r-1, r, 8, c)
}

func (mp *Minipaint) FillRoundRect(x, y, w, h, r, c int) {
	if w <= 0 || h <= 0 {
		return
	}
	maxR := w
	if h < w {
		maxR = h
	}
	maxR /= 2
	if r > maxR {
		r = maxR
	}
	mp.writeFillRect(x+r, y, w-2*r, h, c)
	mp.fillCircleHelper(x+w-r-1, y+r, r, 1, h-2*r-1, c)
	mp.fillCircleHelper(x+r, y+r, r, 2, h-2*r-1, c)
}

func (mp *Minipaint) DrawTriangle(x0, y0, x1, y1, x2, y2, c int) {
	mp.DrawLine(x0, y0, x1, y1, c)
	mp.DrawLine(x1, y1, x2, y2, c)
	mp.DrawLine(x2, y2, x0, y0, c)
}

func (mp *Minipaint) FillTriangle(x0, y0, x1, y1, x2, y2, c int) {
	if y0 > y1 {
		swapInt(&y0, &y1)
		swapInt(&x0, &x1)
	}
	if y1 > y2 {
		swapInt(&y2, &y1)
		swapInt(&x2, &x1)
	}
	if y0 > y1 {
		swapInt(&y0, &y1)
		swapInt(&x0, &x1)
	}
	if y0 == y2 {
		a, b := x0, x0
		if x1 < a {
			a = x1
		} else if x1 > b {
			b = x1
		}
		if x2 < a {
			a = x2
		} else if x2 > b {
			b = x2
		}
		mp.writeFastHLine(a, y0, b-a+1, c)
		return
	}
	dx01, dy01 := x1-x0, y1-y0
	dx02, dy02 := x2-x0, y2-y0
	dx12, dy12 := x2-x1, y2-y1
	sa, sb := 0, 0
	last := y1 - 1
	if y1 == y2 {
		last = y1
	}
	y := y0
	for ; y <= last; y++ {
		a := x0 + sa/dy01
		b := x0 + sb/dy02
		sa += dx01
		sb += dx02
		if a > b {
			swapInt(&a, &b)
		}
		mp.writeFastHLine(a, y, b-a+1, c)
	}
	sa = dx12 * (y - y1)
	sb = dx02 * (y - y0)
	for ; y <= y2; y++ {
		a := x1 + sa/dy12
		b := x0 + sb/dy02
		sa += dx12
		sb += dx02
		if a > b {
			swapInt(&a, &b)
		}
		mp.writeFastHLine(a, y, b-a+1, c)
	}
}

// DrawChar draws classic 5×7 glyph c at (x,y) using ink/bg and textSize.
// cp437 is always on (181.1): no classic charset remap for codes ≥176.
func (mp *Minipaint) DrawChar(x, y int, c byte) {
	size := mp.textSize
	if size < 1 {
		size = 1
	}
	color, bg := mp.ink, mp.bg
	w, h := mp.Width(), mp.Height()
	if x >= w || y >= h ||
		(x+minipaintCharAdvance*size-1) < 0 ||
		(y+minipaintFontHeight*size-1) < 0 {
		return
	}
	base := int(c) * minipaintFontWidth
	for i := 0; i < minipaintFontWidth; i++ {
		line := minipaintFont[base+i]
		for j := 0; j < minipaintFontHeight; j++ {
			if line&1 != 0 {
				if size == 1 {
					mp.PutPixel(x+i, y+j, color)
				} else {
					mp.writeFillRect(x+i*size, y+j*size, size, size, color)
				}
			} else if bg != color {
				if size == 1 {
					mp.PutPixel(x+i, y+j, bg)
				} else {
					mp.writeFillRect(x+i*size, y+j*size, size, size, bg)
				}
			}
			line >>= 1
		}
	}
	if bg != color {
		if size == 1 {
			mp.writeFastVLine(x+minipaintFontWidth, y, minipaintFontHeight, bg)
		} else {
			mp.writeFillRect(x+minipaintFontWidth*size, y, size, minipaintFontHeight*size, bg)
		}
	}
}

// WriteChar advances the text cursor like Adafruit_GFX::write (classic font).
func (mp *Minipaint) WriteChar(c byte) {
	size := mp.textSize
	if size < 1 {
		size = 1
	}
	if c == '\n' {
		mp.cursorX = 0
		mp.cursorY += size * minipaintFontHeight
		return
	}
	if c == '\r' {
		return
	}
	if mp.wrap && (mp.cursorX+size*minipaintCharAdvance) > mp.Width() {
		mp.cursorX = 0
		mp.cursorY += size * minipaintFontHeight
	}
	mp.DrawChar(mp.cursorX, mp.cursorY, c)
	mp.cursorX += size * minipaintCharAdvance
}

// Print writes s (UTF-8 runes in 0..255) at the current cursor.
func (mp *Minipaint) Print(s string) {
	for _, r := range s {
		if r > 255 {
			continue
		}
		mp.WriteChar(byte(r))
	}
}

func popMinipaintInt(ds *Stack[any]) int {
	return int(ds.Pop().(Int))
}

func registerMinipaintMembers(t *Type) {
	t.registerProperty("blitSrc", "BlitSrc",
		func(inst Any) Any { return inst.(*Minipaint).BlitSrc() },
		nil,
	)
	t.registerProperty("ink", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).Ink()) },
		func(inst Any, v Any) { inst.(*Minipaint).SetInk(int(v.(Int))) },
	)
	t.registerProperty("bg", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).Bg()) },
		func(inst Any, v Any) { inst.(*Minipaint).SetBg(int(v.(Int))) },
	)
	t.registerProperty("cursorX", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).CursorX()) },
		func(inst Any, v Any) {
			mp := inst.(*Minipaint)
			mp.SetCursor(int(v.(Int)), mp.cursorY)
		},
	)
	t.registerProperty("cursorY", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).CursorY()) },
		func(inst Any, v Any) {
			mp := inst.(*Minipaint)
			mp.SetCursor(mp.cursorX, int(v.(Int)))
		},
	)
	t.registerProperty("textSize", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).TextSize()) },
		func(inst Any, v Any) { inst.(*Minipaint).SetTextSize(int(v.(Int))) },
	)
	t.registerProperty("wrap", "Bool",
		func(inst Any) Any { return Bool(inst.(*Minipaint).Wrap()) },
		func(inst Any, v Any) { inst.(*Minipaint).SetWrap(bool(v.(Bool))) },
	)
	t.registerProperty("width", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).Width()) },
		nil,
	)
	t.registerProperty("height", "Int",
		func(inst Any) Any { return Int(inst.(*Minipaint).Height()) },
		nil,
	)

	// ( x:i y:i c:i -- )
	t.registerNativeMethod("putPixel", func(f *RunFrame, ths *TypeObjPair) {
		c := int(f.RP.DataStack.Pop().(Int))
		y := int(f.RP.DataStack.Pop().(Int))
		x := int(f.RP.DataStack.Pop().(Int))
		ths.Obj.(*Minipaint).PutPixel(x, y, c)
	})
	// ( x:i y:i -- )  uses .ink
	t.registerNativeMethod("pixel", func(f *RunFrame, ths *TypeObjPair) {
		y := int(f.RP.DataStack.Pop().(Int))
		x := int(f.RP.DataStack.Pop().(Int))
		ths.Obj.(*Minipaint).Pixel(x, y)
	})
	dsInk := func(f *RunFrame, ths *TypeObjPair) (*Minipaint, *Stack[any], int) {
		mp := ths.Obj.(*Minipaint)
		return mp, f.RP.DataStack, mp.ink
	}
	// ( x0:i y0:i x1:i y1:i -- )
	t.registerNativeMethod("drawLine", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		y1 := popMinipaintInt(ds)
		x1 := popMinipaintInt(ds)
		y0 := popMinipaintInt(ds)
		x0 := popMinipaintInt(ds)
		mp.DrawLine(x0, y0, x1, y1, c)
	})
	// ( x:i y:i h:i -- )
	t.registerNativeMethod("drawFastVLine", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		h := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawFastVLine(x, y, h, c)
	})
	// ( x:i y:i w:i -- )
	t.registerNativeMethod("drawFastHLine", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		w := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawFastHLine(x, y, w, c)
	})
	// ( x:i y:i w:i h:i -- )
	t.registerNativeMethod("drawRect", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		h := popMinipaintInt(ds)
		w := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawRect(x, y, w, h, c)
	})
	// ( x:i y:i w:i h:i -- )
	t.registerNativeMethod("fillRect", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		h := popMinipaintInt(ds)
		w := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.FillRect(x, y, w, h, c)
	})
	// ( -- )
	t.registerNativeMethod("fillScreen", func(f *RunFrame, ths *TypeObjPair) {
		mp := ths.Obj.(*Minipaint)
		mp.FillScreen(mp.ink)
	})
	// ( x:i y:i r:i -- )
	t.registerNativeMethod("drawCircle", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		r := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawCircle(x, y, r, c)
	})
	// ( x:i y:i r:i -- )
	t.registerNativeMethod("fillCircle", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		r := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.FillCircle(x, y, r, c)
	})
	// ( x:i y:i rx:i ry:i -- )
	t.registerNativeMethod("drawEllipse", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		ry := popMinipaintInt(ds)
		rx := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawEllipse(x, y, rx, ry, c)
	})
	// ( x:i y:i rx:i ry:i -- )
	t.registerNativeMethod("fillEllipse", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		ry := popMinipaintInt(ds)
		rx := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.FillEllipse(x, y, rx, ry, c)
	})
	// ( x0:i y0:i x1:i y1:i x2:i y2:i -- )
	t.registerNativeMethod("drawTriangle", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		y2 := popMinipaintInt(ds)
		x2 := popMinipaintInt(ds)
		y1 := popMinipaintInt(ds)
		x1 := popMinipaintInt(ds)
		y0 := popMinipaintInt(ds)
		x0 := popMinipaintInt(ds)
		mp.DrawTriangle(x0, y0, x1, y1, x2, y2, c)
	})
	// ( x0:i y0:i x1:i y1:i x2:i y2:i -- )
	t.registerNativeMethod("fillTriangle", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		y2 := popMinipaintInt(ds)
		x2 := popMinipaintInt(ds)
		y1 := popMinipaintInt(ds)
		x1 := popMinipaintInt(ds)
		y0 := popMinipaintInt(ds)
		x0 := popMinipaintInt(ds)
		mp.FillTriangle(x0, y0, x1, y1, x2, y2, c)
	})
	// ( x:i y:i w:i h:i r:i -- )
	t.registerNativeMethod("drawRoundRect", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		r := popMinipaintInt(ds)
		h := popMinipaintInt(ds)
		w := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.DrawRoundRect(x, y, w, h, r, c)
	})
	// ( x:i y:i w:i h:i r:i -- )
	t.registerNativeMethod("fillRoundRect", func(f *RunFrame, ths *TypeObjPair) {
		mp, ds, c := dsInk(f, ths)
		r := popMinipaintInt(ds)
		h := popMinipaintInt(ds)
		w := popMinipaintInt(ds)
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		mp.FillRoundRect(x, y, w, h, r, c)
	})
	// ( x:i y:i -- )
	t.registerNativeMethod("setCursor", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		ths.Obj.(*Minipaint).SetCursor(x, y)
	})
	// ( x:i y:i c:i -- )  classic glyph; uses .ink / .bg / .textSize
	t.registerNativeMethod("drawChar", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		c := byte(popMinipaintInt(ds))
		y := popMinipaintInt(ds)
		x := popMinipaintInt(ds)
		ths.Obj.(*Minipaint).DrawChar(x, y, c)
	})
	// ( c:i -- )  write at cursor and advance
	t.registerNativeMethod("writeChar", func(f *RunFrame, ths *TypeObjPair) {
		c := byte(popMinipaintInt(f.RP.DataStack))
		ths.Obj.(*Minipaint).WriteChar(c)
	})
	// ( s:s -- )  print at cursor
	t.registerNativeMethod("print", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*Minipaint).Print(s)
	})
	// ( src:BlitSrc -- mp:Minipaint )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		src := f.RP.DataStack.Pop().(*BlitSrc)
		f.RP.DataStack.Push(NewMinipaint(src))
	})
}

func CreateRegisterMinipaintType() {
	t := NewRegisteringType("Minipaint",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Minipaint]("Minipaint"),
	)
	registerMinipaintMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	minipaintType = t
}

var (
	minipaintType *Type
	_             Typed = (*Minipaint)(nil)
)

func init() {
	CreateRegisterMinipaintType()
}
