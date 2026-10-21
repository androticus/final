package core

import "time"

// EventTimer mints one-shot EventLoop Events. Wire Events target #_onFire (system);
// the user MethodRef is invoked with an empty stack ( -- ).
type EventTimer struct {
	userHandler *MethodRef
	mintFn      HandlerMintEventFn
	timer       *time.Timer
	closed      bool
	done        chan struct{}
}

// NewEventTimer binds userHandler. Does not fire until EventAfter / EventAt.
func NewEventTimer(userHandler *MethodRef) *EventTimer {
	ev := &EventTimer{userHandler: userHandler}
	el := GetEventLoop()
	ev.mintFn = el.CreateEventFn(NewMethodRef(ev, "_onFire"), 0)
	return ev
}

func (ev *EventTimer) FinalType() *Type { return eventTimerType }

func (ev *EventTimer) fire() {
	if ev == nil || ev.closed {
		return
	}
	if methodRefForgotten(ev.userHandler) {
		ev.Close()
		return
	}
	ev.mintFn(ev)
}

func (ev *EventTimer) EventAfter(interval Interval) {
	if ev.closed {
		return
	}
	if ev.timer == nil {
		ev.done = make(chan struct{})
		ev.timer = time.NewTimer(interval.Duration)
		go func() {
			c := ev.timer.C
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
		ev.timer.Reset(interval.Duration)
	}
}

// EventAt dispatches an event at the given time.
// If the time is in the past, the event is dispatched immediately.
func (ev *EventTimer) EventAt(when time.Time) {
	if ev.closed {
		return
	}
	duration := time.Until(when)
	if duration <= 0 {
		ev.fire()
		return
	}
	ev.EventAfter(Interval{Duration: duration})
}

func (ev *EventTimer) Stop() {
	if ev.timer != nil {
		ev.timer.Stop()
	}
}

func (ev *EventTimer) Close() {
	if ev.closed {
		return
	}
	ev.closed = true
	ev.Stop()
	if ev.done != nil {
		close(ev.done)
	}
	ev.timer = nil
}

func registerEventTimerMembers(t *Type) {
	// ( handler:MethodRef -- timer:EventTimer )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		handler := f.RP.DataStack.Pop().(*MethodRef)
		f.RP.DataStack.Push(NewEventTimer(handler))
	})
	// ( interval:Interval -- )
	t.registerNativeMethod("eventAfter", func(f *RunFrame, ths *TypeObjPair) {
		interval := f.RP.DataStack.Pop().(Interval)
		ths.Obj.(*EventTimer).EventAfter(interval)
	})
	// ( when:DateTime -- ) — Go time.Time under DateTime
	t.registerNativeMethod("eventAt", func(f *RunFrame, ths *TypeObjPair) {
		when := f.RP.DataStack.Pop().(time.Time)
		ths.Obj.(*EventTimer).EventAt(when)
	})
	// ( -- )
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventTimer).Stop()
	})
	// ( -- )
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventTimer).Close()
	})
	// System: EventLoop dispatch. Drops Event; invokes user MethodRef ( -- ).
	t.registerNativeMethod("_onFire", func(f *RunFrame, ths *TypeObjPair) {
		_ = f.RP.DataStack.Pop().(*Event)
		ev := ths.Obj.(*EventTimer)
		if ev.closed || methodRefForgotten(ev.userHandler) {
			return
		}
		ev.userHandler.Invoke(f)
	})
}

func CreateRegisterEventTimerType() {
	t := NewRegisteringType("EventTimer",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventTimer]("EventTimer"),
	)
	registerEventTimerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventTimerType = t
}

func init() {
	CreateRegisterEventTimerType()
}

var (
	eventTimerType *Type
	_              Typed = (*EventTimer)(nil)
)
