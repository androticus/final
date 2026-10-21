package core

import (
	"context"
	"sync"
)

// EventSysExec runs a tier-2 capture (merged stdout/stderr + exit code) off the
// EventLoop and demuxes completion to a user MethodRef with the same modal stack
// as Sys.exec?:
//
//	( out:s code:i true | out:s err:s false -- )
//
// Lifecycle:
//   new   — starts the process immediately
//   close — cancel wait; do not invoke the user handler (idempotent)
//
// Wire Events target EventSysExec#_onComplete (system); authors never see Event.
type EventSysExec struct {
	mu          sync.Mutex
	userHandler *MethodRef
	mintFn      HandlerMintEventFn
	cancel      context.CancelFunc
	closed      bool
}

// NewEventSysExec starts program with args under timeoutS (0 = no deadline).
// Completion is delivered via userHandler with the unified exec modal stack.
func NewEventSysExec(program string, args []string, timeoutS int, userHandler *MethodRef) *EventSysExec {
	ev := &EventSysExec{userHandler: userHandler}
	el := GetEventLoop()
	ev.mintFn = el.CreateEventFn(NewMethodRef(ev, "_onComplete"), 0)

	ctx, cancel := sysExecContext(timeoutS)
	ev.cancel = cancel

	go func() {
		out, code, err := runSysExecCombined(ctx, program, args)
		ev.mu.Lock()
		if ev.closed {
			ev.mu.Unlock()
			return
		}
		if methodRefForgotten(ev.userHandler) {
			ev.closed = true
			ev.mu.Unlock()
			return
		}
		mint := ev.mintFn
		ev.mu.Unlock()
		if err != nil {
			mint(ev, String(out), String(err.Error()), Bool(false))
			return
		}
		mint(ev, String(out), Int(code), Bool(true))
	}()
	return ev
}

func (ev *EventSysExec) FinalType() *Type { return eventSysExecType }

// Close cancels the wait and suppresses handler delivery. Idempotent.
func (ev *EventSysExec) Close() {
	ev.mu.Lock()
	if ev.closed {
		ev.mu.Unlock()
		return
	}
	ev.closed = true
	cancel := ev.cancel
	ev.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func registerEventSysExecMembers(t *Type) {
	// ( program:s args:StringList timeoutS:i handler:MethodRef -- e:EventSysExec )
	t.registerNativeClassMethodSec("new", func(f *RunFrame, _ *TypeObjPair) {
		ds := f.RP.DataStack
		handler := ds.Pop().(*MethodRef)
		timeoutS := int(ds.Pop().(Int))
		argsList := ds.Pop().(*StringList)
		program := string(ds.Pop().(String))
		args := make([]string, argsList.Len())
		for i := 0; i < argsList.Len(); i++ {
			args[i] = argsList.At(i)
		}
		ds.Push(NewEventSysExec(program, args, timeoutS, handler))
	}, SR_System)

	// ( -- )
	t.registerNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventSysExec).Close()
	})

	// System: EventLoop dispatch. Demuxes Event.params to user MethodRef.
	// ( e:Event -- ) o( inst:EventSysExec )
	t.registerNativeMethod("_onComplete", func(f *RunFrame, ths *TypeObjPair) {
		e := f.RP.DataStack.Pop().(*Event)
		ev := ths.Obj.(*EventSysExec)
		ev.mu.Lock()
		if ev.closed || methodRefForgotten(ev.userHandler) {
			ev.mu.Unlock()
			return
		}
		handler := ev.userHandler
		ev.mu.Unlock()

		params := e.Params
		if params == nil || params.Len() != 3 {
			panic("EventSysExec._onComplete: expected 3 params")
		}
		ds := f.RP.DataStack
		ds.Push(params.At(0))
		ds.Push(params.At(1))
		ds.Push(params.At(2))
		handler.Invoke(f)
	})
}

func CreateRegisterEventSysExecType() {
	t := NewRegisteringType("EventSysExec",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventSysExec]("EventSysExec"),
	)
	registerEventSysExecMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventSysExecType = t
}

func init() {
	CreateRegisterEventSysExecType()
}

var (
	eventSysExecType *Type
	_                Typed = (*EventSysExec)(nil)
)
