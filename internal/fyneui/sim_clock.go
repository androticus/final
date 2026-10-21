package fyneui

import (
	"fmt"
	"math"
	"time"

	"github.com/androticus/final/internal/core"
)

// SimClockState is the high-level sim lifecycle visible on SimClock.
type SimClockState = int

const (
	SimClockStateStopped SimClockState = iota
	SimClockStateRunning
	SimClockStatePaused
)

// SimSpeed clamps for SimClock.simSpeed (true rate factor; Pause is Suspend/state).
const (
	SimSpeedMin     = 0.1
	SimSpeedMax     = 10.0
	SimSpeedDefault = 1.0
)

// SimClockMaxCatchUp caps pending 1ms deltas so a stall cannot unbounded-grow debt.
const SimClockMaxCatchUp = 250

// simCostAvgSpan is the MovingAverage span for update/render (and MHE step) cost EMAs.
const simCostAvgSpan = 20

// SimClock is the user-facing sim clock (sg.clock) and owns the 1ms metronome
// that advances ticks. Motion/Hit is a separate engine (SimMotionHit); this type
// does not run physics. Final EventLoop serializes Suspend/Resume/Start/Stop.
// The metronome goroutine owns timer + drain; cross-thread fields (suspended,
// deltasToDo, engineSpeed, ticks) rely on aligned word-sized loads/stores being
// atomic on our 64-bit platforms.
type SimClock struct {
	state    SimClockState
	ticks    int     // accumulated sim ms; metronome may write, Final may read
	simSpeed float64 // user feel / ceiling; 1=realtime; NOT read by metronome

	updateRate     int // 1ms-tick counts between updates; 0=off, 1=asap; try 200 (~5/s)
	mheRate        int // author target ms between SimMHEngine fused steps; 0=off, 1=asap
	lastUpdate     int
	lastUpdateCost int
	avgUpdateCost  *core.MovingAverage // samples are nanoseconds
	updateStart    time.Time

	renderRate     int  // 1ms-tick counts between renders; 0=off, 1=asap; try 40 (~25fps)
	autoRender     bool // paint after update / #onMotion; not mouse-move; independent of renderRate
	lastRender     int
	lastRenderCost int                 // rounded ms (from MovingAverage mean ns, or last sample — see RenderEnded)
	avgRenderCost  *core.MovingAverage // samples are nanoseconds
	renderStart    time.Time

	// Metronome (always on while Start..Stop; SimGame always wants a clock).
	deltaRate   int     // metronome period ms; default 1
	engineSpeed float64 // tick advance rate; starts 1.0; published from SimGame when warm
	suspended   int     // depth; Final-only writers
	deltasToDo  int     // metronome increments; drain decrements
	realTime    *time.Timer
	lastTime    time.Time
	accumNS     int64 // leftover ns toward next whole ms at engineSpeed
	nextTrigger chan struct{}
	done        chan struct{}
	running     bool

	// User timer events (sim-ms After/At/Every). AtomicList of *timerReg.
	timers *core.AtomicAnyList
}

func NewSimClock() *SimClock {
	return &SimClock{
		state:         SimClockStateStopped,
		simSpeed:      SimSpeedDefault,
		renderRate:    0, // try: 40,
		updateRate:    0, // try: 200,
		mheRate:       0,
		avgRenderCost: core.NewMovingAverage(simCostAvgSpan),
		avgUpdateCost: core.NewMovingAverage(simCostAvgSpan),
		deltaRate:     1,
		engineSpeed:   SimSpeedDefault,
		timers:        core.NewAtomicAnyList(),
	}
}

func (c *SimClock) FinalType() *core.Type { return simClockType }

func (c *SimClock) State() SimClockState { return c.state }

func (c *SimClock) SetState(s SimClockState) { c.state = s }

func (c *SimClock) Ticks() int { return c.ticks }

// AddTicks advances sim clock by n ms (metronome drain / tests).
func (c *SimClock) AddTicks(n int) {
	if n <= 0 {
		return
	}
	c.ticks += n
}

