package fyneui

import (
	"fmt"

	"github.com/androticus/final/internal/core"
)

// AnimBlitDrawer stamps one of N Mat2c frames at entity.at.
// Frame index = cycler.currentPhase after library Advance.
// Transparent cells (A==0) are masked off.
type AnimBlitDrawer struct {
	frames    []*core.Mat2C
	phaseRate int16 // default ticks/phase for newCycler
}

func NewAnimBlitDrawer(frames []*core.Mat2C, phaseRate int16) *AnimBlitDrawer {
	if len(frames) == 0 {
		panic("AnimBlitDrawer: need at least one frame")
	}
	if phaseRate <= 0 {
		panic(fmt.Errorf("AnimBlitDrawer: phaseRate must be > 0, got %d", phaseRate))
	}
	for i, m := range frames {
		if m == nil {
			panic(fmt.Errorf("AnimBlitDrawer: frame %d is nil", i))
		}
	}
	return &AnimBlitDrawer{frames: frames, phaseRate: phaseRate}
}

func (d *AnimBlitDrawer) FinalType() *core.Type { return animBlitDrawerType }

func (d *AnimBlitDrawer) Frames() []*core.Mat2C { return d.frames }

func (d *AnimBlitDrawer) PhaseRate() int16 { return d.phaseRate }

func (d *AnimBlitDrawer) SetPhaseRate(r int16) {
	if r <= 0 {
		panic(fmt.Errorf("AnimBlitDrawer.phaseRate: must be > 0, got %d", r))
	}
	d.phaseRate = r
}

// NewCycler returns a looping cycler (cyclesToRun=0) with phaseCount=len(frames).
func (d *AnimBlitDrawer) NewCycler() *SimDrawerCycler {
	c := NewSimDrawerCycler()
	c.phaseCount = int16(len(d.frames))
	c.phaseRate = d.phaseRate
	return c
}

// HandleRender blits frames[currentPhase] at entity.at (top-left). Masks A==0.
func (d *AnimBlitDrawer) HandleRender(e *SimEntity, cyc *SimDrawerCycler) {
	if e == nil || e.sg == nil || e.sg.pixmap == nil {
		panic("AnimBlitDrawer.handleRender: entity.sg.pixmap required")
	}
	phase := 0
	if cyc != nil {
		phase = int(cyc.currentPhase)
	}
	n := len(d.frames)
	if phase < 0 || phase >= n {
		phase = 0
	}
	over := core.NewBlitParams()
	over.SetMask(core.BlitMaskOn)
	src := core.NewBlitSrcC(d.frames[phase])
	e.sg.pixmap.BlitWithOverrides(src, over, e.At())
}

func mat2CFromImageFile(path string) (*core.Mat2C, error) {
	return core.NewMat2CFromFile(path)
}

func animBlitFramesFromAnyList(list *core.AnyList) []*core.Mat2C {
	if list == nil || list.Len() == 0 {
		panic("AnimBlitDrawer: frames list empty")
	}
	out := make([]*core.Mat2C, list.Len())
	for i := 0; i < list.Len(); i++ {
		m, ok := list.At(i).(*core.Mat2C)
		if !ok || m == nil {
			panic(fmt.Errorf("AnimBlitDrawer: frame %d want Mat2c, got %T", i, list.At(i)))
		}
		out[i] = m
	}
	return out
}

func registerAnimBlitDrawerMembers(t *core.Type) {
	t.RegisterProperty("phaseRate", "Int",
		func(inst any) any { return core.Int(inst.(*AnimBlitDrawer).phaseRate) },
		func(inst any, value any) { inst.(*AnimBlitDrawer).SetPhaseRate(int16(value.(core.Int))) },
	)
	t.RegisterProperty("phaseCount", "Int",
		func(inst any) any { return core.Int(len(inst.(*AnimBlitDrawer).frames)) },
		nil,
	)
	// ( -- frames:List ) snapshot of Mat2c frames
	t.RegisterNativeMethod("frames", func(f *core.RunFrame, ths *core.TypeObjPair) {
		d := ths.Obj.(*AnimBlitDrawer)
		list := core.NewAnyListWithCapacity(len(d.frames))
		for _, m := range d.frames {
			list.Append(m)
		}
		f.RP.DataStack.Push(list)
	})
	// ( -- c:SimDrawerCycler )
	t.RegisterNativeMethod("newCycler", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*AnimBlitDrawer).NewCycler())
	})
	// ( entity:SimEntity cycler:SimDrawerCycler -- )
	t.RegisterNativeMethod("handleRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		cyc := f.RP.DataStack.Pop().(*SimDrawerCycler)
		ent := f.RP.DataStack.Pop().(*SimEntity)
		ths.Obj.(*AnimBlitDrawer).HandleRender(ent, cyc)
	})
	// ( frames:List phaseRate:i -- d:AnimBlitDrawer )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		frames := animBlitFramesFromAnyList(f.RP.DataStack.Pop().(*core.AnyList))
		f.RP.DataStack.Push(NewAnimBlitDrawer(frames, rate))
	})
	// ( paths:List phaseRate:i -- d:AnimBlitDrawer )
	t.RegisterNativeClassMethodSec("newFromPngFiles", func(f *core.RunFrame, _ *core.TypeObjPair) {
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		paths := f.RP.DataStack.Pop().(*core.AnyList)
		if paths == nil || paths.Len() == 0 {
			panic("AnimBlitDrawer.newFromPngFiles: paths empty")
		}
		frames := make([]*core.Mat2C, paths.Len())
		for i := 0; i < paths.Len(); i++ {
			path, ok := paths.At(i).(core.String)
			if !ok {
				panic(fmt.Errorf("AnimBlitDrawer.newFromPngFiles: path %d want String, got %T", i, paths.At(i)))
			}
			m, err := mat2CFromImageFile(string(path))
			if err != nil {
				core.Throw(err.Error(), f)
				return
			}
			frames[i] = m
		}
		f.RP.DataStack.Push(NewAnimBlitDrawer(frames, rate))
	}, core.SR_FileRead)
}

func CreateRegisterAnimBlitDrawerType() {
	t := core.NewRegisteringType("AnimBlitDrawer",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[AnimBlitDrawer]("AnimBlitDrawer"),
	)
	registerAnimBlitDrawerMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	animBlitDrawerType = t
}

var (
	animBlitDrawerType *core.Type
	_                  core.Typed = (*AnimBlitDrawer)(nil)
)

func init() {
	CreateRegisterAnimBlitDrawerType()
}
