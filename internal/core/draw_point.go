package core

import (
	"fmt"
	"image"
	"math"
)

// Point is a continuous (x, y) in drawing space (Y down). Value type, like Color.
// Discrete sinks (pixel, blit, image.Point) use Xi/Yi (math.Round).
type Point struct {
	X float64
	Y float64
}

func (p Point) FinalType() *Type { return pointType }

func NewPoint(x, y int) Point {
	return Point{X: float64(x), Y: float64(y)}
}

func NewPointF(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Xi returns the rounded integral X for pixel-grid / blit / image sinks.
func (p Point) Xi() int { return int(math.Round(p.X)) }

// Yi returns the rounded integral Y for pixel-grid / blit / image sinks.
func (p Point) Yi() int { return int(math.Round(p.Y)) }

func (p Point) String() string {
	return fmt.Sprintf("(%g, %g)", p.X, p.Y)
}

func (p Point) ToImagePoint() image.Point {
	return image.Point{X: p.Xi(), Y: p.Yi()}
}

// popPoint consumes cs( x y | Point ) — TOS-backward pair is y then x.
// Numbers may be Int, Uint, or Float (same flex as popPathXY).
func popPoint(cs *Stack[any]) Point {
	switch v := cs.Pop().(type) {
	case Point:
		return v
	case Int, Uint, Float:
		y := float64(checkCoerceNumericToFloat(v))
		x := float64(checkCoerceNumericToFloat(cs.Pop()))
		return NewPointF(x, y)
	default:
		panic(fmt.Sprintf("expected Point or numeric, got %T", v))
	}
}

func registerPointMembers(t *Type) {
	t.registerProperty("x", "Float",
		func(inst Any) Any { return Float(inst.(Point).X) },
		nil,
	)
	t.registerProperty("y", "Float",
		func(inst Any) Any { return Float(inst.(Point).Y) },
		nil,
	)
	t.registerProperty("xi", "Int",
		func(inst Any) Any { return Int(inst.(Point).Xi()) },
		nil,
	)
	t.registerProperty("yi", "Int",
		func(inst Any) Any { return Int(inst.(Point).Yi()) },
		nil,
	)

	t.registerClassProperty("zero", "Point",
		func(inst Any) Any { return Point{} },
		nil,
	)
	// ( x:n y:n -- p:Point )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		y := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		x := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		f.RP.DataStack.Push(NewPointF(x, y))
	})
}

func CreateRegisterPointType() {
	t := &Type{
		name:            "Point",
		getZero:         MakeHostTypeZero[Point](),
		checkCoerceType: MakeHostTypeCheckCoerce[Point]("Point"),
	}
	registerPointMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	pointType = t
}

var pointType *Type

func init() {
	CreateRegisterPointType()
}

var _ Typed = (*Point)(nil)
