package core

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/fogleman/gg"
)

const (
	defaultPainterLineWidth = 1
	painterStateStackMax    = 64
)

// PainterState is the current Stroke/Fill/Font. Push/pop swaps this pointer.
type PainterState struct {
	stroke *Stroke
	fill   *Fill
	font   *Font
}

// ClonePainterState value-copies Stroke, Fill, and Font into a new state.
// Copies p and Font.face; does not Set* or rebuildFace.
func ClonePainterState(ps *PainterState) *PainterState {
	sc := *ps.stroke
	fc := *ps.fill
	fn := *ps.font
	return &PainterState{
		stroke: &sc,
		fill:   &fc,
		font:   &fn,
	}
}

// Painter draws onto a Pixmap (*image.RGBA; gg.NewContextForRGBA, no copy).
// Default factory mints a new Pixmap; newWithPixmap or .pixmap! bind an existing one.
type Painter struct {
	pixmap      *Pixmap
	dc          *gg.Context
	ps          *PainterState
	stack       *Stack[*PainterState] // nil until first Push
	pathBuilder *PathBuilder          // singleton; not PainterState
}

func (p *Painter) FinalType() *Type { return painterType }

func (p *Painter) Pixmap() *Pixmap { return p.pixmap }

func (p *Painter) Stroke() *Stroke { return p.ps.stroke }

func (p *Painter) Fill() *Fill { return p.ps.fill }

func (p *Painter) Font() *Font { return p.ps.font }

func (p *Painter) PathBuilder() *PathBuilder { return p.pathBuilder }

func (p *Painter) Push() {
	p.requirePathReady("push")
	neu := ClonePainterState(p.ps)
	if p.stack == nil {
		p.stack = NewStack[*PainterState]("Painter.stack", painterStateStackMax)
	}
	p.stack.Push(p.ps)
	p.ps = neu
	p.dc.Push()
}

func (p *Painter) Pop() {
	p.requirePathReady("pop")
	if p.stack == nil || p.stack.Len() == 0 {
		panic(fmt.Errorf("Painter.pop: no saved state"))
	}
	p.ps = p.stack.Pop()
	p.dc.Pop()
}

func setDCColor(dc *gg.Context, c Color) {
	dc.SetRGBA255(int(c.R), int(c.G), int(c.B), int(c.A))
}

func (p *Painter) fillThenStroke(bounds image.Rectangle) {
	fill := p.ps.fill
	stroke := p.ps.stroke
	filled := fill.fills()
	stroked := stroke.lineStyle != LineStyleNone
	if filled {
		fill.apply(p.dc, bounds)
		if stroked {
			p.dc.FillPreserve()
		} else {
			p.dc.Fill()
		}
	}
	if stroked {
		p.dc.SetStrokeStyle(gg.NewSolidPattern(stroke.color))
		p.dc.Stroke()
	} else if !filled {
		p.dc.ClearPath()
	}
}

// mapPt applies Pixmap.viewOrigin world→view translation (180.4).
func (p *Painter) mapPt(at Point) Point {
	return p.pixmap.MapWorldToView(at)
}

func (p *Painter) Line(from, to Point) {
	p.requirePathReady("line")
	p.dc.ClearPath()
	if p.ps.stroke.lineStyle == LineStyleNone {
		return
	}
	from, to = p.mapPt(from), p.mapPt(to)
	p.dc.DrawLine(from.X, from.Y, to.X, to.Y)
	setDCColor(p.dc, p.ps.stroke.color)
	p.dc.Stroke()
}

func (p *Painter) Circle(ctr Point, radius float64) {
	p.requirePathReady("circle")
	if radius < 0 {
		panic(fmt.Errorf("Painter.circle: radius must be >= 0, got %v", radius))
	}
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.DrawCircle(ctr.X, ctr.Y, radius)
	p.fillThenStroke(circleAABB(ctr, radius))
}

func (p *Painter) Ellipse(ctr Point, rx, ry float64) {
	p.requirePathReady("ellipse")
	p.validateEllipseRadii(rx, ry)
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.DrawEllipse(ctr.X, ctr.Y, rx, ry)
	p.fillThenStroke(ellipseAABB(ctr, rx, ry))
}

func (p *Painter) EllipseRot(ctr Point, rx, ry, rotationDegrees float64) {
	if rotationDegrees == 0 {
		p.Ellipse(ctr, rx, ry)
		return
	}
	p.requirePathReady("ellipseRot")
	p.validateEllipseRadii(rx, ry)
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.Push()
	p.dc.Translate(ctr.X, ctr.Y)
	p.dc.Rotate(gg.Radians(rotationDegrees))
	p.dc.DrawEllipse(0, 0, rx, ry)
	p.dc.Pop()
	p.fillThenStroke(rotatedEllipseAABB(ctr, rx, ry, rotationDegrees))
}

