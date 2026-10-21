package core

import (
	"strconv"
)

type Counter struct {
	value Int
}

func NewCounter() *Counter {
	return &Counter{value: 0}
}

func (a *Counter) Value() Int {
	return a.value
}

func (c *Counter) Set(value Int) {
	c.value = value
}

func (c *Counter) Add(value Int) {
	c.value += value
}

func (c *Counter) Sub(value Int) {
	c.value -= value
}

func (c *Counter) Dec() {
	c.value--
}

func (c *Counter) Inc() {
	c.value++
}

func (c *Counter) FinalType() *Type {
	return counterType
}

// NaturalString formats for $ / ds$ / fault cells (same decimal text as Int / String identity).
func (c *Counter) NaturalString() string {
	if c == nil {
		return "Counter(nil)"
	}
	return strconv.Itoa(int(c.value))
}

func CreateRegisterCounterType() {
	t := &Type{
		name:            "Counter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Counter]("Counter"),
	}
	t.registerNativeMethod("_toString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*Counter).NaturalString()))
	})
	t.registerNativeMethod("_renderJson", nativeRenderJson)
	t.registerProperty("value", "Int",
		func(inst Any) Any { return inst.(*Counter).Value() },
		nil,
	)
	t.registerNativeMethod("inc", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Counter).Inc()
	})
	t.registerNativeMethod("dec", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Counter).Dec()
	})
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		ths.Obj.(*Counter).Add(n)
	})
	t.registerNativeMethod("sub", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		ths.Obj.(*Counter).Sub(n)
	})
	t.registerNativeMethod("set", func(f *RunFrame, ths *TypeObjPair) {
		n := f.RP.DataStack.Pop().(Int)
		ths.Obj.(*Counter).Set(n)
	})
	t.registerNativeMethod("zero", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Counter).Set(0)
	})
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewCounter())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	counterType = t
}

var counterType *Type
var _ NaturalStringer = (*Counter)(nil)

func init() {
	CreateRegisterCounterType()
}
