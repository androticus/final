package core

import (
	"strings"
	"testing"
	"time"
)

type nestInterpretHost struct{}

func (nestInterpretHost) FinalType() *Type { return nestInterpretHostType }

var nestInterpretHostType *Type

func installNestInterpretHostType(t *testing.T) *nestInterpretHost {
	t.Helper()
	h := &nestInterpretHost{}
	nestInterpretHostType = &Type{
		name: "NestInterpretHost",
		getZero: func() any {
			return nestInterpretHost{}
		},
		checkCoerceType: func(value any) any {
			return value
		},
	}
	nestInterpretHostType.RegisterNativeMethod("onFault", func(f *RunFrame, _ *TypeObjPair) {
		_ = f.RP.DataStack.Pop() // Event
		// Leave junk that would break EventLoop#run's o:.closed if we returned to foreach.
		f.RP.OStack.Push(Int(99))
		f.RP.XStack.Push(Int(88))
		f.RP.TempStack.Push(Int(77))
		_, err := InterpretSource(f.Task, f.RP, "1 >ts nosuchword")
		if err == nil {
			t.Error("InterpretSource: want fault for nosuchword")
			return
		}
		// Host would report here; escalate so RunEventLoopSession re-enters trampoline.
		AfterNestedInterpretFault()
	})
	nestInterpretHostType.RegisterNativeMethod("onClose", func(f *RunFrame, _ *TypeObjPair) {
		_ = f.RP.DataStack.Pop() // Event
		if f.RP.XStack.Len() != 0 || f.RP.TempStack.Len() != 0 || f.RP.YStack.Len() != 0 || f.RP.ZStack.Len() != 0 {
			panic("onClose: leftover value stacks after nested-interpret restart")
		}
		// During #run, EventLoop is on OS. Junk Int(99) must be gone.
		for i := 0; i < f.RP.OStack.Len(); i++ {
			if v, ok := f.RP.OStack.At(i).(Int); ok && int(v) == 99 {
				panic("onClose: leftover OS Int(99) after nested-interpret restart")
			}
		}
		GetEventLoop().Close()
	})
	return h
}

func TestAfterNestedInterpretFaultPanicsSentinel(t *testing.T) {
	defer func() {
		r := recover()
		if _, ok := r.(errEventLoopRestart); !ok {
			t.Fatalf("recover = %T %v, want errEventLoopRestart", r, r)
		}
	}()
	AfterNestedInterpretFault()
	t.Fatal("expected panic")
}

func TestRunEventLoopSessionRestartsAfterNestedInterpretFault(t *testing.T) {
	ResetTestRuntime()
	TheEventLoop = nil
	task := NewTestTask()
	rp := NewTestRunProvider()
	outSB := NewStringBuilder()
	rp.Stdout = outSB

	h := installNestInterpretHostType(t)
	if _, err := InterpretSource(task, rp, ": elRoot\n  EventLoop.inst.run^\n;\n"); err != nil {
		t.Fatalf("compile elRoot: %v", err)
	}

	el := GetEventLoop()
	faultFn := el.CreateEventFn(NewMethodRef(h, "onFault"), 0)
	closeFn := el.CreateEventFn(NewMethodRef(h, "onClose"), 0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunEventLoopSession(task, rp, "elRoot")
	}()

	faultFn(nil)
	// Allow restart to re-enter #run before close.
	time.Sleep(50 * time.Millisecond)
	closeFn(nil)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("RunEventLoopSession did not exit after Close")
	}

	out := outSB.ToString()
	// Host reports before escalate; outer must not WriteRunFaultReport for the sentinel.
	if strings.Count(out, "Fatal Error") > 0 || strings.Contains(out, "eventloop restart") {
		t.Fatalf("outer should not report sentinel restart; stdout=%q", out)
	}
}
