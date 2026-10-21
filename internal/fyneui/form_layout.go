package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/androticus/final/internal/core"
)

// FormLayout is a class-only factory type. new returns Host *fyne.Container
// with Fyne's two-column form layout (label, widget pairs).

func registerFormLayoutMembers(t *core.Type) {
	// ( objs:List -- form:Container )
	// objs is flat: label0, widget0, label1, widget1, … (even length).
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		list := f.RP.DataStack.Pop().(*core.AnyList)
		if list.Len()%2 != 0 {
			panic(fmt.Errorf("FormLayout.new: objs length %d must be even (label/widget pairs)", list.Len()))
		}
		objs := canvasObjectsFromList(list)
		f.RP.DataStack.Push(container.New(layout.NewFormLayout(), objs...))
	})
}

func CreateRegisterFormLayoutType() {
	t := core.NewRegisteringType("FormLayout",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("FormLayout"),
	)
	registerFormLayoutMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterFormLayoutType()
}