func (c *SimClock) SimSpeed() float64 { return c.simSpeed }

// SetSimSpeed sets the user feel / ceiling rate (1=realtime). Values outside
// [SimSpeedMin, SimSpeedMax] are clamped silently; 0 is not pause.
// Does not write engineSpeed — that is published when cost averages are warm.
func (c *SimClock) SetSimSpeed(s float64) {
	if s < SimSpeedMin {
		s = SimSpeedMin
	}
	if s > SimSpeedMax {
		s = SimSpeedMax
	}
	c.simSpeed = s
}

// IsWarm is true once at least one render cost sample exists.
func (c *SimClock) IsWarm() bool {
	return c.avgRenderCost != nil && c.avgRenderCost.N() > 0
}

// CostBasisMs is updateTerm+renderTerm where each term is 0 unless that MovingAverage has N>0.
// Cold (!IsWarm) callers should treat basis as unused for engine publish.
func (c *SimClock) CostBasisMs() int {
	sum := 0
	if c.avgUpdateCost != nil && c.avgUpdateCost.N() > 0 {
		sum += c.AvgUpdateCostMs()
	}
	if c.avgRenderCost != nil && c.avgRenderCost.N() > 0 {
		sum += c.AvgRenderCostMs()
	}
	return sum
}

func (c *SimClock) RenderRate() int { return c.renderRate }

func (c *SimClock) SetRenderRate(r int) {
	// 0 = off; 1 = asap; >1 = period in sim ticks. Not coupled to autoRender.
	if r < 0 {
		r = 0
	}
	c.renderRate = r
}

func (c *SimClock) AutoRender() bool { return c.autoRender }

func (c *SimClock) SetAutoRender(on bool) { c.autoRender = on }

func (c *SimClock) SetAutoRenderOn() { c.autoRender = true }

func (c *SimClock) UpdateRate() int { return c.updateRate }

func (c *SimClock) SetUpdateRate(r int) {
	if r < 0 {
		r = 0
	}
	c.updateRate = r
}

func (c *SimClock) MheRate() int { return c.mheRate }

// SetMheRate sets the author target period for SimMHEngine (ms ticks).
// 0=off, 1=asap, >1 = period. MHE reads this as the ceiling; actualRate on
// the engine may be coarser under load.
func (c *SimClock) SetMheRate(r int) {
	if r < 0 {
		r = 0
	}
	c.mheRate = r
}

func (c *SimClock) LastRender() int     { return c.lastRender }
func (c *SimClock) LastRenderCost() int { return c.lastRenderCost }
func (c *SimClock) LastUpdate() int     { return c.lastUpdate }
func (c *SimClock) LastUpdateCost() int { return c.lastUpdateCost }

// AvgRenderCostMs is mean render duration as rounded ms (MovingAverage holds ns).
func (c *SimClock) AvgRenderCostMs() int { return avgNsToRoundedMs(c.avgRenderCost) }

// AvgUpdateCostMs is mean update duration as rounded ms (MovingAverage holds ns).
func (c *SimClock) AvgUpdateCostMs() int { return avgNsToRoundedMs(c.avgUpdateCost) }

// RenderStarted begins the render timing sandwich.
func (c *SimClock) RenderStarted() {
	c.renderStart = time.Now()
}

// RenderEnded stops the render timer, adds exact ns to the MovingAverage, sets
// lastRenderCost to this sample rounded to ms, and stamps lastRender = ticks.
func (c *SimClock) RenderEnded() {
	ns := time.Since(c.renderStart).Nanoseconds()
	if ns < 0 {
		ns = 0
	}
	c.avgRenderCost.Add(float64(ns))
	c.lastRenderCost = nsToRoundedMs(ns)
	c.lastRender = c.ticks
}

// UpdateStarted begins the update timing sandwich.
func (c *SimClock) UpdateStarted() {
	c.updateStart = time.Now()
}

