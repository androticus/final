package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

func registerAppTabsMembers(t *core.Type) {
	t.RegisterProperty("selectedIndex", "Int",
		func(inst any) any {
			return core.Int(inst.(*container.AppTabs).SelectedIndex())
		},
		func(inst any, value any) {
			tabs := inst.(*container.AppTabs)
			idx := int(value.(core.Int))
			fyne.Do(func() {
				tabs.SelectIndex(idx)
			})
		},
	)

	// ( title:String content:CanvasObject -- )
	t.RegisterNativeMethod("append", func(f *core.RunFrame, ths *core.TypeObjPair) {
		content := canvasObjectRequired(f.RP.DataStack.Pop())
		title := string(f.RP.DataStack.Pop().(core.String))
		tabs := ths.Obj.(*container.AppTabs)
		fyne.Do(func() {
			tabs.Append(container.NewTabItem(title, content))
		})
	})

	// ( -- tabs:AppTabs )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(container.NewAppTabs())
	})
}

func CreateRegisterAppTabsType() {
	t := core.NewRegisteringType("AppTabs",
		core.MakeHostTypeZero[*container.AppTabs](),
		core.MakeHostTypeCheckCoerce[*container.AppTabs]("AppTabs"),
	)
	registerAppTabsMembers(t)
	applyCanvasMixin[*container.AppTabs](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*container.AppTabs](t)
}

func init() {
	CreateRegisterAppTabsType()
}
