package finaltui

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

// Tui is the dual-context bridge object (Final worker ↔ host UI).
// Dialog rendezvous uses plain Go chans (not Final Pipes) — host machinery only.
// Headed chrome EventLoop handlers (onSend / onQuit) live here like FinalGui.
type Tui struct {
	typ *finalembed.Type

	uiMu     sync.Mutex
	uiReq    chan *UiRequest
	uiRes    chan *UiResponse
	inFlight atomic.Bool
	replied  atomic.Bool // one reply per in-flight raise (Break vs UI race)

	chromeMu sync.Mutex
	chrome   *appUI // set while headed EventLoop chrome is live
}

// FinalType implements the Typed interface for host instances on the Final stack.
func (t *Tui) FinalType() *finalembed.Type { return t.typ }

// Tui_ is the package singleton. Final scripts see the same instance as var Tui_ after Init.
var Tui_ = &Tui{}

func registerTui(ty *finalembed.Type) {
	Tui_.typ = ty
	ty.RegisterNativeClassMethod("instance", func(f *finalembed.RunFrame, _ *finalembed.TypeObjPair) {
		f.RP.DataStack.Push(Tui_)
	})
	reg := func(name string, fn func(*Tui, *finalembed.RunFrame)) {
		ty.RegisterNativeMethod(name, func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
			fn(ths.Obj.(*Tui), f)
		})
	}
	// ( ev:Event -- ) — Event.params[0] may carry prompt/path script strings
	reg("onSend", (*Tui).onSend)
	reg("onQuit", (*Tui).onQuit)
}

func (t *Tui) bindChrome(ui *appUI) {
	t.chromeMu.Lock()
	t.chrome = ui
	t.chromeMu.Unlock()
}

func (t *Tui) unbindChrome() {
	t.chromeMu.Lock()
	t.chrome = nil
	t.chromeMu.Unlock()
}

func (t *Tui) chromeUI() *appUI {
	t.chromeMu.Lock()
	defer t.chromeMu.Unlock()
	return t.chrome
}

func (t *Tui) onSend(f *finalembed.RunFrame) {
	src := takeEventStringParam(f)
	if src == "" {
		return
	}
	ui := t.chromeUI()
	if ui == nil {
		return
	}
	ui.runInterpret(f, src, true)
}

func (t *Tui) onQuit(f *finalembed.RunFrame) {
	popEvent(f)
	if _, err := core.InterpretSource(f.Task, f.RP, "EventLoop.inst.close^"); err != nil {
		if ui := t.chromeUI(); ui != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.appendResponse(fmt.Sprintf("Error: %v\n", err))
				ui.setBusy(false)
			})
		}
		return
	}
	if ui := t.chromeUI(); ui != nil {
		ui.app.QueueUpdateDraw(func() { ui.app.Stop() })
	}
}

// takeEventStringParam pops the Event and returns Params[0] as a trimmed string.
func takeEventStringParam(f *finalembed.RunFrame) string {
	ev := popEvent(f)
	if ev == nil || ev.Params == nil || ev.Params.Len() == 0 {
		return ""
	}
	s, ok := ev.Params.At(0).(core.String)
	if !ok {
		return ""
	}
	return strings.TrimRight(string(s), "\r\n")
}

func popEvent(f *finalembed.RunFrame) *core.Event {
	if f == nil || f.RP == nil || f.RP.DataStack.Len() == 0 {
		return nil
	}
	ev, _ := f.RP.DataStack.Pop().(*core.Event)
	return ev
}

// ArmUI allocates the request/response channels. Call once from the host UI before Run.
func (t *Tui) ArmUI() {
	t.uiMu.Lock()
	defer t.uiMu.Unlock()
	if t.uiReq != nil {
		return
	}
	t.uiReq = make(chan *UiRequest, 1)
	t.uiRes = make(chan *UiResponse, 1)
}

// UIRequests is the host-facing receive end (Final → UI).
func (t *Tui) UIRequests() <-chan *UiRequest {
	t.uiMu.Lock()
	defer t.uiMu.Unlock()
	return t.uiReq
}

// raise blocks the Final worker until the UI replies. Panics if another dialog is in flight
// or if ArmUI was not called.
func (t *Tui) raise(cmd UiCmd, params *finalembed.DynamicObject) *UiResponse {
	t.uiMu.Lock()
	reqCh, resCh := t.uiReq, t.uiRes
	t.uiMu.Unlock()
	if reqCh == nil || resCh == nil {
		panic(fmt.Errorf("finaltui: UI not armed (ArmUI before Run)"))
	}
	if !t.inFlight.CompareAndSwap(false, true) {
		panic(fmt.Errorf("finaltui: dialog already in flight"))
	}
	t.replied.Store(false)
	defer t.inFlight.Store(false)

	reqCh <- &UiRequest{Cmd: cmd, Params: params}
	return <-resCh
}

func (t *Tui) raisePrompt(params *finalembed.DynamicObject) *UiResponse {
	return t.raise(UiCmdPrompt, params)
}

func (t *Tui) raiseField(params *finalembed.DynamicObject) *UiResponse {
	return t.raise(UiCmdField, params)
}

// Complete delivers at most one UiResponse for the current in-flight raise.
// Used by the host UI (button/Esc) and by Break→Cancel.
func (t *Tui) Complete(res *UiResponse) {
	if res == nil {
		return
	}
	if !t.inFlight.Load() {
		return
	}
	if !t.replied.CompareAndSwap(false, true) {
		return
	}
	t.uiMu.Lock()
	resCh := t.uiRes
	t.uiMu.Unlock()
	if resCh == nil {
		return
	}
	resCh <- res
}

// CompleteChoice builds an OK response with Result.choice set (Prompt / Cancel).
func CompleteChoice(btn Btn) *UiResponse {
	result := finalembed.NewDynamicObject()
	result.Set(DOFieldChoice.String(), finalembed.Int(btn))
	return &UiResponse{Status: UiStatusOK, Result: result}
}

// CompleteFields builds an OK response with choice and fields (Field dialog Ok).
func CompleteFields(btn Btn, fields *finalembed.DynamicObject) *UiResponse {
	result := finalembed.NewDynamicObject()
	result.Set(DOFieldChoice.String(), finalembed.Int(btn))
	if fields != nil {
		result.Set(DOFieldFields.String(), fields)
	}
	return &UiResponse{Status: UiStatusOK, Result: result}
}

// TypeRegs returns host type registrations for this package (call from Init).
// Registered into the session's local module DC (not sealed core).
func TypeRegs() []finalembed.TypeReg {
	return []finalembed.TypeReg{
		finalembed.NewTypeReg[*Tui]("Tui", registerTui),
		finalembed.NewTypeReg[*Dialog]("Dialog", registerDialog),
	}
}
