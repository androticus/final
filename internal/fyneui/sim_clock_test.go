package fyneui

import (
	"testing"
	"time"

	"github.com/androticus/final/internal/core"
)

func TestSimClockClassConstants(t *testing.T) {
	typ := simClockType
	if typ == nil {
		t.Fatal("simClockType not registered")
	}
	get := func(name string) any {
		pm := typ.GetClassPropertyMetadata(name)
		if pm == nil {
			t.Fatalf("missing class const %q", name)
		}
		return pm.GetValue(nil)
	}
	if get("stateStopped") != core.Int(SimClockStateStopped) {
		t.Fatal("stateStopped")
	}
	if get("stateRunning") != core.Int(SimClockStateRunning) {
		t.Fatal("stateRunning")
	}
	if get("statePaused") != core.Int(SimClockStatePaused) {
		t.Fatal("statePaused")
	}
	if get("simSpeedMin") != core.Float(SimSpeedMin) {
		t.Fatal("simSpeedMin")
	}
	if get("simSpeedMax") != core.Float(SimSpeedMax) {
		t.Fatal("simSpeedMax")
	}
	if get("simSpeedDefault") != core.Float(SimSpeedDefault) {
		t.Fatal("simSpeedDefault")
	}
}

func TestSimClockSimSpeedClamp(t *testing.T) {
	c := NewSimClock()
	if c.SimSpeed() != SimSpeedDefault {
		t.Fatalf("default=%v", c.SimSpeed())
	}
	c.SetSimSpeed(0)
	if c.SimSpeed() != SimSpeedMin {
		t.Fatalf("floor want %v got %v", SimSpeedMin, c.SimSpeed())
	}
	c.SetSimSpeed(100)
	if c.SimSpeed() != SimSpeedMax {
		t.Fatalf("ceil want %v got %v", SimSpeedMax, c.SimSpeed())
	}
	c.SetSimSpeed(2.5)
	if c.SimSpeed() != 2.5 {
		t.Fatalf("passthrough got %v", c.SimSpeed())
	}
}

func TestSimClockShouldUpdateRender(t *testing.T) {
	c := NewSimClock()
	c.SetUpdateRate(10)
	c.SetRenderRate(5)
	c.AddTicks(5)
	c.UpdateStarted()
	c.UpdateEnded() // lastUpdate = 5
	if c.ShouldUpdate(12) {
		t.Fatal("not due yet (12-5 < 10)")
	}
	if !c.ShouldUpdate(15) {
		t.Fatal("expected update at 15")
	}
	c.SetUpdateRate(0)
	if c.ShouldUpdate(1000) {
		t.Fatal("rate 0 means never")
	}
	c.SetRenderRate(1)
	if !c.ShouldRender(0) {
		t.Fatal("asap render")
	}
}

func TestSimClockMheRate(t *testing.T) {
	c := NewSimClock()
	if c.MheRate() != 0 {
		t.Fatal("default mheRate 0")
	}
	c.SetMheRate(-3)
	if c.MheRate() != 0 {
		t.Fatal("negative clamps to 0")
	}
	c.SetMheRate(16)
	if c.MheRate() != 16 {
		t.Fatalf("got %d", c.MheRate())
	}
}

func TestSimClockRenderSandwichRoundsMs(t *testing.T) {
	c := NewSimClock()
	c.RenderStarted()
	time.Sleep(2 * time.Millisecond)
	c.RenderEnded()
	if c.LastRenderCost() < 1 {
		t.Fatalf("expected >=1ms cost, got %d", c.LastRenderCost())
	}
	if c.AvgRenderCostMs() < 1 {
		t.Fatalf("expected avg >=1ms, got %d", c.AvgRenderCostMs())
	}
}

func TestSimClockAutoRender(t *testing.T) {
	c := NewSimClock()
	if c.AutoRender() {
		t.Fatal("default off")
	}
	c.SetUpdateRate(100)
	c.SetAutoRenderOn()
	if !c.AutoRender() {
		t.Fatal("expected auto after setAutoRenderOn")
	}
	if c.RenderRate() != 0 {
		t.Fatalf("setAutoRenderOn must not change renderRate, got %d", c.RenderRate())
	}
	if c.ShouldRender(1000) {
		t.Fatal("renderRate 0: shouldRender stays false")
	}
	// timer driven only by update when renderRate is 0
	c.ticks = 0
	c.lastUpdate = 0
	until, ok := c.simTicksUntilNext(0)
	if !ok || until != 100 {
		t.Fatalf("until=%d ok=%v want 100", until, ok)
	}

	c.SetRenderRate(40)
	c.lastRender = 0
	if !c.AutoRender() {
		t.Fatal("autoRender must stay on when renderRate is set")
	}
	if !c.ShouldRender(40) {
		t.Fatal("periodic lane independent of autoRender")
	}
	c.SetAutoRender(false)
	if c.AutoRender() {
		t.Fatal("expected off")
	}
	if c.RenderRate() != 40 {
		t.Fatalf("clearing autoRender must not change renderRate, got %d", c.RenderRate())
	}
}

