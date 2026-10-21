package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewRadioGroup(handler *core.MethodRef, index int, options *core.StringList) *widget.RadioGroup {
	n := options.Len()
	opts := make([]string, n)
	for i := 0; i < n; i++ {
		opts[i] = options.At(i)
	}
	var rg *widget.RadioGroup
	onChanged := func(string) {}
	if handler != nil {
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		onChanged = func(s string) {
			eventFn(rg, core.String(s))
		}
	}
	rg = widget.NewRadioGroup(opts, onChanged)
	return rg
}

func registerRadioGroupMembers(t *core.Type) {
	t.RegisterProperty("selected", "String",
		func(inst any) any {
			return core.String(inst.(*widget.RadioGroup).Selected)
		},
		func(inst any, value any) {
			rg := inst.(*widget.RadioGroup)
			sel := string(value.(core.String))
			fyne.Do(func() {
				rg.SetSelected(sel)
			})
		},
	)

	// ( handler:MethodRef index:Int options:StringList -- w:RadioGroup )
	// handler may be nil to skip event minting.
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		options := f.RP.DataStack.Pop().(*core.StringList)
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		f.RP.DataStack.Push(NewRadioGroup(handler, index, options))
	})
}

func CreateRegisterRadioGroupType() {
	t := core.NewRegisteringType("RadioGroup",
		core.MakeHostTypeZero[*widget.RadioGroup](),
		core.MakeHostTypeCheckCoerce[*widget.RadioGroup]("RadioGroup"),
	)
	registerRadioGroupMembers(t)
	applyCanvasMixin[*widget.RadioGroup](t)
	applyDisableableMixin[*widget.RadioGroup](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.RadioGroup](t)
}

func init() {
	CreateRegisterRadioGroupType()
}
