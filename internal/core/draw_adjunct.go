package core

import (
	"fmt"
	"image"
	"math"

	"github.com/fogleman/gg"
)

// LineCap is a stroke end cap. Class consts: round, butt, square.
type LineCap string

const (
	LineCapRound  LineCap = "round"
	LineCapButt   LineCap = "butt"
	LineCapSquare LineCap = "square"
)

func (c LineCap) FinalType() *Type { return lineCapType }

func (c LineCap) ggCap() gg.LineCap {
	switch c {
	case LineCapRound:
		return gg.LineCapRound
	case LineCapButt:
		return gg.LineCapButt
	case LineCapSquare:
		return gg.LineCapSquare
	default:
		panic(fmt.Errorf("LineCap: unknown %q", c))
	}
}

// LineJoin is a stroke corner join. Class consts: round, bevel.
// gg v1.3.0 has no miter (only Round and Bevel).
type LineJoin string

const (
	LineJoinRound LineJoin = "round"
	LineJoinBevel LineJoin = "bevel"
)

func (j LineJoin) FinalType() *Type { return lineJoinType }

func (j LineJoin) ggJoin() gg.LineJoin {
	switch j {
	case LineJoinRound:
		return gg.LineJoinRound
	case LineJoinBevel:
		return gg.LineJoinBevel
	default:
		panic(fmt.Errorf("LineJoin: unknown %q", j))
	}
}

// LineStyle is a stroke dash style. Class consts: none, solid, dot, dash, dotDash.
type LineStyle string

const (
	LineStyleNone    LineStyle = "none"
	LineStyleSolid   LineStyle = "solid"
	LineStyleDot     LineStyle = "dot"
	LineStyleDash    LineStyle = "dash"
	LineStyleDotDash LineStyle = "dotDash"
)

func (s LineStyle) FinalType() *Type { return lineStyleType }

// dashPattern is in user units, scaled to line width. Offset is always 0.
// nil means SetDash() with no args (solid / none). unknown style: ok false.
func dashPattern(st LineStyle, width float64) (dashes []float64, ok bool) {
	switch st {
	case LineStyleNone, LineStyleSolid:
		return nil, true
	case LineStyleDot:
		return []float64{width, 3.5 * width}, true
	case LineStyleDash:
		return []float64{4 * width, 3.5 * width}, true
	case LineStyleDotDash:
		return []float64{4 * width, 4 * width, width, 4 * width}, true
	default:
		return nil, false
	}
}

// FillStyle is a fill mode. Phase 3.0: none, solid.
type FillStyle string

const (
	FillStyleNone         FillStyle = "none"
	FillStyleSolid        FillStyle = "solid"
	FillStyleCheckerboard FillStyle = "checkerboard"
	FillStyleGradient     FillStyle = "gradient"
	FillStyleTile         FillStyle = "tile"
)

const (
	defaultCheckerSize = 8
	defaultTileZoom    = 1.0
)

type fillGradientKind int

const (
	fillGradientUnset fillGradientKind = iota
	fillGradientFromTo
	fillGradientDir
)

func (s FillStyle) FinalType() *Type { return fillStyleType }

// Stroke is the Painter outline pen. Minted only with a Painter; p is never nil.
type Stroke struct {
	p         *Painter
	color     Color
	lineWidth float64
	lineCap   LineCap
	lineJoin  LineJoin
	lineStyle LineStyle
}

func (s *Stroke) FinalType() *Type { return strokeType }

func (s *Stroke) Color() Color { return s.color }

func (s *Stroke) SetColor(c Color) { s.color = c }

func (s *Stroke) LineWidth() float64 { return s.lineWidth }

func (s *Stroke) SetLineWidth(w float64) {
	if w <= 0 {
		panic(fmt.Errorf("Stroke.lineWidth: must be positive, got %v", w))
	}
	if s.lineWidth == w {
		return
	}
	s.lineWidth = w
	s.p.dc.SetLineWidth(w)
	s.applyDash()
}

func (s *Stroke) LineCap() LineCap { return s.lineCap }

func (s *Stroke) SetLineCap(c LineCap) {
	_ = c.ggCap()
	if s.lineCap == c {
		return
	}
	s.lineCap = c
	s.p.dc.SetLineCap(c.ggCap())
}

