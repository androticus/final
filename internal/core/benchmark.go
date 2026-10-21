package core

import (
	"fmt"
	"runtime"
	"time"
)

type BenchResults struct {
	Name       string // from Benchmark.Name
	Duration   time.Duration
	MemDeltaMB float64 // difference in memory usage between start and end
	Allocs     uint64  // difference in allocations between start and end
	NumGC      uint32  // number of garbage collections between start and end
}

func (r *BenchResults) FinalType() *Type {
	return benchResultsType
}

type Benchmark struct {
	Name          string // any text, for reporting
	before, after runtime.MemStats

	start time.Time
	end   time.Time
}

func NewBenchmark() *Benchmark {
	return &Benchmark{}
}

func (b *Benchmark) FinalType() *Type {
	return benchmarkType
}

func (b *Benchmark) Start() {
	runtime.ReadMemStats(&b.before)
	b.start = time.Now()
}

func (b *Benchmark) End() *BenchResults {
	b.end = time.Now()
	runtime.ReadMemStats(&b.after)
	return &BenchResults{
		Name:       b.Name,
		Duration:   time.Since(b.start),
		MemDeltaMB: float64(b.after.Alloc-b.before.Alloc) / (1024 * 1024),
		Allocs:     b.after.Mallocs - b.before.Mallocs,
		NumGC:      b.after.NumGC - b.before.NumGC,
	}
}

func registerBenchmarkMembers(t *Type) {
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*Benchmark).Name) },
		func(inst Any, value Any) { inst.(*Benchmark).Name = value.(string) },
	)
	t.registerNativeMethod("start", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Benchmark).Start()
	})
	t.registerNativeMethod("end", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Benchmark).End())
	})
}

func registerBenchResultsMembers(t *Type) {
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*BenchResults).Name) },
		nil,
	)
	t.registerProperty("duration", "Interval",
		func(inst Any) Any { return Interval{inst.(*BenchResults).Duration} },
		nil,
	)
	t.registerProperty("memDeltaMB", "Float",
		func(inst Any) Any { return Float(inst.(*BenchResults).MemDeltaMB) },
		nil,
	)
	t.registerProperty("allocs", "Int",
		func(inst Any) Any { return Int(inst.(*BenchResults).Allocs) },
		nil,
	)
	t.registerProperty("numGC", "Int",
		func(inst Any) Any { return Int(inst.(*BenchResults).NumGC) },
		nil,
	)
	t.registerNativeMethod("fmt", func(f *RunFrame, ths *TypeObjPair) {
		bm := ths.Obj.(*BenchResults)
		f.RP.DataStack.Push(String(fmt.Sprintf("BenchResults: name: %s, duration: %s, memDeltaMB: %f, allocs: %d, numGC: %d)", bm.Name, bm.Duration, bm.MemDeltaMB, bm.Allocs, bm.NumGC)))
	})
}

func CreateRegisterBenchmarkType() {
	t := &Type{
		name:            "Benchmark",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Benchmark]("Benchmark"),
	}
	registerBenchmarkMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewBenchmark())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	benchmarkType = t
}

func CreateRegisterBenchResultsType() {
	t := &Type{
		name:            "BenchResults",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[BenchResults]("BenchResults"),
	}
	registerBenchResultsMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	benchResultsType = t
}

var (
	benchmarkType    *Type
	benchResultsType *Type
)

func init() {
	CreateRegisterBenchmarkType()
	CreateRegisterBenchResultsType()
}

var (
	_ Typed = (*Benchmark)(nil)
	_ Typed = (*BenchResults)(nil)
)
