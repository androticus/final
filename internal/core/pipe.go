package core

import (
	"fmt"
	"sync/atomic"
	"time"
)

type pipeMode int

const (
	pipeModeVanilla pipeMode = iota
	pipeModeTicker
	pipeModeTimer
)

type Pipe struct {
	channel chan any
	closed  atomic.Bool
	mode    pipeMode
	ticker  *time.Ticker
	timer   *time.Timer
	done    chan struct{}
}

func NewPipe(capacity int) *Pipe {
	if capacity < 0 {
		panic("Pipe.NewPipe: capacity must be >= 0")
	}
	return &Pipe{channel: make(chan any, capacity)}
}

// NewTickerPipe creates an unarmed ticker-mode pipe (capacity 1).
// Arm with TickerStart / pipe::tickerStart.
func NewTickerPipe() *Pipe {
	return &Pipe{channel: make(chan any, 1), mode: pipeModeTicker}
}

// NewTimerPipe creates an unarmed timer-mode pipe (capacity 1).
// Arm with TimerAfter / TimerWhen / pipe::timerAfter / pipe::timerWhen.
func NewTimerPipe() *Pipe {
	return &Pipe{channel: make(chan any, 1), mode: pipeModeTimer}
}

func (p *Pipe) requireMode(want pipeMode, op string) {
	if p.mode != want {
		panic(fmt.Errorf("%s: incompatible Pipe mode", op))
	}
}

// TickerStart starts or restarts the ticker at interval.
func (p *Pipe) TickerStart(interval Interval) {
	p.requireMode(pipeModeTicker, "pipe::tickerStart")
	if p.ticker == nil {
		p.done = make(chan struct{})
		p.ticker = time.NewTicker(interval.Duration)
		go func() {
			c := p.ticker.C
			for {
				select {
				case t := <-c:
					p.channel <- t
				case <-p.done:
					return
				}
			}
		}()
	} else {
		p.ticker.Reset(interval.Duration)
	}
}

// TickerStop stops ticks; the pipe stays open. No-op if not started.
func (p *Pipe) TickerStop() {
	p.requireMode(pipeModeTicker, "pipe::tickerStop")
	if p.ticker != nil {
		p.ticker.Stop()
	}
}

// TimerAfter arms or re-arms a one-shot after interval.
func (p *Pipe) TimerAfter(interval Interval) {
	p.requireMode(pipeModeTimer, "pipe::timerAfter")
	if p.timer == nil {
		p.done = make(chan struct{})
		p.timer = time.NewTimer(interval.Duration)
		go func() {
			c := p.timer.C
			for {
				select {
				case t := <-c:
					p.channel <- t
				case <-p.done:
					return
				}
			}
		}()
	} else {
		p.timer.Reset(interval.Duration)
	}
}

// TimerWhen arms or re-arms a one-shot at when.
// If when is in the past, pushes time.Now() immediately.
func (p *Pipe) TimerWhen(when time.Time) {
	p.requireMode(pipeModeTimer, "pipe::timerWhen")
	d := time.Until(when)
	if d <= 0 {
		if p.timer != nil {
			p.timer.Stop()
		}
		p.channel <- time.Now()
		return
	}
	p.TimerAfter(Interval{Duration: d})
}

func (p *Pipe) FinalType() *Type {
	return pipeType
}

func (p *Pipe) GetIter() Iterator {
	return &PipeIterator{pipe: p, current: nil}
}

// Len returns the number of values in the pipe
func (p *Pipe) Len() int {
	return len(p.channel)
}

// Cap returns the capacity of the pipe
func (p *Pipe) Cap() int {
	return cap(p.channel)
}

// IsClosed returns true if the pipe is closed
func (p *Pipe) IsClosed() bool {
	return p.closed.Load()
}

// Push pushes a value onto the pipe
func (p *Pipe) Push(value any) {
	p.channel <- value
}

// Pull pulls a value from the pipe
// It blocks until a value is available or the pipe is closed
// Returns a value and true, or nil and false if the pipe is closed
func (p *Pipe) Pull() (any, bool) {
	value, ok := <-p.channel
	return value, ok
}

func (p *Pipe) Close() {
	if p.mode == pipeModeTicker || p.mode == pipeModeTimer {
		panic(fmt.Errorf("pipe::close: cannot close ticker/timer pipe"))
	}
	if p.closed.CompareAndSwap(false, true) {
		close(p.channel)
	}
}

type PipeIterator struct {
	pipe    *Pipe
	current any
}

// GetNext will block until a value is available or the pipe is closed
// Returns true if a value is available, false if the pipe is closed
func (p *PipeIterator) GetNext() bool {
	value, ok := p.pipe.Pull()
	if ok {
		p.current = value
		return true
	}
	return false
}

func (p *PipeIterator) Current() any {
	return p.current
}

func registerPipeMembers(t *Type) {
	// ( capacity:i -- p:Pipe )
	t.registerNativeClassMethod("newWithCapacity", func(f *RunFrame, _ *TypeObjPair) {
		cap := f.RP.DataStack.Pop().(Int)
		f.RP.DataStack.Push(NewPipe(int(cap)))
	})
	// ( -- p:Pipe )
	t.registerNativeClassMethod("newTicker", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTickerPipe())
	})
	// ( -- p:Pipe )
	t.registerNativeClassMethod("newTimer", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTimerPipe())
	})
}

