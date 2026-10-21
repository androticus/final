package core

import (
	"fmt"
	"math"
)

// LA vec:: ops — shape-partitioned opcode tables.
// Subcode ordinals are append-only (git-friendly expansion).
// First wave (160.1): + − scale ·dot ·norm ·copy.
// Stack: left-to-right source order, last arg TOS.

func popVecNf(ds *Stack[any], op string) *VecNf {
	v := ds.Pop()
	u, ok := v.(*VecNf)
	if !ok || u == nil {
		panic(fmt.Errorf("%s: want *VecNf, got %T", op, v))
	}
	return u
}

func popFloat(ds *Stack[any], op string) Float {
	v := ds.Pop()
	f, ok := v.(Float)
	if !ok {
		panic(fmt.Errorf("%s: want Float, got %T", op, v))
	}
	return f
}

func requireSameVecLen(op string, a, b *VecNf) {
	if a.Len() != b.Len() {
		panic(fmt.Errorf("%s: length mismatch %d vs %d", op, a.Len(), b.Len()))
	}
}

func newVecNfCopyOf(src *VecNf) *VecNf {
	out := NewVecNf(src.Len())
	copy(out.data, src.data)
	return out
}

// --- OpVecUnary ( v -- … ) ---------------------------------------------------

type SCVecUnary = int

const (
	SC_vec_norm SCVecUnary = iota // first wave
	SC_vec_copy

	SCVecUnaryCount
)

type VecUnaryFn func(ds *Stack[any])

var vecUnaryFns = [SCVecUnaryCount]VecUnaryFn{
	SC_vec_norm: runVecNorm,
	SC_vec_copy: runVecCopy,
}

func runVecNorm(ds *Stack[any]) {
	v := popVecNf(ds, "vec::norm")
	var sum float64
	for _, x := range v.data {
		sum += x * x
	}
	ds.Push(Float(math.Sqrt(sum)))
}

func runVecCopy(ds *Stack[any]) {
	v := popVecNf(ds, "vec::copy")
	ds.Push(newVecNfCopyOf(v))
}

// --- OpVecVec ( u v -- … ) — v on TOS ----------------------------------------

type SCVecVec = int

const (
	SC_vec_add SCVecVec = iota // first wave
	SC_vec_sub
	SC_vec_dot

	SCVecVecCount
)

type VecVecFn func(ds *Stack[any])

var vecVecFns = [SCVecVecCount]VecVecFn{
	SC_vec_add: runVecAdd,
	SC_vec_sub: runVecSub,
	SC_vec_dot: runVecDot,
}

func runVecAdd(ds *Stack[any]) {
	b := popVecNf(ds, "vec::+")
	a := popVecNf(ds, "vec::+")
	requireSameVecLen("vec::+", a, b)
	out := NewVecNf(a.Len())
	for i := range out.data {
		out.data[i] = a.data[i] + b.data[i]
	}
	ds.Push(out)
}

func runVecSub(ds *Stack[any]) {
	b := popVecNf(ds, "vec::-")
	a := popVecNf(ds, "vec::-")
	requireSameVecLen("vec::-", a, b)
	out := NewVecNf(a.Len())
	for i := range out.data {
		out.data[i] = a.data[i] - b.data[i]
	}
	ds.Push(out)
}

func runVecDot(ds *Stack[any]) {
	b := popVecNf(ds, "vec::dot")
	a := popVecNf(ds, "vec::dot")
	requireSameVecLen("vec::dot", a, b)
	var sum float64
	for i := range a.data {
		sum += a.data[i] * b.data[i]
	}
	ds.Push(Float(sum))
}

// --- OpVecScalar ( v s:f -- … ) — s on TOS -----------------------------------

type SCVecScalar = int

const (
	SC_vec_scale SCVecScalar = iota // first wave

	SCVecScalarCount
)

type VecScalarFn func(ds *Stack[any])

var vecScalarFns = [SCVecScalarCount]VecScalarFn{
	SC_vec_scale: runVecScale,
}

func runVecScale(ds *Stack[any]) {
	s := float64(popFloat(ds, "vec::scale"))
	v := popVecNf(ds, "vec::scale")
	out := NewVecNf(v.Len())
	for i := range out.data {
		out.data[i] = v.data[i] * s
	}
	ds.Push(out)
}

func registerVecOps(d *[]*Word) {
	// OpVecUnary — wave order: norm, copy
	regSubcodeAxiom(d, "vec::norm", OpVecUnary, SC_vec_norm, WordFlagAxiom)
	regSubcodeAxiom(d, "vec::copy", OpVecUnary, SC_vec_copy, WordFlagAxiom)
	// OpVecVec — wave order: +, -, dot
	regSubcodeAxiom(d, "vec::+", OpVecVec, SC_vec_add, WordFlagAxiom)
	regSubcodeAxiom(d, "vec::-", OpVecVec, SC_vec_sub, WordFlagAxiom)
	regSubcodeAxiom(d, "vec::dot", OpVecVec, SC_vec_dot, WordFlagAxiom)
	// OpVecScalar — wave order: scale
	regSubcodeAxiom(d, "vec::scale", OpVecScalar, SC_vec_scale, WordFlagAxiom)
}
