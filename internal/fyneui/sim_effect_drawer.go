package fyneui

import (
	"fmt"

	"github.com/androticus/final/internal/core"
)

// EffectDrawer is programmatic paint driven by SimDrawerCycler.
// Recipe fields feed #newCycler. Paint is hybrid: MethodRef render and/or
// TypedObject del (del#handleRender). Prefer render when both set.
// Stack for paint: ( entity:SimEntity cycler:SimDrawerCycler -- ).
type EffectDrawer struct {
	phaseCount  int16
	phaseRate   int16
	phaseOffset int16
	cyclesToRun int16 // 0 = continuous
	del         *core.TypedObject
	render      *core.MethodRef
}

func NewEffectDrawer(phaseCount, phaseRate, phaseOffset, cyclesToRun int16) *EffectDrawer {
	if phaseCount < 0 {
		panic(fmt.Errorf("EffectDrawer: phaseCount must be >= 0, got %d", phaseCount))
	}
	if phaseRate < 0 {
		panic(fmt.Errorf("EffectDrawer: phaseRate must be >= 0, got %d", phaseRate))
	}
	return &EffectDrawer{
		phaseCount:  phaseCount,
		phaseRate:   phaseRate,
		phaseOffset: phaseOffset,
		cyclesToRun: cyclesToRun,
	}
}

func NewEffectDrawerWithDelegate(del *core.TypedObject, phaseCount, phaseRate, phaseOffset, cyclesToRun int16) *EffectDrawer {
	if del == nil {
		panic("EffectDrawer.newWithDelegate: del is nil")
	}
	d := NewEffectDrawer(phaseCount, phaseRate, phaseOffset, cyclesToRun)
	d.del = del
	return d
}

func NewEffectDrawerWithRender(render *core.MethodRef, phaseCount, phaseRate, phaseOffset, cyclesToRun int16) *EffectDrawer {
	if render == nil {
		panic("EffectDrawer.newWithRender: render is nil")
	}
	d := NewEffectDrawer(phaseCount, phaseRate, phaseOffset, cyclesToRun)
	d.render = render
	return d
}

func (d *EffectDrawer) FinalType() *core.Type { return effectDrawerType }

// NewCycler copies the recipe onto a fresh cycler (dynamic fields zero).
func (d *EffectDrawer) NewCycler() *SimDrawerCycler {
	c := NewSimDrawerCycler()
	c.phaseCount = d.phaseCount
	c.phaseRate = d.phaseRate
	c.phaseOffset = d.phaseOffset
	c.cyclesToRun = d.cyclesToRun
	return c
}

// HandleRender forwards to render MethodRef, else del#handleRender.
// Data stack must hold ( entity cycler -- ) for the callee.
func (d *EffectDrawer) HandleRender(f *core.RunFrame) {
	if d == nil {
		panic("EffectDrawer.handleRender: nil drawer")
	}
	if d.render != nil {
		d.render.Invoke(f)
		return
	}
	if d.del != nil {
		core.NewMethodRef(d.del, "handleRender").Invoke(f)
		return
	}
	panic("EffectDrawer.handleRender: need render MethodRef or del TypedObject")
}

func registerEffectDrawerMembers(t *core.Type) {
	t.RegisterProperty("phaseCount", "Int",
		func(inst any) any { return core.Int(inst.(*EffectDrawer).phaseCount) },
		func(inst any, value any) {
			n := int16(value.(core.Int))
			if n < 0 {
				panic(fmt.Errorf("EffectDrawer.phaseCount: must be >= 0, got %d", n))
			}
			inst.(*EffectDrawer).phaseCount = n
		},
	)
	t.RegisterProperty("phaseRate", "Int",
		func(inst any) any { return core.Int(inst.(*EffectDrawer).phaseRate) },
		func(inst any, value any) {
			n := int16(value.(core.Int))
			if n < 0 {
				panic(fmt.Errorf("EffectDrawer.phaseRate: must be >= 0, got %d", n))
			}
			inst.(*EffectDrawer).phaseRate = n
		},
	)
	t.RegisterProperty("phaseOffset", "Int",
		func(inst any) any { return core.Int(inst.(*EffectDrawer).phaseOffset) },
		func(inst any, value any) { inst.(*EffectDrawer).phaseOffset = int16(value.(core.Int)) },
	)
	t.RegisterProperty("cyclesToRun", "Int",
		func(inst any) any { return core.Int(inst.(*EffectDrawer).cyclesToRun) },
		func(inst any, value any) { inst.(*EffectDrawer).cyclesToRun = int16(value.(core.Int)) },
	)
	t.RegisterProperty("del", "TypedObject",
		func(inst any) any { return inst.(*EffectDrawer).del },
		func(inst any, value any) {
			if value == nil {
				inst.(*EffectDrawer).del = nil
				return
			}
			inst.(*EffectDrawer).del = value.(*core.TypedObject)
		},
	)
	t.RegisterProperty("render", "MethodRef",
		func(inst any) any { return inst.(*EffectDrawer).render },
		func(inst any, value any) {
			if value == nil {
				inst.(*EffectDrawer).render = nil
				return
			}
			inst.(*EffectDrawer).render = value.(*core.MethodRef)
		},
	)
	// ( -- c:SimDrawerCycler )
	t.RegisterNativeMethod("newCycler", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*EffectDrawer).NewCycler())
	})
	// ( entity:SimEntity cycler:SimDrawerCycler -- )
	t.RegisterNativeMethod("handleRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*EffectDrawer).HandleRender(f)
	})
	// ( phaseCount:i phaseRate:i phaseOffset:i cyclesToRun:i -- d:EffectDrawer )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ctr := int16(f.RP.DataStack.Pop().(core.Int))
		off := int16(f.RP.DataStack.Pop().(core.Int))
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		count := int16(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(NewEffectDrawer(count, rate, off, ctr))
	})
	// ( phaseCount:i phaseRate:i phaseOffset:i cyclesToRun:i del:TypedObject -- d )
	t.RegisterNativeClassMethod("newWithDelegate", func(f *core.RunFrame, _ *core.TypeObjPair) {
		del := f.RP.DataStack.Pop().(*core.TypedObject)
		ctr := int16(f.RP.DataStack.Pop().(core.Int))
		off := int16(f.RP.DataStack.Pop().(core.Int))
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		count := int16(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(NewEffectDrawerWithDelegate(del, count, rate, off, ctr))
	})
	// ( phaseCount:i phaseRate:i phaseOffset:i cyclesToRun:i render:MethodRef -- d )
	t.RegisterNativeClassMethod("newWithRender", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ref := f.RP.DataStack.Pop().(*core.MethodRef)
		ctr := int16(f.RP.DataStack.Pop().(core.Int))
		off := int16(f.RP.DataStack.Pop().(core.Int))
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		count := int16(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(NewEffectDrawerWithRender(ref, count, rate, off, ctr))
	})
}

func CreateRegisterEffectDrawerType() {
	t := core.NewRegisteringType("EffectDrawer",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[EffectDrawer]("EffectDrawer"),
	)
	registerEffectDrawerMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	effectDrawerType = t
}

var (
	effectDrawerType *core.Type
	_                core.Typed = (*EffectDrawer)(nil)
)

func init() {
	CreateRegisterEffectDrawerType()
}
