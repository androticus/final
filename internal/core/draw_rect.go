package core

import (
	"fmt"
	"image"
)

// Rect is top-left + Size in drawing space. Value type.
// ToImageRect maps to Go image.Rectangle (Min/Max) for blit — not for gg.
type Rect struct {
	TopLeft Point
	Size    Size
}

func (r Rect) FinalType() *Type { return rectType }

func NewRect(topLeft Point, size Size) Rect {
	return Rect{TopLeft: topLeft, Size: size}
}

func (r Rect) String() string {
	return fmt.Sprintf("[%s %s]", r.TopLeft, r.Size)
}

// ToImageRect maps to Go image.Rectangle for blit/clip.
// Origin uses Xi/Yi (round); size must be exactly integral (panic otherwise).
func (r Rect) ToImageRect() image.Rectangle {
	w := RequireIntegralDim("Rect.ToImageRect width", r.Size.W)
	h := RequireIntegralDim("Rect.ToImageRect height", r.Size.H)
	x := r.TopLeft.Xi()
	y := r.TopLeft.Yi()
	return image.Rect(x, y, x+w, y+h)
}

func registerRectMembers(t *Type) {
	t.registerProperty("topLeft", "Point",
		func(inst Any) Any { return inst.(Rect).TopLeft },
		nil,
	)
	t.registerProperty("size", "Size",
		func(inst Any) Any { return inst.(Rect).Size },
		nil,
	)
	// ( topLeft:Point size:Size -- r:Rect )
	// Each slot is polymorphic: Point|x y and Size|w h.
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		sz := popSize(f.RP.DataStack)
		tl := popPoint(f.RP.DataStack)
		f.RP.DataStack.Push(NewRect(tl, sz))
	})
}

func CreateRegisterRectType() {
	t := &Type{
		name:            "Rect",
		getZero:         MakeHostTypeZero[Rect](),
		checkCoerceType: MakeHostTypeCheckCoerce[Rect]("Rect"),
	}
	registerRectMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	rectType = t
}

var rectType *Type

func init() {
	CreateRegisterRectType()
}

var _ Typed = (*Rect)(nil)
