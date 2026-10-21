package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

func registerScrollMembers(t *core.Type) {
	// ( content:CanvasObject -- scroll:Scroll )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		content := canvasObjectRequired(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(container.NewScroll(content))
	})
	// ( w:n h:n -- )  Int/Uint/Float → fyne float32
	t.RegisterNativeMethod("setMinSize", func(f *core.RunFrame, ths *core.TypeObjPair) {
		h := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		w := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		scroll := ths.Obj.(*container.Scroll)
		fyne.Do(func() {
			scroll.SetMinSize(fyne.NewSize(w, h))
			scroll.Refresh()
		})
	})
}

func CreateRegisterScrollType() {
	t := core.NewRegisteringType("Scroll",
		core.MakeHostTypeZero[*container.Scroll](),
		core.MakeHostTypeCheckCoerce[*container.Scroll]("Scroll"),
	)
	registerScrollMembers(t)
	applyCanvasMixin[*container.Scroll](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*container.Scroll](t)
}

func init() {
	CreateRegisterScrollType()
}
