package core

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

// Pixmap is a rectangular raster of color.RGBA (premultiplied) cells.
type Pixmap struct {
	P          *image.RGBA
	params     *BlitParams
	viewOrigin *Point // nil = identity; exact (0,0) stores as nil (180.4)
}

func (pm *Pixmap) FinalType() *Type { return pixmapType }

func (pm *Pixmap) Image() image.Image { return pm.P }

// ViewOrigin returns the world point mapped to pixmap (0,0).
// When unset / identity, returns Point.zero (Final never sees null).
func (pm *Pixmap) ViewOrigin() Point {
	if pm.viewOrigin == nil {
		return Point{}
	}
	return *pm.viewOrigin
}

// SetViewOrigin sets the world→view translation origin.
// Exact (0,0) disables translation (nil fast path). Point is copied; callers may reuse.
func (pm *Pixmap) SetViewOrigin(at Point) {
	if at.X == 0 && at.Y == 0 {
		pm.viewOrigin = nil
		return
	}
	cp := at
	pm.viewOrigin = &cp
}

// MapWorldToView maps a world point to pixmap coordinates: pixmap = world − viewOrigin.
// Identity when viewOrigin is unset.
func (pm *Pixmap) MapWorldToView(at Point) Point {
	if pm == nil || pm.viewOrigin == nil {
		return at
	}
	o := pm.viewOrigin
	return Point{X: at.X - o.X, Y: at.Y - o.Y}
}

// MapWorldXYToView is MapWorldToView for raw coordinates (PathBuilder vertices).
func (pm *Pixmap) MapWorldXYToView(x, y float64) (float64, float64) {
	if pm == nil || pm.viewOrigin == nil {
		return x, y
	}
	o := pm.viewOrigin
	return x - o.X, y - o.Y
}

// MapViewToWorld maps pixmap / pointer coordinates to world: world = pixmap + viewOrigin.
func (pm *Pixmap) MapViewToWorld(at Point) Point {
	if pm == nil || pm.viewOrigin == nil {
		return at
	}
	o := pm.viewOrigin
	return Point{X: at.X + o.X, Y: at.Y + o.Y}
}

func (pm *Pixmap) At(x, y int) color.Color {
	return pm.P.RGBAAt(x, y)
}
func (pm *Pixmap) Bounds() image.Rectangle {
	return pm.P.Bounds()
}
func (pm *Pixmap) ColorModel() color.Model {
	return pm.P.ColorModel()
}

func NewPixmap(rows, cols int) *Pixmap {
	return newPixmapFromRGBA(image.NewRGBA(image.Rect(0, 0, cols, rows)))
}

func newPixmapFromRGBA(p *image.RGBA) *Pixmap {
	return &Pixmap{P: p, params: NewBlitParamsDefaults()}
}

// Clear zeroes the pixmap (transparent). Does not replace P.
func (pm *Pixmap) Clear() {
	draw.Draw(pm.P, pm.P.Bounds(), &image.Uniform{colorTransparent}, image.Point{}, draw.Src)
}

