package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

// Box is a class-only factory type. newV / newH return a Host *fyne.Container
// (Final type Container), not a Box instance — so Fyne can walk/layout/paint.

func canvasObjectOrNil(v any) fyne.CanvasObject {
	if v == nil {
		return nil
	}
	obj, ok := v.(fyne.CanvasObject)
	if !ok {
		panic(fmt.Errorf("want fyne.CanvasObject, got %T", v))
	}
	return obj
}

func canvasObjectRequired(v any) fyne.CanvasObject {
	obj := canvasObjectOrNil(v)
	if obj == nil {
		panic(fmt.Errorf("want fyne.CanvasObject, got nil"))
	}
	return obj
}

func canvasObjectsFromList(list *core.AnyList) []fyne.CanvasObject {
	n := list.Len()
	objs := make([]fyne.CanvasObject, n)
	for i := 0; i < n; i++ {
		objs[i] = canvasObjectRequired(list.At(i))
	}
	return objs
}

func classOnlyCheckCoerce(typeName string) func(any) any {
	return func(value any) any {
		panic(fmt.Errorf("%s.checkCoerceType: %s is a class-only type (got %T)", typeName, typeName, value))
	}
}

func registerBoxMembers(t *core.Type) {
	// ( objs:List -- box:Container )
	t.RegisterNativeClassMethod("newV", func(f *core.RunFrame, _ *core.TypeObjPair) {
		objs := canvasObjectsFromList(f.RP.DataStack.Pop().(*core.AnyList))
		f.RP.DataStack.Push(container.NewVBox(objs...))
	})
	// ( objs:List -- box:Container )
	t.RegisterNativeClassMethod("newH", func(f *core.RunFrame, _ *core.TypeObjPair) {
		objs := canvasObjectsFromList(f.RP.DataStack.Pop().(*core.AnyList))
		f.RP.DataStack.Push(container.NewHBox(objs...))
	})
}

func CreateRegisterBoxType() {
	t := core.NewRegisteringType("Box",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Box"),
	)
	registerBoxMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterBoxType()
}
