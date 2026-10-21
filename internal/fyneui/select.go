package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewSelect(handler *core.MethodRef, index int, options *core.StringList) *widget.Select {
	n := options.Len()
	opts := make([]string, n)
	for i := 0; i < n; i++ {
		opts[i] = options.At(i)
	}
	var sel *widget.Select
	onChanged := func(string) {}
	if handler != nil {
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		onChanged = func(s string) {
			eventFn(sel, core.String(s))
		}
	}
	sel = widget.NewSelect(opts, onChanged)
	return sel
}

func registerSelectMembers(t *core.Type) {
	t.RegisterProperty("selectedIndex", "Int",
		func(inst any) any {
			return core.Int(inst.(*widget.Select).SelectedIndex())
		},
		func(inst any, value any) {
			sel := inst.(*widget.Select)
			idx := int(value.(core.Int))
			fyne.Do(func() {
				sel.SetSelectedIndex(idx)
			})
		},
	)

	// ( handler:MethodRef index:Int options:StringList -- w:Select )
	// handler may be nil to skip event minting.
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		options := f.RP.DataStack.Pop().(*core.StringList)
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		f.RP.DataStack.Push(NewSelect(handler, index, options))
	})
}

func CreateRegisterSelectType() {
	t := core.NewRegisteringType("Select",
		core.MakeHostTypeZero[*widget.Select](),
		core.MakeHostTypeCheckCoerce[*widget.Select]("Select"),
	)
	registerSelectMembers(t)
	applyCanvasMixin[*widget.Select](t)
	applyDisableableMixin[*widget.Select](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Select](t)
}

func init() {
	CreateRegisterSelectType()
}