// Fill stomps every pixel to c. Does not replace P.
func (pm *Pixmap) Fill(c Color) {
	draw.Draw(pm.P, pm.P.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
}

func (pm *Pixmap) Rows() int { return pm.P.Bounds().Dy() }
func (pm *Pixmap) Cols() int { return pm.P.Bounds().Dx() }

func (pm *Pixmap) xy(row, col int) (x, y int) {
	b := pm.P.Bounds()
	rows, cols := b.Dy(), b.Dx()
	if row < 0 || row >= rows || col < 0 || col >= cols {
		panic(fmt.Errorf("Pixmap: index (%d,%d) out of range [0,%d)×[0,%d)", row, col, rows, cols))
	}
	return b.Min.X + col, b.Min.Y + row
}

func (pm *Pixmap) xyFlat(i int) (x, y int) {
	b := pm.P.Bounds()
	rows, cols := b.Dy(), b.Dx()
	n := rows * cols
	if i < 0 || i >= n {
		panic(fmt.Errorf("Pixmap: flat index %d out of range [0,%d)", i, n))
	}
	row, col := i/cols, i%cols
	return b.Min.X + col, b.Min.Y + row
}

func colorToPremulRGBA(c Color) color.RGBA {
	return color.RGBA{R: c.rp, G: c.gp, B: c.bp, A: c.ap}
}

// colorFromPremulRGBA reads a premul store cell. Premul bytes are ground truth;
// straight R,G,B,A are derived (un-premul). A=0 straight channels are zero.
func colorFromPremulRGBA(px color.RGBA) Color {
	rp, gp, bp, ap := px.R, px.G, px.B, px.A
	var r, g, b, a uint8
	switch {
	case ap == 0:
		// straight RGB not recoverable from premul when alpha is zero
	case ap == 0xff:
		r, g, b, a = rp, gp, bp, ap
	default:
		a = ap
		r = uint8(uint16(rp) * 255 / uint16(ap))
		g = uint8(uint16(gp) * 255 / uint16(ap))
		b = uint8(uint16(bp) * 255 / uint16(ap))
	}
	return Color{R: r, G: g, B: b, A: a, rp: rp, gp: gp, bp: bp, ap: ap}
}

func (pm *Pixmap) putXY(x, y int, c Color) {
	pm.P.SetRGBA(x, y, colorToPremulRGBA(c))
}

func (pm *Pixmap) Put(row, col int, c Color) {
	x, y := pm.xy(row, col)
	pm.putXY(x, y, c)
}

func (pm *Pixmap) PutFlat(i int, c Color) {
	x, y := pm.xyFlat(i)
	pm.putXY(x, y, c)
}

// ToMat2C snapshots the pixmap premul buffer into a new Mat2C (row-major Color).
// A=0 store cells mint transparent Color; straight RGB from the buffer is not recovered.
func (pm *Pixmap) ToMat2C() *Mat2C {
	rows, cols := pm.Rows(), pm.Cols()
	m := NewMat2C(rows, cols)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			m.Put(row, col, colorFromPremulRGBA(pm.P.RGBAAt(col, row)))
		}
	}
	return m
}

