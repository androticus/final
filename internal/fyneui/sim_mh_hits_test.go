package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func addCircleEntity(sg *SimGame, kind int32, at core.Point, r float64, vx, vy float64) *SimEntity {
	e := NewSimEntity()
	e.kind = kind
	e.SetAt(at)
	e.vx, e.vy = vx, vy
	e.SetHitArea(NewCircleHitArea(kind, core.Point{}, r))
	sg.AddEntity(e)
	return e
}

// addSquareEntity: at is center; local RectHitArea (-half,-half)×(2half,2half).
func addSquareEntity(sg *SimGame, kind int32, at core.Point, half int, vx, vy float64) *SimEntity {
	e := NewSimEntity()
	e.kind = kind
	e.SetAt(at)
	e.vx, e.vy = vx, vy
	r := core.NewRect(
		core.NewPoint(-half, -half),
		core.NewSize(2*half, 2*half),
	)
	e.SetHitArea(NewRectHitArea(kind, r))
	sg.AddEntity(e)
	return e
}

func captureHitEmit(got **SimHitEvent) core.HandlerMintEventFn {
	return func(_ any, params ...any) {
		if len(params) > 1 {
			if ev, ok := params[1].(*SimHitEvent); ok {
				*got = ev
			}
		}
	}
}

func TestHits_CircleCircleBounceOnce(t *testing.T) {
	sg := testMHGame(200, 200)
	// Overlap: centers 10 apart, radii 8+8
	a := addCircleEntity(sg, 1, core.Point{X: 50, Y: 50}, 8, 10, 0)
	b := addCircleEntity(sg, 2, core.Point{X: 60, Y: 50}, 8, -10, 0)

	hits := sg.mhe.stepHits(sg.entities.Snapshot())
	if hits != nil && hits.Len() > 0 {
		t.Fatal("default bounce should not emit without HitEvents")
	}
	// Equal-mass head-on: velocities exchange along x
	if a.vx >= 0 || b.vx <= 0 {
		t.Fatalf("after bounce vx a=%v b=%v want a leftward / b rightward", a.vx, b.vx)
	}
	// Separated
	g := hitTestPair(a, b)
	if g.ok {
		t.Fatalf("still overlapping after separate pen=%v", g.pen)
	}

	// Stay: re-overlap; velocity into wall must be fixed (no Event).
	a.SetAt(core.Point{X: 50, Y: 50})
	b.SetAt(core.Point{X: 60, Y: 50})
	a.vx, b.vx = 10, -10
	sg.mhe.stepHits(sg.entities.Snapshot())
	if a.vx >= 0 || b.vx <= 0 {
		t.Fatalf("stay while closing should re-bounce: vx a=%v b=%v", a.vx, b.vx)
	}
	// Resting / separating stay: no Event spam path already covered; nudge apart
	a.SetAt(core.Point{X: 50, Y: 50})
	b.SetAt(core.Point{X: 60, Y: 50})
	a.vx, b.vx = -1, 1 // separating along x
	sg.mhe.stepHits(sg.entities.Snapshot())
	if a.vx != -1 || b.vx != 1 {
		t.Fatalf("stay while separating must not flip: vx a=%v b=%v", a.vx, b.vx)
	}
}

func TestHits_CircleWallBounceSeed(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setAllBoundaryModes(BoundaryHitActionBounce)
	sg.EnsureHitBoundaries()

	e := addCircleEntity(sg, 1, core.Point{X: 2, Y: 40}, 5, -40, 0)
	// Place so circle overlaps left slab (x∈[-50,0))
	e.SetAt(core.Point{X: 2, Y: 40})

	sg.mhe.stepHits(sg.entities.Snapshot())
	if e.vx <= 0 {
		t.Fatalf("wall bounce: vx=%v want >0", e.vx)
	}
}

func TestHits_RectWallBounceSeed(t *testing.T) {
	// Square movers vs border HitRects (was deferred → squares drifted OOB).
	sg := testMHGame(100, 80)
	sg.options.setAllBoundaryModes(BoundaryHitActionBounce)
	sg.EnsureHitBoundaries()

	e := addSquareEntity(sg, 1, core.Point{X: 4, Y: 40}, 8, -40, 0)
	sg.mhe.stepHits(sg.entities.Snapshot())
	if e.vx <= 0 {
		t.Fatalf("rect wall bounce: vx=%v want >0", e.vx)
	}
	if e.At().X < 8 {
		// half=8; after separate should clear left slab (x>=0 for left edge of square)
		t.Fatalf("should separate into field: at.X=%v", e.At().X)
	}
}

func TestHits_RectRectBounce(t *testing.T) {
	sg := testMHGame(200, 200)
	a := addSquareEntity(sg, 1, core.Point{X: 50, Y: 50}, 8, 10, 0)
	b := addSquareEntity(sg, 2, core.Point{X: 60, Y: 50}, 8, -10, 0)
	sg.mhe.stepHits(sg.entities.Snapshot())
	if a.vx >= 0 || b.vx <= 0 {
		t.Fatalf("rect↔rect bounce vx a=%v b=%v", a.vx, b.vx)
	}
	if hitTestPair(a, b).ok {
		t.Fatal("still overlapping after separate")
	}
}

