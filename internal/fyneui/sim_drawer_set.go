package fyneui

import (
	"github.com/androticus/final/internal/core"
)

// SimDrawerSet pairs a Drawer (may be shared) with a live SimDrawerCycler and
// the paint MethodRef. Library walk always invokes render (never calls
// drawer.handleRender directly). SimDrawerSet.new auto-binds render to
// drawer#handleRender; newWithRender supplies an explicit MethodRef (e.g. a
// method on the game TypedObject for starfield paint).
type SimDrawerSet struct {
	drawer SimDrawer
	cycler *SimDrawerCycler
	render *core.MethodRef
}

func NewSimDrawerSet(drawer SimDrawer, cycler *SimDrawerCycler) *SimDrawerSet {
	return NewSimDrawerSetWithRender(drawer, cycler, core.NewMethodRef(drawer, "handleRender"))
}

func NewSimDrawerSetWithRender(drawer SimDrawer, cycler *SimDrawerCycler, render *core.MethodRef) *SimDrawerSet {
	if render == nil {
		panic("SimDrawerSet: render MethodRef is required")
	}
	return &SimDrawerSet{drawer: drawer, cycler: cycler, render: render}
}

func (s *SimDrawerSet) FinalType() *core.Type { return simDrawerSetType }

func (s *SimDrawerSet) Drawer() SimDrawer { return s.drawer }

func (s *SimDrawerSet) SetDrawer(d SimDrawer) { s.drawer = d }

func (s *SimDrawerSet) Cycler() *SimDrawerCycler { return s.cycler }

func (s *SimDrawerSet) SetCycler(c *SimDrawerCycler) { s.cycler = c }

func (s *SimDrawerSet) Render() *core.MethodRef { return s.render }

func (s *SimDrawerSet) SetRender(r *core.MethodRef) {
	if r == nil {
		panic("SimDrawerSet.render: MethodRef is required")
	}
	s.render = r
}

// InvokeRender runs render with ( entity cycler -- ) already on the data stack.
func (s *SimDrawerSet) InvokeRender(f *core.RunFrame) {
	if s == nil || s.render == nil {
		panic("SimDrawerSet.invokeRender: nil render MethodRef")
	}
	s.render.Invoke(f)
}

func registerSimDrawerSetMembers(t *core.Type) {
	t.RegisterProperty("drawer", "Any",
		func(inst any) any { return inst.(*SimDrawerSet).drawer },
		func(inst any, value any) { inst.(*SimDrawerSet).drawer = value },
	)
	t.RegisterProperty("cycler", "SimDrawerCycler",
		func(inst any) any { return inst.(*SimDrawerSet).cycler },
		func(inst any, value any) {
			if value == nil {
				inst.(*SimDrawerSet).cycler = nil
				return
			}
			inst.(*SimDrawerSet).cycler = value.(*SimDrawerCycler)
		},
	)
	t.RegisterProperty("render", "MethodRef",
		func(inst any) any { return inst.(*SimDrawerSet).render },
		func(inst any, value any) {
			if value == nil {
				panic("SimDrawerSet.render: MethodRef is required")
			}
			inst.(*SimDrawerSet).render = value.(*core.MethodRef)
		},
	)
	// ( entity:SimEntity cycler:SimDrawerCycler -- ) o( set )
	t.RegisterNativeMethod("invokeRender", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimDrawerSet).InvokeRender(f)
	})
	// ( drawer:Any cycler:SimDrawerCycler -- set:SimDrawerSet )
	// Auto-binds render → drawer#handleRender.
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		cyc := f.RP.DataStack.Pop().(*SimDrawerCycler)
		drw := f.RP.DataStack.Pop()
		f.RP.DataStack.Push(NewSimDrawerSet(drw, cyc))
	})
	// ( drawer:Any cycler:SimDrawerCycler render:MethodRef -- set:SimDrawerSet )
	// drawer may be null when paint is entirely the MethodRef (e.g. game method).
	t.RegisterNativeClassMethod("newWithRender", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ref := f.RP.DataStack.Pop().(*core.MethodRef)
		cyc := f.RP.DataStack.Pop().(*SimDrawerCycler)
		drw := f.RP.DataStack.Pop()
		f.RP.DataStack.Push(NewSimDrawerSetWithRender(drw, cyc, ref))
	})
}

func CreateRegisterSimDrawerSetType() {
	t := core.NewRegisteringType("SimDrawerSet",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimDrawerSet]("SimDrawerSet"),
	)
	registerSimDrawerSetMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simDrawerSetType = t
}

var (
	simDrawerSetType *core.Type
	_                core.Typed = (*SimDrawerSet)(nil)
)

func init() {
	CreateRegisterSimDrawerSetType()
}
