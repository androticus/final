package fyneui

import (
	"github.com/androticus/final/internal/core"
)

// BlitDrawer stamps one Mat2c at entity.at (top-left). Transparent cells (A==0) masked.
// Static look — use AnimBlitDrawer for multi-frame.
type BlitDrawer struct {
	mat *core.Mat2C
}

func NewBlitDrawer(mat *core.Mat2C) *BlitDrawer {
	if mat == nil {
		panic("BlitDrawer: mat is nil")
	}
	return &BlitDrawer{mat: mat}
}

func (d *BlitDrawer) FinalType() *core.Type { return blitDrawerType }

func (d *BlitDrawer) Mat() *core.Mat2C { return d.mat }

// NewCycler returns an idle cycler (phaseCount 0).
func (d *BlitDrawer) NewCycler() *SimDrawerCycler {
	return NewSimDrawerCycler()
}

// HandleRender blits mat at entity.at. Masks A==0. cycler unused.
func (d *BlitDrawer) HandleRender(e *SimEntity, _ *SimDrawerCycler) {
	if e == nil || e.sg == nil || e.sg.pixmap == nil {
		panic("BlitDrawer.handleRender: entity.sg.pixmap required")
	}
	over := core.NewBlitParams()
	over.SetMask(core.BlitMaskOn)
	src := core.NewBlitSrcC(d.mat)
	e.sg.pixmap.BlitWithOverrides(src, over, e.At())
}

func registerBlitDrawerMembers(t *core.Type) {
	t.RegisterProperty("mat", "Mat2c",
		func(inst any) any { return inst.(*BlitDrawer).mat },
		nil,
	)
	t.RegisterProperty("rows", "Int",
		func(inst any) any { return core.Int(inst.(*BlitDrawer).mat.Rows()) },
		nil,
	)
	t.RegisterProperty("cols", "Int",
		func(inst any) any { return core.Int(inst.(*BlitDrawer).mat.Cols()) },
		nil,
	)
	// ( -- c:SimDrawerCycler )
	t.RegisterNativeMethod("newCycler", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*BlitDrawer).NewCycler())
	})
	// ( entity:SimEntity cycler:SimDrawerCycler -- )
	t.RegisterNativeMethod("handleRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		cyc := f.RP.DataStack.Pop().(*SimDrawerCycler)
		ent := f.RP.DataStack.Pop().(*SimEntity)
		ths.Obj.(*BlitDrawer).HandleRender(ent, cyc)
	})
	// ( mat:Mat2c -- d:BlitDrawer )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		mat := f.RP.DataStack.Pop().(*core.Mat2C)
		f.RP.DataStack.Push(NewBlitDrawer(mat))
	})
	// ( path:s -- d:BlitDrawer )
	t.RegisterNativeClassMethodSec("newFromPngFile", func(f *core.RunFrame, _ *core.TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(core.String))
		m, err := mat2CFromImageFile(path)
		if err != nil {
			core.Throw(err.Error(), f)
			return
		}
		f.RP.DataStack.Push(NewBlitDrawer(m))
	}, core.SR_FileRead)
}

func CreateRegisterBlitDrawerType() {
	t := core.NewRegisteringType("BlitDrawer",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[BlitDrawer]("BlitDrawer"),
	)
	registerBlitDrawerMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	blitDrawerType = t
}

var (
	blitDrawerType *core.Type
	_              core.Typed = (*BlitDrawer)(nil)
)

func init() {
	CreateRegisterBlitDrawerType()
}