func (p *Painter) Rect(ctr Point, size Size) {
	p.requirePathReady("rect")
	p.validateRectSize(size, "rect")
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.DrawRectangle(ctr.X-size.W/2, ctr.Y-size.H/2, size.W, size.H)
	p.fillThenStroke(rectAABB(ctr, size))
}

func (p *Painter) RectRot(ctr Point, size Size, rotationDegrees float64) {
	if rotationDegrees == 0 {
		p.Rect(ctr, size)
		return
	}
	p.requirePathReady("rectRot")
	p.validateRectSize(size, "rectRot")
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.Push()
	p.dc.Translate(ctr.X, ctr.Y)
	p.dc.Rotate(gg.Radians(rotationDegrees))
	p.dc.DrawRectangle(-size.W/2, -size.H/2, size.W, size.H)
	p.dc.Pop()
	p.fillThenStroke(rotatedRectAABB(ctr, size, rotationDegrees))
}

func (p *Painter) RoundedRect(ctr Point, size Size, radius float64) {
	p.requirePathReady("roundedRect")
	p.validateRoundedRect(size, radius, "roundedRect")
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.DrawRoundedRectangle(
		ctr.X-size.W/2, ctr.Y-size.H/2,
		size.W, size.H,
		radius,
	)
	p.fillThenStroke(rectAABB(ctr, size))
}

func (p *Painter) RoundedRectRot(ctr Point, size Size, radius, rotationDegrees float64) {
	if rotationDegrees == 0 {
		p.RoundedRect(ctr, size, radius)
		return
	}
	p.requirePathReady("roundedRectRot")
	p.validateRoundedRect(size, radius, "roundedRectRot")
	p.dc.ClearPath()
	ctr = p.mapPt(ctr)
	p.dc.Push()
	p.dc.Translate(ctr.X, ctr.Y)
	p.dc.Rotate(gg.Radians(rotationDegrees))
	p.dc.DrawRoundedRectangle(-size.W/2, -size.H/2, size.W, size.H, radius)
	p.dc.Pop()
	p.fillThenStroke(rotatedRectAABB(ctr, size, rotationDegrees))
}

func (p *Painter) RegularPolygon(sides int, center Point, radius float64) {
	p.RegularPolygonRot(sides, center, radius, 0)
}

func (p *Painter) RegularPolygonRot(sides int, center Point, radius, rotationDegrees float64) {
	p.requirePathReady("regularPolygon")
	if sides < 3 {
		panic(fmt.Errorf("Painter.regularPolygon: sides must be >= 3, got %d", sides))
	}
	if radius < 0 {
		panic(fmt.Errorf("Painter.regularPolygon: radius must be >= 0, got %v", radius))
	}
	p.dc.ClearPath()
	center = p.mapPt(center)
	p.dc.DrawRegularPolygon(
		sides,
		center.X,
		center.Y,
		radius,
		gg.Radians(rotationDegrees),
	)
	p.fillThenStroke(circleAABB(center, radius))
}

func (p *Painter) validateRectSize(size Size, what string) {
	if size.W < 0 || size.H < 0 {
		panic(fmt.Errorf("Painter.%s: size must be non-negative, got %s", what, size))
	}
}

func (p *Painter) validateRoundedRect(size Size, radius float64, what string) {
	p.validateRectSize(size, what)
	if radius < 0 {
		panic(fmt.Errorf("Painter.%s: radius must be >= 0, got %v", what, radius))
	}
}

func (p *Painter) validateEllipseRadii(rx, ry float64) {
	if rx < 0 || ry < 0 {
		panic(fmt.Errorf("Painter.ellipse: radii must be >= 0, got rx=%v ry=%v", rx, ry))
	}
}

// Clear wipes the existing RGBA to transparent zeros (same buffer as the
// minter). Pixmap, P, and dc stay the same; Stroke/Fill/Font colors unchanged.
func (p *Painter) Clear() {
	img := p.pixmap.P
	draw.Draw(img, img.Bounds(), &image.Uniform{colorTransparent}, image.Point{}, draw.Src)
}

// Pixel writes one cell on the Painter's RGBA pixmap (not gg.DrawPoint).
// No-op if stroke.lineStyle is none. Does not change Stroke/Fill/Font fields.
func (p *Painter) Pixel(at Point) {
	stroke := p.ps.stroke
	if stroke.lineStyle == LineStyleNone {
		return
	}
	p.writePixel(at, stroke.color)
}