func (s *Stroke) LineJoin() LineJoin { return s.lineJoin }

func (s *Stroke) SetLineJoin(j LineJoin) {
	_ = j.ggJoin()
	if s.lineJoin == j {
		return
	}
	s.lineJoin = j
	s.p.dc.SetLineJoin(j.ggJoin())
}

func (s *Stroke) LineStyle() LineStyle { return s.lineStyle }

func (s *Stroke) SetLineStyle(st LineStyle) {
	if _, ok := dashPattern(st, s.lineWidth); !ok {
		panic(fmt.Errorf("Stroke.lineStyle: unknown %q", st))
	}
	if s.lineStyle == st {
		return
	}
	s.lineStyle = st
	s.applyDash()
}

func (s *Stroke) applyDash() {
	d, ok := dashPattern(s.lineStyle, s.lineWidth)
	if !ok {
		panic(fmt.Errorf("Stroke.lineStyle: unknown %q", s.lineStyle))
	}
	if d == nil {
		s.p.dc.SetDash()
		return
	}
	s.p.dc.SetDash(d...)
}

func newStroke(p *Painter) *Stroke {
	s := &Stroke{
		p:         p,
		color:     WebColors[ColorBlack],
		lineWidth: defaultPainterLineWidth,
		lineCap:   LineCapRound,
		lineJoin:  LineJoinRound,
		lineStyle: LineStyleSolid,
	}
	s.p.dc.SetLineWidth(s.lineWidth)
	s.p.dc.SetLineCap(s.lineCap.ggCap())
	s.p.dc.SetLineJoin(s.lineJoin.ggJoin())
	s.p.dc.SetDash()
	return s
}

// Fill is the Painter interior pen. Minted only with a Painter; p is never nil.
type Fill struct {
	p           *Painter
	color       Color
	color2      Color
	style       FillStyle
	checkerSize int
	tileImage   image.Image
	tileZoom    float64
	tileScaled  image.Image // cache; nil when zoom is 1 or dirty
	gradKind    fillGradientKind
	gradFrom    Point
	gradTo      Point
	gradDX      float64
	gradDY      float64
}

func (f *Fill) FinalType() *Type { return fillType }

func (f *Fill) Color() Color { return f.color }

func (f *Fill) SetColor(c Color) { f.color = c }

func (f *Fill) Style() FillStyle { return f.style }

func (f *Fill) SetStyle(st FillStyle) {
	switch st {
	case FillStyleNone, FillStyleSolid, FillStyleCheckerboard, FillStyleGradient, FillStyleTile:
		f.style = st
	default:
		panic(fmt.Errorf("Fill.style: unknown %q", st))
	}
}

func (f *Fill) Color2() Color { return f.color2 }

func (f *Fill) SetColor2(c Color) { f.color2 = c }

func (f *Fill) CheckerSize() int { return f.checkerSize }

func (f *Fill) SetCheckerSize(n int) {
	if n <= 0 {
		panic(fmt.Errorf("Fill.checkerSize: must be positive, got %d", n))
	}
	f.checkerSize = n
}

func (f *Fill) TileImage() image.Image { return f.tileImage }

func (f *Fill) SetTileImage(im image.Image) {
	if im == nil {
		panic(fmt.Errorf("Fill.tileImage: nil"))
	}
	f.tileImage = im
	f.tileScaled = nil
}

func (f *Fill) TileZoom() float64 { return f.tileZoom }

func (f *Fill) SetTileZoom(z float64) {
	if z <= 0 {
		panic(fmt.Errorf("Fill.tileZoom: must be positive, got %v", z))
	}
	if f.tileZoom == z {
		return
	}
	f.tileZoom = z
	f.tileScaled = nil
}

func (f *Fill) SetGradientFromTo(from, to Point) {
	f.gradKind = fillGradientFromTo
	f.gradFrom = from
	f.gradTo = to
}

func (f *Fill) SetGradientRect(r Rect) {
	from := r.TopLeft
	to := NewPointF(r.TopLeft.X+r.Size.W, r.TopLeft.Y+r.Size.H)
	f.SetGradientFromTo(from, to)
}

