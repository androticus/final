package fyneui

import (
	"fmt"

	"github.com/androticus/final/internal/core"
)

// Entity kinds: negative = system-reserved; non-negative = user.
// Author render/update loops typically skip kind < 0.
const (
	EntityKindBorder = -1 // playfield edge HitRect; Final: SimEntity.kindBorder
)

// simEntityLastID is the last debug stamp assigned by NewSimEntity (0 = none yet).
// Next mint uses lastID+1. EventLoop-only; no mutex.
var simEntityLastID int

// SimEntity is a visible or invisible game object.
type SimEntity struct {
	data      any           // user data; lazy DynamicObject on get if nil
	id        int           // debug stamp (monotonic); not a live-store key
	sg        *SimGame      // optional backref; authors usually hold the game separately
	kind      int32         // user+system defined kind
	subkind   int32         // user defined subkind
	state     int32         // user defined state
	_at       *core.Point   // natural anchor in world coordinates; managed as a Value
	hitArea   HitArea       // optional shape recipe (offset from entity); per entity — do not share across entities
	drawerSet *SimDrawerSet // optional look + cycler; nil = not in library paint walk
	vx, vy    float64       // velocity in world coordinates
	zorder    int16         // <0 background, 0 base, >0 foreground
	anchored  bool          // if true, anchored to the world at the natural anchor point
	visible   bool          // if false, not visible; default true
	enabled   bool          // soft switch; default true — ephemeral off without list remove
	active    bool          // true after add to SimGame.entities, false after remove
	stopped   bool          // no motion updates; set by `stop` processing, or user
}

// NewSimEntity creates a new SimEntity (not yet in the sim list; active false).
// Assigns the next debug id starting at 1.
func NewSimEntity() *SimEntity {
	simEntityLastID++
	return &SimEntity{
		id:      simEntityLastID,
		_at:     &core.Point{},
		visible: true,
		enabled: true,
	}
}

func (e *SimEntity) FinalType() *core.Type { return simEntityType }

func (e *SimEntity) ID() int { return e.id }

// Data returns user data, minting a DynamicObject if nil.
func (e *SimEntity) Data() any {
	if e.data == nil {
		e.data = core.NewDynamicObject()
	}
	return e.data
}

func (e *SimEntity) SetData(v any) { e.data = v }

func (e *SimEntity) HitArea() HitArea { return e.hitArea }

// SetHitArea sets the collider recipe. While active, only clearing (h == nil) is allowed —
// replace-while-live panics. Do not share one HitArea instance across entities.
func (e *SimEntity) SetHitArea(h HitArea) {
	if e.active && h != nil {
		panic(fmt.Errorf("SimEntity.setHitArea: entity is active"))
	}
	e.hitArea = h
}

// Point uses value semantics, but is a pointer for atomicity.
func (e *SimEntity) At() core.Point { return *e._at }

func (e *SimEntity) SetAt(p core.Point) {
	e._at = &p
}

// as close to atomic as possible, but not atomic
func (e *SimEntity) SetVelocity(vx, vy float64) {
	e.vx = vx
	e.vy = vy
}

// MoveInBounds moves the entity by (dx, dy), keeping it in the playfield.
// at is a hard stop: the anchor stays in (0,0)…(w-1,h-1).
// A HitArea further constrains the move so the volume stays on the pixmap
// when it fits; if it cannot fit, at still stays in bounds (best effort).
func (e *SimEntity) MoveInBounds(dx, dy float64) {
	if e.sg == nil {
		panic(fmt.Errorf("SimEntity.moveInBounds: entity has no SimGame"))
	}
	maxAtX := e.sg.pixSize.W - 1
	maxAtY := e.sg.pixSize.H - 1
	if maxAtX < 0 {
		maxAtX = 0
	}
	if maxAtY < 0 {
		maxAtY = 0
	}
	x := clampFloat(e._at.X+dx, 0, maxAtX)
	y := clampFloat(e._at.Y+dy, 0, maxAtY)
	if minX, maxX, minY, maxY, ok := e.hitAtLimits(e.sg.pixSize.W, e.sg.pixSize.H); ok {
		loX, hiX := max(0, minX), min(maxAtX, maxX)
		loY, hiY := max(0, minY), min(maxAtY, maxY)
		if loX <= hiX {
			x = clampFloat(x, loX, hiX)
		}
		if loY <= hiY {
			y = clampFloat(y, loY, hiY)
		}
	}
	e.SetAt(core.Point{X: x, Y: y})
}

// hitAtLimits is the at range that keeps the HitArea inside [0, fieldW]×[0, fieldH].
// ok is false when there is no HitArea (or an unknown shape). min>max means it cannot fit.
func (e *SimEntity) hitAtLimits(fieldW, fieldH float64) (minX, maxX, minY, maxY float64, ok bool) {
	if e.hitArea == nil {
		return 0, 0, 0, 0, false
	}
	switch e.hitArea.Shape() {
	case HitAreaShapeCircle:
		c := e.hitArea.(*CircleHitArea)
		return c.radius - c.center.X, fieldW - c.radius - c.center.X,
			c.radius - c.center.Y, fieldH - c.radius - c.center.Y, true
	case HitAreaShapeRect:
		r := e.hitArea.(*RectHitArea).rect
		rw, rh := r.Size.W, r.Size.H
		return -r.TopLeft.X, fieldW - r.TopLeft.X - rw,
			-r.TopLeft.Y, fieldH - r.TopLeft.Y - rh, true
	default:
		return 0, 0, 0, 0, false
	}
}

