package fyneui

import (
	"github.com/androticus/final/internal/core"
)

// SimDrawerCycler manages phase/cycle timing for a Drawer.
// Drawers can supply a copy with phaseCount etc filled in.
// All temporal values are in ticks unless otherwise noted.
// One-shot: phaseCount>=1, phaseRate>0, cyclesToRun=N>0.
type SimDrawerCycler struct {
	tag   any  // typically a number or string for later switch
	start int  // start time in ticks; Advance sets if still unset (begun false)
	begun bool // true after first Advance (allows start==0 as a real latch)
	// static recipe (ticks)
	phaseCount  int16 // phases per cycle; 0 = not phased
	phaseRate   int16 // ticks per phase; 0 = not phased
	phaseOffset int16 // ticks before phase 0 starts; usually 0
	cyclesToRun int16 // 0 = continuous; else N cycles then done
	// dynamic
	currentPhase int16 // 0..phaseCount-1
	currentCycle int16 // 0..cyclesToRun-1 when finite
	cyclesDone   bool  // true after finite run completes; poll in update
	// useCycleDoneEvent bool // deferred
}

func NewSimDrawerCycler() *SimDrawerCycler {
	return &SimDrawerCycler{}
}

// Clone returns a shallow copy (tag shared). Dynamic fields are reset.
func (c *SimDrawerCycler) Clone() *SimDrawerCycler {
	if c == nil {
		return NewSimDrawerCycler()
	}
	return &SimDrawerCycler{
		tag:         c.tag,
		phaseCount:  c.phaseCount,
		phaseRate:   c.phaseRate,
		phaseOffset: c.phaseOffset,
		cyclesToRun: c.cyclesToRun,
		// start / current* / cyclesDone zero — fresh run
	}
}

// Advance updates currentPhase / currentCycle / cyclesDone from nowTicks.
// On first call, if start is still 0, latches start = nowTicks (tick 0 is a valid latch).
// Idle when phaseCount or phaseRate is 0.
func (c *SimDrawerCycler) Advance(nowTicks int) {
	if c == nil {
		return
	}
	if c.phaseCount <= 0 || c.phaseRate <= 0 {
		return
	}
	if !c.begun {
		if c.start == 0 {
			c.start = nowTicks
		}
		c.begun = true
	}
	elapsed := nowTicks - c.start
	if elapsed < 0 {
		elapsed = 0
	}
	off := int(c.phaseOffset)
	if elapsed < off {
		c.currentPhase = 0
		c.currentCycle = 0
		c.cyclesDone = false
		return
	}
	elapsed -= off
	rate := int(c.phaseRate)
	count := int(c.phaseCount)
	phaseIdx := elapsed / rate
	c.currentPhase = int16(phaseIdx % count)
	cycleIdx := phaseIdx / count
	if c.cyclesToRun == 0 {
		c.currentCycle = 0
		c.cyclesDone = false
		return
	}
	n := int(c.cyclesToRun)
	if cycleIdx >= n {
		c.cyclesDone = true
		c.currentCycle = int16(n - 1)
		c.currentPhase = int16(count - 1)
		return
	}
	c.cyclesDone = false
	c.currentCycle = int16(cycleIdx)
}

// Reset clears the run (keeps static recipe). Next Advance re-latches start if still 0.
func (c *SimDrawerCycler) Reset() {
	if c == nil {
		return
	}
	c.start = 0
	c.begun = false
	c.currentPhase = 0
	c.currentCycle = 0
	c.cyclesDone = false
}

func (c *SimDrawerCycler) FinalType() *core.Type { return simDrawerCyclerType }

func registerSimDrawerCyclerMembers(t *core.Type) {
	t.RegisterProperty("tag", "Any",
		func(inst any) any { return inst.(*SimDrawerCycler).tag },
		func(inst any, value any) { inst.(*SimDrawerCycler).tag = value },
	)
	t.RegisterProperty("start", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).start) },
		func(inst any, value any) { inst.(*SimDrawerCycler).start = int(value.(core.Int)) },
	)
	t.RegisterProperty("phaseCount", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).phaseCount) },
		func(inst any, value any) { inst.(*SimDrawerCycler).phaseCount = int16(value.(core.Int)) },
	)
	t.RegisterProperty("phaseRate", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).phaseRate) },
		func(inst any, value any) { inst.(*SimDrawerCycler).phaseRate = int16(value.(core.Int)) },
	)
	t.RegisterProperty("phaseOffset", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).phaseOffset) },
		func(inst any, value any) { inst.(*SimDrawerCycler).phaseOffset = int16(value.(core.Int)) },
	)
	t.RegisterProperty("cyclesToRun", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).cyclesToRun) },
		func(inst any, value any) { inst.(*SimDrawerCycler).cyclesToRun = int16(value.(core.Int)) },
	)
	t.RegisterProperty("currentPhase", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).currentPhase) },
		nil,
	)
	t.RegisterProperty("currentCycle", "Int",
		func(inst any) any { return core.Int(inst.(*SimDrawerCycler).currentCycle) },
		nil,
	)
	t.RegisterProperty("cyclesDone", "Bool",
		func(inst any) any { return core.Bool(inst.(*SimDrawerCycler).cyclesDone) },
		nil,
	)

	// ( -- )
	t.RegisterNativeMethod("reset", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimDrawerCycler).Reset()
	})
	// ( nowTicks:i -- )
	t.RegisterNativeMethod("advance", func(f *core.RunFrame, ths *core.TypeObjPair) {
		now := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimDrawerCycler).Advance(now)
	})
	// ( -- copy:SimDrawerCycler )
	t.RegisterNativeMethod("clone", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*SimDrawerCycler).Clone())
	})
	// ( -- c:SimDrawerCycler )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewSimDrawerCycler())
	})
}

func CreateRegisterSimDrawerCyclerType() {
	t := core.NewRegisteringType("SimDrawerCycler",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimDrawerCycler]("SimDrawerCycler"),
	)
	registerSimDrawerCyclerMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simDrawerCyclerType = t
}

var (
	simDrawerCyclerType *core.Type
	_                   core.Typed = (*SimDrawerCycler)(nil)
)

func init() {
	CreateRegisterSimDrawerCyclerType()
}
