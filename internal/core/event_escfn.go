package core

import "sync"

// EventEscFn mints EventLoop Events when Escape / F1–F12 are latched by a headed
// host (TUI or GUI call LatchEscFn). Narrow by design — not a general keyboard API.
//
// Lifecycle (parity with EventTimer / EventTicker pause/resume):
//   new   — linked, not started (latches ignored)
//   start — accept presses
//   stop  — ignore presses (same as post-new); may start again
//   close — unlink permanently
//
// Wire Events target #_onEscFn (system). The user MethodRef receives ( code:i -- )
// where code is 0=Esc, 1..12=F1..F12. Authors never unpack Event.params.
//
// After forget:/type rollback, Final method slots are poisoned. deliver drops and
// unlinks subscribers whose MethodRef targets a forgotten type so EventLoop does
// not panic on orphaned controllers left open across reload.
type EventEscFn struct {
	mu          sync.Mutex
	userHandler *MethodRef
	mintFn      HandlerMintEventFn
	started     bool
	closed      bool
}

var (
	escFnSubMu sync.Mutex
	escFnSubs  []*EventEscFn
)

func linkEventEscFn(ev *EventEscFn) {
	escFnSubMu.Lock()
	defer escFnSubMu.Unlock()
	escFnSubs = append(escFnSubs, ev)
}

func unlinkEventEscFn(ev *EventEscFn) {
	escFnSubMu.Lock()
	defer escFnSubMu.Unlock()
	for i, s := range escFnSubs {
		if s == ev {
			escFnSubs = append(escFnSubs[:i], escFnSubs[i+1:]...)
			return
		}
	}
}

func notifyEventEscFnSubs(n int) {
	escFnSubMu.Lock()
	subs := append([]*EventEscFn(nil), escFnSubs...)
	escFnSubMu.Unlock()
	for _, s := range subs {
		s.deliver(n)
	}
}

func methodRefForgotten(mr *MethodRef) bool {
	return mr == nil || mr.MM == nil || mr.MM.forgotten
}

// NewEventEscFn creates a subscriber. Does not deliver until Start.
func NewEventEscFn(userHandler *MethodRef) *EventEscFn {
	ev := &EventEscFn{userHandler: userHandler}
	el := GetEventLoop()
	ev.mintFn = el.CreateEventFn(NewMethodRef(ev, "_onEscFn"), 0)
	linkEventEscFn(ev)
	return ev
}

func (ev *EventEscFn) FinalType() *Type { return eventEscFnType }

func (ev *EventEscFn) deliver(n int) {
	ev.mu.Lock()
	if ev.closed || !ev.started {
		ev.mu.Unlock()
		return
	}
	if methodRefForgotten(ev.userHandler) {
		ev.closed = true
		ev.started = false
		ev.mu.Unlock()
		unlinkEventEscFn(ev)
		return
	}
	mint := ev.mintFn
	ev.mu.Unlock()
	mint(ev, Int(n))
}

// Start begins delivering latched Esc/Fn presses as Events.
func (ev *EventEscFn) Start() {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if ev.closed {
		return
	}
	ev.started = true
}

// Stop ignores further presses (same as post-new). May Start again.
func (ev *EventEscFn) Stop() {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.started = false
}

// Close permanently unlinks this subscriber. Further Start/Stop are no-ops.
func (ev *EventEscFn) Close() {
	ev.mu.Lock()
	if ev.closed {
		ev.mu.Unlock()
		return
	}
	ev.closed = true
	ev.started = false
	ev.mu.Unlock()
	unlinkEventEscFn(ev)
}

func registerEventEscFnMembers(t *Type) {
	// ( handler:MethodRef -- esc:EventEscFn )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		handler := f.RP.DataStack.Pop().(*MethodRef)
		f.RP.DataStack.Push(NewEventEscFn(handler))
	})
	// ( -- )
	t.registerNativeMethod("start", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventEscFn).Start()
	})
	// ( -- )
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventEscFn).Stop()
	})
	// ( -- )
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventEscFn).Close()
	})
	// System: EventLoop dispatch. Demuxes params[0] code to user MethodRef.
	// ( e:Event -- ) o( inst:EventEscFn )
	t.registerNativeMethod("_onEscFn", func(f *RunFrame, ths *TypeObjPair) {
		e := f.RP.DataStack.Pop().(*Event)
		ev := ths.Obj.(*EventEscFn)
		ev.mu.Lock()
		if ev.closed || methodRefForgotten(ev.userHandler) {
			ev.mu.Unlock()
			return
		}
		handler := ev.userHandler
		ev.mu.Unlock()
		if e.Params == nil || e.Params.Len() < 1 {
			panic("EventEscFn._onEscFn: missing code param")
		}
		f.RP.DataStack.Push(e.Params.At(0))
		handler.Invoke(f)
	})
	// Global latch helpers (class methods; bare words in 0086-events.fnl).
	t.registerNativeClassMethod("clearEscFn", func(f *RunFrame, _ *TypeObjPair) {
		ClearEscFn()
	})
	t.registerNativeClassMethod("escFn?", func(f *RunFrame, _ *TypeObjPair) {
		if n, ok := TakeEscFn(); ok {
			f.RP.DataStack.Push(Int(n))
			f.RP.DataStack.Push(Bool(true))
			return
		}
		f.RP.DataStack.Push(Bool(false))
	})
}

func CreateRegisterEventEscFnType() {
	t := NewRegisteringType("EventEscFn",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventEscFn]("EventEscFn"),
	)
	registerEventEscFnMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventEscFnType = t
}

func init() {
	CreateRegisterEventEscFnType()
}

var (
	eventEscFnType *Type
	_              Typed = (*EventEscFn)(nil)
)