// UpdateEnded stops the update timer, adds exact ns to the MovingAverage, sets
// lastUpdateCost to this sample rounded to ms, and stamps lastUpdate = ticks.
func (c *SimClock) UpdateEnded() {
	ns := time.Since(c.updateStart).Nanoseconds()
	if ns < 0 {
		ns = 0
	}
	c.avgUpdateCost.Add(float64(ns))
	c.lastUpdateCost = nsToRoundedMs(ns)
	c.lastUpdate = c.ticks
}

// ShouldUpdate reports whether an update is due at nowTicks (usually clock.Ticks()).
// updateRate 0 = never; 1 = every call (asap).
func (c *SimClock) ShouldUpdate(nowTicks int) bool {
	if c.updateRate == 0 {
		return false
	}
	if c.updateRate == 1 {
		return true
	}
	return nowTicks-c.lastUpdate >= c.updateRate
}

// ShouldRender reports whether a periodic render is due. renderRate 0 = never; 1 = asap.
func (c *SimClock) ShouldRender(nowTicks int) bool {
	if c.renderRate <= 0 {
		return false
	}
	if c.renderRate == 1 {
		return true
	}
	return nowTicks-c.lastRender >= c.renderRate
}

// simTicksUntilNext is sim-ms until the next due update or render lane.
// ok is false when both scheduleable lanes are off. asap / overdue use 1 sim-ms.
// autoRender is not a timer lane — event/update paint; periodic uses renderRate.
func (c *SimClock) simTicksUntilNext(now int) (until int, ok bool) {
	best := -1
	consider := func(rate, last int) {
		if rate <= 0 {
			return
		}
		var u int
		if rate == 1 {
			u = 1
		} else {
			u = last + rate - now
			if u < 1 {
				u = 1
			}
		}
		if best < 0 || u < best {
			best = u
		}
	}
	consider(c.updateRate, c.lastUpdate)
	consider(c.renderRate, c.lastRender)
	if best < 0 {
		return 0, false
	}
	return best, true
}

// ReprimeInterval is the wall-clock Interval until the next update/render due.
// Both lanes off → Interval.zero (caller must not schedule).
// engineSpeed is the metronome rate factor (not user simSpeed):
//
//	≤1 → nominal simTicksUntil wall ms (extra no-op timer fires OK under slo-mo);
//	>1 → wallMs = ceil(simTicksUntil / engineSpeed), floored at 1ms.
func (c *SimClock) ReprimeInterval(engineSpeed float64) core.Interval {
	simUntil, ok := c.simTicksUntilNext(c.ticks)
	if !ok {
		return core.Interval{}
	}
	wallMs := simUntil
	if engineSpeed > 1 {
		wallMs = int(math.Ceil(float64(simUntil) / engineSpeed))
	}
	if wallMs < 1 {
		wallMs = 1
	}
	return core.Interval{Duration: time.Duration(wallMs) * time.Millisecond}
}

func (c *SimClock) Suspend() {
	c.suspended++
}

func (c *SimClock) Resume() {
	was := c.suspended
	if c.suspended > 0 {
		c.suspended--
	}
	if was > 0 && c.suspended == 0 && c.running {
		c.signal()
	}
}

func (c *SimClock) IsSuspended() bool {
	return c.suspended > 0
}

func (c *SimClock) SuspendedDepth() int {
	return c.suspended
}

func (c *SimClock) DeltasToDo() int {
	return c.deltasToDo
}

func (c *SimClock) DeltaRate() int {
	return c.deltaRate
}

func (c *SimClock) SetDeltaRate(ms int) {
	if ms < 1 {
		ms = 1
	}
	c.deltaRate = ms
}

func (c *SimClock) EngineSpeed() float64 { return c.engineSpeed }

// SetEngineSpeed sets the metronome rate factor. Values outside
// [SimSpeedMin, SimSpeedMax] are clamped silently. No-op if unchanged.
func (c *SimClock) SetEngineSpeed(s float64) {
	if s < SimSpeedMin {
		s = SimSpeedMin
	}
	if s > SimSpeedMax {
		s = SimSpeedMax
	}
	if c.engineSpeed == s {
		return
	}
	c.engineSpeed = s
}