func registerPixmapMembers(t *Type) {
	t.registerProperty("image", "Any",
		func(inst Any) Any { return inst.(*Pixmap).Image() },
		nil,
	)
	t.registerProperty("rows", "Int",
		func(inst Any) Any { return Int(inst.(*Pixmap).Rows()) },
		nil,
	)
	t.registerProperty("cols", "Int",
		func(inst Any) Any { return Int(inst.(*Pixmap).Cols()) },
		nil,
	)
	t.registerProperty("params", "BlitParams",
		func(inst Any) Any { return inst.(*Pixmap).params },
		nil,
	)
	// viewOrigin: world point → pixmap (0,0). Exact (0,0) / unset = identity (180.4).
	t.registerProperty("viewOrigin", "Point",
		func(inst Any) Any { return inst.(*Pixmap).ViewOrigin() },
		func(inst Any, value Any) { inst.(*Pixmap).SetViewOrigin(value.(Point)) },
	)
	// ( x:n y:n | at:Point -- ) — flex set matching blit/paint point slots
	t.registerNativeMethod("setViewOrigin", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Pixmap).SetViewOrigin(popPoint(f.RP.DataStack))
	})
	// ( at:Point | x y -- world:Point )
	t.registerNativeMethod("viewToWorld", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		f.RP.DataStack.Push(ths.Obj.(*Pixmap).MapViewToWorld(at))
	})
	// ( at:Point | x y -- view:Point )
	t.registerNativeMethod("worldToView", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		f.RP.DataStack.Push(ths.Obj.(*Pixmap).MapWorldToView(at))
	})
	// Write-only !() / ![] (no @() / @[]).
	t.registerItemsData2("Color",
		nil,
		func(inst Any, idx Any, value Any) {
			inst.(*Pixmap).PutFlat(int(idx.(Int)), value.(Color))
		},
		nil, nil, nil,
		nil,
		func(inst Any, value Any, idxs ...Any) {
			if len(idxs) != 2 {
				panic(fmt.Errorf("Pixmap ![]: need 2 indexes, got %d", len(idxs)))
			}
			inst.(*Pixmap).Put(int(idxs[0].(Int)), int(idxs[1].(Int)), value.(Color))
		},
		2,
	)
	t.registerNativeMethod("wrap", func(f *RunFrame, ths *TypeObjPair) {
		pm := ths.Obj.(*Pixmap)
		wrapMat2dOnFrame(f, pm.Rows(), pm.Cols())
	})
	// ( -- )
	t.registerNativeMethod("clear", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Pixmap).Clear()
	})
	// ( c:Color -- )
	t.registerNativeMethod("fill", func(f *RunFrame, ths *TypeObjPair) {
		c := f.RP.DataStack.Pop().(Color)
		ths.Obj.(*Pixmap).Fill(c)
	})
	// ( src:BlitSrc at:Point -- )
	t.registerNativeMethod("blit", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		src := f.RP.DataStack.Pop().(*BlitSrc)
		ths.Obj.(*Pixmap).Blit(src, at)
	})
	// ( src:BlitSrc over:BlitParams at:Point -- )
	t.registerNativeMethod("blitWithOverrides", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		over := f.RP.DataStack.Pop().(*BlitParams)
		src := f.RP.DataStack.Pop().(*BlitSrc)
		ths.Obj.(*Pixmap).BlitWithOverrides(src, over, at)
	})
	// ( pal:Palette1 mat:Mat2i1 at:Point -- )
	t.registerNativeMethod("blit1", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		mat := f.RP.DataStack.Pop().(*Mat2i1)
		pal := f.RP.DataStack.Pop().(*Palette1)
		ths.Obj.(*Pixmap).Blit1(pal, mat, at)
	})
	// ( pal:Palette4 mat:Mat2i4 at:Point -- )
	t.registerNativeMethod("blit4", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		mat := f.RP.DataStack.Pop().(*Mat2i4)
		pal := f.RP.DataStack.Pop().(*Palette4)
		ths.Obj.(*Pixmap).Blit4(pal, mat, at)
	})
	// ( pal:Palette8 mat:Mat2u8 at:Point -- )
	t.registerNativeMethod("blit8", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		mat := f.RP.DataStack.Pop().(*Mat2u8)
		pal := f.RP.DataStack.Pop().(*Palette8)
		ths.Obj.(*Pixmap).Blit8(pal, mat, at)
	})
	// ( mat:Mat2c at:Point -- )
	t.registerNativeMethod("blitC", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		mat := f.RP.DataStack.Pop().(*Mat2C)
		ths.Obj.(*Pixmap).BlitC(mat, at)
	})
	// ( -- mat:Mat2c )
	t.registerNativeMethod("toMat2C", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Pixmap).ToMat2C())
	})
	// ( rows:n cols:n -- pm:Pixmap )  dims must be exact integral
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		cols := RequireIntegralDim("Pixmap.new cols", float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop())))
		rows := RequireIntegralDim("Pixmap.new rows", float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop())))
		f.RP.DataStack.Push(NewPixmap(rows, cols))
	})
}

func CreateRegisterPixmapType() {
	t := NewRegisteringType("Pixmap",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Pixmap]("Pixmap"),
	)
	registerPixmapMembers(t)
	ApplyMixin(t, imageType, func(t *Type) {
		registerImageCodecMembers[*Pixmap](t)
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	pixmapType = t
}

var (
	pixmapType *Type
	_          Typed       = (*Pixmap)(nil)
	_          image.Image = (*Pixmap)(nil)
)

func init() {
	CreateRegisterPixmapType()
}
