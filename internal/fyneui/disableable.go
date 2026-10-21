package fyneui

import (
	"fyne.io/fyne/v2"
	"github.com/androticus/final/internal/core"
)

// Disableable is a class-only mixin catalog type. ApplyMixin copies disabled
// onto Host types whose Obj is a fyne.Disableable.
var disableableType *core.Type

func registerDisableableMembers[T fyne.Disableable](t *core.Type) {
	t.RegisterProperty("disabled", "Bool",
		func(inst any) any {
			return core.Bool(inst.(T).Disabled())
		},
		func(inst any, value any) {
			obj := inst.(T)
			disabled := bool(value.(core.Bool))
			fyne.Do(func() {
				if disabled {
					obj.Disable()
				} else {
					obj.Enable()
				}
			})
		},
	)
}

func applyDisableableMixin[T fyne.Disableable](t *core.Type) {
	core.ApplyMixin(t, disableableType, func(t *core.Type) {
		registerDisableableMembers[T](t)
	})
}

func CreateRegisterDisableableType() {
	t := core.NewRegisteringType("Disableable",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Disableable"),
	)
	registerDisableableMembers[fyne.Disableable](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	disableableType = t
}