func TestSimClockReprimeInterval(t *testing.T) {
	c := NewSimClock()
	if c.ReprimeInterval(1).Duration != 0 {
		t.Fatal("both lanes off → zero")
	}
	c.SetUpdateRate(100)
	c.SetRenderRate(40)
	c.lastUpdate = 0
	c.lastRender = 0
	c.ticks = 0
	// next render at 40 → 40ms wall at engineSpeed 1
	iv := c.ReprimeInterval(1)
	if iv.Milliseconds() != 40 {
		t.Fatalf("want 40ms got %d", iv.Milliseconds())
	}
	// slo-mo engine: still nominal (no /scale)
	iv = c.ReprimeInterval(0.5)
	if iv.Milliseconds() != 40 {
		t.Fatalf("engine<=1 want nominal 40ms got %d", iv.Milliseconds())
	}
	// fast engine: shortened wait
	iv = c.ReprimeInterval(2)
	if iv.Milliseconds() != 20 {
		t.Fatalf("engine 2 want 20ms got %d", iv.Milliseconds())
	}
	c.SetRenderRate(1) // asap → 1ms floor
	iv = c.ReprimeInterval(1)
	if iv.Milliseconds() != 1 {
		t.Fatalf("asap want 1ms got %d", iv.Milliseconds())
	}
}

func TestSimClockWarmCostBasis(t *testing.T) {
	c := NewSimClock()
	if c.IsWarm() || c.CostBasisMs() != 0 {
		t.Fatal("cold")
	}
	c.RenderStarted()
	time.Sleep(2 * time.Millisecond)
	c.RenderEnded()
	if !c.IsWarm() {
		t.Fatal("warm after render sample")
	}
	if c.CostBasisMs() < 1 {
		t.Fatalf("basis=%d", c.CostBasisMs())
	}
}

func TestPublishEngineSpeedColdWarm(t *testing.T) {
	sg := newTestSimGame()
	sg.clock.SetSimSpeed(2)
	if sg.clock.EngineSpeed() != SimSpeedDefault {
		t.Fatal("engine starts at 1")
	}
	sg.PublishEngineSpeed() // cold
	if sg.clock.EngineSpeed() != SimSpeedDefault {
		t.Fatal("cold must not publish")
	}
	sg.clock.RenderStarted()
	time.Sleep(2 * time.Millisecond)
	sg.clock.RenderEnded()
	sg.PublishEngineSpeed()
	if sg.clock.EngineSpeed() != 2 {
		t.Fatalf("warm publish want 2 got %v", sg.clock.EngineSpeed())
	}
}

func TestSimClockMaxCatchUp(t *testing.T) {
	clock := NewSimClock()
	clock.SetEngineSpeed(10)
	clock.Suspend()
	clock.Start()
	defer clock.Close()
	time.Sleep(50 * time.Millisecond)
	if clock.DeltasToDo() > SimClockMaxCatchUp {
		t.Fatalf("deltasToDo=%d > max %d", clock.DeltasToDo(), SimClockMaxCatchUp)
	}
}

func TestSimClockSuspendAccumulatesThenCatchUp(t *testing.T) {
	clock := NewSimClock()
	clock.SetDeltaRate(1)

	clock.Suspend()
	clock.Start()
	defer clock.Close()

	time.Sleep(25 * time.Millisecond)
	if clock.Ticks() != 0 {
		t.Fatalf("ticks advanced while suspended: %d", clock.Ticks())
	}
	if clock.DeltasToDo() == 0 {
		t.Fatal("expected deltasToDo > 0 while suspended")
	}

	clock.Resume()
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) && clock.Ticks() == 0 {
		time.Sleep(2 * time.Millisecond)
	}
	if clock.Ticks() == 0 {
		t.Fatalf("expected ticks after resume; deltasToDo=%d", clock.DeltasToDo())
	}
}

