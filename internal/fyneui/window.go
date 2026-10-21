package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/androticus/final/internal/core"
)

type Window struct {
	fyne.Window
	menuBar    *MenuBar
	contentSet bool
	autoRepeat bool // for key events, readonly
	keyAR      *keyAutoRepeat
}

func NewWindow(app *App, title string) *Window {
	var w fyne.Window
	fyne.DoAndWait(func() {
		w = app.NewWindow(title)
	})
	return &Window{Window: w}
}

func (a *Window) FinalType() *core.Type {
	return windowType
}

func (w *Window) setAutoRepeat(autoRepeat bool) {
	if w.autoRepeat == autoRepeat {
		return
	}
	w.autoRepeat = autoRepeat
	if autoRepeat {
		w.keyAR = newKeyAutoRepeat(w)
		return
	}
	if w.keyAR != nil {
		w.keyAR.close()
		w.keyAR = nil
	}
}
func (w *Window) setMenuBar(bar *MenuBar) {
	if w.contentSet {
		panic(fmt.Errorf("Window.setMenuBar: must be called before setContent"))
	}
	if w.menuBar != nil {
		panic(fmt.Errorf("Window.setMenuBar: already set"))
	}
	if bar == nil {
		panic(fmt.Errorf("Window.setMenuBar: MenuBar is nil"))
	}
	w.menuBar = bar
}

// userContentTree records that content was set. If a menu bar is present,
// wrap as Border (bar at top, user widget in center).
func (w *Window) userContentTree(content fyne.CanvasObject) fyne.CanvasObject {
	w.contentSet = true
	if w.menuBar == nil {
		return content
	}
	return container.NewBorder(w.menuBar, nil, nil, nil, content)
}

func registerWindowMembers(t *core.Type) {
	t.RegisterProperty("autoRepeat", "Bool",
		func(inst any) any { return core.Bool(inst.(*Window).autoRepeat) },
		nil,
	)
	// ( bar:MenuBar -- )
	t.RegisterNativeMethod("setMenuBar", func(f *core.RunFrame, ths *core.TypeObjPair) {
		bar := f.RP.DataStack.Pop().(*MenuBar)
		ths.Obj.(*Window).setMenuBar(bar)
	})
	// ( widget:any -- )
	t.RegisterNativeMethod("setContent", func(f *core.RunFrame, ths *core.TypeObjPair) {
		content := f.RP.DataStack.Pop().(fyne.CanvasObject)
		win := ths.Obj.(*Window)
		tree := win.userContentTree(content)
		fyne.Do(func() {
			win.SetContent(tree)
		})
	})
	t.RegisterNativeMethod("show", func(f *core.RunFrame, ths *core.TypeObjPair) {
		win := ths.Obj.(*Window)
		fyne.Do(func() {
			win.Show()
		})
	})
	// ( -- )
	t.RegisterNativeMethod("hide", func(f *core.RunFrame, ths *core.TypeObjPair) {
		win := ths.Obj.(*Window)
		fyne.Do(func() {
			win.Hide()
		})
	})
	// ( -- )  Chrome close hides instead of destroying the native viewport.
	t.RegisterNativeMethod("setHideOnClose", func(f *core.RunFrame, ths *core.TypeObjPair) {
		win := ths.Obj.(*Window)
		fyne.Do(func() {
			win.SetCloseIntercept(func() {
				win.Hide()
			})
		})
	})
	// ( w:n h:n -- )  Int/Uint/Float → fyne float32
	t.RegisterNativeMethod("resize", func(f *core.RunFrame, ths *core.TypeObjPair) {
		h := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		w := float32(core.CheckCoerceNumericToFloat(f.RP.DataStack.Pop()))
		win := ths.Obj.(*Window)
		fyne.Do(func() {
			win.Resize(fyne.NewSize(w, h))
		})
	})
	// ( b:Bool -- )
	t.RegisterNativeMethod("setFixedSize", func(f *core.RunFrame, ths *core.TypeObjPair) {
		fixed := bool(f.RP.DataStack.Pop().(core.Bool))
		win := ths.Obj.(*Window)
		fyne.Do(func() {
			win.SetFixedSize(fixed)
		})
	})
	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnClosed", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		win := ths.Obj.(*Window)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			win.SetOnClosed(func() {
				eventFn(win, nil)
			})
		})
	})
	// ( down:MethodRef downIdx:Int up:MethodRef upIdx:Int autoRepeat:Bool -- )
	t.RegisterNativeMethod("setOnKeyDownUp", func(f *core.RunFrame, ths *core.TypeObjPair) {
		win := ths.Obj.(*Window)
		autoRepeat := bool(f.RP.DataStack.Pop().(core.Bool))
		indexUp := int(f.RP.DataStack.Pop().(core.Int))
		handlerUp, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		indexDown := int(f.RP.DataStack.Pop().(core.Int))
		handlerDown, _ := f.RP.DataStack.Pop().(*core.MethodRef)
		win.setAutoRepeat(autoRepeat) // must be called before bindWindowKeyDownUp
		bindWindowKeyDownUp(win, handlerDown, indexDown, handlerUp, indexUp, autoRepeat)
	})

	// ( app:App title:String -- window:Window )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		title := string(f.RP.DataStack.Pop().(core.String))
		app := f.RP.DataStack.Pop().(*App)
		f.RP.DataStack.Push(NewWindow(app, title))
	})
}

func CreateRegisterWindowType() {
	t := core.NewRegisteringType("Window",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[Window]("Window"),
	)
	registerWindowMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	windowType = t
}

var (
	windowType *core.Type
	_          core.Typed = (*Window)(nil)
)

func init() {
	CreateRegisterWindowType()
}
