package fyneui

import "testing"

func TestSimDrawerCycler_AdvanceIdle(t *testing.T) {
	c := NewSimDrawerCycler()
	c.Advance(100)
	if c.start != 0 || c.currentPhase != 0 {
		t.Fatalf("idle should not latch: start=%d phase=%d", c.start, c.currentPhase)
	}
}

func TestSimDrawerCycler_AdvanceLoop(t *testing.T) {
	c := NewSimDrawerCycler()
	c.phaseCount = 4
	c.phaseRate = 10
	c.Advance(0)
	if c.start != 0 {
		t.Fatalf("start=%d", c.start)
	}
	c.Advance(0) // latch
	if c.start != 0 || c.currentPhase != 0 {
		t.Fatalf("at start: start=%d phase=%d", c.start, c.currentPhase)
	}
	c.Advance(25) // phase 2
	if c.currentPhase != 2 || c.cyclesDone {
		t.Fatalf("phase=%d done=%v", c.currentPhase, c.cyclesDone)
	}
	c.Advance(40) // wrap to phase 0, continuous
	if c.currentPhase != 0 || c.cyclesDone {
		t.Fatalf("wrap phase=%d done=%v", c.currentPhase, c.cyclesDone)
	}
}

func TestSimDrawerCycler_AdvanceFinite(t *testing.T) {
	c := NewSimDrawerCycler()
	c.phaseCount = 2
	c.phaseRate = 10
	c.cyclesToRun = 2
	c.Advance(0)
	c.Advance(15) // cycle 0 phase 1
	if c.currentCycle != 0 || c.currentPhase != 1 || c.cyclesDone {
		t.Fatalf("mid: cycle=%d phase=%d done=%v", c.currentCycle, c.currentPhase, c.cyclesDone)
	}
	c.Advance(40) // past 2 cycles (period 20)
	if !c.cyclesDone || c.currentCycle != 1 || c.currentPhase != 1 {
		t.Fatalf("done: cycle=%d phase=%d done=%v", c.currentCycle, c.currentPhase, c.cyclesDone)
	}
}

func TestSimDrawerCycler_PhaseOffset(t *testing.T) {
	c := NewSimDrawerCycler()
	c.phaseCount = 2
	c.phaseRate = 10
	c.phaseOffset = 5
	c.Advance(0)
	c.Advance(3) // still in offset
	if c.currentPhase != 0 {
		t.Fatalf("during offset phase=%d", c.currentPhase)
	}
	c.Advance(15) // elapsed 15-5=10 → phase 1
	if c.currentPhase != 1 {
		t.Fatalf("after offset phase=%d", c.currentPhase)
	}
}

func TestSimDrawerCycler_Clone(t *testing.T) {
	c := NewSimDrawerCycler()
	c.phaseCount = 3
	c.phaseRate = 5
	c.cyclesToRun = 2
	c.start = 99
	c.currentPhase = 1
	c.cyclesDone = true
	n := c.Clone()
	if n.phaseCount != 3 || n.phaseRate != 5 || n.cyclesToRun != 2 {
		t.Fatal("static fields not copied")
	}
	if n.start != 0 || n.currentPhase != 0 || n.cyclesDone {
		t.Fatal("dynamic should reset on clone")
	}
}