func TestSimClockNestedSuspend(t *testing.T) {
	c := NewSimClock()
	c.Suspend()
	c.Suspend()
	if c.SuspendedDepth() != 2 {
		t.Fatalf("depth=%d", c.SuspendedDepth())
	}
	c.Resume()
	if !c.IsSuspended() {
		t.Fatal("still suspended")
	}
	c.Resume()
	if c.IsSuspended() {
		t.Fatal("should be clear")
	}
	c.Resume() // extra resume must not go negative
	if c.SuspendedDepth() != 0 {
		t.Fatalf("depth=%d", c.SuspendedDepth())
	}
}

func TestSimClockStartStopClose(t *testing.T) {
	clock := NewSimClock()
	clock.Start()
	clock.Start() // idempotent
	time.Sleep(15 * time.Millisecond)
	clock.Stop()
	clock.Stop()
	ticks := clock.Ticks()
	time.Sleep(20 * time.Millisecond)
	if clock.Ticks() != ticks {
		t.Fatalf("ticks moved after Stop: %d -> %d", ticks, clock.Ticks())
	}
	clock.Close()
}

func newTestSimGame() *SimGame {
	return &SimGame{
		clock: NewSimClock(),
	}
}

func waitTicksAdvance(t *testing.T, clock *SimClock, min int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if clock.Ticks() >= min {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("ticks did not reach %d (got %d)", min, clock.Ticks())
}

func TestSimGameStartPauseStop(t *testing.T) {
	sg := newTestSimGame()
	if sg.clock.State() != SimClockStateStopped {
		t.Fatal("initial state")
	}

	sg.Start()
	if sg.clock.State() != SimClockStateRunning {
		t.Fatal("after Start")
	}
	waitTicksAdvance(t, sg.clock, 5, 200*time.Millisecond)

	sg.Pause()
	if sg.clock.State() != SimClockStatePaused {
		t.Fatal("after Pause")
	}
	pausedAt := sg.clock.Ticks()
	time.Sleep(40 * time.Millisecond)
	if sg.clock.Ticks() != pausedAt {
		t.Fatalf("ticks advanced while paused: %d -> %d", pausedAt, sg.clock.Ticks())
	}

	sg.Start() // resume from pause
	if sg.clock.State() != SimClockStateRunning {
		t.Fatal("after Start from Pause")
	}
	waitTicksAdvance(t, sg.clock, pausedAt+5, 200*time.Millisecond)

	sg.Stop()
	if sg.clock.State() != SimClockStateStopped {
		t.Fatal("after Stop")
	}
	stoppedAt := sg.clock.Ticks()
	time.Sleep(40 * time.Millisecond)
	if sg.clock.Ticks() != stoppedAt {
		t.Fatalf("ticks advanced after Stop: %d -> %d", stoppedAt, sg.clock.Ticks())
	}

	sg.Start()
	if sg.clock.State() != SimClockStateRunning {
		t.Fatal("restart")
	}
	waitTicksAdvance(t, sg.clock, stoppedAt+3, 200*time.Millisecond)
	sg.Close()
	if sg.clock.State() != SimClockStateStopped {
		t.Fatal("after Close")
	}
	if sg.clock.DeltasToDo() != 0 || sg.clock.SuspendedDepth() != 0 {
		t.Fatal("clock metronome not cleared on Close")
	}
	sg.Close() // idempotent
}

func TestSimGameStopResetsCostStats(t *testing.T) {
	sg := newTestSimGame()
	sg.mhe = NewSimMHEngine()
	sg.clock.avgUpdateCost.Add(1_000_000)
	sg.clock.avgRenderCost.Add(2_000_000)
	sg.clock.lastUpdateCost = 1
	sg.clock.lastRenderCost = 2
	sg.mhe.avgStepNs.Add(3_000_000)
	sg.Start()
	sg.Stop()
	if sg.clock.avgUpdateCost.N() != 0 || sg.clock.avgRenderCost.N() != 0 {
		t.Fatal("cost averages should reset on Stop")
	}
	if sg.clock.lastUpdateCost != 0 || sg.clock.lastRenderCost != 0 {
		t.Fatal("last cost snapshots should reset on Stop")
	}
	if sg.mhe.avgStepNs.N() != 0 {
		t.Fatal("mhe avgStepNs should reset on Stop")
	}
	if sg.clock.IsWarm() {
		t.Fatal("clock should not be warm after Stop")
	}
	sg.Close()
}

func timerRecordEmit(got *[]int, index int) core.HandlerMintEventFn {
	return func(source any, params ...any) {
		*got = append(*got, index)
	}
}

func runningClock() *SimClock {
	c := NewSimClock()
	c.SetState(SimClockStateRunning)
	return c
}

func TestSimClockTimerEventAfter(t *testing.T) {
	c := runningClock()
	var got []int
	c.TimerEventAfter(timerRecordEmit(&got, 3), nil, 3, 50)
	c.fireDueTimers()
	if len(got) != 0 {
		t.Fatalf("fired early: %v", got)
	}
	c.AddTicks(50)
	c.fireDueTimers()
	if len(got) != 1 || got[0] != 3 {
		t.Fatalf("got %v want [3]", got)
	}
	if c.timers.Len() != 0 {
		t.Fatal("one-shot should be removed")
	}
	c.fireDueTimers()
	if len(got) != 1 {
		t.Fatalf("second fire: %v", got)
	}
}

func TestSimClockTimerEventAtPastIsAsap(t *testing.T) {
	c := runningClock()
	c.AddTicks(10)
	var got []int
	c.TimerEventAt(timerRecordEmit(&got, 1), nil, 1, 5)
	c.fireDueTimers()
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("got %v want [1]", got)
	}
}