func (f *Fill) SetGradientDegrees(deg float64) {
	rad := gg.Radians(deg)
	f.SetGradientDir(math.Cos(rad), math.Sin(rad))
}

func (f *Fill) SetGradientDir(dx, dy float64) {
	if dx == 0 && dy == 0 {
		panic(fmt.Errorf("Fill.setGradientDir: zero vector"))
	}
	f.gradKind = fillGradientDir
	f.gradDX = dx
	f.gradDY = dy
}

func (f *Fill) fills() bool {
	return f.style != FillStyleNone
}

func newFill(p *Painter) *Fill {
	return &Fill{
		p:           p,
		color:       WebColors[ColorBlack],
		color2:      WebColors[ColorWhite],
		style:       FillStyleNone,
		checkerSize: defaultCheckerSize,
		tileZoom:    defaultTileZoom,
	}
}

func registerNamedConst[T any](t *Type, typeName, name string, v T) {
	t.registerClassProperty(name, typeName,
		func(_ Any) Any { return v },
		nil,
	)
}

func registerLineCapMembers(t *Type) {
	registerNamedConst(t, "LineCap", "round", LineCapRound)
	registerNamedConst(t, "LineCap", "butt", LineCapButt)
	registerNamedConst(t, "LineCap", "square", LineCapSquare)
}

func registerLineJoinMembers(t *Type) {
	registerNamedConst(t, "LineJoin", "round", LineJoinRound)
	registerNamedConst(t, "LineJoin", "bevel", LineJoinBevel)
}

func registerLineStyleMembers(t *Type) {
	registerNamedConst(t, "LineStyle", "none", LineStyleNone)
	registerNamedConst(t, "LineStyle", "solid", LineStyleSolid)
	registerNamedConst(t, "LineStyle", "dot", LineStyleDot)
	registerNamedConst(t, "LineStyle", "dash", LineStyleDash)
	registerNamedConst(t, "LineStyle", "dotDash", LineStyleDotDash)
}

func registerFillStyleMembers(t *Type) {
	registerNamedConst(t, "FillStyle", "none", FillStyleNone)
	registerNamedConst(t, "FillStyle", "solid", FillStyleSolid)
	registerNamedConst(t, "FillStyle", "checkerboard", FillStyleCheckerboard)
	registerNamedConst(t, "FillStyle", "gradient", FillStyleGradient)
	registerNamedConst(t, "FillStyle", "tile", FillStyleTile)
}

func registerStrokeMembers(t *Type) {
	t.registerProperty("color", "Color",
		func(inst Any) Any { return inst.(*Stroke).color },
		func(inst Any, value Any) { inst.(*Stroke).SetColor(value.(Color)) },
	)
	t.registerProperty("lineWidth", "Float",
		func(inst Any) Any { return Float(inst.(*Stroke).lineWidth) },
		func(inst Any, value Any) {
			inst.(*Stroke).SetLineWidth(float64(checkCoerceNumericToFloat(value)))
		},
	)
	t.registerProperty("lineCap", "LineCap",
		func(inst Any) Any { return inst.(*Stroke).lineCap },
		func(inst Any, value Any) { inst.(*Stroke).SetLineCap(value.(LineCap)) },
	)
	t.registerProperty("lineJoin", "LineJoin",
		func(inst Any) Any { return inst.(*Stroke).lineJoin },
		func(inst Any, value Any) { inst.(*Stroke).SetLineJoin(value.(LineJoin)) },
	)
	t.registerProperty("lineStyle", "LineStyle",
		func(inst Any) Any { return inst.(*Stroke).lineStyle },
		func(inst Any, value Any) { inst.(*Stroke).SetLineStyle(value.(LineStyle)) },
	)
}

