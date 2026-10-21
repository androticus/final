package core

// EventAgain mints one EventLoop Event per Again() call.
// Use for CPU-bound controllers that should yield between frames
// without a timer delay: handler does work, then .again^ to re-arm.
//
// Lifecycle:
//   new   — bound; not delivering until again
//   again — enqueue one Event (first kick or re-arm); no-op if closed
//   close — permanent; further again are no-ops
//
// Wire Events target #_onFire (system). The user MethodRef is invoked with
// an empty stack ( -- ). Authors never see Event.
type EventAgain struct {
	userHandler *MethodRef
	mintFn      HandlerMintEventFn
	closed      bool
}

// NewEventAgain binds userHandler. Does not deliver until Again.
func NewEventAgain(userHandler *MethodRef) *EventAgain {
	ev := &EventAgain{userHandler: userHandler}
	el := GetEventLoop()
	ev.mintFn = el.CreateEventFn(NewMethodRef(ev, "_onFire"), 0)
	return ev
}

func (ev *EventAgain) FinalType() *Type { return eventAgainType }

// Again enqueues one Event onto the EventLoop pipe. No-op if closed.
func (ev *EventAgain) Again() {
	if ev == nil || ev.closed {
		return
	}
	if methodRefForgotten(ev.userHandler) {
		ev.Close()
		return
	}
	ev.mintFn(ev)
}

// Close permanently disables Again. Idempotent.
func (ev *EventAgain) Close() {
	if ev == nil || ev.closed {
		return
	}
	ev.closed = true
}

func registerEventAgainMembers(t *Type) {
	// ( handler:MethodRef -- ea:EventAgain )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		handler := f.RP.DataStack.Pop().(*MethodRef)
		f.RP.DataStack.Push(NewEventAgain(handler))
	})
	// ( -- )
	t.registerNativeMethod("again", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventAgain).Again()
	})
	// ( -- )
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventAgain).Close()
	})
	// System: EventLoop dispatch. Drops Event; invokes user MethodRef ( -- ).
	t.registerNativeMethod("_onFire", func(f *RunFrame, ths *TypeObjPair) {
		_ = f.RP.DataStack.Pop().(*Event)
		ev := ths.Obj.(*EventAgain)
		if ev.closed || methodRefForgotten(ev.userHandler) {
			return
		}
		ev.userHandler.Invoke(f)
	})
}

func CreateRegisterEventAgainType() {
	t := NewRegisteringType("EventAgain",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventAgain]("EventAgain"),
	)
	registerEventAgainMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventAgainType = t
}

func init() {
	CreateRegisterEventAgainType()
}

var (
	eventAgainType *Type
	_              Typed = (*EventAgain)(nil)
)