// PixelC writes one cell with an explicit color (not stroke). Always writes;
// lineStyle is ignored. Does not change Stroke/Fill/Font fields.
func (p *Painter) PixelC(c Color, at Point) {
	p.writePixel(at, c)
}

func (p *Painter) writePixel(at Point, c Color) {
	img := p.pixmap.P
	b := img.Bounds()
	at = p.mapPt(at)
	x, y := at.Xi(), at.Yi()
	// Clip like blit dest: world points outside the current view are no-ops (180.4).
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	p.pixmap.putXY(x, y, c)
}

func (p *Painter) Text(s string, at Point) {
	setDCColor(p.dc, p.ps.font.color)
	p.drawAlignedText(s, p.mapPt(at))
}

func (p *Painter) TextWrap(s string, at Point, width float64) {
	setDCColor(p.dc, p.ps.font.color)
	p.drawWrappedText(s, p.mapPt(at), width)
}

// popDrawFloat consumes a numeric cell (Int, Uint, or Float).
func popDrawFloat(cs *Stack[any]) float64 {
	return float64(checkCoerceNumericToFloat(cs.Pop()))
}

func requirePixmap(pm *Pixmap, op string) {
	if pm == nil || pm.P == nil {
		panic(fmt.Errorf("%s: pixmap must not be nil", op))
	}
}

func popPixmap(ds *Stack[any], op string) *Pixmap {
	v := ds.Pop()
	pm, ok := v.(*Pixmap)
	if !ok || pm == nil {
		panic(fmt.Errorf("%s: want Pixmap, got %T", op, v))
	}
	requirePixmap(pm, op)
	return pm
}

func newPainterWithPixmap(pm *Pixmap) *Painter {
	requirePixmap(pm, "Painter")
	p := &Painter{
		pixmap: pm,
		dc:     gg.NewContextForRGBA(pm.P),
	}
	p.ps = &PainterState{
		stroke: newStroke(p),
		fill:   newFill(p),
		font:   newFont(p),
	}
	p.pathBuilder = newPathBuilder(p)
	return p
}

func (p *Painter) bindPixmap(pm *Pixmap) {
	requirePixmap(pm, "Painter.pixmap")
	p.pixmap = pm
	p.dc = gg.NewContextForRGBA(pm.P)
	if p.pathBuilder != nil {
		p.pathBuilder.mode = pathBuilderReady
		p.pathBuilder.resetBounds()
	}
}

func setPainterPixmapFromAny(p *Painter, val Any) {
	if val == nil {
		panic(fmt.Errorf("Painter.pixmap: pixmap must not be nil"))
	}
	pm, ok := val.(*Pixmap)
	if !ok || pm == nil {
		panic(fmt.Errorf("Painter.pixmap: want Pixmap, got %T", val))
	}
	p.bindPixmap(pm)
}

// SetPixmap rebinds drawing to an existing Pixmap (same Painter, new gg context).
func (p *Painter) SetPixmap(pm *Pixmap) {
	p.bindPixmap(pm)
}