func registerFillMembers(t *Type) {
	t.registerProperty("color", "Color",
		func(inst Any) Any { return inst.(*Fill).color },
		func(inst Any, value Any) { inst.(*Fill).SetColor(value.(Color)) },
	)
	t.registerProperty("color2", "Color",
		func(inst Any) Any { return inst.(*Fill).color2 },
		func(inst Any, value Any) { inst.(*Fill).SetColor2(value.(Color)) },
	)
	t.registerProperty("style", "FillStyle",
		func(inst Any) Any { return inst.(*Fill).style },
		func(inst Any, value Any) { inst.(*Fill).SetStyle(value.(FillStyle)) },
	)
	t.registerProperty("checkerSize", "Int",
		func(inst Any) Any { return Int(inst.(*Fill).checkerSize) },
		func(inst Any, value Any) { inst.(*Fill).SetCheckerSize(int(value.(Int))) },
	)
	t.registerProperty("tileImage", "Any",
		func(inst Any) Any { return inst.(*Fill).tileImage },
		func(inst Any, value Any) { inst.(*Fill).SetTileImage(asTileImage(value)) },
	)
	t.registerProperty("tileZoom", "Float",
		func(inst Any) Any { return Float(inst.(*Fill).tileZoom) },
		func(inst Any, value Any) {
			inst.(*Fill).SetTileZoom(float64(checkCoerceNumericToFloat(value)))
		},
	)
	// ( from:Point to:Point -- )
	t.registerNativeMethod("setGradientFromTo", func(f *RunFrame, ths *TypeObjPair) {
		to := popPoint(f.RP.DataStack)
		from := popPoint(f.RP.DataStack)
		ths.Obj.(*Fill).SetGradientFromTo(from, to)
	})
	// ( r:Rect -- )
	t.registerNativeMethod("setGradientRect", func(f *RunFrame, ths *TypeObjPair) {
		r := f.RP.DataStack.Pop().(Rect)
		ths.Obj.(*Fill).SetGradientRect(r)
	})
	// ( deg:n -- )  0° = right, 90° = down
	t.registerNativeMethod("setGradientDegrees", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Fill).SetGradientDegrees(popDrawFloat(f.RP.DataStack))
	})
	// ( dir:Point | dx dy -- )
	t.registerNativeMethod("setGradientDir", func(f *RunFrame, ths *TypeObjPair) {
		d := popPoint(f.RP.DataStack)
		ths.Obj.(*Fill).SetGradientDir(d.X, d.Y)
	})
}

func CreateRegisterLineCapType() {
	t := &Type{
		name:            "LineCap",
		getZero:         MakeHostTypeZero[LineCap](),
		checkCoerceType: MakeHostTypeCheckCoerce[LineCap]("LineCap"),
	}
	registerLineCapMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	lineCapType = t
}

func CreateRegisterLineJoinType() {
	t := &Type{
		name:            "LineJoin",
		getZero:         MakeHostTypeZero[LineJoin](),
		checkCoerceType: MakeHostTypeCheckCoerce[LineJoin]("LineJoin"),
	}
	registerLineJoinMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	lineJoinType = t
}

func CreateRegisterLineStyleType() {
	t := &Type{
		name:            "LineStyle",
		getZero:         MakeHostTypeZero[LineStyle](),
		checkCoerceType: MakeHostTypeCheckCoerce[LineStyle]("LineStyle"),
	}
	registerLineStyleMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	lineStyleType = t
}

func CreateRegisterFillStyleType() {
	t := &Type{
		name:            "FillStyle",
		getZero:         MakeHostTypeZero[FillStyle](),
		checkCoerceType: MakeHostTypeCheckCoerce[FillStyle]("FillStyle"),
	}
	registerFillStyleMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	fillStyleType = t
}

func CreateRegisterStrokeType() {
	t := NewRegisteringType("Stroke",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Stroke]("Stroke"),
	)
	registerStrokeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	strokeType = t
}

func CreateRegisterFillType() {
	t := NewRegisteringType("Fill",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Fill]("Fill"),
	)
	registerFillMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	fillType = t
}

var (
	lineCapType   *Type
	lineJoinType  *Type
	lineStyleType *Type
	fillStyleType *Type
	strokeType    *Type
	fillType      *Type
	_             Typed = (*LineCap)(nil)
	_             Typed = (*LineJoin)(nil)
	_             Typed = (*LineStyle)(nil)
	_             Typed = (*FillStyle)(nil)
	_             Typed = (*Stroke)(nil)
	_             Typed = (*Fill)(nil)
)

func init() {
	CreateRegisterLineCapType()
	CreateRegisterLineJoinType()
	CreateRegisterLineStyleType()
	CreateRegisterFillStyleType()
	CreateRegisterStrokeType()
	CreateRegisterFillType()
}
