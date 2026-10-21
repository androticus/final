package core

import (
	"fmt"
	"image"
	"math"

	"github.com/fogleman/gg"
)

type pathBuilderMode int

const (
	pathBuilderReady pathBuilderMode = iota
	pathBuilderPath
	pathBuilderSolid
)

func (m pathBuilderMode) String() string {
	switch m {
	case pathBuilderReady:
		return "ready"
	case pathBuilderPath:
		return "path"
	case pathBuilderSolid:
		return "solid"
	default:
		return fmt.Sprintf("pathBuilderMode(%d)", int(m))
	}
}

// PathBuilder is Painter.pathBuilder: one session object, not PainterState.
// startPath / startSolid / end; vertex verbs only while a session is live.
type PathBuilder struct {
	p     *Painter
	mode  pathBuilderMode
	hasBB bool
	minX  float64
	minY  float64
	maxX  float64
	maxY  float64
}

func (b *PathBuilder) FinalType() *Type { return pathBuilderType }

func (b *PathBuilder) resetBounds() {
	b.hasBB = false
}

func (b *PathBuilder) include(x, y float64) {
	if !b.hasBB {
		b.minX, b.maxX = x, x
		b.minY, b.maxY = y, y
		b.hasBB = true
		return
	}
	if x < b.minX {
		b.minX = x
	}
	if x > b.maxX {
		b.maxX = x
	}
	if y < b.minY {
		b.minY = y
	}
	if y > b.maxY {
		b.maxY = y
	}
}

func (b *PathBuilder) aabb() image.Rectangle {
	if !b.hasBB {
		return image.Rectangle{}
	}
	return image.Rect(
		int(math.Floor(b.minX)),
		int(math.Floor(b.minY)),
		int(math.Ceil(b.maxX)),
		int(math.Ceil(b.maxY)),
	)
}

func newPathBuilder(p *Painter) *PathBuilder {
	return &PathBuilder{p: p, mode: pathBuilderReady}
}

func (p *Painter) requirePathReady(op string) {
	if p.pathBuilder.mode != pathBuilderReady {
		panic(fmt.Errorf("Painter.%s: pathBuilder is %s (call end first)", op, p.pathBuilder.mode))
	}
}

func (b *PathBuilder) requireReady(op string) {
	if b.mode != pathBuilderReady {
		panic(fmt.Errorf("PathBuilder.%s: already in %s (call end first)", op, b.mode))
	}
}

func (b *PathBuilder) requireSession(op string) {
	if b.mode == pathBuilderReady {
		panic(fmt.Errorf("PathBuilder.%s: not started (startPath or startSolid)", op))
	}
}

func (b *PathBuilder) StartPath() {
	b.requireReady("startPath")
	b.p.dc.ClearPath()
	b.resetBounds()
	b.mode = pathBuilderPath
}

func (b *PathBuilder) StartSolid() {
	b.requireReady("startSolid")
	b.p.dc.ClearPath()
	b.resetBounds()
	b.mode = pathBuilderSolid
}

func (b *PathBuilder) End() {
	switch b.mode {
	case pathBuilderReady:
		panic(fmt.Errorf("PathBuilder.end: not started (startPath or startSolid)"))
	case pathBuilderPath:
		stroke := b.p.ps.stroke
		if stroke.lineStyle != LineStyleNone {
			b.p.dc.SetStrokeStyle(gg.NewSolidPattern(stroke.color))
			b.p.dc.Stroke()
		} else {
			b.p.dc.ClearPath()
		}
	case pathBuilderSolid:
		b.p.dc.ClosePath()
		b.p.fillThenStroke(b.aabb())
	}
	b.p.dc.ClearPath()
	b.mode = pathBuilderReady
}

func (b *PathBuilder) MoveTo(x, y float64) {
	b.requireSession("moveTo")
	x, y = b.p.pixmap.MapWorldXYToView(x, y)
	b.include(x, y)
	b.p.dc.MoveTo(x, y)
}

func (b *PathBuilder) LineTo(x, y float64) {
	b.requireSession("lineTo")
	x, y = b.p.pixmap.MapWorldXYToView(x, y)
	b.include(x, y)
	b.p.dc.LineTo(x, y)
}

