package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func testMHGame(w, h int) *SimGame {
	sg := &SimGame{
		clock:    NewSimClock(),
		options:  NewSimGameOptions(),
		mhe:      NewSimMHEngine(),
		pixSize:  core.NewSize(w, h),
		entities: NewSimEntityList(),
	}
	sg.mhe.sg = sg
	sg.clock.SetMheRate(100) // 100ms → dt=0.1s
	sg.mhe.actualRate = 100
	return sg
}

func captureCrossingEmit(got **SimCrossingEvent) core.HandlerMintEventFn {
	return func(_ any, params ...any) {
		if len(params) > 0 {
			if ev, ok := params[0].(*SimCrossingEvent); ok {
				*got = ev
			}
		}
	}
}

func TestSimMHEngine_IntegratePixelsPerSecond(t *testing.T) {
	sg := testMHGame(100, 80)
	e := NewSimEntity()
	e.SetAt(core.Point{X: 10, Y: 20})
	e.vx = 50 // px/s
	e.vy = -20
	sg.AddEntity(e)

	sg.mhe.Step()

	// dt=0.1 → +5, -2
	if e.At() != (core.Point{X: 15, Y: 18}) {
		t.Fatalf("at=%v want (15,18)", e.At())
	}
}

func TestSimMHEngine_SkipAnchoredStoppedInactive(t *testing.T) {
	sg := testMHGame(100, 80)

	anch := NewSimEntity()
	anch.SetAt(core.Point{X: 1, Y: 1})
	anch.vx = 100
	anch.anchored = true
	sg.AddEntity(anch)

	stop := NewSimEntity()
	stop.SetAt(core.Point{X: 2, Y: 2})
	stop.vx = 100
	stop.stopped = true
	sg.AddEntity(stop)

	off := NewSimEntity()
	off.SetAt(core.Point{X: 3, Y: 3})
	off.vx = 100
	off.enabled = false
	sg.AddEntity(off)

	sg.mhe.Step()
	if anch.At().X != 1 || stop.At().X != 2 || off.At().X != 3 {
		t.Fatalf("skipped entities moved: %v %v %v", anch.At(), stop.At(), off.At())
	}
}

func TestSimMHEngine_StopClampsAndSetsStopped(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexLeft, BoundaryCrossActionStop)

	e := NewSimEntity()
	e.SetAt(core.Point{X: 5, Y: 40})
	e.vx = -100 // 10px left in dt=0.1 → x=-5
	e.vy = 0
	sg.AddEntity(e)

	sg.mhe.Step()

	if e.At() != (core.Point{X: 0, Y: 40}) {
		t.Fatalf("at=%v want (0,40)", e.At())
	}
	if !e.stopped {
		t.Fatal("expected stopped")
	}
	if e.vx != -100 {
		t.Fatalf("vx cleared: %v", e.vx)
	}
}

func TestSimMHEngine_CrossEmitsOnceOnLeave(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexRight, BoundaryCrossEvents)

	var got *SimCrossingEvent
	sg.mhe.emitMotion = captureCrossingEmit(&got)

	e := NewSimEntity()
	e.SetAt(core.Point{X: 95, Y: 10})
	e.vx = 100 // → x=105 past maxX=99
	sg.AddEntity(e)

	sg.mhe.Step()
	if got == nil || got.crossings.Len() != 1 {
		t.Fatalf("want 1 crossing, got %#v", got)
	}
	c := got.crossings.items[0]
	if c.boundary != BoundaryIndexRight || c.entity != e {
		t.Fatalf("crossing=%+v", c)
	}
	if c.at.X != 105 {
		t.Fatalf("snapshot at.X=%g want 105", c.at.X)
	}
	if e.stopped {
		t.Fatal("cross must not set stopped")
	}
	if e.At().X != 105 {
		t.Fatalf("cross must not clamp, at=%v", e.At())
	}

	// Already outside: no second fire
	got = nil
	sg.mhe.Step()
	if got != nil {
		t.Fatal("already-outside must not re-fire")
	}
}

