package fyneui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/androticus/final/internal/core"
)

// Key names: fyne.io/fyne/v2/key.go (portable) and driver/desktop/key.go (modifiers).
// Letters "A"–"Z" and digits "0"–"9" are not class constants — compare to those strings.

// KeyEvent is a decoded keyboard Event. Name and Down come from params.
// isLeft etc. are const properties that compare Name to fyne.KeyName constants.
type KeyEvent struct {
	Name string
	Down bool
}

func (e KeyEvent) FinalType() *core.Type { return keyEventType }

func DecodeKeyEvent(ev *core.Event) KeyEvent {
	if ev == nil || ev.Params == nil || ev.Params.Len() < 2 {
		panic(fmt.Errorf("Keys.decode: Event.params must be [ name:String down:Bool ]"))
	}
	return KeyEvent{
		Name: string(ev.Params.At(0).(core.String)),
		Down: bool(ev.Params.At(1).(core.Bool)),
	}
}

func bindWindowKeyDownUp(win *Window, handlerDown *core.MethodRef, indexDown int, handlerUp *core.MethodRef, indexUp int, autoRepeat bool) {
	if autoRepeat && (handlerDown == nil || handlerUp == nil) {
		panic(fmt.Errorf("Keys.bindWindowKeyDownUp: autoRepeat is true, but handlerDown or handlerUp is nil"))
	}
	desk := win.Canvas().(desktop.Canvas)
	var eventFnDown core.HandlerMintEventFn
	if handlerDown != nil {
		eventFnDown = core.GetEventLoop().CreateEventFn(handlerDown, indexDown)
		fnDown := func(ev *fyne.KeyEvent) {
			name := string(ev.Name)
			eventFnDown(win, core.String(name), core.Bool(true))
			if ar := win.keyAR; ar != nil {
				ar.onDown(name)
			}
		}
		desk.SetOnKeyDown(fnDown)
	}
	if handlerUp != nil {
		eventFnUp := core.GetEventLoop().CreateEventFn(handlerUp, indexUp)
		fnUp := func(ev *fyne.KeyEvent) {
			name := string(ev.Name)
			eventFnUp(win, core.String(name), core.Bool(false))
			if ar := win.keyAR; ar != nil {
				ar.onUp(name)
			}
		}
		desk.SetOnKeyUp(fnUp)
	}
	if win.keyAR != nil && eventFnDown != nil {
		win.keyAR.setMintDown(eventFnDown)
	}
}

func keysStringConst(t *core.Type, name string, val fyne.KeyName) {
	s := core.String(val)
	t.RegisterClassProperty(name, "String",
		func(_ any) any { return s },
		nil,
	)
}

