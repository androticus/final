package fyneui

import (
	"sync"
	"time"

	"github.com/androticus/final/internal/core"
)

// SimMHEngine is the v1 MotionHit engine on SimGame (sg.mhe).
//
// Cadence (locked): background/semi-autonomous. When options.useMotionEngine and
// clock is Running (not suspended), Step runs on a Go goroutine. Period ceiling
// is clock.mheRate (author target); actualRate is the working copy and may be
// coarsened so the engine does not fall behind realtime (coarsen deferred).
// updateRate / #handleUpdate are asynchronous and unordered vs MH Events.
//
// Integrate: _atf += (vx,vy)*dt with dt=actualRate/1000 (vx/vy = pixels/sec).
// Never call Final from this thread. Enqueue at most one Event per nonempty
// cycle via CreateEventFn (params: SimCrossingEvent|nil, SimHitEvent|nil).
// SimGame #onMotion demuxes to del#handle*, acks once, then optional autoRender.
//
// Hits: Circle↔Circle / Circle↔Rect; contacts; Hit Matrix (+ border seed); bounce/stop.
type SimMHEngine struct {
	sg         *SimGame
	emitMotion core.HandlerMintEventFn

	actualRate     int // working period ms; >= clock.mheRate when both >1; coarsened under load
	latestAsOfTick int // asOfTick of the last emitted cycle
	ackedAsOfTick  int // asOfTick of the last fully acked cycle
	latestCycle    int // fused-cycle seq of last emit with events
	ackedCycle     int // fused-cycle seq fully acked
	pendingAcks    int // 0 or 1: fused Event outstanding for latestCycle
	avgStepNs      *core.MovingAverage

	// Engine-owned playfield HitRect slabs (not on SimGame.entities).
	borders       []*SimEntity
	borderW       int
	borderH       int
	borderHitMask uint8 // bit i set ⇒ side i has HitAction ≠ none
	bordersBuilt  bool  // true after at least one EnsureHitBoundaries

	// Active HitArea-pair contacts (mark/sweep by generation).
	contacts   map[hitPairKey]uint64
	contactGen uint64

	mu      sync.Mutex
	running bool
	done    chan struct{}
}

func NewSimMHEngine() *SimMHEngine {
	return &SimMHEngine{
		avgStepNs: core.NewMovingAverage(simCostAvgSpan),
	}
}

func (e *SimMHEngine) FinalType() *core.Type { return simMHEngineType }

func (e *SimMHEngine) BindOnMotion(handler *core.MethodRef) {
	e.emitMotion = core.GetEventLoop().CreateEventFn(handler, 0)
}

func (e *SimMHEngine) ActualRate() int { return e.actualRate }

// TargetRate is the author ceiling from SimClock.mheRate (0=off).
func (e *SimMHEngine) TargetRate() int {
	if e.sg == nil || e.sg.clock == nil {
		return 0
	}
	return e.sg.clock.MheRate()
}

// Backlog is true when more than one fused cycle is unacked
// (latestCycle - ackedCycle > 1). Do not use raw tick subtraction for this:
// asOfTick is ms, not a cycle index.
func (e *SimMHEngine) Backlog() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.latestCycle-e.ackedCycle > 1
}

// Ack records that SimGame#onMotion finished the fused delivery.
// asOfTick is the token from that event. Cross-thread with the MH goroutine.
func (e *SimMHEngine) Ack(asOfTick int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.pendingAcks > 0 {
		e.pendingAcks--
	}
	if e.pendingAcks == 0 {
		e.ackedAsOfTick = asOfTick
		e.ackedCycle = e.latestCycle
	}
}

// Start begins the MH background loop. Idempotent. Call from SimGame.Start.
func (e *SimMHEngine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.done = make(chan struct{})
	e.running = true
	done := e.done
	go e.loop(done)
}

// Stop halts the MH background loop. Idempotent. Call from Pause/Stop/Close.
func (e *SimMHEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	done := e.done
	e.done = nil
	if done != nil {
		close(done)
	}
}

