package fyneui

import (
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

// Border is a class-only factory type. new returns Host *fyne.Container.
// Edge slots may be nil; center is the middle fill (also may be nil).

func registerBorderMembers(t *core.Type) {
	// ( top bottom left right center -- border:Container )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		center := canvasObjectOrNil(f.RP.DataStack.Pop())
		right := canvasObjectOrNil(f.RP.DataStack.Pop())
		left := canvasObjectOrNil(f.RP.DataStack.Pop())
		bottom := canvasObjectOrNil(f.RP.DataStack.Pop())
		top := canvasObjectOrNil(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(container.NewBorder(top, bottom, left, right, center))
	})
}

func CreateRegisterBorderType() {
	t := core.NewRegisteringType("Border",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Border"),
	)
	registerBorderMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterBorderType()
}
