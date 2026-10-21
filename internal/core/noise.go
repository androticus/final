package core

import (
	"math/rand"

	"github.com/aquilax/go-perlin"
)

const (
	noiseAlpha = 2.0
	noiseBeta  = 2.0
	noiseN     = 3
)

// Noise is a Perlin noise generator. Nearby coordinates yield related values.
type Noise struct {
	*perlin.Perlin
}

func NewNoise() *Noise {
	return NewNoiseWithSeed(rand.Int63())
}

func NewNoiseWithSeed(seed int64) *Noise {
	return &Noise{Perlin: perlin.NewPerlin(noiseAlpha, noiseBeta, noiseN, seed)}
}

func (n *Noise) FinalType() *Type {
	return noiseType
}

func registerNoiseMembers(t *Type) {
	// ( -- n:Noise )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewNoise())
	})
	// ( seed:i -- n:Noise )
	t.registerNativeClassMethod("newWithSeed", func(f *RunFrame, _ *TypeObjPair) {
		seed := int64(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewNoiseWithSeed(seed))
	})
	// ( x:f -- v:f )
	t.registerNativeMethod("at1", func(f *RunFrame, ths *TypeObjPair) {
		x := float64(f.RP.DataStack.Pop().(Float))
		f.RP.DataStack.Push(Float(ths.Obj.(*Noise).Noise1D(x)))
	})
	// ( x:f y:f -- v:f )
	t.registerNativeMethod("at2", func(f *RunFrame, ths *TypeObjPair) {
		y := float64(f.RP.DataStack.Pop().(Float))
		x := float64(f.RP.DataStack.Pop().(Float))
		f.RP.DataStack.Push(Float(ths.Obj.(*Noise).Noise2D(x, y)))
	})
	// ( x:f y:f z:f -- v:f )
	t.registerNativeMethod("at3", func(f *RunFrame, ths *TypeObjPair) {
		z := float64(f.RP.DataStack.Pop().(Float))
		y := float64(f.RP.DataStack.Pop().(Float))
		x := float64(f.RP.DataStack.Pop().(Float))
		f.RP.DataStack.Push(Float(ths.Obj.(*Noise).Noise3D(x, y, z)))
	})
}

func CreateRegisterNoiseType() {
	t := NewRegisteringType("Noise",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Noise]("Noise"),
	)
	registerNoiseMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	noiseType = t
}

var noiseType *Type

func init() {
	CreateRegisterNoiseType()
}
