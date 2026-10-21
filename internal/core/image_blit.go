package core

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// BlitType is the per-pixel composite. Ordinals match Go's draw.Op so
// workers can use the value directly: blend=Over=0, stomp=Src=1.
type BlitType int

const (
	BlitBlend BlitType = BlitType(draw.Over) // 0
	BlitStomp BlitType = BlitType(draw.Src)  // 1
)

func (bt BlitType) FinalType() *Type { return blitTypeType }

// Params is the concrete Pixmap blit base (no Inherit). Final: .params
func (pm *Pixmap) Params() *BlitParams { return pm.params }

// blitWin is the once-at-entry stamp ∩ pixmap window (like image/draw).
// Cell loops cover [r0,r1)×[c0,c1); edge cells may be partial when cellPx > 1.
// ax/ay are rounded integral dest origin (from Point.Xi/Yi at clip entry).
type blitWin struct {
	ax, ay int
	clip   image.Rectangle // absolute dest pixels to write
	r0, c0 int
	r1, c1 int
	cellPx int
}

// blitClipWindow intersects the stamp with the pixmap. empty ⇒ no-op.
// Dest at is world coordinates; viewOrigin translates before clip (180.4).
func (pm *Pixmap) blitClipWindow(at Point, rows, cols, cellPx int) (blitWin, bool) {
	b := pm.P.Bounds()
	at = pm.MapWorldToView(at)
	ax, ay := at.Xi(), at.Yi()
	stamp := image.Rect(ax, ay, ax+cols*cellPx, ay+rows*cellPx)
	clip := stamp.Intersect(b)
	if clip.Empty() {
		return blitWin{}, false
	}
	ox0 := clip.Min.X - ax
	oy0 := clip.Min.Y - ay
	ox1 := clip.Max.X - ax
	oy1 := clip.Max.Y - ay
	c0 := ox0 / cellPx
	r0 := oy0 / cellPx
	c1 := (ox1 + cellPx - 1) / cellPx
	r1 := (oy1 + cellPx - 1) / cellPx
	if c0 < 0 {
		c0 = 0
	}
	if r0 < 0 {
		r0 = 0
	}
	if c1 > cols {
		c1 = cols
	}
	if r1 > rows {
		r1 = rows
	}
	if c0 >= c1 || r0 >= r1 {
		return blitWin{}, false
	}
	return blitWin{ax: ax, ay: ay, clip: clip, r0: r0, c0: c0, r1: r1, c1: c1, cellPx: cellPx}, true
}

// blitSink writes color.RGBA cells into the Pixmap store.
// Stomp copies 4 bytes; blend is Porter-Duff Over in the 8-bit premul store.
type blitSink struct {
	rgba *image.RGBA
	op   BlitType
}

func newBlitSink(pm *Pixmap, op BlitType) blitSink {
	return blitSink{rgba: pm.P, op: op}
}

func (s *blitSink) put(x, y int, src color.RGBA) {
	i := s.rgba.PixOffset(x, y)
	pix := s.rgba.Pix
	if s.op == BlitStomp {
		pix[i+0] = src.R
		pix[i+1] = src.G
		pix[i+2] = src.B
		pix[i+3] = src.A
		return
	}
	dst := color.RGBA{R: pix[i+0], G: pix[i+1], B: pix[i+2], A: pix[i+3]}
	out := rgbaOver(dst, src)
	pix[i+0] = out.R
	pix[i+1] = out.G
	pix[i+2] = out.B
	pix[i+3] = out.A
}

func (s *blitSink) fillCell(x0, y0, cellPx int, src color.RGBA, clip image.Rectangle) {
	cell := image.Rect(x0, y0, x0+cellPx, y0+cellPx).Intersect(clip)
	for y := cell.Min.Y; y < cell.Max.Y; y++ {
		for x := cell.Min.X; x < cell.Max.X; x++ {
			s.put(x, y, src)
		}
	}
}

// rgbaOver is src-over-dst on premul 8-bit RGBA (same arithmetic as draw.Over onto *image.RGBA).
func rgbaOver(dst, src color.RGBA) color.RGBA {
	if src.A == 0 {
		return dst
	}
	if src.A == 0xff {
		return src
	}
	sr, sg, sb, sa := src.RGBA()
	a := (0xffff - sa) * 0x101
	return color.RGBA{
		R: uint8((uint32(dst.R)*a + sr*0xffff) / 0xffff >> 8),
		G: uint8((uint32(dst.G)*a + sg*0xffff) / 0xffff >> 8),
		B: uint8((uint32(dst.B)*a + sb*0xffff) / 0xffff >> 8),
		A: uint8((uint32(dst.A)*a + sa*0xffff) / 0xffff >> 8),
	}
}

// Blit stamps src using the Pixmap's concrete params base.
// Final: ( src:BlitSrc at:Point -- )
func (pm *Pixmap) Blit(src *BlitSrc, at Point) {
	pm.blitInner(src, pm.params, at)
}

// BlitWithOverrides resolves over against the Pixmap params base, then stamps.
// Final: ( src:BlitSrc over:BlitParams at:Point -- )
func (pm *Pixmap) BlitWithOverrides(src *BlitSrc, over *BlitParams, at Point) {
	pm.blitInner(src, resolveBlitParams(pm.params, over), at)
}

