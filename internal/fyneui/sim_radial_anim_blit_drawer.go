package fyneui

import (
	"fmt"

	"github.com/androticus/final/internal/core"
)

// RadialAnimBlitDrawer blits every Mat2c cell of mats each paint.
// Placement: packed grid origin + vels[r][c] * currentPhase (Point as δ/phase).
// Cycler currentPhase is time (not tile index). Tiles may differ in size;
// origins use max height per row and max width per column.
type RadialAnimBlitDrawer struct {
	mats      *core.AnyMat
	vels      *core.AnyMat
	phaseRate int16 // default for newCycler; phaseCount/cyclesToRun set by author
}

func NewRadialAnimBlitDrawer(mats, vels *core.AnyMat, phaseRate int16) *RadialAnimBlitDrawer {
	if mats == nil || vels == nil {
		panic("RadialAnimBlitDrawer: mats and vels required")
	}
	if mats.Rows() != vels.Rows() || mats.Cols() != vels.Cols() {
		panic(fmt.Errorf("RadialAnimBlitDrawer: mats %dx%d vs vels %dx%d",
			mats.Rows(), mats.Cols(), vels.Rows(), vels.Cols()))
	}
	if phaseRate <= 0 {
		panic(fmt.Errorf("RadialAnimBlitDrawer: phaseRate must be > 0, got %d", phaseRate))
	}
	return &RadialAnimBlitDrawer{mats: mats, vels: vels, phaseRate: phaseRate}
}

func (d *RadialAnimBlitDrawer) FinalType() *core.Type { return radialAnimBlitDrawerType }

func (d *RadialAnimBlitDrawer) Mats() *core.AnyMat { return d.mats }
func (d *RadialAnimBlitDrawer) Vels() *core.AnyMat { return d.vels }

func (d *RadialAnimBlitDrawer) PhaseRate() int16 { return d.phaseRate }

func (d *RadialAnimBlitDrawer) SetPhaseRate(r int16) {
	if r <= 0 {
		panic(fmt.Errorf("RadialAnimBlitDrawer.phaseRate: must be > 0, got %d", r))
	}
	d.phaseRate = r
}

func (d *RadialAnimBlitDrawer) SetVels(vels *core.AnyMat) {
	if vels == nil {
		panic("RadialAnimBlitDrawer.setVels: nil")
	}
	if d.mats.Rows() != vels.Rows() || d.mats.Cols() != vels.Cols() {
		panic(fmt.Errorf("RadialAnimBlitDrawer.setVels: shape %dx%d want %dx%d",
			vels.Rows(), vels.Cols(), d.mats.Rows(), d.mats.Cols()))
	}
	d.vels = vels
}

// NewCycler returns a cycler with phaseRate set; author sets phaseCount / cyclesToRun.
func (d *RadialAnimBlitDrawer) NewCycler() *SimDrawerCycler {
	c := NewSimDrawerCycler()
	c.phaseRate = d.phaseRate
	return c
}

func radialMatAt(mats *core.AnyMat, r, c int) *core.Mat2C {
	v := mats.At_(r, c)
	if v == nil {
		return nil
	}
	m, ok := v.(*core.Mat2C)
	if !ok || m == nil {
		panic(fmt.Errorf("RadialAnimBlitDrawer: mats[%d][%d] want Mat2c, got %T", r, c, v))
	}
	return m
}

func radialVelAt(vels *core.AnyMat, r, c int) core.Point {
	v := vels.At_(r, c)
	if v == nil {
		return core.Point{}
	}
	p, ok := v.(core.Point)
	if !ok {
		panic(fmt.Errorf("RadialAnimBlitDrawer: vels[%d][%d] want Point, got %T", r, c, v))
	}
	return p
}

func radialRowHeight(mats *core.AnyMat, r, cols int) int {
	max := 0
	for c := 0; c < cols; c++ {
		if m := radialMatAt(mats, r, c); m != nil && m.Rows() > max {
			max = m.Rows()
		}
	}
	return max
}