func (b *PathBuilder) BezierTo(cx, cy, x, y float64) {
	b.requireSession("bezierTo")
	cx, cy = b.p.pixmap.MapWorldXYToView(cx, cy)
	x, y = b.p.pixmap.MapWorldXYToView(x, y)
	b.include(cx, cy)
	b.include(x, y)
	b.p.dc.QuadraticTo(cx, cy, x, y)
}

func (b *PathBuilder) Arc(x, y, r, angle1Deg, angle2Deg float64) {
	b.requireSession("arc")
	if r < 0 {
		panic(fmt.Errorf("PathBuilder.arc: radius must be >= 0, got %v", r))
	}
	x, y = b.p.pixmap.MapWorldXYToView(x, y)
	b.include(x-r, y-r)
	b.include(x+r, y+r)
	b.p.dc.DrawArc(x, y, r, gg.Radians(angle1Deg), gg.Radians(angle2Deg))
}

func (b *PathBuilder) EllipticalArc(x, y, rx, ry, angle1Deg, angle2Deg float64) {
	b.requireSession("ellipticalArc")
	if rx < 0 || ry < 0 {
		panic(fmt.Errorf("PathBuilder.ellipticalArc: radii must be >= 0, got %v %v", rx, ry))
	}
	x, y = b.p.pixmap.MapWorldXYToView(x, y)
	b.include(x-rx, y-ry)
	b.include(x+rx, y+ry)
	b.p.dc.DrawEllipticalArc(x, y, rx, ry, gg.Radians(angle1Deg), gg.Radians(angle2Deg))
}

// popPathXY consumes cs( x y | Point ). Numbers may be Int, Uint, or Float.
func popPathXY(cs *Stack[any]) (x, y float64) {
	switch v := cs.Pop().(type) {
	case Point:
		return v.X, v.Y
	case Int, Uint, Float:
		y = float64(checkCoerceNumericToFloat(v))
		x = float64(checkCoerceNumericToFloat(cs.Pop()))
		return x, y
	default:
		panic(fmt.Sprintf("expected Point or numeric, got %T", v))
	}
}

func registerPathBuilderMembers(t *Type) {
	// ( -- )
	t.registerNativeMethod("startPath", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*PathBuilder).StartPath()
	})
	// ( -- )
	t.registerNativeMethod("startSolid", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*PathBuilder).StartSolid()
	})
	// ( -- )
	t.registerNativeMethod("end", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*PathBuilder).End()
	})
	// ( at:Point | x y -- )
	t.registerNativeMethod("moveTo", func(f *RunFrame, ths *TypeObjPair) {
		x, y := popPathXY(f.RP.DataStack)
		ths.Obj.(*PathBuilder).MoveTo(x, y)
	})
	// ( at:Point | x y -- )
	t.registerNativeMethod("lineTo", func(f *RunFrame, ths *TypeObjPair) {
		x, y := popPathXY(f.RP.DataStack)
		ths.Obj.(*PathBuilder).LineTo(x, y)
	})
	// ( control end -- ) each slot Point or x y (flex numeric)
	t.registerNativeMethod("bezierTo", func(f *RunFrame, ths *TypeObjPair) {
		x, y := popPathXY(f.RP.DataStack)
		cx, cy := popPathXY(f.RP.DataStack)
		ths.Obj.(*PathBuilder).BezierTo(cx, cy, x, y)
	})
	// ( center radius:n angle1:n angle2:n -- ) angles in degrees
	t.registerNativeMethod("arc", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		a2 := popDrawFloat(ds)
		a1 := popDrawFloat(ds)
		r := popDrawFloat(ds)
		x, y := popPathXY(ds)
		ths.Obj.(*PathBuilder).Arc(x, y, r, a1, a2)
	})
	// ( center rx:n ry:n angle1:n angle2:n -- ) angles in degrees
	t.registerNativeMethod("ellipticalArc", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		a2 := popDrawFloat(ds)
		a1 := popDrawFloat(ds)
		ry := popDrawFloat(ds)
		rx := popDrawFloat(ds)
		x, y := popPathXY(ds)
		ths.Obj.(*PathBuilder).EllipticalArc(x, y, rx, ry, a1, a2)
	})
}

func CreateRegisterPathBuilderType() {
	t := NewRegisteringType("PathBuilder",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[PathBuilder]("PathBuilder"),
	)
	registerPathBuilderMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	pathBuilderType = t
}

var (
	pathBuilderType *Type
	_               Typed = (*PathBuilder)(nil)
)

func init() {
	CreateRegisterPathBuilderType()
}
