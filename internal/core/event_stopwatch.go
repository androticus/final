package core

import "time"

type Stopwatch struct {
	start   time.Time // aka Final DateTime
	elapsed time.Duration
	running bool
}

func NewStopwatch() *Stopwatch {
	return &Stopwatch{}
}

func (s *Stopwatch) Start() {
	s.elapsed = 0
	s.start = time.Now()
	s.running = true
}

func (s *Stopwatch) Stop() {
	if !s.running {
		return
	}
	s.elapsed = time.Since(s.start)
	s.running = false
}

func (s *Stopwatch) Elapsed() time.Duration {
	if s.running {
		return time.Since(s.start)
	}
	return s.elapsed
}

func (s *Stopwatch) FinalType() *Type { return stopwatchType }

func registerStopwatchMembers(t *Type) {
	// ( -- sw:Stopwatch )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewStopwatch())
	})
	// ( -- )
	t.registerNativeMethod("start", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Stopwatch).Start()
	})
	// ( -- )
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Stopwatch).Stop()
	})
	t.registerProperty("elapsed", "Interval",
		func(inst Any) Any { return Interval{Duration: inst.(*Stopwatch).Elapsed()} },
		nil,
	)
}

func CreateRegisterStopwatchType() {
	t := NewRegisteringType("Stopwatch",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Stopwatch]("Stopwatch"),
	)
	registerStopwatchMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	stopwatchType = t
}

func init() {
	CreateRegisterStopwatchType()
}

var (
	stopwatchType *Type
	_             Typed = (*Stopwatch)(nil)
)
