package core

import (
	"strconv"
	"sync/atomic"
)

type AtomicInt struct {
	value *atomic.Int64
}

func NewAtomicInt() *AtomicInt {
	return &AtomicInt{value: &atomic.Int64{}}
}

func (a *AtomicInt) Value() Int {
	return Int(a.value.Load())
}

func (a *AtomicInt) Set(value Int) {
	a.value.Store(int64(value))
}

func (a *AtomicInt) Add(value Int) Int {
	return Int(a.value.Add(int64(value)))
}

func (a *AtomicInt) Sub(value Int) Int {
	return Int(a.value.Add(-int64(value)))
}

func (a *AtomicInt) Dec() Int {
	return a.Add(-1)
}

func (a *AtomicInt) Inc() Int {
	return a.Add(1)
}

func (a *AtomicInt) FinalType() *Type {
	return atomicIntType
}

// NaturalString formats for $ / ds$ / fault cells (same decimal text as Int / String identity).
func (a *AtomicInt) NaturalString() string {
	if a == nil {
		return "AtomicInt(nil)"
	}
	return strconv.Itoa(int(a.Value()))
}

var _ NaturalStringer = (*AtomicInt)(nil)

func CreateRegisterAtomicIntType() {
	t := &Type{
		name:            "AtomicInt",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[AtomicInt]("AtomicInt"),
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*AtomicInt).NaturalString()))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	t.registerProperty("value", "Int",
		func(inst Any) Any { return inst.(*AtomicInt).Value() },
		nil,
	)
	t.registerNativeMethod("inc", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*AtomicInt).Inc())
	})
	t.registerNativeMethod("dec", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*AtomicInt).Dec())
	})
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		f.RP.DataStack.Push(ths.Obj.(*AtomicInt).Add(n))
	})
	t.registerNativeMethod("sub", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		f.RP.DataStack.Push(ths.Obj.(*AtomicInt).Sub(n))
	})
	t.registerNativeMethod("set", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		ths.Obj.(*AtomicInt).Set(n)
	})
	t.registerNativeMethod("zero", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*AtomicInt).Set(0)
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewAtomicInt())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	atomicIntType = t
}

var atomicIntType *Type

func init() {
	CreateRegisterAtomicIntType()
}
