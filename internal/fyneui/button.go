package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func newButtonHandler(handler *core.MethodRef, index int) func() {
	el := core.GetEventLoop()
	eventFn := el.CreateEventFn(handler, index)
	handlerFn := func() {
		eventFn(nil, nil)
	}
	return handlerFn
}

func NewButton(handler *core.MethodRef, index int, text string) *widget.Button {
	return widget.NewButton(text, newButtonHandler(handler, index))
}

func NewButtonWithIcon(handler *core.MethodRef, index int, text string, icon fyne.Resource) *widget.Button {
	return widget.NewButtonWithIcon(text, icon, newButtonHandler(handler, index))
}

func registerButtonMembers(t *core.Type) {

	t.RegisterNativeMethod("setText", func(f *core.RunFrame, ths *core.TypeObjPair) {
		text := f.RP.DataStack.Pop().(core.String)
		btn := ths.Obj.(*widget.Button)
		fyne.Do(func() {
			btn.SetText(string(text))
		})
	})

	// ( handler:MethodRef index:Int text:String -- btn:Button )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		text := string(f.RP.DataStack.Pop().(core.String))
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		f.RP.DataStack.Push(NewButton(handler, index, text))
	})
	// ( handler:MethodRef index:Int text:String icon:Icon -- btn:Button )
	t.RegisterNativeClassMethod("newWithIcon", func(f *core.RunFrame, _ *core.TypeObjPair) {
		icon := f.RP.DataStack.Pop().(fyne.Resource)
		text := string(f.RP.DataStack.Pop().(core.String))
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		f.RP.DataStack.Push(NewButtonWithIcon(handler, index, text, icon))
	})
}

func CreateRegisterButtonType() {
	t := core.NewRegisteringType("Button",
		core.MakeHostTypeZero[*widget.Button](),
		core.MakeHostTypeCheckCoerce[*widget.Button]("Button"),
	)
	registerButtonMembers(t)
	applyCanvasMixin[*widget.Button](t)
	applyDisableableMixin[*widget.Button](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Button](t)
}

func init() {
	CreateRegisterButtonType()
}