func (c *SimClock) signal() {
	if c.nextTrigger == nil {
		return
	}
	select {
	case c.nextTrigger <- struct{}{}:
	default:
		// coalesce
	}
}

// Start begins the metronome+drain goroutine. Idempotent. Call from Final only.
func (c *SimClock) Start() {
	if c.running {
		return
	}
	period := time.Duration(c.deltaRate) * time.Millisecond
	if period < time.Millisecond {
		period = time.Millisecond
	}
	c.nextTrigger = make(chan struct{}, 1)
	c.done = make(chan struct{})
	c.lastTime = time.Now()
	c.accumNS = 0
	c.deltasToDo = 0
	c.realTime = time.NewTimer(period)
	c.running = true

	done := c.done
	trig := c.nextTrigger
	timer := c.realTime
	go c.loop(timer, period, done, trig)
}

// loop owns timer fires and drain (single goroutine — no mutex with itself).
func (c *SimClock) loop(timer *time.Timer, period time.Duration, done <-chan struct{}, trig <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case now := <-timer.C:
			c.onMetronomeFire(now)
			c.drain()
			c.fireDueTimers()
			timer.Reset(period)
		case <-trig:
			c.drain()
			c.fireDueTimers()
		}
	}
}

func (c *SimClock) onMetronomeFire(now time.Time) {
	if !c.running {
		return
	}
	elapsed := now.Sub(c.lastTime)
	c.lastTime = now
	if elapsed < 0 {
		elapsed = 0
	}
	// Paused: keep lastTime current so Resume/Start does not catch up a pause gap.
	// Suspended drain is separate; Pause is state, not suspend depth.
	if c.state == SimClockStatePaused {
		return
	}
	scale := c.engineSpeed
	if scale < SimSpeedMin {
		scale = SimSpeedMin
	}
	c.accumNS += int64(float64(elapsed.Nanoseconds()) * scale)
	const msNS = int64(time.Millisecond)
	for c.accumNS >= msNS {
		c.accumNS -= msNS
		if c.deltasToDo < SimClockMaxCatchUp {
			c.deltasToDo++
		}
	}
}

// drain applies pending 1ms deltas when not suspended / not Paused.
func (c *SimClock) drain() {
	for {
		if !c.running || c.suspended > 0 || c.deltasToDo <= 0 {
			return
		}
		if c.state == SimClockStatePaused {
			return
		}
		c.deltasToDo--
		c.ticks++ // word-sized write; safe vs Final reads of ticks
	}
}

// Stop stops the metronome goroutine; leaves state alone. Idempotent.
func (c *SimClock) Stop() {
	if !c.running {
		return
	}
	c.running = false
	done := c.done
	timer := c.realTime
	c.done = nil
	c.realTime = nil
	c.nextTrigger = nil

	if timer != nil {
		timer.Stop()
	}
	if done != nil {
		close(done)
	}
}

// ResetCostStats clears update/render cost averages and last-cost snapshots
// so a later Start warms engine speed from a fresh run.
func (c *SimClock) ResetCostStats() {
	if c.avgUpdateCost != nil {
		c.avgUpdateCost.Reset()
	}
	if c.avgRenderCost != nil {
		c.avgRenderCost.Reset()
	}
	c.lastUpdateCost = 0
	c.lastRenderCost = 0
}

// Close stops and clears pending deltas. Call from Final only.
func (c *SimClock) Close() {
	c.Stop()
	c.deltasToDo = 0
	c.accumNS = 0
	c.suspended = 0
	c.engineSpeed = SimSpeedDefault
	if c.timers != nil {
		c.timers.Clear()
	}
}

// timerReg is one After/At/Every registration. Stored as a pointer in timers
// so deleteItem can use identity.
type timerReg struct {
	emit    core.HandlerMintEventFn // CreateEventFn(sg.onTimerEvent, index)
	source  any                     // Event.Source (SimGame); tests may pass nil
	index   int
	nextDue int
	period  int // 0 = one-shot After/At; >0 = Every
}

