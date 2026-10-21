package core

import (
	"math"
	"sync"
)

// Average is an exact running arithmetic mean with variance.
// It is thread-safe for updates, and read via GetMeanAndN()
type Average struct {
	mu   sync.Mutex
	mean float64
	m2   float64
	n    uint64
}

func NewAverage() *Average { return &Average{} }

func (a *Average) FinalType() *Type { return averageType }

// Add incorporates one new sample.
func (a *Average) Add(x float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.n++
	delta := x - a.mean
	a.mean += delta / float64(a.n)
	a.m2 += delta * (x - a.mean)
}

// note: fetching a 64-bit float is atomic
func (a *Average) Mean() float64 { return a.mean }

// note:fetching an int is atomic
func (a *Average) N() int { return int(a.n) }

func (a *Average) Variance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.n < 1 {
		return 0
	}
	return a.m2 / float64(a.n)
}
func (a *Average) StandardDeviation() float64 {
	// note: do NOT use a lock here; Variance() does the locking
	return math.Sqrt(a.Variance())
}

func (a *Average) GetMeanAndN() (float64, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.mean, int(a.n)
}

func (a *Average) GetMeanStdDevAndN() (float64, float64, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.mean, a.StandardDeviation(), int(a.n)
}

func (a *Average) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.mean = 0
	a.m2 = 0
	a.n = 0
}

func registerAverageMembers(t *Type) {
	t.registerProperty("mean", "Float",
		func(inst Any) Any { return Float(inst.(*Average).Mean()) },
		nil,
	)
	t.registerProperty("stdDev", "Float",
		func(inst Any) Any { return Float(inst.(*Average).StandardDeviation()) },
		nil,
	)
	t.registerProperty("variance", "Float",
		func(inst Any) Any { return Float(inst.(*Average).Variance()) },
		nil,
	)
	t.registerProperty("n", "Int",
		func(inst Any) Any { return Int(inst.(*Average).N()) },
		nil,
	)
	// ( value:n -- ) any numeric, converted to float64
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		x := float64(checkCoerceNumericToFloat(f.RP.DataStack.Pop()))
		ths.Obj.(*Average).Add(x)
	})
	t.registerNativeMethod("reset", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Average).Reset()
	})
	// ( -- mean:f n:int )
	t.registerNativeMethod("getMeanAndN", func(f *RunFrame, ths *TypeObjPair) {
		mean, n := ths.Obj.(*Average).GetMeanAndN()
		f.RP.DataStack.Push(Float(mean))
		f.RP.DataStack.Push(Int(n))
	})
	// ( -- mean:f n:int )
	t.registerNativeMethod("getMeanStdDevAndN", func(f *RunFrame, ths *TypeObjPair) {
		mean, stddev, n := ths.Obj.(*Average).GetMeanStdDevAndN()
		f.RP.DataStack.Push(Float(mean))
		f.RP.DataStack.Push(Float(stddev))
		f.RP.DataStack.Push(Int(n))
	})
	// ( -- mean:Average )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewAverage())
	})
}

func CreateRegisterAverageType() {
	t := NewRegisteringType("Average",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Average]("Average"),
	)
	registerAverageMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	averageType = t
}

func init() {
	CreateRegisterAverageType()
}

var (
	averageType *Type
	_           Typed = (*Average)(nil)
)
