package core

import (
	"fmt"
	"math"
)

// Size is continuous W×H in drawing space (columns × rows). Value type.
// Discrete sinks (pixmap / mat / Painter.new) require exact integral dims.
type Size struct {
	W float64
	H float64
}

func (s Size) FinalType() *Type { return sizeType }

func NewSize(w, h int) Size {
	return Size{W: float64(w), H: float64(h)}
}

func NewSizeF(w, h float64) Size {
	return Size{W: w, H: h}
}

// Wi returns the rounded integral W (pixel-span sample; not for buffer mint).
func (s Size) Wi() int { return int(math.Round(s.W)) }

// Hi returns the rounded integral H (pixel-span sample; not for buffer mint).
func (s Size) Hi() int { return int(math.Round(s.H)) }

func (s Size) String() string {
	return fmt.Sprintf("%gx%g", s.W, s.H)
}

// RequireIntegralDim panics unless v is finite and exactly integral.
// Used at discrete sinks (pixmap / mat / Painter.new / ToImageRect size).
func RequireIntegralDim(what string, v float64) int {
	if math.IsNaN(v) || math.IsInf(v, 0) || v != math.Trunc(v) {
		panic(fmt.Errorf("%s: must be integral, got %g", what, v))
	}
	return int(v)
}

// popSize consumes cs( w h | Size ) — TOS-backward pair is h then w.
// Numbers may be Int, Uint, or Float (same flex as popPoint).
func popSize(cs *Stack[any]) Size {
	switch v := cs.Pop().(type) {
	case Size:
		return v
	case Int, Uint, Float:
		h := float64(checkCoerceNumericToFloat(v))
		w := float64(checkCoerceNumericToFloat(cs.Pop()))
		return NewSizeF(w, h)
	default:
		panic(fmt.Sprintf("expected Size or numeric, got %T", v))
	}
}

func registerSizeMembers(t *Type) {
	t.registerProperty("w", "Float",
		func(inst Any) Any { return Float(inst.(Size).W) },
		nil,
	)
	t.registerProperty("h", "Float",
		func(inst Any) Any { return Float(inst.(Size).H) },
		nil,
	)
	t.registerProperty("wi", "Int",
		func(inst Any) Any { return Int(inst.(Size).Wi()) },
		nil,
	)
	t.registerProperty("hi", "Int",
		func(inst Any) Any { return Int(inst.(Size).Hi()) },
		nil,
	)
	// ( w:n h:n -- s:Size )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		h := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		w := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		f.RP.DataStack.Push(NewSizeF(w, h))
	})
}

func CreateRegisterSizeType() {
	t := &Type{
		name:            "Size",
		getZero:         MakeHostTypeZero[Size](),
		checkCoerceType: MakeHostTypeCheckCoerce[Size]("Size"),
	}
	registerSizeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	sizeType = t
}

var sizeType *Type

func init() {
	CreateRegisterSizeType()
}

var _ Typed = (*Size)(nil)