func TestSimMHEngine_StopEmitsWhenEventsSet(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexLeft, BoundaryCrossActionStop|BoundaryCrossEvents)

	var got *SimCrossingEvent
	sg.mhe.emitMotion = captureCrossingEmit(&got)

	e := NewSimEntity()
	e.SetAt(core.Point{X: 5, Y: 40})
	e.vx = -100
	sg.AddEntity(e)

	sg.mhe.Step()
	if got == nil || got.crossings.Len() != 1 {
		t.Fatalf("want 1 crossing, got %#v", got)
	}
	if got.crossings.items[0].at.X != 0 {
		t.Fatalf("stop snapshot at.X=%g want 0", got.crossings.items[0].at.X)
	}
	if !e.stopped {
		t.Fatal("expected stopped")
	}
}

func TestSimMHEngine_CrossBounceReversesAndClamps(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexRight, BoundaryCrossActionBounce)

	e := NewSimEntity()
	e.SetAt(core.Point{X: 95, Y: 10})
	e.vx = 100
	sg.AddEntity(e)

	sg.mhe.Step()
	if e.At().X != 99 {
		t.Fatalf("at.X=%g want 99", e.At().X)
	}
	if e.vx != -100 {
		t.Fatalf("vx=%v want -100", e.vx)
	}
	if e.stopped {
		t.Fatal("bounce must not set stopped")
	}
}

func TestSimMHEngine_CrossCornerTwoReports(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexLeft, BoundaryCrossEvents)
	sg.options.setBoundaryMode(BoundaryIndexTop, BoundaryCrossEvents)

	var got *SimCrossingEvent
	sg.mhe.emitMotion = captureCrossingEmit(&got)

	e := NewSimEntity()
	e.SetAt(core.Point{X: 5, Y: 5})
	e.vx = -100
	e.vy = -100
	sg.AddEntity(e)

	sg.mhe.Step()
	if got == nil || got.crossings.Len() != 2 {
		t.Fatalf("want 2 crossings, got %#v len=%d", got, crossingLen(got))
	}
}

func crossingLen(ev *SimCrossingEvent) int {
	if ev == nil || ev.crossings == nil {
		return 0
	}
	return ev.crossings.Len()
}

func TestSimMHEngine_EmitIfAnyFusedSlots(t *testing.T) {
	sg := testMHGame(100, 80)
	var params []any
	sg.mhe.emitMotion = func(_ any, p ...any) {
		params = p
	}

	live := func() *SimEntity {
		e := NewSimEntity()
		e.active = true
		return e
	}

	cross := NewSimCrossings()
	cross.Add(NewSimCrossing(live(), BoundaryIndexLeft, core.Point{}))
	sg.mhe.emitIfAny(cross, nil)
	if len(params) != 2 {
		t.Fatalf("len=%d want 2", len(params))
	}
	if _, ok := params[0].(*SimCrossingEvent); !ok {
		t.Fatalf("params[0]=%T", params[0])
	}
	if params[1] != nil {
		t.Fatalf("params[1] must be untyped nil, got %#v", params[1])
	}
	if sg.mhe.pendingAckCount() != 1 {
		t.Fatalf("pending=%d want 1", sg.mhe.pendingAckCount())
	}

	hits := NewSimHits()
	hits.Add(NewSimHit(live(), live(), core.Point{}, core.Point{}))
	sg.mhe.emitIfAny(cross, hits)
	if len(params) != 2 {
		t.Fatalf("both len=%d want 2", len(params))
	}
	if _, ok := params[0].(*SimCrossingEvent); !ok {
		t.Fatalf("both params[0]=%T", params[0])
	}
	if _, ok := params[1].(*SimHitEvent); !ok {
		t.Fatalf("both params[1]=%T", params[1])
	}
	if sg.mhe.pendingAckCount() != 2 {
		t.Fatalf("pending after second emit=%d want 2", sg.mhe.pendingAckCount())
	}
}

