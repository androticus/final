package fyneui

import (
	"github.com/androticus/final/internal/core"
)

// HitAreaShape is the geometric recipe (rect vs circle).
type HitAreaShape int32

const (
	HitAreaShapeRect HitAreaShape = iota
	HitAreaShapeCircle
)

// Hit kinds: negative = system-reserved; non-negative = user.
// Parallel to SimEntity kinds (borders stay −1). Matrix keys on HitArea.kind.
const (
	HitKindBorder int32 = -1 // playfield edge HitRect; Final: HitArea.kindBorder / SimEntity.kindBorder
	HitKindAll    int32 = -2 // Hit Matrix wildcard only; Final: HitArea.kindAll
)

type HitArea interface {
	Shape() HitAreaShape
	Kind() int32
}

type HitAreaBase struct {
	shape HitAreaShape
	kind  int32 // user/system hit kind (matrix); not geometry
}

func (h *HitAreaBase) Shape() HitAreaShape {
	return h.shape
}

func (h *HitAreaBase) Kind() int32 {
	return h.kind
}

// RectHitArea is an immutable axis-aligned hit recipe (local to entity `at`).
type RectHitArea struct {
	HitAreaBase
	rect core.Rect
}

// CircleHitArea is an immutable circle hit recipe (local to entity `at`).
type CircleHitArea struct {
	HitAreaBase
	center core.Point
	radius float64
}

func NewRectHitArea(kind int32, rect core.Rect) *RectHitArea {
	return &RectHitArea{
		HitAreaBase: HitAreaBase{shape: HitAreaShapeRect, kind: kind},
		rect:        rect,
	}
}

func NewCircleHitArea(kind int32, center core.Point, radius float64) *CircleHitArea {
	return &CircleHitArea{
		HitAreaBase: HitAreaBase{shape: HitAreaShapeCircle, kind: kind},
		center:      center,
		radius:      radius,
	}
}

func (h *RectHitArea) FinalType() *core.Type   { return rectHitAreaType }
func (h *CircleHitArea) FinalType() *core.Type { return circleHitAreaType }

func (h *RectHitArea) Rect() core.Rect { return h.rect }

func (h *CircleHitArea) Center() core.Point { return h.center }
func (h *CircleHitArea) Radius() float64    { return h.radius }

func registerHitAreaKindConsts(t *core.Type) {
	hitClassConstInt := func(name string, v int32) {
		t.RegisterClassProperty(name, "Int",
			func(_ any) any { return core.Int(v) },
			nil,
		)
	}
	hitClassConstInt("kindBorder", HitKindBorder)
	hitClassConstInt("kindAll", HitKindAll)
}

func registerRectHitAreaMembers(t *core.Type) {
	registerHitAreaKindConsts(t)
	t.RegisterProperty("shape", "Int",
		func(inst any) any { return core.Int(inst.(*RectHitArea).shape) },
		nil,
	)
	t.RegisterProperty("kind", "Int",
		func(inst any) any { return core.Int(inst.(*RectHitArea).kind) },
		nil,
	)
	t.RegisterProperty("rect", "Rect",
		func(inst any) any { return inst.(*RectHitArea).rect },
		nil,
	)
	// ( r:Rect kind:Int -- rha:RectHitArea )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		kind := int32(f.RP.DataStack.Pop().(core.Int))
		r := f.RP.DataStack.Pop().(core.Rect)
		f.RP.DataStack.Push(NewRectHitArea(kind, r))
	})
}

func registerCircleHitAreaMembers(t *core.Type) {
	registerHitAreaKindConsts(t)
	t.RegisterProperty("shape", "Int",
		func(inst any) any { return core.Int(inst.(*CircleHitArea).shape) },
		nil,
	)
	t.RegisterProperty("kind", "Int",
		func(inst any) any { return core.Int(inst.(*CircleHitArea).kind) },
		nil,
	)
	t.RegisterProperty("center", "Point",
		func(inst any) any { return inst.(*CircleHitArea).center },
		nil,
	)
	t.RegisterProperty("radius", "Float",
		func(inst any) any { return core.Float(inst.(*CircleHitArea).radius) },
		nil,
	)
	// ( center:Point radius:n kind:Int -- cha:CircleHitArea )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		kind := int32(f.RP.DataStack.Pop().(core.Int))
		radius := float64(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		center := f.RP.DataStack.Pop().(core.Point)
		f.RP.DataStack.Push(NewCircleHitArea(kind, center, radius))
	})
}

func CreateRegisterRectHitAreaType() {
	t := core.NewRegisteringType("RectHitArea",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[RectHitArea]("RectHitArea"),
	)
	registerRectHitAreaMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	rectHitAreaType = t
}

func CreateRegisterCircleHitAreaType() {
	t := core.NewRegisteringType("CircleHitArea",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[CircleHitArea]("CircleHitArea"),
	)
	registerCircleHitAreaMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	circleHitAreaType = t
}

var (
	rectHitAreaType   *core.Type
	circleHitAreaType *core.Type
	_                 core.Typed = (*RectHitArea)(nil)
	_                 core.Typed = (*CircleHitArea)(nil)
	_                 HitArea    = (*RectHitArea)(nil)
	_                 HitArea    = (*CircleHitArea)(nil)
)

func init() {
	CreateRegisterRectHitAreaType()
	CreateRegisterCircleHitAreaType()
}
