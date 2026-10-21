package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewLabel(title string) *widget.Label {
	return widget.NewLabel(title)
}

func registerLabelMembers(t *core.Type) {
	t.RegisterNativeMethod("setText", func(f *core.RunFrame, ths *core.TypeObjPair) {
		text := f.RP.DataStack.Pop().(core.String)
		lbl := ths.Obj.(*widget.Label)
		fyne.Do(func() {
			lbl.SetText(string(text))
		})
	})
	// ( title:String -- label:Label )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		title := string(f.RP.DataStack.Pop().(core.String))
		f.RP.DataStack.Push(NewLabel(title))
	})
}

func CreateRegisterLabelType() {
	t := core.NewRegisteringType("Label",
		core.MakeHostTypeZero[*widget.Label](),
		core.MakeHostTypeCheckCoerce[*widget.Label]("Label"),
	)
	registerLabelMembers(t)
	applyCanvasMixin[*widget.Label](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Label](t)
}

func init() {
	CreateRegisterLabelType()
}
