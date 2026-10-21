package fyneui

import (
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewSeparator() *widget.Separator {
	return widget.NewSeparator()
}

func registerSeparatorMembers(t *core.Type) {
	// ( -- sep:Separator )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewSeparator())
	})
}

func CreateRegisterSeparatorType() {
	t := core.NewRegisteringType("Separator",
		core.MakeHostTypeZero[*widget.Separator](),
		core.MakeHostTypeCheckCoerce[*widget.Separator]("Separator"),
	)
	registerSeparatorMembers(t)
	applyCanvasMixin[*widget.Separator](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Separator](t)
}

func init() {
	CreateRegisterSeparatorType()
}
