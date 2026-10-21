package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewForm() *widget.Form {
	return widget.NewForm()
}

func registerFormMembers(t *core.Type) {
	// ( label:String widget:CanvasObject -- )
	t.RegisterNativeMethod("append", func(f *core.RunFrame, ths *core.TypeObjPair) {
		item := f.RP.DataStack.Pop()
		label := string(f.RP.DataStack.Pop().(core.String))
		obj, ok := item.(fyne.CanvasObject)
		if !ok {
			panic(fmt.Errorf("Form.append: widget is %T, want fyne.CanvasObject", item))
		}
		form := ths.Obj.(*widget.Form)
		fyne.Do(func() {
			form.Append(label, obj)
		})
	})

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnCancel", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		form := ths.Obj.(*widget.Form)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			form.OnCancel = func() {
				eventFn(form)
			}
			form.Refresh()
		})
	})

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnSubmit", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		form := ths.Obj.(*widget.Form)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			form.OnSubmit = func() {
				eventFn(form)
			}
			form.Refresh()
		})
	})

	// ( parent:Window title:s -- closer:i )
	// Modal wait: Submit → btnOk, Cancel/X → btnCancel. Caller harvests fields after Ok.
	// Non-modal setOnSubmit / setOnCancel are restored when the wait ends.
	t.RegisterNativeMethod("showModal", func(f *core.RunFrame, ths *core.TypeObjPair) {
		title := string(f.RP.DataStack.Pop().(core.String))
		parent := dialogParent(f.RP.DataStack.Pop())
		form := ths.Obj.(*widget.Form)
		w := beginModalWait()
		var d dialog.Dialog
		var prevSubmit, prevCancel func()
		fyne.DoAndWait(func() {
			prevSubmit = form.OnSubmit
			prevCancel = form.OnCancel
			d = dialog.NewCustomWithoutButtons(title, form, parent.Window)
			form.OnSubmit = func() {
				w.complete(btnOk, nil)
				d.Hide()
			}
			form.OnCancel = func() {
				w.complete(btnCancel, nil)
				d.Hide()
			}
			d.SetOnClosed(func() {
				w.complete(btnCancel, nil)
			})
			form.Refresh()
			d.Show()
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
			form.OnSubmit = prevSubmit
			form.OnCancel = prevCancel
			form.Refresh()
		})
		f.RP.DataStack.Push(core.Int(res.CloserIndex))
	})

	// ( -- form:Form )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewForm())
	})
}

func CreateRegisterFormType() {
	t := core.NewRegisteringType("Form",
		core.MakeHostTypeZero[*widget.Form](),
		core.MakeHostTypeCheckCoerce[*widget.Form]("Form"),
	)
	registerFormMembers(t)
	applyCanvasMixin[*widget.Form](t)
	applyDisableableMixin[*widget.Form](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Form](t)
}

func init() {
	CreateRegisterFormType()
}