func registerSimEntityMembers(t *core.Type) {
	entClassConstInt := func(name string, v int) {
		t.RegisterClassProperty(name, "Int",
			func(_ any) any { return core.Int(v) },
			nil,
		)
	}
	entClassConstInt("kindBorder", EntityKindBorder)

	// ( -- data:Any )  lazy DynamicObject if nil
	t.RegisterProperty("data", "Any",
		func(inst any) any { return inst.(*SimEntity).Data() },
		func(inst any, value any) { inst.(*SimEntity).SetData(value) },
	)
	t.RegisterProperty("id", "Int",
		func(inst any) any { return core.Int(inst.(*SimEntity).id) },
		nil,
	)
	t.RegisterProperty("sg", "SimGame",
		func(inst any) any { return inst.(*SimEntity).sg },
		nil,
	)
	t.RegisterProperty("kind", "Int",
		func(inst any) any { return core.Int(inst.(*SimEntity).kind) },
		func(inst any, value any) { inst.(*SimEntity).kind = int32(value.(core.Int)) },
	)
	t.RegisterProperty("subkind", "Int",
		func(inst any) any { return core.Int(inst.(*SimEntity).subkind) },
		func(inst any, value any) { inst.(*SimEntity).subkind = int32(value.(core.Int)) },
	)
	t.RegisterProperty("state", "Int",
		func(inst any) any { return core.Int(inst.(*SimEntity).state) },
		func(inst any, value any) { inst.(*SimEntity).state = int32(value.(core.Int)) },
	)
	t.RegisterProperty("at", "Point",
		func(inst any) any { return inst.(*SimEntity).At() },
		func(inst any, value any) { inst.(*SimEntity).SetAt(value.(core.Point)) },
	)
	// HitArea concrete types: RectHitArea / CircleHitArea (immutable).
	t.RegisterProperty("hitArea", "Any",
		func(inst any) any { return inst.(*SimEntity).hitArea },
		func(inst any, value any) {
			if value == nil {
				inst.(*SimEntity).SetHitArea(nil)
				return
			}
			inst.(*SimEntity).SetHitArea(value.(HitArea))
		},
	)
	t.RegisterProperty("drawerSet", "SimDrawerSet",
		func(inst any) any { return inst.(*SimEntity).drawerSet },
		func(inst any, value any) {
			if value == nil {
				inst.(*SimEntity).drawerSet = nil
				return
			}
			inst.(*SimEntity).drawerSet = value.(*SimDrawerSet)
		},
	)
	t.RegisterProperty("vx", "Float",
		func(inst any) any { return core.Float(inst.(*SimEntity).vx) },
		func(inst any, value any) { inst.(*SimEntity).vx = float64(value.(core.Float)) },
	)
	t.RegisterProperty("vy", "Float",
		func(inst any) any { return core.Float(inst.(*SimEntity).vy) },
		func(inst any, value any) { inst.(*SimEntity).vy = float64(value.(core.Float)) },
	)
	t.RegisterProperty("zorder", "Int",
		func(inst any) any { return core.Int(inst.(*SimEntity).zorder) },
		func(inst any, value any) { inst.(*SimEntity).zorder = int16(value.(core.Int)) },
	)
	t.RegisterProperty("anchored", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimEntity).anchored) },
		func(inst any, value any) { inst.(*SimEntity).anchored = bool(value.(core.Bool)) },
	)
	t.RegisterProperty("visible", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimEntity).visible) },
		func(inst any, value any) { inst.(*SimEntity).visible = bool(value.(core.Bool)) },
	)
	t.RegisterProperty("enabled", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimEntity).enabled) },
		func(inst any, value any) { inst.(*SimEntity).enabled = bool(value.(core.Bool)) },
	)
	t.RegisterProperty("active", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimEntity).active) },
		nil,
	)
	t.RegisterProperty("stopped", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimEntity).stopped) },
		func(inst any, value any) { inst.(*SimEntity).stopped = bool(value.(core.Bool)) },
	)

	// ( dx:n dy:n -- )
	t.RegisterNativeMethod("moveInBounds", func(f *core.RunFrame, ths *core.TypeObjPair) {
		dy := float64(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		dx := float64(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		ths.Obj.(*SimEntity).MoveInBounds(dx, dy)
	})

	// ( dx:n dy:n -- )
	t.RegisterNativeMethod("setVelocity", func(f *core.RunFrame, ths *core.TypeObjPair) {
		vy := float64(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		vx := float64(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		ths.Obj.(*SimEntity).SetVelocity(vx, vy)
	})

	// ( -- entity:SimEntity )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewSimEntity())
	})
}

func CreateRegisterSimEntityType() {
	t := core.NewRegisteringType("SimEntity",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimEntity]("SimEntity"),
	)
	registerSimEntityMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simEntityType = t
}

var (
	simEntityType *core.Type
	_             core.Typed = (*SimEntity)(nil)
)

func init() {
	CreateRegisterSimEntityType()
}
