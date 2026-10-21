package fyneui

import (
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

// Grid is a class-only factory type. Factories return Host *fyne.Container.

func registerGridMembers(t *core.Type) {
	// ( cols:Int objs:List -- grid:Container )
	t.RegisterNativeClassMethod("newColumns", func(f *core.RunFrame, _ *core.TypeObjPair) {
		objs := canvasObjectsFromList(f.RP.DataStack.Pop().(*core.AnyList))
		cols := int(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(container.NewGridWithColumns(cols, objs...))
	})
	// ( rows:Int objs:List -- grid:Container )
	t.RegisterNativeClassMethod("newRows", func(f *core.RunFrame, _ *core.TypeObjPair) {
		objs := canvasObjectsFromList(f.RP.DataStack.Pop().(*core.AnyList))
		rows := int(f.RP.DataStack.Pop().(core.Int))
		f.RP.DataStack.Push(container.NewGridWithRows(rows, objs...))
	})
}

func CreateRegisterGridType() {
	t := core.NewRegisteringType("Grid",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Grid"),
	)
	registerGridMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterGridType()
}
