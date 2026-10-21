package core

import (
	"testing"
	"time"
)

func TestPipeOpWordsRegistered(t *testing.T) {
	ResetTestRuntime()
	want := map[string]int{
		"pipe::new":         SC_pipe_new,
		"pipe::cap":         SC_pipe_cap,
		"pipe::len":         SC_pipe_len,
		">pipe":             SC_pipe_push,
		"pipe>":             SC_pipe_pull,
		"pipe>?":            SC_pipe_pullQ,
		"pipe::close":       SC_pipe_close,
		"pipe::closed?":     SC_pipe_closedQ,
		"pipe::tickerStart": SC_pipe_tickerStart,
		"pipe::tickerStop":  SC_pipe_tickerStop,
		"pipe::timerAfter":  SC_pipe_timerAfter,
		"pipe::timerWhen":   SC_pipe_timerWhen,
	}
	for name, sc := range want {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("missing word %q", name)
		}
		if w.opCode != OpPipeOp {
			t.Fatalf("%q opCode = %v, want OpPipeOp", name, w.opCode)
		}
		if w.subCode != sc {
			t.Fatalf("%q subCode = %d, want %d", name, w.subCode, sc)
		}
	}
}

func TestOpPipeNewCapLen(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(3))
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_new, OpPipeOp, OpStop)
	p, ok := rp.DataStack.Pop().(*Pipe)
	if !ok || p == nil {
		t.Fatalf("pipe::new = %T, want *Pipe", rp.DataStack.Peek())
	}
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_cap, OpPipeOp, OpStop)
	if got := rp.DataStack.Pop(); got != Int(3) {
		t.Fatalf("cap = %v, want 3", got)
	}
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_len, OpPipeOp, OpStop)
	if got := rp.DataStack.Pop(); got != Int(0) {
		t.Fatalf("len = %v, want 0", got)
	}
}

func TestOpPipeNewZeroCapacity(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(Int(0))
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_new, OpPipeOp, OpStop)
	p, ok := rp.DataStack.Pop().(*Pipe)
	if !ok || p == nil {
		t.Fatalf("pipe::new 0 = %T", rp.DataStack.Peek())
	}
	if p.Cap() != 0 {
		t.Fatalf("cap = %d, want 0", p.Cap())
	}
}

func TestOpPipePushPull(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewPipe(1)
	rp.DataStack.Push(String("hi"))
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_push, OpPipeOp, OpStop)
	if rp.DataStack.Len() != 0 {
		t.Fatalf("DS depth after >pipe = %d, want 0", rp.DataStack.Len())
	}
	if p.Len() != 1 {
		t.Fatalf("len after push = %d, want 1", p.Len())
	}
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_pull, OpPipeOp, OpStop)
	if got := rp.DataStack.Pop(); got != String("hi") {
		t.Fatalf("pipe> = %v, want hi", got)
	}
}

func TestOpPipePullQAndClose(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewPipe(2)
	p.Push(Int(7))
	p.Close()

	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_closedQ, OpPipeOp, OpStop)
	if got := rp.DataStack.Pop(); got != Bool(true) {
		t.Fatalf("closed? = %v, want true", got)
	}

	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_pullQ, OpPipeOp, OpStop)
	ok := rp.DataStack.Pop()
	v := rp.DataStack.Pop()
	if ok != Bool(true) || v != Int(7) {
		t.Fatalf("pipe>? residual = %v %v, want 7 true", v, ok)
	}

	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_pullQ, OpPipeOp, OpStop)
	ok = rp.DataStack.Pop()
	v = rp.DataStack.Pop()
	if ok != Bool(false) || v != nil {
		t.Fatalf("pipe>? after drain = %v %v, want nil false", v, ok)
	}
}

func TestOpPipePullClosedPanics(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewPipe(1)
	p.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected pipe> panic on closed empty pipe")
		}
	}()
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_pull, OpPipeOp, OpStop)
}

func TestOpPipePushAfterClosePanics(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewPipe(1)
	p.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected >pipe panic on closed pipe")
		}
	}()
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_push, OpPipeOp, OpStop)
}

func TestOpPipeCloseIdempotent(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewPipe(1)
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_close, OpPipeOp, OpStop)
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_close, OpPipeOp, OpStop)
	if !p.IsClosed() {
		t.Fatal("want closed after double close")
	}
}

func TestPipeNewTickerUnarmedThenStart(t *testing.T) {
	ResetTestRuntime()
	mm := pipeType.ClassMethodByName("newTicker")
	if mm == nil {
		t.Fatal("missing Pipe.newTicker")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pipeType)
	mm.invokeCall(f, ths)
	p, ok := rp.DataStack.Pop().(*Pipe)
	if !ok || p == nil || p.mode != pipeModeTicker {
		t.Fatalf("newTicker = %v mode=%v", p, p.mode)
	}
	if p.Len() != 0 {
		t.Fatal("unarmed ticker should not push")
	}
	p.TickerStart(Interval{Duration: time.Millisecond})
	got, ok := p.Pull()
	if !ok {
		t.Fatal("Ticker pipe unexpectedly closed")
	}
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("Ticker value = %T, want time.Time", got)
	}
	p.TickerStop()
}

func TestPipeNewTimerAfter(t *testing.T) {
	ResetTestRuntime()
	mm := pipeType.ClassMethodByName("newTimer")
	if mm == nil {
		t.Fatal("missing Pipe.newTimer")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, pipeType)
	mm.invokeCall(f, ths)
	p, ok := rp.DataStack.Pop().(*Pipe)
	if !ok || p == nil || p.mode != pipeModeTimer {
		t.Fatalf("newTimer = %v", p)
	}
	p.TimerAfter(Interval{Duration: time.Millisecond})
	got, ok := p.Pull()
	if !ok {
		t.Fatal("Timer pipe unexpectedly closed")
	}
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("Timer value = %T, want time.Time", got)
	}
	if p.IsClosed() {
		t.Fatal("timer pipe must not self-close after fire")
	}
}

func TestPipeTimerWhenPast(t *testing.T) {
	ResetTestRuntime()
	p := NewTimerPipe()
	p.TimerWhen(time.Now().Add(-time.Millisecond))
	got, ok := p.Pull()
	if !ok {
		t.Fatal("expected immediate value")
	}
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("got %T", got)
	}
}

func TestPipeCloseTickerPanics(t *testing.T) {
	ResetTestRuntime()
	p := NewTickerPipe()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected close panic on ticker pipe")
		}
	}()
	p.Close()
}

func TestPipeCloseTimerPanics(t *testing.T) {
	ResetTestRuntime()
	p := NewTimerPipe()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected close panic on timer pipe")
		}
	}()
	p.Close()
}

func TestPipeTickerOpsWrongModePanics(t *testing.T) {
	ResetTestRuntime()
	p := NewPipe(1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected tickerStart panic on vanilla pipe")
		}
	}()
	p.TickerStart(Interval{Duration: time.Millisecond})
}

func TestOpPipeTickerStartStop(t *testing.T) {
	ResetTestRuntime()
	rp := NewTestRunProvider()
	p := NewTickerPipe()
	rp.DataStack.Push(Interval{Duration: time.Millisecond})
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_tickerStart, OpPipeOp, OpStop)
	got, ok := p.Pull()
	if !ok {
		t.Fatal("no tick")
	}
	if _, ok := got.(time.Time); !ok {
		t.Fatalf("got %T", got)
	}
	rp.DataStack.Push(p)
	RunOpsWithSubCode(NewTestTask(), rp, SC_pipe_tickerStop, OpPipeOp, OpStop)
}