func radialColWidth(mats *core.AnyMat, rows, c int) int {
	max := 0
	for r := 0; r < rows; r++ {
		if m := radialMatAt(mats, r, c); m != nil && m.Cols() > max {
			max = m.Cols()
		}
	}
	return max
}

func radialCellOrigin(mats *core.AnyMat, r, c int) (ox, oy int) {
	rows, cols := mats.Rows(), mats.Cols()
	for rr := 0; rr < r; rr++ {
		oy += radialRowHeight(mats, rr, cols)
	}
	for cc := 0; cc < c; cc++ {
		ox += radialColWidth(mats, rows, cc)
	}
	return
}

// HandleRender blits all non-nil tiles. phase = cycler.currentPhase (time).
func (d *RadialAnimBlitDrawer) HandleRender(e *SimEntity, cyc *SimDrawerCycler) {
	if e == nil || e.sg == nil || e.sg.pixmap == nil {
		panic("RadialAnimBlitDrawer.handleRender: entity.sg.pixmap required")
	}
	phase := 0
	if cyc != nil {
		phase = int(cyc.currentPhase)
	}
	at := e.At()
	over := core.NewBlitParams()
	over.SetMask(core.BlitMaskOn)
	rows, cols := d.mats.Rows(), d.mats.Cols()
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			tile := radialMatAt(d.mats, r, c)
			if tile == nil {
				continue
			}
			ox, oy := radialCellOrigin(d.mats, r, c)
			v := radialVelAt(d.vels, r, c)
			pos := core.NewPointF(
				at.X+float64(ox)+v.X*float64(phase),
				at.Y+float64(oy)+v.Y*float64(phase),
			)
			src := core.NewBlitSrcC(tile)
			e.sg.pixmap.BlitWithOverrides(src, over, pos)
		}
	}
}

func registerRadialAnimBlitDrawerMembers(t *core.Type) {
	t.RegisterProperty("phaseRate", "Int",
		func(inst any) any { return core.Int(inst.(*RadialAnimBlitDrawer).phaseRate) },
		func(inst any, value any) { inst.(*RadialAnimBlitDrawer).SetPhaseRate(int16(value.(core.Int))) },
	)
	t.RegisterProperty("mats", "Mat",
		func(inst any) any { return inst.(*RadialAnimBlitDrawer).mats },
		nil,
	)
	t.RegisterProperty("vels", "Mat",
		func(inst any) any { return inst.(*RadialAnimBlitDrawer).vels },
		func(inst any, value any) { inst.(*RadialAnimBlitDrawer).SetVels(value.(*core.AnyMat)) },
	)
	// ( -- c:SimDrawerCycler )
	t.RegisterNativeMethod("newCycler", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*RadialAnimBlitDrawer).NewCycler())
	})
	// ( entity:SimEntity cycler:SimDrawerCycler -- )
	t.RegisterNativeMethod("handleRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		cyc := f.RP.DataStack.Pop().(*SimDrawerCycler)
		ent := f.RP.DataStack.Pop().(*SimEntity)
		ths.Obj.(*RadialAnimBlitDrawer).HandleRender(ent, cyc)
	})
	// ( mats:Mat vels:Mat phaseRate:i -- d:RadialAnimBlitDrawer )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		rate := int16(f.RP.DataStack.Pop().(core.Int))
		vels := f.RP.DataStack.Pop().(*core.AnyMat)
		mats := f.RP.DataStack.Pop().(*core.AnyMat)
		f.RP.DataStack.Push(NewRadialAnimBlitDrawer(mats, vels, rate))
	})
}

func CreateRegisterRadialAnimBlitDrawerType() {
	t := core.NewRegisteringType("RadialAnimBlitDrawer",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[RadialAnimBlitDrawer]("RadialAnimBlitDrawer"),
	)
	registerRadialAnimBlitDrawerMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	radialAnimBlitDrawerType = t
}

var (
	radialAnimBlitDrawerType *core.Type
	_                        core.Typed = (*RadialAnimBlitDrawer)(nil)
)

func init() {
	CreateRegisterRadialAnimBlitDrawerType()
}
