package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestEnsureHitBoundaries_NoneWhenNoHitAction(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.EnsureHitBoundaries()
	if n := len(sg.mhe.BorderEntities()); n != 0 {
		t.Fatalf("borders=%d want 0", n)
	}
	if len(sg.entities.Snapshot()) != 0 {
		t.Fatal("gameplay entities must stay empty")
	}
}

func TestEnsureHitBoundaries_FourBounceOutsideOverlap(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setAllBoundaryModes(BoundaryHitActionBounce)
	sg.EnsureHitBoundaries()
	borders := sg.mhe.BorderEntities()
	if len(borders) != 4 {
		t.Fatalf("borders=%d want 4", len(borders))
	}
	if len(sg.entities.Snapshot()) != 0 {
		t.Fatal("borders must not be on SimGame.entities")
	}

	tx, ty := 50, 40 // half of 100×80
	var left, top, right, bottom *SimEntity
	for _, b := range borders {
		if b.kind != EntityKindBorder || !b.anchored || !b.active || b.visible {
			t.Fatalf("border meta: kind=%d anch=%v active=%v vis=%v", b.kind, b.anchored, b.active, b.visible)
		}
		ha, ok := b.HitArea().(*RectHitArea)
		if !ok || ha.Kind() != HitKindBorder {
			t.Fatalf("hitArea=%T kind=%v", b.HitArea(), b.HitArea().Kind())
		}
		switch BoundaryIndex(b.subkind) {
		case BoundaryIndexLeft:
			left = b
		case BoundaryIndexTop:
			top = b
		case BoundaryIndexRight:
			right = b
		case BoundaryIndexBottom:
			bottom = b
		}
	}
	if left == nil || top == nil || right == nil || bottom == nil {
		t.Fatal("missing side")
	}

	// Flush outside: left covers x∈[-tx,0); right x∈[w,w+tx)
	assertSlab(t, left, core.NewPoint(-tx, -ty), core.NewSize(tx, 80+2*ty))
	assertSlab(t, top, core.NewPoint(-tx, -ty), core.NewSize(100+2*tx, ty))
	assertSlab(t, right, core.NewPoint(100, -ty), core.NewSize(tx, 80+2*ty))
	assertSlab(t, bottom, core.NewPoint(-tx, 80), core.NewSize(100+2*tx, ty))

	// Corner overlap: NW point (-1,-1) in both left and top world AABBs
	if !worldAABBContains(left, -1, -1) || !worldAABBContains(top, -1, -1) {
		t.Fatal("NW corner should be in left and top slabs")
	}
	if !worldAABBContains(right, 100, -1) || !worldAABBContains(top, 100, -1) {
		t.Fatal("NE corner should be in right and top slabs")
	}
}

func TestEnsureHitBoundaries_IdempotentAndRebuild(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexBottom, BoundaryHitActionStop|BoundaryHitEvents)
	sg.EnsureHitBoundaries()
	first := sg.mhe.BorderEntities()
	if len(first) != 1 {
		t.Fatalf("borders=%d want 1", len(first))
	}
	ptr := first[0]
	sg.EnsureHitBoundaries()
	if sg.mhe.BorderEntities()[0] != ptr {
		t.Fatal("idempotent: expected same entity pointer")
	}

	sg.pixSize = core.NewSize(200, 80)
	sg.EnsureHitBoundaries()
	second := sg.mhe.BorderEntities()
	if len(second) != 1 {
		t.Fatalf("after resize borders=%d", len(second))
	}
	if second[0] == ptr {
		t.Fatal("resize should mint a new slab")
	}
	if ptr.active {
		t.Fatal("old slab should be inactive")
	}
	assertSlab(t, second[0], core.Point{X: -100, Y: 80}, core.NewSize(200+200, 40))

	sg.options.setAllBoundaryModes(0)
	sg.EnsureHitBoundaries()
	if len(sg.mhe.BorderEntities()) != 0 {
		t.Fatal("clear HitAction → no borders")
	}
	if second[0].active {
		t.Fatal("cleared slab should be inactive")
	}
}

func assertSlab(t *testing.T, e *SimEntity, at core.Point, size core.Size) {
	t.Helper()
	if e.At() != at {
		t.Fatalf("at=%v want %v", e.At(), at)
	}
	r := e.HitArea().(*RectHitArea).Rect()
	want := core.NewRect(core.Point{}, size)
	if r != want {
		t.Fatalf("local rect=%v want %v", r, want)
	}
}

func worldAABBContains(e *SimEntity, x, y int) bool {
	at := e.At()
	r := e.HitArea().(*RectHitArea).Rect()
	minX, minY := at.X+r.TopLeft.X, at.Y+r.TopLeft.Y
	maxX, maxY := minX+r.Size.W, minY+r.Size.H
	xf, yf := float64(x), float64(y)
	return xf >= minX && xf < maxX && yf >= minY && yf < maxY
}
