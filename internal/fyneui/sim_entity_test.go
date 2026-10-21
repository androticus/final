package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestSimEntity_DataLazyMint(t *testing.T) {
	e := NewSimEntity()
	if e.data != nil {
		t.Fatalf("expected nil data initially")
	}
	d1 := e.Data()
	if d1 == nil {
		t.Fatalf("Data() returned nil")
	}
	if _, ok := d1.(*core.DynamicObject); !ok {
		t.Fatalf("want *DynamicObject, got %T", d1)
	}
	d2 := e.Data()
	if d1 != d2 {
		t.Fatalf("Data() should return same instance")
	}
}

func TestSimEntity_ActiveOnListAddRemove(t *testing.T) {
	e := NewSimEntity()
	h := NewCircleHitArea(0, core.Point{}, 4)
	e.SetHitArea(h)
	if e.HitArea() != h {
		t.Fatalf("hitArea not set")
	}
	if e.active {
		t.Fatalf("new entity should not be active")
	}

	sg := &SimGame{entities: NewSimEntityList()}
	sg.AddEntity(e)
	if !e.active {
		t.Fatalf("expected active after AddEntity")
	}
	if e.sg != sg {
		t.Fatalf("expected sg set on AddEntity")
	}

	sg.entities.Remove(e)
	if e.active {
		t.Fatalf("expected inactive after Remove")
	}
	if e.sg != nil {
		t.Fatalf("expected sg nil after Remove")
	}
}

func TestSimEntity_SetHitArea_ReplaceWhileActivePanics(t *testing.T) {
	e := NewSimEntity()
	e.SetHitArea(NewCircleHitArea(0, core.Point{}, 4))
	sg := &SimGame{entities: NewSimEntityList()}
	sg.AddEntity(e)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic replacing HitArea while active")
		}
	}()
	e.SetHitArea(NewCircleHitArea(1, core.Point{}, 2))
}

func TestSimEntity_SetHitArea_NilClearsWhileActive(t *testing.T) {
	e := NewSimEntity()
	e.SetHitArea(NewCircleHitArea(0, core.Point{}, 4))
	sg := &SimGame{entities: NewSimEntityList()}
	sg.AddEntity(e)
	e.SetHitArea(nil)
	if e.HitArea() != nil {
		t.Fatal("expected nil hitArea after clear")
	}
}

func TestSimEntity_NewAssignsMonotonicDebugID(t *testing.T) {
	before := simEntityLastID
	e1 := NewSimEntity()
	e2 := NewSimEntity()
	if e1.id != before+1 || e2.id != before+2 {
		t.Fatalf("ids: %d %d (last before %d)", e1.id, e2.id, before)
	}
	if e2.id-e1.id != 1 {
		t.Fatalf("id step: got %d want 1", e2.id-e1.id)
	}
	if e1.active {
		t.Fatalf("new entity should not be active until list add")
	}
}

func TestSimEntity_MoveInBounds_AtHardStop(t *testing.T) {
	sg := testMHGame(100, 80)
	e := NewSimEntity()
	e.SetAt(core.Point{X: 10, Y: 20})
	sg.AddEntity(e)

	e.MoveInBounds(5, -3)
	if e.At() != (core.Point{X: 15, Y: 17}) {
		t.Fatalf("interior: at=%v want (15,17)", e.At())
	}

	e.SetAt(core.Point{X: 95, Y: 70})
	e.MoveInBounds(20, 20)
	if e.At() != (core.Point{X: 99, Y: 79}) {
		t.Fatalf("clamp high: at=%v want (99,79)", e.At())
	}

	e.SetAt(core.Point{X: 2, Y: 3})
	e.MoveInBounds(-10, -10)
	if e.At() != (core.Point{X: 0, Y: 0}) {
		t.Fatalf("clamp low: at=%v want (0,0)", e.At())
	}
}

func TestSimEntity_MoveInBounds_HitAreaConstrains(t *testing.T) {
	sg := testMHGame(100, 80)
	e := NewSimEntity()
	e.SetAt(core.Point{X: 20, Y: 40})
	e.SetHitArea(NewCircleHitArea(0, core.Point{}, 8))
	sg.AddEntity(e)

	e.MoveInBounds(-20, 0)
	if e.At().X != 8 || e.At().Y != 40 {
		t.Fatalf("circle left: at=%v want (8,40)", e.At())
	}

	e.SetAt(core.Point{X: 80, Y: 40})
	e.MoveInBounds(50, 0)
	if e.At().X != 92 || e.At().Y != 40 {
		t.Fatalf("circle right: at=%v want (92,40)", e.At())
	}
}

func TestSimEntity_MoveInBounds_RectHitArea(t *testing.T) {
	sg := testMHGame(100, 80)
	e := NewSimEntity()
	e.SetAt(core.Point{X: 50, Y: 40})
	e.SetHitArea(NewRectHitArea(0, core.NewRect(core.NewPoint(-8, -8), core.Size{W: 16, H: 16})))
	sg.AddEntity(e)

	e.MoveInBounds(0, 50)
	if e.At().X != 50 || e.At().Y != 72 {
		t.Fatalf("rect bottom: at=%v want (50,72)", e.At())
	}
}

func TestSimEntity_MoveInBounds_HitAreaTooLargeKeepsAtInBounds(t *testing.T) {
	sg := testMHGame(100, 80)
	e := NewSimEntity()
	e.SetAt(core.Point{X: 50, Y: 40})
	e.SetHitArea(NewCircleHitArea(0, core.Point{}, 200))
	sg.AddEntity(e)

	e.MoveInBounds(-10, 5)
	if e.At() != (core.Point{X: 40, Y: 45}) {
		t.Fatalf("oversized hit: at=%v want (40,45)", e.At())
	}

	e.MoveInBounds(-100, -100)
	if e.At() != (core.Point{X: 0, Y: 0}) {
		t.Fatalf("oversized hit clamp: at=%v want (0,0)", e.At())
	}
}

func TestSimEntity_MoveInBounds_RequiresGame(t *testing.T) {
	e := NewSimEntity()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic without SimGame")
		}
	}()
	e.MoveInBounds(1, 0)
}