func registerKeysMembers(t *core.Type) {
	// Named specials and punctuation (not A–Z / 0–9). Values are fyne.KeyName strings.
	keysStringConst(t, "escape", fyne.KeyEscape)
	keysStringConst(t, "return", fyne.KeyReturn)
	keysStringConst(t, "tab", fyne.KeyTab)
	keysStringConst(t, "backspace", fyne.KeyBackspace)
	keysStringConst(t, "insert", fyne.KeyInsert)
	keysStringConst(t, "delete", fyne.KeyDelete)
	keysStringConst(t, "right", fyne.KeyRight)
	keysStringConst(t, "left", fyne.KeyLeft)
	keysStringConst(t, "down", fyne.KeyDown)
	keysStringConst(t, "up", fyne.KeyUp)
	keysStringConst(t, "pageUp", fyne.KeyPageUp)     // "Prior"
	keysStringConst(t, "pageDown", fyne.KeyPageDown) // "Next"
	keysStringConst(t, "home", fyne.KeyHome)
	keysStringConst(t, "end", fyne.KeyEnd)
	keysStringConst(t, "f1", fyne.KeyF1)
	keysStringConst(t, "f2", fyne.KeyF2)
	keysStringConst(t, "f3", fyne.KeyF3)
	keysStringConst(t, "f4", fyne.KeyF4)
	keysStringConst(t, "f5", fyne.KeyF5)
	keysStringConst(t, "f6", fyne.KeyF6)
	keysStringConst(t, "f7", fyne.KeyF7)
	keysStringConst(t, "f8", fyne.KeyF8)
	keysStringConst(t, "f9", fyne.KeyF9)
	keysStringConst(t, "f10", fyne.KeyF10)
	keysStringConst(t, "f11", fyne.KeyF11)
	keysStringConst(t, "f12", fyne.KeyF12)
	keysStringConst(t, "enter", fyne.KeyEnter) // keypad "KP_Enter"
	keysStringConst(t, "space", fyne.KeySpace)
	keysStringConst(t, "apostrophe", fyne.KeyApostrophe)
	keysStringConst(t, "comma", fyne.KeyComma)
	keysStringConst(t, "minus", fyne.KeyMinus)
	keysStringConst(t, "period", fyne.KeyPeriod)
	keysStringConst(t, "slash", fyne.KeySlash)
	keysStringConst(t, "backslash", fyne.KeyBackslash)
	keysStringConst(t, "leftBracket", fyne.KeyLeftBracket)
	keysStringConst(t, "rightBracket", fyne.KeyRightBracket)
	keysStringConst(t, "semicolon", fyne.KeySemicolon)
	keysStringConst(t, "equal", fyne.KeyEqual)
	keysStringConst(t, "asterisk", fyne.KeyAsterisk)
	keysStringConst(t, "plus", fyne.KeyPlus)
	keysStringConst(t, "backtick", fyne.KeyBackTick)
	keysStringConst(t, "leftShift", desktop.KeyShiftLeft)
	keysStringConst(t, "rightShift", desktop.KeyShiftRight)
	keysStringConst(t, "leftControl", desktop.KeyControlLeft)
	keysStringConst(t, "rightControl", desktop.KeyControlRight)
	keysStringConst(t, "leftAlt", desktop.KeyAltLeft)
	keysStringConst(t, "rightAlt", desktop.KeyAltRight)
	keysStringConst(t, "leftSuper", desktop.KeySuperLeft)
	keysStringConst(t, "rightSuper", desktop.KeySuperRight)
	keysStringConst(t, "menu", desktop.KeyMenu)
	keysStringConst(t, "printScreen", desktop.KeyPrintScreen)
	keysStringConst(t, "capsLock", desktop.KeyCapsLock)

	// ( ev:Event -- k:KeyEvent )
	t.RegisterNativeClassMethod("decode", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ev := f.RP.DataStack.Pop().(*core.Event)
		f.RP.DataStack.Push(DecodeKeyEvent(ev))
	})
}

func registerKeyEventMembers(t *core.Type) {
	t.RegisterProperty("name", "String",
		func(inst any) any { return core.String(inst.(KeyEvent).Name) },
		nil,
	)
	t.RegisterProperty("down", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Down) },
		nil,
	)
	t.RegisterProperty("isLeft", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeyLeft)) },
		nil,
	)
	t.RegisterProperty("isRight", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeyRight)) },
		nil,
	)
	t.RegisterProperty("isUp", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeyUp)) },
		nil,
	)
	t.RegisterProperty("isDown", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeyDown)) },
		nil,
	)
	t.RegisterProperty("isSpace", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeySpace)) },
		nil,
	)
	t.RegisterProperty("isEscape", "Bool",
		func(inst any) any { return core.Bool(inst.(KeyEvent).Name == string(fyne.KeyEscape)) },
		nil,
	)
}

func CreateRegisterKeyEventType() {
	t := core.NewRegisteringType("KeyEvent",
		core.MakeHostTypeZero[KeyEvent](),
		core.MakeHostTypeCheckCoerce[KeyEvent]("KeyEvent"),
	)
	registerKeyEventMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	keyEventType = t
}

func CreateRegisterKeysType() {
	t := core.NewRegisteringType("Keys",
		core.ReferenceTypeZeroFn,
		func(value any) any {
			panic(fmt.Errorf("Keys.checkCoerceType: Keys is a class-only type (got %T)", value))
		},
	)
	registerKeysMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	keysType = t
}

var (
	keyEventType *core.Type
	keysType     *core.Type
	_            core.Typed = (*KeyEvent)(nil)
)

func init() {
	CreateRegisterKeyEventType()
	CreateRegisterKeysType()
}
