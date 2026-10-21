package core

import (
	"math/rand"
)

// Rand is a pseudo type namespace for random number generation; it has no instances.

func registerRandMembers(t *Type) {
	t.registerNativeClassMethod("int", nativeRandInt)
	t.registerNativeClassMethod("float", nativeRandFloat)
	t.registerNativeClassMethod("normFloat", nativeRandNormFloat)
}

func CreateRegisterRandType() {
	t := &Type{
		name: "Rand",
		getZero: func() any {
			panic("Rand: no instances")
		},
	}
	registerRandMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

// ( -- n:Float )
func nativeRandFloat(f *RunFrame, _ *TypeObjPair) {
	f.RP.DataStack.Push(Float(rand.Float64()))
}

func nativeRandNormFloat(f *RunFrame, _ *TypeObjPair) {
	f.RP.DataStack.Push(Float(rand.NormFloat64()))
}
// ( n:Int -- result:Int ) generate a random integer between 0 and n-1
func nativeRandInt(f *RunFrame, _ *TypeObjPair) {
	n := f.RP.DataStack.Pop().(Int)
	f.RP.DataStack.Push(Int(rand.Intn(int(n))))
}

func init() {
	CreateRegisterRandType()
}
