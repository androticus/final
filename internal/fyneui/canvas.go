package fyneui

import (
	"fyne.io/fyne/v2"
	"github.com/androticus/final/internal/core"
)

// Canvas is a class-only mixin catalog type. ApplyMixin copies visible / refresh
// onto Host types whose Obj is a fyne.CanvasObject.
var canvasType *core.Type

func registerCanvasMembers[T fyne.CanvasObject](t *core.Type) {
	t.RegisterProperty("visible", "Bool",
		func(inst any) any {
			return core.Bool(inst.(T).Visible())
		},
		func(inst any, value any) {
			obj := inst.(T)
			vis := bool(value.(core.Bool))
			fyne.Do(func() {
				if vis {
					obj.Show()
				} else {
					obj.Hide()
				}
			})
		},
	)
	// ( -- )
	t.RegisterNativeMethod("refresh", func(_ *core.RunFrame, ths *core.TypeObjPair) {
		obj := ths.Obj.(T)
		fyne.Do(func() {
			obj.Refresh()
		})
	})
}

func applyCanvasMixin[T fyne.CanvasObject](t *core.Type) {
	core.ApplyMixin(t, canvasType, func(t *core.Type) {
		registerCanvasMembers[T](t)
	})
}

func CreateRegisterCanvasType() {
	t := core.NewRegisteringType("Canvas",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Canvas"),
	)
	registerCanvasMembers[fyne.CanvasObject](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	canvasType = t
}
