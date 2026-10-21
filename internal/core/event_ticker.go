package core

import "time"

// EventTicker mints periodic EventLoop Events. Wire Events target #_onFire (system);
// the user MethodRef is invoked with an empty stack ( -- ).
type EventTicker struct {
	userHandler *MethodRef
	mintFn      HandlerMintEventFn
	ticker      *time.Ticker
	closed      bool
	done        chan struct{}
}

// NewEventTicker binds userHandler. Does not fire until EventEvery.
func NewEventTicker(userHandler *MethodRef) *EventTicker {
	ev := &EventTicker{userHandler: userHandler}
	el := GetEventLoop()
	ev.mintFn = el.CreateEventFn(NewMethodRef(ev, "_onFire"), 0)
	return ev
}

func (ev *EventTicker) FinalType() *Type { return eventTickerType }

func (ev *EventTicker) fire() {
	if ev == nil || ev.closed {
		return
	}
	if methodRefForgotten(ev.userHandler) {
		ev.Close()
		return
	}
	ev.mintFn(ev)
}

func (ev *EventTicker) EventEvery(interval Interval) {
	if ev.closed {
		return
	}
	if ev.ticker == nil {
		ev.done = make(chan struct{})
		ev.ticker = time.NewTicker(interval.Duration)
		go func() {
			c := ev.ticker.C
			for {
				select {
				case <-c:
					if ev.closed {
						return
					}
					ev.fire()
				case <-ev.done:
					return
				}
			}
		}()
	} else {
		ev.ticker.Reset(interval.Duration)
	}
}

func (ev *EventTicker) Stop() {
	if ev.ticker != nil {
		ev.ticker.Stop()
	}
}

func (ev *EventTicker) Close() {
	if ev.closed {
		return
	}
	ev.closed = true
	ev.Stop()
	if ev.done != nil {
		close(ev.done)
	}
	ev.ticker = nil
}

func registerEventTickerMembers(t *Type) {
	// ( handler:MethodRef -- ticker:EventTicker )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		handler := f.RP.DataStack.Pop().(*MethodRef)
		f.RP.DataStack.Push(NewEventTicker(handler))
	})
	// ( interval:Interval -- )
	t.registerNativeMethod("eventEvery", func(f *RunFrame, ths *TypeObjPair) {
		interval := f.RP.DataStack.Pop().(Interval)
		ths.Obj.(*EventTicker).EventEvery(interval)
	})
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventTicker).Stop()
	})
	// ( -- )
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventTicker).Close()
	})
	// System: EventLoop dispatch. Drops Event; invokes user MethodRef ( -- ).
	t.registerNativeMethod("_onFire", func(f *RunFrame, ths *TypeObjPair) {
		_ = f.RP.DataStack.Pop().(*Event)
		ev := ths.Obj.(*EventTicker)
		if ev.closed || methodRefForgotten(ev.userHandler) {
			return
		}
		ev.userHandler.Invoke(f)
	})
}

func CreateRegisterEventTickerType() {
	t := NewRegisteringType("EventTicker",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventTicker]("EventTicker"),
	)
	registerEventTickerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventTickerType = t
}

func init() {
	CreateRegisterEventTickerType()
}

var (
	eventTickerType *Type
	_               Typed = (*EventTicker)(nil)
)
