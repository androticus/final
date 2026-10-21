package fyneui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewEntry() *widget.Entry {
	return widget.NewEntry()
}

func NewMultiLineEntry() *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapWord
	return e
}

func setEntryText(e *widget.Entry, text string) {
	if e.Disabled() {
		e.Enable()
		e.SetText(text)
		e.CursorRow = strings.Count(text, "\n")
		e.CursorColumn = 0
		e.Refresh()
		e.Disable()
		return
	}
	e.SetText(text)
}

func registerEntryMembers(t *core.Type) {
	t.RegisterProperty("text", "String",
		func(inst any) any {
			return core.String(inst.(*widget.Entry).Text)
		},
		func(inst any, value any) {
			e := inst.(*widget.Entry)
			text := string(value.(core.String))
			fyne.Do(func() {
				setEntryText(e, text)
			})
		},
	)

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnChanged", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		e := ths.Obj.(*widget.Entry)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			e.OnChanged = func(s string) {
				eventFn(e, core.String(s))
			}
		})
	})

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnSubmitted", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		e := ths.Obj.(*widget.Entry)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			e.OnSubmitted = func(s string) {
				eventFn(e, core.String(s))
			}
		})
	})

	// ( -- entry:Entry )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewEntry())
	})

	// ( -- entry:Entry )
	t.RegisterNativeClassMethod("newMultiLine", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewMultiLineEntry())
	})
}

func CreateRegisterEntryType() {
	t := core.NewRegisteringType("Entry",
		core.MakeHostTypeZero[*widget.Entry](),
		core.MakeHostTypeCheckCoerce[*widget.Entry]("Entry"),
	)
	registerEntryMembers(t)
	applyCanvasMixin[*widget.Entry](t)
	applyDisableableMixin[*widget.Entry](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Entry](t)
}

func init() {
	CreateRegisterEntryType()
}