func (c *SimClock) ensureTimers() {
	if c.timers == nil {
		c.timers = core.NewAtomicAnyList()
	}
}

func (c *SimClock) addTimer(emit core.HandlerMintEventFn, source any, index, nextDue, period int) {
	c.ensureTimers()
	c.timers.Append(&timerReg{
		emit:    emit,
		source:  source,
		index:   index,
		nextDue: nextDue,
		period:  period,
	})
}

// TimerEventAfter fires once after `after` sim-ms from now. after must be > 0.
func (c *SimClock) TimerEventAfter(emit core.HandlerMintEventFn, source any, index, after int) {
	if after <= 0 {
		panic(fmt.Errorf("SimClock.timerEventAfter: after must be > 0, got %d", after))
	}
	c.addTimer(emit, source, index, c.ticks+after, 0)
}

// TimerEventAt fires once as soon as ticks >= tick (past tick → asap).
func (c *SimClock) TimerEventAt(emit core.HandlerMintEventFn, source any, index, tick int) {
	c.addTimer(emit, source, index, tick, 0)
}

// TimerEventEvery first fires at now+period, then every period sim-ms.
// period must be > 0. Catch-up coalesces to one mint per metronome wake.
func (c *SimClock) TimerEventEvery(emit core.HandlerMintEventFn, source any, index, period int) {
	if period <= 0 {
		panic(fmt.Errorf("SimClock.timerEventEvery: period must be > 0, got %d", period))
	}
	c.addTimer(emit, source, index, c.ticks+period, period)
}

// CancelTimerEvent removes all registrations with index (snapshot + deleteItem).
func (c *SimClock) CancelTimerEvent(index int) {
	c.ensureTimers()
	it := c.timers.GetIter()
	for it.GetNext() {
		reg, ok := it.Current().(*timerReg)
		if !ok || reg == nil || reg.index != index {
			continue
		}
		c.timers.DeleteItemFromAny(reg)
	}
}

// fireDueTimers mints coalesced Events for registrations with nextDue <= ticks.
// Skips Paused/Stopped. Suspended still mints (coalesced) so catch-up cannot flood.
func (c *SimClock) fireDueTimers() {
	if c.timers == nil || c.timers.Len() == 0 {
		return
	}
	if c.state != SimClockStateRunning {
		return
	}
	now := c.ticks
	it := c.timers.GetIter()
	for it.GetNext() {
		reg, ok := it.Current().(*timerReg)
		if !ok || reg == nil || reg.nextDue > now {
			continue
		}
		if reg.period > 0 {
			for reg.nextDue <= now {
				reg.nextDue += reg.period
			}
		} else {
			c.timers.DeleteItemFromAny(reg)
		}
		if reg.emit != nil {
			reg.emit(reg.source, nil)
		}
	}
}

// nsToRoundedMs: ns + 500µs, then integer ms.
func nsToRoundedMs(ns int64) int {
	return int((ns + 500_000) / 1_000_000)
}

// avgNsToRoundedMs: MovingAverage mean is ns; add 500µs then convert to int ms.
func avgNsToRoundedMs(a *core.MovingAverage) int {
	if a == nil || a.N() == 0 {
		return 0
	}
	meanNs, _ := a.GetMeanAndN()
	return nsToRoundedMs(int64(meanNs))
}

