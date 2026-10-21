package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewAccordion() *widget.Accordion {
	return widget.NewAccordion()
}

func registerAccordionMembers(t *core.Type) {
	// ( title:String detail:CanvasObject -- )
	t.RegisterNativeMethod("append", func(f *core.RunFrame, ths *core.TypeObjPair) {
		item := f.RP.DataStack.Pop()
		title := string(f.RP.DataStack.Pop().(core.String))
		obj, ok := item.(fyne.CanvasObject)
		if !ok {
			panic(fmt.Errorf("Accordion.append: detail is %T, want fyne.CanvasObject", item))
		}
		acc := ths.Obj.(*widget.Accordion)
		fyne.Do(func() {
			acc.Append(widget.NewAccordionItem(title, obj))
		})
	})

	// ( -- accordion:Accordion )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewAccordion())
	})
}

func CreateRegisterAccordionType() {
	t := core.NewRegisteringType("Accordion",
		core.MakeHostTypeZero[*widget.Accordion](),
		core.MakeHostTypeCheckCoerce[*widget.Accordion]("Accordion"),
	)
	registerAccordionMembers(t)
	applyCanvasMixin[*widget.Accordion](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Accordion](t)
}

func init() {
	CreateRegisterAccordionType()
}