// NewPainter mints an RGBA Pixmap of sz (W×H, both > 0, exact integral) and binds gg to it.
func NewPainter(sz Size) *Painter {
	w := RequireIntegralDim("Painter.new width", sz.W)
	h := RequireIntegralDim("Painter.new height", sz.H)
	if w <= 0 || h <= 0 {
		panic(fmt.Errorf("Painter.new: size must be positive, got %s", sz))
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	return newPainterWithPixmap(newPixmapFromRGBA(rgba))
}

// NewPainterWithPixmap binds gg to an existing Pixmap (no new buffer).
func NewPainterWithPixmap(pm *Pixmap) *Painter {
	requirePixmap(pm, "Painter.newWithPixmap")
	return newPainterWithPixmap(pm)
}

func registerPainterMembers(t *Type) {
	t.registerProperty("pixmap", "Pixmap",
		func(inst Any) Any { return inst.(*Painter).pixmap },
		func(inst Any, val Any) { setPainterPixmapFromAny(inst.(*Painter), val) },
	)
	t.registerProperty("stroke", "Stroke",
		func(inst Any) Any { return inst.(*Painter).ps.stroke },
		nil,
	)
	t.registerProperty("fill", "Fill",
		func(inst Any) Any { return inst.(*Painter).ps.fill },
		nil,
	)
	t.registerProperty("font", "Font",
		func(inst Any) Any { return inst.(*Painter).ps.font },
		nil,
	)
	t.registerProperty("pathBuilder", "PathBuilder",
		func(inst Any) Any { return inst.(*Painter).pathBuilder },
		nil,
	)

	// ( -- )
	t.registerNativeMethod("clear", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Painter).Clear()
	})
	// ( -- )
	t.registerNativeMethod("push", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Painter).Push()
	})
	// ( -- )
	t.registerNativeMethod("pop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Painter).Pop()
	})
	// ( at:Point -- )
	t.registerNativeMethod("pixel", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Painter).Pixel(popPoint(f.RP.DataStack))
	})
	// ( c:Color at:Point -- )
	t.registerNativeMethod("pixelC", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		c := f.RP.DataStack.Pop().(Color)
		ths.Obj.(*Painter).PixelC(c, at)
	})
	// ( from:Point to:Point -- )
	t.registerNativeMethod("line", func(f *RunFrame, ths *TypeObjPair) {
		to := popPoint(f.RP.DataStack)
		from := popPoint(f.RP.DataStack)
		ths.Obj.(*Painter).Line(from, to)
	})
	// ( ctr:Point radius:n -- )  radius Int/Uint/Float → float64 at boundary
	t.registerNativeMethod("circle", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		r := popDrawFloat(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).Circle(ctr, r)
	})
	// ( ctr:Point rx:n ry:n -- )  radii Int/Uint/Float → float64 at boundary
	t.registerNativeMethod("ellipse", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		ry := popDrawFloat(ds)
		rx := popDrawFloat(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).Ellipse(ctr, rx, ry)
	})
	// ( ctr:Point rx:n ry:n rotationDegrees:n -- )
	t.registerNativeMethod("ellipseRot", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		rot := popDrawFloat(ds)
		ry := popDrawFloat(ds)
		rx := popDrawFloat(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).EllipseRot(ctr, rx, ry, rot)
	})
	// ( ctr:Point size:Size -- )
	t.registerNativeMethod("rect", func(f *RunFrame, ths *TypeObjPair) {
		sz := popSize(f.RP.DataStack)
		ctr := popPoint(f.RP.DataStack)
		ths.Obj.(*Painter).Rect(ctr, sz)
	})
	// ( ctr:Point size:Size rotationDegrees:n -- )
	t.registerNativeMethod("rectRot", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		rot := popDrawFloat(ds)
		sz := popSize(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).RectRot(ctr, sz, rot)
	})
	// ( ctr:Point size:Size radius:n -- )  corner radius Int/Uint/Float → float64
	t.registerNativeMethod("roundedRect", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		r := popDrawFloat(ds)
		sz := popSize(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).RoundedRect(ctr, sz, r)
	})
	// ( ctr:Point size:Size radius:n rotationDegrees:n -- )
	t.registerNativeMethod("roundedRectRot", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		rot := popDrawFloat(ds)
		r := popDrawFloat(ds)
		sz := popSize(ds)
		ctr := popPoint(ds)
		ths.Obj.(*Painter).RoundedRectRot(ctr, sz, r, rot)
	})
	// ( sides:i center:Point radius:n -- )  sides Int-only; radius flex
	t.registerNativeMethod("regularPolygon", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		radius := popDrawFloat(ds)
		ctr := popPoint(ds)
		sides := int(ds.Pop().(Int))
		ths.Obj.(*Painter).RegularPolygon(sides, ctr, radius)
	})
	// ( sides:i center:Point radius:n rotationDegrees:n -- )
	t.registerNativeMethod("regularPolygonRot", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		rot := popDrawFloat(ds)
		radius := popDrawFloat(ds)
		ctr := popPoint(ds)
		sides := int(ds.Pop().(Int))
		ths.Obj.(*Painter).RegularPolygonRot(sides, ctr, radius, rot)
	})
	// ( s:s at:Point -- )
	t.registerNativeMethod("text", func(f *RunFrame, ths *TypeObjPair) {
		at := popPoint(f.RP.DataStack)
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*Painter).Text(s, at)
	})
	// ( s:s at:Point width:n -- )
	t.registerNativeMethod("textWrap", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		width := popDrawFloat(ds)
		at := popPoint(ds)
		s := string(ds.Pop().(String))
		ths.Obj.(*Painter).TextWrap(s, at, width)
	})

	// ( size:Size | w h -- p:Painter )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewPainter(popSize(f.RP.DataStack)))
	})
	// ( pm:Pixmap -- p:Painter )
	t.registerNativeClassMethod("newWithPixmap", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewPainterWithPixmap(popPixmap(f.RP.DataStack, "Painter.newWithPixmap")))
	})
}

func CreateRegisterPainterType() {
	t := NewRegisteringType("Painter",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Painter]("Painter"),
	)
	registerPainterMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	painterType = t
}

var (
	painterType *Type
	_           Typed = (*Painter)(nil)
)

func init() {
	CreateRegisterPainterType()
}