func registerSimClockMembers(t *core.Type) {
	simClockClassConstInt := func(name string, v int) {
		t.RegisterClassProperty(name, "Int",
			func(_ any) any { return core.Int(v) },
			nil,
		)
	}
	simClockClassConstFloat := func(name string, v float64) {
		t.RegisterClassProperty(name, "Float",
			func(_ any) any { return core.Float(v) },
			nil,
		)
	}
	simClockClassConstInt("stateStopped", SimClockStateStopped)
	simClockClassConstInt("stateRunning", SimClockStateRunning)
	simClockClassConstInt("statePaused", SimClockStatePaused)
	simClockClassConstFloat("simSpeedMin", SimSpeedMin)
	simClockClassConstFloat("simSpeedMax", SimSpeedMax)
	simClockClassConstFloat("simSpeedDefault", SimSpeedDefault)

	t.RegisterProperty("state", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).State()) },
		nil,
	)
	t.RegisterProperty("ticks", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).Ticks()) },
		nil,
	)
	t.RegisterProperty("simSpeed", "Float",
		func(inst any) any { return core.Float(inst.(*SimClock).SimSpeed()) },
		func(inst any, value any) {
			inst.(*SimClock).SetSimSpeed(float64(value.(core.Float)))
		},
	)
	t.RegisterProperty("autoRender", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimClock).AutoRender()) },
		func(inst any, value any) {
			inst.(*SimClock).SetAutoRender(bool(value.(core.Bool)))
		},
	)
	t.RegisterProperty("renderRate", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).RenderRate()) },
		func(inst any, value any) {
			inst.(*SimClock).SetRenderRate(int(value.(core.Int)))
		},
	)
	t.RegisterProperty("lastRender", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).LastRender()) },
		nil,
	)
	t.RegisterProperty("lastRenderCost", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).LastRenderCost()) },
		nil,
	)
	t.RegisterProperty("avgRenderCost", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).AvgRenderCostMs()) },
		nil,
	)
	t.RegisterProperty("avgRenderCostNs", "MovingAverage",
		func(inst any) any { return inst.(*SimClock).avgRenderCost },
		nil,
	)
	t.RegisterProperty("updateRate", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).UpdateRate()) },
		func(inst any, value any) {
			inst.(*SimClock).SetUpdateRate(int(value.(core.Int)))
		},
	)
	t.RegisterProperty("mheRate", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).MheRate()) },
		func(inst any, value any) {
			inst.(*SimClock).SetMheRate(int(value.(core.Int)))
		},
	)
	t.RegisterProperty("lastUpdate", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).LastUpdate()) },
		nil,
	)
	t.RegisterProperty("lastUpdateCost", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).LastUpdateCost()) },
		nil,
	)
	t.RegisterProperty("avgUpdateCost", "Int",
		func(inst any) any { return core.Int(inst.(*SimClock).AvgUpdateCostMs()) },
		nil,
	)
	t.RegisterProperty("avgUpdateCostNs", "MovingAverage",
		func(inst any) any { return inst.(*SimClock).avgUpdateCost },
		nil,
	)

	// ( -- )
	t.RegisterNativeMethod("setAutoRenderOn", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimClock).SetAutoRenderOn()
	})
	// ( -- )
	t.RegisterNativeMethod("renderStarted", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimClock).RenderStarted()
	})
	// ( -- )
	t.RegisterNativeMethod("renderEnded", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimClock).RenderEnded()
	})
	// ( -- )
	t.RegisterNativeMethod("updateStarted", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimClock).UpdateStarted()
	})
	// ( -- )
	t.RegisterNativeMethod("updateEnded", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimClock).UpdateEnded()
	})
	// ( nowTicks:Int -- due:Bool )
	t.RegisterNativeMethod("shouldUpdate", func(f *core.RunFrame, ths *core.TypeObjPair) {
		now := int(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(core.Bool(ths.Obj.(*SimClock).ShouldUpdate(now)))
	})
	// ( nowTicks:Int -- due:Bool )
	t.RegisterNativeMethod("shouldRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		now := int(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(core.Bool(ths.Obj.(*SimClock).ShouldRender(now)))
	})
	// ( engineSpeed:Float -- iv:Interval )
	t.RegisterNativeMethod("reprimeInterval", func(f *core.RunFrame, ths *core.TypeObjPair) {
		speed := float64(f.RP.DataStack.Pop().(core.Float))
		f.RP.DataStack.Push(ths.Obj.(*SimClock).ReprimeInterval(speed))
	})
}

func CreateRegisterSimClockType() {
	t := core.NewRegisteringType("SimClock",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimClock]("SimClock"),
	)
	registerSimClockMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simClockType = t
}

var (
	simClockType *core.Type
	_            core.Typed = (*SimClock)(nil)
)

func init() {
	CreateRegisterSimClockType()
}
