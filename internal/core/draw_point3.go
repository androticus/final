package core

import (
	"fmt"
	"math"
)

// Point3 is a continuous (x, y) in drawing space (Y down). Value type, like Color.
// Discrete sinks (pixel, blit, image.Point3) use Xi/Yi (math.Round).
type Point3 struct {
	X float64
	Y float64
	Z float64
}

func (p Point3) FinalType() *Type { return point3Type }

func NewPoint3(x, y, z int) Point3 {
	return Point3{X: float64(x), Y: float64(y), Z: float64(z)}
}

func NewPoint3F(x, y, z float64) Point3 {
	return Point3{X: x, Y: y, Z: z}
}

// Xi returns the rounded integral X for pixel-grid / blit / image sinks.
func (p Point3) Xi() int { return int(math.Round(p.X)) }

// Yi returns the rounded integral Y for pixel-grid / blit / image sinks.
func (p Point3) Yi() int { return int(math.Round(p.Y)) }

func (p Point3) Zi() int { return int(math.Round(p.Z)) }

func (p Point3) String() string {
	return fmt.Sprintf("(%g, %g, %g)", p.X, p.Y, p.Z)
}

/*if needed // popPoint3 consumes cs( x y z | Point3 ) — TOS-backward triple is z then y then x.
// Numbers may be Int, Uint, or Float (same flex as popPathXY).
func popPoint3(cs *Stack[any]) Point3 {
	switch v := cs.Pop().(type) {
	case Point3:
		return v
	case Int, Uint, Float:
		z := float64(checkCoerceNumericToFloat(v))
		y := float64(checkCoerceNumericToFloat(cs.Pop()))
		x := float64(checkCoerceNumericToFloat(cs.Pop()))
		return NewPoint3F(x, y, z)
	default:
		panic(fmt.Sprintf("expected Point3 or numeric, got %T", v))
	}
}*/

func registerPoint3Members(t *Type) {
	t.registerProperty("x", "Float",
		func(inst Any) Any { return Float(inst.(Point3).X) },
		nil,
	)
	t.registerProperty("y", "Float",
		func(inst Any) Any { return Float(inst.(Point3).Y) },
		nil,
	)
	t.registerProperty("z", "Float",
		func(inst Any) Any { return Float(inst.(Point3).Z) },
		nil,
	)
	t.registerProperty("xi", "Int",
		func(inst Any) Any { return Int(inst.(Point3).Xi()) },
		nil,
	)
	t.registerProperty("yi", "Int",
		func(inst Any) Any { return Int(inst.(Point3).Yi()) },
		nil,
	)
	t.registerProperty("zi", "Int",
		func(inst Any) Any { return Int(inst.(Point3).Zi()) },
		nil,
	)

	t.registerClassProperty("zero", "Point3",
		func(inst Any) Any { return Point3{} },
		nil,
	)
	// ( x:n y:n z:n -- p:Point3 )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		z := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		y := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		x := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		f.RP.DataStack.Push(NewPoint3F(x, y, z))
	})
}

func CreateRegisterPoint3Type() {
	t := &Type{
		name:            "Point3",
		getZero:         MakeHostTypeZero[Point3](),
		checkCoerceType: MakeHostTypeCheckCoerce[Point3]("Point3"),
	}
	registerPoint3Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	point3Type = t
}

var point3Type *Type

func init() {
	CreateRegisterPoint3Type()
}

var _ Typed = (*Point3)(nil)
