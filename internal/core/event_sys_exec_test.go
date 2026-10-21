package core

import (
	"strings"
	"testing"
	"time"
)

func TestEventSysExecTypeRegistered(t *testing.T) {
	InitTestRuntime()
	typ := GetTypeRegistrar().Get("EventSysExec")
	if typ == nil || typ != eventSysExecType {
		t.Fatal("EventSysExec type not registered")
	}
	mm := typ.GetClassMethodMetadata("new")
	if mm == nil {
		t.Fatal("EventSysExec missing #new")
	}
	if mm.restrict != SR_System {
		t.Fatalf("EventSysExec.new restrict = %#x, want SR_System", mm.restrict)
	}
	if typ.GetMethodMetadata("_onComplete") == nil {
		t.Fatal("EventSysExec missing #_onComplete")
	}
	if typ.GetMethodMetadata("close") == nil {
		t.Fatal("EventSysExec missing #close")
	}
}

func waitEventSysExecEvent(t *testing.T, timeout time.Duration) *Event {
	t.Helper()
	select {
	case v := <-TheEventLoop.Chan():
		ev, ok := v.(*Event)
		if !ok {
			t.Fatalf("got %T, want *Event", v)
		}
		return ev
	case <-time.After(timeout):
		t.Fatal("timeout waiting for EventSysExec Event")
		return nil
	}
}

func TestEventSysExecSuccessMints(t *testing.T) {
	InitTestRuntime()
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	mm := &MethodMetadata{name: "h"}
	handler := &MethodRef{Parent: NewTypeObjPair(sys), MM: mm}
	ese := NewEventSysExec("echo", []string{"hello-async"}, 0, handler)
	defer ese.Close()

	ev := waitEventSysExecEvent(t, 3*time.Second)
	if ev.Source != ese {
		t.Fatalf("source = %v, want EventSysExec", ev.Source)
	}
	if ev.Params == nil || ev.Params.Len() != 3 {
		t.Fatalf("params len = %v, want 3", ev.Params)
	}
	out := string(ev.Params.At(0).(String))
	code := int(ev.Params.At(1).(Int))
	ok := bool(ev.Params.At(2).(Bool))
	if !ok || code != 0 || !strings.Contains(out, "hello-async") {
		t.Fatalf("params = out=%q code=%d ok=%v", out, code, ok)
	}
}

func TestEventSysExecOnCompleteDemux(t *testing.T) {
	InitTestRuntime()
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	var gotOut, gotErr string
	var gotCode int
	var gotOk bool
	userMM := &MethodMetadata{
		name: "handleExec",
		call: func(f *RunFrame, _ *TypeObjPair) {
			gotOk = bool(f.RP.DataStack.Pop().(Bool))
			if gotOk {
				gotCode = int(f.RP.DataStack.Pop().(Int))
				gotOut = string(f.RP.DataStack.Pop().(String))
				return
			}
			gotErr = string(f.RP.DataStack.Pop().(String))
			gotOut = string(f.RP.DataStack.Pop().(String))
		},
	}
	handler := &MethodRef{Parent: NewTypeObjPair(sys), MM: userMM}
	ese := NewEventSysExec("echo", []string{"demux-ok"}, 0, handler)
	defer ese.Close()

	wire := waitEventSysExecEvent(t, 3*time.Second)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, ese)
	rp.DataStack.Push(wire)
	eventSysExecType.GetMethodMetadata("_onComplete").invokeCall(f, ths)

	if !gotOk || gotCode != 0 || !strings.Contains(gotOut, "demux-ok") {
		t.Fatalf("demux got ok=%v code=%d out=%q err=%q", gotOk, gotCode, gotOut, gotErr)
	}
}

func TestEventSysExecTimeout(t *testing.T) {
	InitTestRuntime()
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	mm := &MethodMetadata{name: "h"}
	handler := &MethodRef{Parent: NewTypeObjPair(sys), MM: mm}
	// Single process — see TestSysExecQTimeout.
	ese := NewEventSysExec("python3", []string{
		"-c", "import sys,time; sys.stdout.write('partial-async'); sys.stdout.flush(); time.sleep(5)",
	}, 1, handler)
	defer ese.Close()

	ev := waitEventSysExecEvent(t, 4*time.Second)
	out := string(ev.Params.At(0).(String))
	errStr := string(ev.Params.At(1).(String))
	ok := bool(ev.Params.At(2).(Bool))
	if ok {
		t.Fatal("timeout should be modal false")
	}
	if errStr != sysExecTimeoutErr {
		t.Fatalf("err = %q, want %q", errStr, sysExecTimeoutErr)
	}
	if !strings.Contains(out, "partial-async") {
		t.Fatalf("out = %q, want to contain partial-async", out)
	}
}

func TestEventSysExecCloseSuppresses(t *testing.T) {
	InitTestRuntime()
	TheEventLoop = NewEventLoop(8)
	defer func() { TheEventLoop = nil }()

	mm := &MethodMetadata{name: "h"}
	handler := &MethodRef{Parent: NewTypeObjPair(sys), MM: mm}
	ese := NewEventSysExec("sleep", []string{"5"}, 0, handler)
	ese.Close()
	ese.Close() // idempotent

	time.Sleep(200 * time.Millisecond)
	select {
	case v := <-TheEventLoop.Chan():
		t.Fatalf("close must not mint, got %T", v)
	default:
	}
}

func TestEventSysExecNewRespectsSRSystem(t *testing.T) {
	InitTestRuntime()
	task := NewTestTask()
	task.restrict = SR_System
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	rp.DataStack.Push(String("echo"))
	rp.DataStack.Push(NewStringList())
	rp.DataStack.Push(Int(0))
	rp.DataStack.Push(&MethodRef{Parent: NewTypeObjPair(sys), MM: &MethodMetadata{name: "h"}})
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("recover = %v, want operation not allowed", r)
		}
	}()
	eventSysExecType.GetClassMethodMetadata("new").invokeCall(f, NewTypeObjPair(eventSysExecType))
	t.Fatal("expected SR_System panic")
}