// blitInner stamps with concrete params (no Inherit). Dest overhang is
// intersection-clipped once at entry (image/draw style); mask skips index 0 / A==0 when on.
func (pm *Pixmap) blitInner(src *BlitSrc, params *BlitParams, at Point) {
	if src == nil {
		panic(fmt.Errorf("Pixmap.blit: src is nil"))
	}
	if params == nil {
		panic(fmt.Errorf("Pixmap.blit: params is nil"))
	}
	cellPx := params.cellPx
	if cellPx < 1 {
		panic(fmt.Errorf("Pixmap.blit.cellPx: must be >= 1, got %d", cellPx))
	}
	rows, cols := src.rowsCols()
	win, ok := pm.blitClipWindow(at, rows, cols, cellPx)
	if !ok {
		return
	}
	s := newBlitSink(pm, BlitType(params.op))
	mask := params.mask == BlitMaskOn
	switch src.kind {
	case blitSrc1:
		pal := src.pal.(*Palette1).premulRGBA()
		mat := src.mat.(*Mat2i1)
		blit1(&s, pal, mat, win, mask)
	case blitSrc4:
		pal := src.pal.(*Palette4).premulRGBA()
		mat := src.mat.(*Mat2i4)
		blit4(&s, pal, mat, win, mask)
	case blitSrc8:
		pal := src.pal.(*Palette8).premulRGBA()
		mat := src.mat.(*Mat2u8)
		blit8(&s, pal, mat, win, mask)
	case blitSrcC:
		mat := src.mat.(*Mat2C)
		blitC(&s, mat, win, mask)
	default:
		panic(fmt.Errorf("Pixmap.blit: unknown BlitSrc kind %d", src.kind))
	}
}

// Compatibility wrappers — prefer Blit / BlitWithOverrides + BlitSrc.
func (pm *Pixmap) Blit1(pal *Palette1, mat *Mat2i1, at Point) {
	pm.Blit(NewBlitSrc1(pal, mat), at)
}

func (pm *Pixmap) Blit4(pal *Palette4, mat *Mat2i4, at Point) {
	pm.Blit(NewBlitSrc4(pal, mat), at)
}

func (pm *Pixmap) Blit8(pal *Palette8, mat *Mat2u8, at Point) {
	pm.Blit(NewBlitSrc8(pal, mat), at)
}

func (pm *Pixmap) BlitC(mat *Mat2C, at Point) {
	pm.Blit(NewBlitSrcC(mat), at)
}

func blit1(s *blitSink, pal []color.RGBA, mat *Mat2i1, win blitWin, mask bool) {
	cp := win.cellPx
	for r := win.r0; r < win.r1; r++ {
		y := win.ay + r*cp
		for c := win.c0; c < win.c1; c++ {
			idx := int(mat.At(r, c))
			if mask && idx == 0 {
				continue
			}
			x := win.ax + c*cp
			if cp == 1 {
				s.put(x, y, pal[idx])
			} else {
				s.fillCell(x, y, cp, pal[idx], win.clip)
			}
		}
	}
}

func blit4(s *blitSink, pal []color.RGBA, mat *Mat2i4, win blitWin, mask bool) {
	cp := win.cellPx
	for r := win.r0; r < win.r1; r++ {
		y := win.ay + r*cp
		for c := win.c0; c < win.c1; c++ {
			idx := int(mat.At(r, c))
			if mask && idx == 0 {
				continue
			}
			x := win.ax + c*cp
			if cp == 1 {
				s.put(x, y, pal[idx])
			} else {
				s.fillCell(x, y, cp, pal[idx], win.clip)
			}
		}
	}
}

func blit8(s *blitSink, pal []color.RGBA, mat *Mat2u8, win blitWin, mask bool) {
	cp := win.cellPx
	for r := win.r0; r < win.r1; r++ {
		y := win.ay + r*cp
		for c := win.c0; c < win.c1; c++ {
			idx := int(mat.At_(r, c))
			if mask && idx == 0 {
				continue
			}
			x := win.ax + c*cp
			if cp == 1 {
				s.put(x, y, pal[idx])
			} else {
				s.fillCell(x, y, cp, pal[idx], win.clip)
			}
		}
	}
}

func blitC(s *blitSink, mat *Mat2C, win blitWin, mask bool) {
	cp := win.cellPx
	for r := win.r0; r < win.r1; r++ {
		y := win.ay + r*cp
		for c := win.c0; c < win.c1; c++ {
			col := mat.At_(r, c)
			if mask && col.A == 0 {
				continue
			}
			x := win.ax + c*cp
			src := colorToPremulRGBA(col)
			if cp == 1 {
				s.put(x, y, src)
			} else {
				s.fillCell(x, y, cp, src, win.clip)
			}
		}
	}
}

func registerBlitTypeMembers(t *Type) {
	registerNamedConst(t, "BlitType", "blend", BlitBlend)
	registerNamedConst(t, "BlitType", "stomp", BlitStomp)
}

func CreateRegisterBlitTypeType() {
	t := &Type{
		name:            "BlitType",
		getZero:         MakeHostTypeZero[BlitType](),
		checkCoerceType: MakeHostTypeCheckCoerce[BlitType]("BlitType"),
	}
	registerBlitTypeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	blitTypeType = t
}

var (
	blitTypeType *Type
	_            Typed = (*BlitType)(nil)
)

func init() {
	CreateRegisterBlitTypeType()
}
