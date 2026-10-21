package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewActivity() *widget.Activity {
	return widget.NewActivity()
}

func registerActivityMembers(t *core.Type) {
	// ( -- )
	t.RegisterNativeMethod("start", func(f *core.RunFrame, ths *core.TypeObjPair) {
		a := ths.Obj.(*widget.Activity)
		fyne.Do(func() {
			a.Start()
		})
	})
	// ( -- )
	t.RegisterNativeMethod("stop", func(f *core.RunFrame, ths *core.TypeObjPair) {
		a := ths.Obj.(*widget.Activity)
		fyne.Do(func() {
			a.Stop()
		})
	})

	// ( -- activity:Activity )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewActivity())
	})
}

func CreateRegisterActivityType() {
	t := core.NewRegisteringType("Activity",
		core.MakeHostTypeZero[*widget.Activity](),
		core.MakeHostTypeCheckCoerce[*widget.Activity]("Activity"),
	)
	registerActivityMembers(t)
	applyCanvasMixin[*widget.Activity](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Activity](t)
}

func init() {
	CreateRegisterActivityType()
}