func TestSimMHEngine_EmitIfAnyDropsInactive(t *testing.T) {
	sg := testMHGame(100, 80)
	emitted := 0
	sg.mhe.emitMotion = func(_ any, _ ...any) { emitted++ }

	dead := NewSimEntity() // active false
	live := NewSimEntity()
	live.active = true

	cross := NewSimCrossings()
	cross.Add(NewSimCrossing(dead, BoundaryIndexLeft, core.Point{}))
	cross.Add(NewSimCrossing(live, BoundaryIndexRight, core.Point{}))
	sg.mhe.emitIfAny(cross, nil)
	if emitted != 1 {
		t.Fatalf("emitted=%d want 1", emitted)
	}
	if cross.Len() != 1 || cross.items[0].entity != live {
		t.Fatalf("want only live crossing, got %#v", cross.items)
	}

	emitted = 0
	hits := NewSimHits()
	hits.Add(NewSimHit(dead, live, core.Point{}, core.Point{}))
	hits.Add(NewSimHit(live, live, core.Point{}, core.Point{}))
	sg.mhe.emitIfAny(nil, hits)
	if emitted != 1 {
		t.Fatalf("hits emitted=%d want 1", emitted)
	}
	if hits.Len() != 1 {
		t.Fatalf("want 1 live hit, got %d", hits.Len())
	}

	emitted = 0
	sg.mhe.emitIfAny(NewSimCrossings(), NewSimHits()) // empty after filter
	deadOnly := NewSimCrossings()
	deadOnly.Add(NewSimCrossing(dead, BoundaryIndexTop, core.Point{}))
	sg.mhe.emitIfAny(deadOnly, nil)
	if emitted != 0 {
		t.Fatalf("inactive-only must not emit, got %d", emitted)
	}
}

func TestSimMHEngine_AckClearsPending(t *testing.T) {
	sg := testMHGame(100, 80)
	sg.options.setBoundaryMode(BoundaryIndexBottom, BoundaryCrossEvents)
	sg.mhe.emitMotion = func(_ any, params ...any) {}

	e := NewSimEntity()
	e.SetAt(core.Point{X: 10, Y: 75})
	e.vy = 100
	sg.AddEntity(e)

	sg.mhe.Step()
	if sg.mhe.pendingAckCount() != 1 {
		t.Fatalf("pending=%d", sg.mhe.pendingAckCount())
	}
	sg.mhe.Ack(sg.clock.Ticks())
	if sg.mhe.pendingAckCount() != 0 {
		t.Fatalf("pending after ack=%d", sg.mhe.pendingAckCount())
	}
}

func TestBoundaryMode_PackedUint8(t *testing.T) {
	if BoundaryHitEvents != 0x80 {
		t.Fatalf("HitEvents=0x%02x want 0x80", BoundaryHitEvents)
	}
	m := BoundaryCrossActionStop | BoundaryCrossEvents | BoundaryHitActionBounce | BoundaryHitEvents
	if m.CrossAction() != BoundaryCrossActionStop || !m.CrossEvents() {
		t.Fatalf("cross unpack %02x", m)
	}
	if m.HitAction() != BoundaryHitActionBounce || !m.HitEvents() {
		t.Fatalf("hit unpack %02x", m)
	}
	o := NewSimGameOptions()
	o.setAllBoundaryModes(BoundaryCrossActionStop)
	if o.useCrossings {
		t.Fatal("silent stop must not set useCrossings")
	}
	o.setAllBoundaryModes(BoundaryCrossEvents)
	if !o.useCrossings {
		t.Fatal("CrossEvents must set useCrossings")
	}
	if o.useMotionEngine {
		t.Fatal("useMotionEngine is author-set, not derived")
	}
}
