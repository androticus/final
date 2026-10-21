package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

func registerSplitMembers(t *core.Type) {
	// ( offset:Float -- )
	t.RegisterNativeMethod("setOffset", func(f *core.RunFrame, ths *core.TypeObjPair) {
		offset := float64(f.RP.DataStack.Pop().(core.Float))
		split := ths.Obj.(*container.Split)
		fyne.Do(func() {
			split.SetOffset(offset)
		})
	})

	// ( leading trailing -- split:Split )
	t.RegisterNativeClassMethod("newH", func(f *core.RunFrame, _ *core.TypeObjPair) {
		trailing := canvasObjectRequired(f.RP.DataStack.Pop())
		leading := canvasObjectRequired(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(container.NewHSplit(leading, trailing))
	})
	// ( top bottom -- split:Split )
	t.RegisterNativeClassMethod("newV", func(f *core.RunFrame, _ *core.TypeObjPair) {
		bottom := canvasObjectRequired(f.RP.DataStack.Pop())
		top := canvasObjectRequired(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(container.NewVSplit(top, bottom))
	})
}

func CreateRegisterSplitType() {
	t := core.NewRegisteringType("Split",
		core.MakeHostTypeZero[*container.Split](),
		core.MakeHostTypeCheckCoerce[*container.Split]("Split"),
	)
	registerSplitMembers(t)
	applyCanvasMixin[*container.Split](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*container.Split](t)
}

func init() {
	CreateRegisterSplitType()
}
