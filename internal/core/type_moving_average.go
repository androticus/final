package core

import "sync"

// MovingAverage is an exponential moving average (EMA) parameterized by span.
// α = 2/(span+1). No variance — use Average for exact mean/variance.
// Thread-safe for updates; read via GetMeanAndN() when atomicity matters.
type MovingAverage struct {
	mu    sync.Mutex
	span  int
	alpha float64
	mean  float64
	n     uint64
}

// NewMovingAverage creates an EMA with the given span (effective window).
// span < 1 is clamped to 1 (α = 1, last sample only).
func NewMovingAverage(span int) *MovingAverage {
	if span < 1 {
		span = 1
	}
	return &MovingAverage{
		span:  span,
		alpha: 2.0 / float64(span+1),
	}
}

func (m *MovingAverage) FinalType() *Type { return movingAverageType }

func (m *MovingAverage) Span() int { return m.span }

// note: fetching a 64-bit float is atomic
func (m *MovingAverage) Mean() float64 { return m.mean }

// note: fetching an int is atomic
func (m *MovingAverage) N() int { return int(m.n) }

// Add incorporates one new sample.
func (m *MovingAverage) Add(x float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.n == 0 {
		m.mean = x
	} else {
		m.mean = m.alpha*x + (1-m.alpha)*m.mean
	}
	m.n++
}

func (m *MovingAverage) GetMeanAndN() (float64, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mean, int(m.n)
}

func (m *MovingAverage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mean = 0
	m.n = 0
}

func registerMovingAverageMembers(t *Type) {
	t.registerProperty("mean", "Float",
		func(inst Any) Any { return Float(inst.(*MovingAverage).Mean()) },
		nil,
	)
	t.registerProperty("n", "Int",
		func(inst Any) Any { return Int(inst.(*MovingAverage).N()) },
		nil,
	)
	t.registerProperty("span", "Int",
		func(inst Any) Any { return Int(inst.(*MovingAverage).Span()) },
		nil,
	)
	// ( value:n -- ) any numeric, converted to float64
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		x := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		ths.Obj.(*MovingAverage).Add(x)
	})
	t.registerNativeMethod("reset", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*MovingAverage).Reset()
	})
	// ( -- mean:f n:int )
	t.registerNativeMethod("getMeanAndN", func(f *RunFrame, ths *TypeObjPair) {
		mean, n := ths.Obj.(*MovingAverage).GetMeanAndN()
		f.RP.DataStack.Push(Float(mean))
		f.RP.DataStack.Push(Int(n))
	})
	// ( span:i -- ma:MovingAverage )
	t.registerNativeClassMethod("newWithSpan", func(f *RunFrame, _ *TypeObjPair) {
		span := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewMovingAverage(span))
	})
}

func CreateRegisterMovingAverageType() {
	t := NewRegisteringType("MovingAverage",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[MovingAverage]("MovingAverage"),
	)
	registerMovingAverageMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	movingAverageType = t
}

func init() {
	CreateRegisterMovingAverageType()
}

var (
	movingAverageType *Type
	_                 Typed = (*MovingAverage)(nil)
)
