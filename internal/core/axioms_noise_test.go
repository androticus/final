package core

import (
	"math"
	"testing"
)

func TestNoiseNewWithSeedDeterministic(t *testing.T) {
	typ := GetTypeRegistrar().Get("Noise")
	if typ == nil {
		t.Fatal("Noise type not registered")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)

	rp.DataStack.Push(Int(42))
	typ.ClassMethodByName("newWithSeed").invokeCall(f, ths)
	n1 := rp.DataStack.Pop().(*Noise)

	rp.DataStack.Push(Int(42))
	typ.ClassMethodByName("newWithSeed").invokeCall(f, ths)
	n2 := rp.DataStack.Pop().(*Noise)

	a := n1.Noise2D(0.1, 0.2)
	b := n2.Noise2D(0.1, 0.2)
	if a != b {
		t.Fatalf("same seed: got %v and %v", a, b)
	}
}

func TestNoiseAtMethods(t *testing.T) {
	typ := GetTypeRegistrar().Get("Noise")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)

	rp.DataStack.Push(Int(7))
	typ.ClassMethodByName("newWithSeed").invokeCall(f, ths)
	n := rp.DataStack.Pop().(*Noise)
	thsN := NewTypeObjPair(n)

	rp.DataStack.Push(Float(0.5))
	typ.MethodByName("at1").invokeCall(f, thsN)
	v1 := float64(rp.DataStack.Pop().(Float))
	if math.IsNaN(v1) {
		t.Fatal("at1 returned NaN")
	}

	rp.DataStack.Push(Float(0.1))
	rp.DataStack.Push(Float(0.2))
	typ.MethodByName("at2").invokeCall(f, thsN)
	near := float64(rp.DataStack.Pop().(Float))

	rp.DataStack.Push(Float(0.11))
	rp.DataStack.Push(Float(0.21))
	typ.MethodByName("at2").invokeCall(f, thsN)
	near2 := float64(rp.DataStack.Pop().(Float))

	if math.Abs(near-near2) > 0.05 {
		t.Fatalf("nearby samples should be close: |%v-%v|=%v", near, near2, math.Abs(near-near2))
	}
	if near != n.Noise2D(0.1, 0.2) {
		t.Fatalf("at2 != Noise2D: %v vs %v", near, n.Noise2D(0.1, 0.2))
	}

	rp.DataStack.Push(Float(0.1))
	rp.DataStack.Push(Float(0.2))
	rp.DataStack.Push(Float(0.3))
	typ.MethodByName("at3").invokeCall(f, thsN)
	v3 := float64(rp.DataStack.Pop().(Float))
	if math.IsNaN(v3) {
		t.Fatal("at3 returned NaN")
	}
}

func TestNoiseNewClassMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Noise")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	typ.ClassMethodByName("new").invokeCall(f, ths)
	n := rp.DataStack.Pop().(*Noise)
	if n == nil || n.Perlin == nil {
		t.Fatal("Noise.new returned nil")
	}
}