func TestHits_StopAndHitEvents(t *testing.T) {
	sg := testMHGame(200, 200)
	sg.options.AddHitMatrixEntry(1, 2, BoundaryHitActionStop|BoundaryHitEvents)
	a := addCircleEntity(sg, 1, core.Point{X: 50, Y: 50}, 8, 0, 0)
	b := addCircleEntity(sg, 2, core.Point{X: 60, Y: 50}, 8, 0, 0)

	var got *SimHitEvent
	sg.mhe.emitMotion = captureHitEmit(&got)
	hits := sg.mhe.stepHits(sg.entities.Snapshot())
	if hits == nil || hits.Len() != 1 {
		t.Fatalf("hits=%v", hits)
	}
	if !a.stopped || !b.stopped {
		t.Fatal("stop should set stopped on both free bodies")
	}
	sg.mhe.emitIfAny(nil, hits)
	if got == nil || got.hits.Len() != 1 {
		t.Fatal("expected SimHitEvent")
	}
	h := got.hits.hits[0]
	if h.entity1.kind > h.entity2.kind {
		t.Fatalf("slot order kind %d > %d", h.entity1.kind, h.entity2.kind)
	}
}

func TestHits_BorderSeedHitEvents(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexBottom, BoundaryHitActionStop|BoundaryHitEvents)
	sg.EnsureHitBoundaries()

	e := addCircleEntity(sg, 3, core.Point{X: 50, Y: 78}, 5, 0, 20)
	var got *SimHitEvent
	sg.mhe.emitMotion = captureHitEmit(&got)
	hits := sg.mhe.stepHits(sg.entities.Snapshot())
	if hits == nil || hits.Len() != 1 {
		t.Fatalf("expected border HitEvent, hits=%v", hits)
	}
	if !e.stopped {
		t.Fatal("bottom stop seed should stop mover")
	}
	sg.mhe.emitIfAny(nil, hits)
	if got == nil {
		t.Fatal("expected emit")
	}
	h := got.hits.hits[0]
	if h.entity1.kind != EntityKindBorder && h.entity2.kind != EntityKindBorder {
		t.Fatal("SimHit should include kindBorder entity")
	}
}

func TestHits_MatrixOverridesBorderSeed(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setAllBoundaryModes(BoundaryHitActionBounce)
	sg.options.AddHitMatrixEntry(1, HitKindBorder, BoundaryHitActionStop|BoundaryHitEvents)
	sg.EnsureHitBoundaries()

	e := addCircleEntity(sg, 1, core.Point{X: 2, Y: 40}, 5, -10, 0)
	hits := sg.mhe.stepHits(sg.entities.Snapshot())
	if !e.stopped {
		t.Fatal("matrix override stop should win over bounce seed")
	}
	if hits == nil || hits.Len() != 1 {
		t.Fatal("HitEvents from matrix cell")
	}
	if e.vx != -10 {
		t.Fatalf("stop keeps velocity: vx=%v", e.vx)
	}
}

func TestHits_SilentDefaultNoEmit(t *testing.T) {
	sg := testMHGame(200, 200)
	a := addCircleEntity(sg, 1, core.Point{X: 50, Y: 50}, 8, 0, 0)
	_ = a
	b := addCircleEntity(sg, 2, core.Point{X: 60, Y: 50}, 8, 0, 0)
	_ = b
	var emitted int
	sg.mhe.emitMotion = func(_ any, _ ...any) { emitted++ }
	hits := sg.mhe.stepHits(sg.entities.Snapshot())
	sg.mhe.emitIfAny(nil, hits)
	if emitted != 0 {
		t.Fatal("default bounce must not emit")
	}
}

func TestHits_StayWallReBounceAfterNudge(t *testing.T) {
	// Wall contact stays active; a peer bounce can push vx back into the wall.
	// Stay must re-reflect so the circle cannot tunnel out of bounds.
	sg := testMHGame(100, 80)
	sg.options.setAllBoundaryModes(BoundaryHitActionBounce)
	sg.EnsureHitBoundaries()

	e := addCircleEntity(sg, 1, core.Point{X: 2, Y: 40}, 5, -40, 0)
	sg.mhe.stepHits(sg.entities.Snapshot()) // enter
	if e.vx <= 0 {
		t.Fatalf("enter bounce vx=%v", e.vx)
	}
	// Still overlapping left slab; shove back into the wall as a rock would.
	e.SetAt(core.Point{X: 2, Y: 40})
	e.vx = -50
	sg.mhe.stepHits(sg.entities.Snapshot()) // stay
	if e.vx <= 0 {
		t.Fatalf("stay closing into wall: vx=%v want >0", e.vx)
	}
	if e.At().X < 0 {
		t.Fatalf("should separate back into field: at.X=%v", e.At().X)
	}
}