func TestSimClockTimerEventEveryFirstAndCoalesce(t *testing.T) {
	c := runningClock()
	var got []int
	c.TimerEventEvery(timerRecordEmit(&got, 7), nil, 7, 100)
	c.AddTicks(50)
	c.fireDueTimers()
	if len(got) != 0 {
		t.Fatalf("first fire is now+period, got %v", got)
	}
	c.AddTicks(300) // ticks=350; periods 100,200,300 due
	c.fireDueTimers()
	if len(got) != 1 {
		t.Fatalf("coalesce want 1 mint, got %v", got)
	}
	c.fireDueTimers()
	if len(got) != 1 {
		t.Fatalf("nextDue should be in the future, got %v", got)
	}
}

func TestSimClockCancelTimerEventByIndex(t *testing.T) {
	c := runningClock()
	var got []int
	c.TimerEventEvery(timerRecordEmit(&got, 1), nil, 1, 10)
	c.TimerEventEvery(timerRecordEmit(&got, 1), nil, 1, 20)
	c.TimerEventEvery(timerRecordEmit(&got, 2), nil, 2, 10)
	c.CancelTimerEvent(1)
	if c.timers.Len() != 1 {
		t.Fatalf("len=%d want 1", c.timers.Len())
	}
	c.AddTicks(10)
	c.fireDueTimers()
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("got %v want [2]", got)
	}
}

func TestSimClockTimerEventPauseSkipsMint(t *testing.T) {
	c := runningClock()
	var got []int
	c.TimerEventAfter(timerRecordEmit(&got, 4), nil, 4, 10)
	c.AddTicks(10)
	c.SetState(SimClockStatePaused)
	c.fireDueTimers()
	if len(got) != 0 {
		t.Fatalf("paused minted %v", got)
	}
	c.SetState(SimClockStateRunning)
	c.fireDueTimers()
	if len(got) != 1 {
		t.Fatalf("resume mint got %v", got)
	}
}

func TestSimClockTimerEventSuspendedStillMints(t *testing.T) {
	c := runningClock()
	var got []int
	c.TimerEventAfter(timerRecordEmit(&got, 9), nil, 9, 5)
	c.Suspend()
	c.AddTicks(5)
	c.fireDueTimers()
	if len(got) != 1 {
		t.Fatalf("suspended should still mint, got %v", got)
	}
}

func TestSimClockTimerEventAfterPanicsNonPositive(t *testing.T) {
	c := runningClock()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	c.TimerEventAfter(nil, nil, 0, 0)
}

func TestSimClockTimerEventEveryPanicsNonPositive(t *testing.T) {
	c := runningClock()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	c.TimerEventEvery(nil, nil, 0, -1)
}

func TestSimClockCloseClearsTimers(t *testing.T) {
	c := runningClock()
	c.TimerEventEvery(timerRecordEmit(new([]int), 1), nil, 1, 10)
	c.Close()
	if c.timers.Len() != 0 {
		t.Fatalf("len=%d want 0", c.timers.Len())
	}
}