func (e *SimMHEngine) shouldStep() bool {
	sg := e.sg
	if sg == nil || sg.options == nil || sg.clock == nil {
		return false
	}
	if !sg.options.useMotionEngine {
		return false
	}
	if sg.clock.MheRate() == 0 {
		return false
	}
	if sg.clock.State() != SimClockStateRunning {
		return false
	}
	if sg.clock.IsSuspended() {
		return false
	}
	return true
}

func (e *SimMHEngine) syncActualRate() int {
	target := e.TargetRate()
	if target < 1 {
		e.actualRate = 0
		return 0
	}
	// Coarsen under load later; v1 copies the author ceiling.
	e.actualRate = target
	return e.actualRate
}

func (e *SimMHEngine) pendingAckCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.pendingAcks
}

func (e *SimMHEngine) loop(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}
		if !e.shouldStep() {
			select {
			case <-done:
				return
			case <-time.After(time.Millisecond):
			}
			continue
		}
		// Back-pressure: do not mint another Event until prior cycle acked.
		for e.pendingAckCount() > 0 {
			select {
			case <-done:
				return
			case <-time.After(time.Millisecond):
			}
			if !e.shouldStep() {
				break
			}
		}
		if !e.shouldStep() || e.pendingAckCount() > 0 {
			continue
		}

		periodMs := e.syncActualRate()
		if periodMs < 1 {
			continue
		}
		start := time.Now()
		e.Step()
		elapsed := time.Since(start)
		e.avgStepNs.Add(float64(elapsed.Nanoseconds()))
		wait := time.Duration(periodMs)*time.Millisecond - elapsed
		if wait < time.Millisecond {
			wait = time.Millisecond
		}
		select {
		case <-done:
			return
		case <-time.After(wait):
		}
	}
}

// Step is one fused engine cycle: integrate → World Bounds / Crossings → Hits.
func (e *SimMHEngine) Step() {
	sg := e.sg
	if sg == nil || sg.options == nil || sg.entities == nil {
		return
	}
	periodMs := e.actualRate
	if periodMs < 1 {
		periodMs = e.syncActualRate()
	}
	if periodMs < 1 {
		return
	}
	dt := float64(periodMs) / 1000.0
	modes := sg.options.boundaryModes
	maxX := sg.pixSize.W - 1
	maxY := sg.pixSize.H - 1
	if sg.pixSize.W < 1 {
		maxX = 0
	}
	if sg.pixSize.H < 1 {
		maxY = 0
	}

	var crossings *SimCrossings
	ents := sg.entities.Snapshot()
	for _, ent := range ents {
		if ent == nil || !ent.active || !ent.enabled || ent.anchored || ent.stopped {
			continue
		}
		preX, preY := ent._at.X, ent._at.Y
		ent.SetAt(core.Point{X: preX + ent.vx*dt, Y: preY + ent.vy*dt})
		postX, postY := ent._at.X, ent._at.Y
		x, y := postX, postY
		moved := false
		var newly [4]bool
		for i := BoundaryIndexLeft; i <= BoundaryIndexBottom; i++ {
			newly[i] = !boundaryPast(i, preX, preY, maxX, maxY) &&
				boundaryPast(i, postX, postY, maxX, maxY)
			m := modes[i]
			if !boundaryPast(i, x, y, maxX, maxY) {
				continue
			}
			switch m.CrossAction() {
			case BoundaryCrossActionStop:
				x, y = clampToBoundary(i, x, y, maxX, maxY)
				ent.stopped = true
				moved = true
			case BoundaryCrossActionBounce:
				bounceBoundary(i, ent)
				x, y = clampToBoundary(i, x, y, maxX, maxY)
				moved = true
			}
		}
		if moved {
			ent.SetAt(core.Point{X: x, Y: y})
		}
		atSnap := ent.At()
		for i := BoundaryIndexLeft; i <= BoundaryIndexBottom; i++ {
			if newly[i] && modes[i].CrossEvents() {
				crossings = ensureCrossings(crossings)
				crossings.Add(NewSimCrossing(ent, i, atSnap))
			}
		}
	}

	hits := e.stepHits(ents)
	e.emitIfAny(crossings, hits)
}

