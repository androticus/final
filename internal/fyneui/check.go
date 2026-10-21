package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewCheck(label string) *widget.Check {
	return widget.NewCheck(label, func(bool) {})
}

func registerCheckMembers(t *core.Type) {
	t.RegisterProperty("checked", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*widget.Check).Checked)
		},
		func(inst any, value any) {
			c := inst.(*widget.Check)
			checked := bool(value.(core.Bool))
			fyne.Do(func() {
				c.SetChecked(checked)
			})
		},
	)

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnChanged", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		c := ths.Obj.(*widget.Check)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			c.OnChanged = func(checked bool) {
				eventFn(c, core.Bool(checked))
			}
		})
	})

	// ( label:String -- check:Check )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		label := string(f.RP.DataStack.Pop().(core.String))
		f.RP.DataStack.Push(NewCheck(label))
	})
}

func CreateRegisterCheckType() {
	t := core.NewRegisteringType("Check",
		core.MakeHostTypeZero[*widget.Check](),
		core.MakeHostTypeCheckCoerce[*widget.Check]("Check"),
	)
	registerCheckMembers(t)
	applyCanvasMixin[*widget.Check](t)
	applyDisableableMixin[*widget.Check](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Check](t)
}

func init() {
	CreateRegisterCheckType()
}