func CreateRegisterPipeType() {
	t := &Type{
		name:            "Pipe",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Pipe]("Pipe"),
		restrict:        SR_Threads,
	}
	registerPipeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	pipeType = t
}

var pipeType *Type

func init() {
	CreateRegisterPipeType()
}

// Pipe stack ops share OpPipeOp; subcode selects the operation (pipe::new, >pipe, …).

type SCPipeOp = int

const (
	SC_pipe_new SCPipeOp = iota
	SC_pipe_cap
	SC_pipe_len
	SC_pipe_push  // >pipe
	SC_pipe_pull  // pipe>
	SC_pipe_pullQ // pipe>?
	SC_pipe_close
	SC_pipe_closedQ
	SC_pipe_tickerStart
	SC_pipe_tickerStop
	SC_pipe_timerAfter
	SC_pipe_timerWhen

	SCPipeOpCount
)

type PipeOpFn func(ds *Stack[any])

var pipeOpFns = [SCPipeOpCount]PipeOpFn{
	SC_pipe_new:         runPipeNew,
	SC_pipe_cap:         runPipeCap,
	SC_pipe_len:         runPipeLen,
	SC_pipe_push:        runPipePush,
	SC_pipe_pull:        runPipePull,
	SC_pipe_pullQ:       runPipePullQ,
	SC_pipe_close:       runPipeClose,
	SC_pipe_closedQ:     runPipeClosedQ,
	SC_pipe_tickerStart: runPipeTickerStart,
	SC_pipe_tickerStop:  runPipeTickerStop,
	SC_pipe_timerAfter:  runPipeTimerAfter,
	SC_pipe_timerWhen:   runPipeTimerWhen,
}

func registerPipeOps(d *[]*Word) {
	regSubcodeAxiom(d, "pipe::new", OpPipeOp, SC_pipe_new, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::cap", OpPipeOp, SC_pipe_cap, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::len", OpPipeOp, SC_pipe_len, WordFlagAxiom)
	regSubcodeAxiom(d, ">pipe", OpPipeOp, SC_pipe_push, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe>", OpPipeOp, SC_pipe_pull, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe>?", OpPipeOp, SC_pipe_pullQ, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::close", OpPipeOp, SC_pipe_close, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::closed?", OpPipeOp, SC_pipe_closedQ, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::tickerStart", OpPipeOp, SC_pipe_tickerStart, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::tickerStop", OpPipeOp, SC_pipe_tickerStop, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::timerAfter", OpPipeOp, SC_pipe_timerAfter, WordFlagAxiom)
	regSubcodeAxiom(d, "pipe::timerWhen", OpPipeOp, SC_pipe_timerWhen, WordFlagAxiom)
}

func popPipe(ds *Stack[any], op string) *Pipe {
	v := ds.Pop()
	p, ok := v.(*Pipe)
	if !ok || p == nil {
		panic(fmt.Errorf("%s: want *Pipe, got %T", op, v))
	}
	return p
}

func runPipeNew(ds *Stack[any]) {
	cap := int(ds.Pop().(Int))
	ds.Push(NewPipe(cap))
}

func runPipeCap(ds *Stack[any]) {
	p := popPipe(ds, "pipe::cap")
	ds.Push(Int(p.Cap()))
}

func runPipeLen(ds *Stack[any]) {
	p := popPipe(ds, "pipe::len")
	ds.Push(Int(p.Len()))
}

func runPipePush(ds *Stack[any]) {
	p := popPipe(ds, ">pipe")
	item := ds.Pop()
	p.Push(item)
}

func runPipePull(ds *Stack[any]) {
	p := popPipe(ds, "pipe>")
	v, ok := p.Pull()
	if !ok {
		panic(fmt.Errorf("pipe>: pipe is closed"))
	}
	ds.Push(v)
}

func runPipePullQ(ds *Stack[any]) {
	p := popPipe(ds, "pipe>?")
	v, ok := p.Pull()
	if ok {
		ds.Push(v)
		ds.Push(Bool(true))
	} else {
		ds.Push(nil)
		ds.Push(Bool(false))
	}
}

func runPipeClose(ds *Stack[any]) {
	popPipe(ds, "pipe::close").Close()
}

func runPipeClosedQ(ds *Stack[any]) {
	p := popPipe(ds, "pipe::closed?")
	ds.Push(Bool(p.IsClosed()))
}

func runPipeTickerStart(ds *Stack[any]) {
	p := popPipe(ds, "pipe::tickerStart")
	iv := ds.Pop().(Interval)
	p.TickerStart(iv)
}

func runPipeTickerStop(ds *Stack[any]) {
	popPipe(ds, "pipe::tickerStop").TickerStop()
}

func runPipeTimerAfter(ds *Stack[any]) {
	p := popPipe(ds, "pipe::timerAfter")
	iv := ds.Pop().(Interval)
	p.TimerAfter(iv)
}

func runPipeTimerWhen(ds *Stack[any]) {
	p := popPipe(ds, "pipe::timerWhen")
	when := ds.Pop().(time.Time)
	p.TimerWhen(when)
}