func boundaryPast(idx BoundaryIndex, x, y, maxX, maxY float64) bool {
	switch idx {
	case BoundaryIndexLeft:
		return x < 0
	case BoundaryIndexTop:
		return y < 0
	case BoundaryIndexRight:
		return x > maxX
	case BoundaryIndexBottom:
		return y > maxY
	default:
		return false
	}
}

func clampToBoundary(idx BoundaryIndex, x, y, maxX, maxY float64) (float64, float64) {
	switch idx {
	case BoundaryIndexLeft:
		if x < 0 {
			x = 0
		}
	case BoundaryIndexTop:
		if y < 0 {
			y = 0
		}
	case BoundaryIndexRight:
		if x > maxX {
			x = maxX
		}
	case BoundaryIndexBottom:
		if y > maxY {
			y = maxY
		}
	}
	return x, y
}

func bounceBoundary(idx BoundaryIndex, e *SimEntity) {
	switch idx {
	case BoundaryIndexLeft, BoundaryIndexRight:
		e.vx = -e.vx
	case BoundaryIndexTop, BoundaryIndexBottom:
		e.vy = -e.vy
	}
}

func ensureCrossings(c *SimCrossings) *SimCrossings {
	if c == nil {
		return NewSimCrossings()
	}
	return c
}

// emitIfAny enqueues one Event when either list is nonempty after dropping
// inactive entities. params[0] = SimCrossingEvent or nil; params[1] = SimHitEvent or nil.
func (e *SimMHEngine) emitIfAny(crossings *SimCrossings, hits *SimHits) {
	if e.sg == nil {
		return
	}
	if crossings != nil {
		crossings.retainLive()
	}
	if hits != nil {
		hits.retainLive()
	}
	tick := 0
	if e.sg.clock != nil {
		tick = e.sg.clock.Ticks()
	}
	var crossEv *SimCrossingEvent
	var hitEv *SimHitEvent
	if crossings != nil && crossings.Len() > 0 {
		crossEv = NewSimCrossingEvent(tick, crossings)
	}
	if hits != nil && hits.Len() > 0 {
		hitEv = NewSimHitEvent(tick, hits)
	}
	if crossEv == nil && hitEv == nil {
		return
	}
	if e.emitMotion == nil {
		panic("SimMHEngine: events to deliver but #onMotion not bound")
	}

	e.mu.Lock()
	e.latestAsOfTick = tick
	e.latestCycle++
	e.pendingAcks++
	e.mu.Unlock()

	// Untyped nil in any so Final Null? is true (typed nil is not Null).
	var p0, p1 any
	if crossEv != nil {
		p0 = crossEv
	}
	if hitEv != nil {
		p1 = hitEv
	}
	e.emitMotion(e.sg, p0, p1)
}

func registerSimMHEngineMembers(t *core.Type) {
	t.RegisterProperty("actualRate", "Int",
		func(inst any) any { return core.Int(inst.(*SimMHEngine).ActualRate()) },
		nil,
	)
	t.RegisterProperty("avgStepNs", "MovingAverage",
		func(inst any) any { return inst.(*SimMHEngine).avgStepNs },
		nil,
	)
	// ( handler:MethodRef -- )
	t.RegisterNativeMethod("bindOnMotion", func(f *core.RunFrame, ths *core.TypeObjPair) {
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		ths.Obj.(*SimMHEngine).BindOnMotion(handler)
	})
	// ( asOfTick:Int -- )
	t.RegisterNativeMethod("ack", func(f *core.RunFrame, ths *core.TypeObjPair) {
		asOfTick := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimMHEngine).Ack(asOfTick)
	})
	// ( -- )
	t.RegisterNativeMethod("step", func(_ *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimMHEngine).Step()
	})
}

func CreateRegisterSimMHEngineType() {
	t := core.NewRegisteringType("SimMHEngine",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimMHEngine]("SimMHEngine"),
	)
	registerSimMHEngineMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simMHEngineType = t
}

var (
	simMHEngineType *core.Type
	_               core.Typed = (*SimMHEngine)(nil)
)

func init() {
	CreateRegisterSimMHEngineType()
}
