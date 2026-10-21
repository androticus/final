package core

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// LA mat:: ops — shape-partitioned opcode tables.
// Subcode ordinals are append-only (git-friendly expansion).
// First wave (160.1): matvec · * · transpose · solve · + · - · scale · row · col · identity · copy.
// Stack: left-to-right source order, last arg TOS.
// Dense product / matvec / solve via gonum; element-wise and views in-package.

func popMat2f(ds *Stack[any], op string) *Mat2f {
	v := ds.Pop()
	m, ok := v.(*Mat2f)
	if !ok || m == nil {
		panic(fmt.Errorf("%s: want *Mat2f, got %T", op, v))
	}
	return m
}

func popInt(ds *Stack[any], op string) int {
	v := ds.Pop()
	i, ok := v.(Int)
	if !ok {
		panic(fmt.Errorf("%s: want Int, got %T", op, v))
	}
	return int(i)
}

func requireSameMatShape(op string, a, b *Mat2f) {
	if a.rows != b.rows || a.cols != b.cols {
		panic(fmt.Errorf("%s: shape mismatch %d×%d vs %d×%d", op, a.rows, a.cols, b.rows, b.cols))
	}
}

func newMat2fCopyOf(src *Mat2f) *Mat2f {
	out := NewMat2f(src.rows, src.cols)
	copy(out.data, src.data)
	return out
}

func matAsDense(m *Mat2f) *mat.Dense {
	return mat.NewDense(m.rows, m.cols, m.data)
}

func vecAsVecDense(v *VecNf) *mat.VecDense {
	return mat.NewVecDense(v.Len(), v.data)
}

func newMat2fFromDense(d *mat.Dense) *Mat2f {
	r, c := d.Dims()
	out := NewMat2f(r, c)
	// Dense stores row-major; RawMatrix().Data is the backing slice.
	copy(out.data, d.RawMatrix().Data)
	return out
}

func newVecNfFromVecDense(v *mat.VecDense) *VecNf {
	n := v.Len()
	out := NewVecNf(n)
	copy(out.data, v.RawVector().Data)
	return out
}

// --- OpMatVec ( A v -- … ) — v on TOS ----------------------------------------
// Wave: matvec, solve

type SCMatVec = int

const (
	SC_mat_matvec SCMatVec = iota
	SC_mat_solve

	SCMatVecCount
)

type MatVecFn func(ds *Stack[any])

var matVecFns = [SCMatVecCount]MatVecFn{
	SC_mat_matvec: runMatMatvec,
	SC_mat_solve:  runMatSolve,
}

func runMatMatvec(ds *Stack[any]) {
	v := popVecNf(ds, "mat::matvec")
	a := popMat2f(ds, "mat::matvec")
	if v.Len() != a.cols {
		panic(fmt.Errorf("mat::matvec: vec len %d != mat cols %d", v.Len(), a.cols))
	}
	var y mat.VecDense
	y.MulVec(matAsDense(a), vecAsVecDense(v))
	ds.Push(newVecNfFromVecDense(&y))
}

func runMatSolve(ds *Stack[any]) {
	b := popVecNf(ds, "mat::solve")
	a := popMat2f(ds, "mat::solve")
	if b.Len() != a.rows {
		panic(fmt.Errorf("mat::solve: rhs len %d != mat rows %d", b.Len(), a.rows))
	}
	// Copy A and b so gonum may overwrite during factorization.
	ac := newMat2fCopyOf(a)
	bc := newVecNfCopyOf(b)
	var x mat.VecDense
	if err := x.SolveVec(matAsDense(ac), vecAsVecDense(bc)); err != nil {
		panic(fmt.Errorf("mat::solve: %w", err))
	}
	ds.Push(newVecNfFromVecDense(&x))
}

// --- OpMatMat ( A B -- … ) — B on TOS ----------------------------------------
// Wave: *, +, -

type SCMatMat = int

const (
	SC_mat_mul SCMatMat = iota // matrix product
	SC_mat_add                 // element-wise
	SC_mat_sub

	SCMatMatCount
)

type MatMatFn func(ds *Stack[any])

var matMatFns = [SCMatMatCount]MatMatFn{
	SC_mat_mul: runMatMul,
	SC_mat_add: runMatAdd,
	SC_mat_sub: runMatSub,
}

func runMatMul(ds *Stack[any]) {
	b := popMat2f(ds, "mat::*")
	a := popMat2f(ds, "mat::*")
	if a.cols != b.rows {
		panic(fmt.Errorf("mat::*: inner dims %d vs %d", a.cols, b.rows))
	}
	var c mat.Dense
	c.Mul(matAsDense(a), matAsDense(b))
	ds.Push(newMat2fFromDense(&c))
}

func runMatAdd(ds *Stack[any]) {
	b := popMat2f(ds, "mat::+")
	a := popMat2f(ds, "mat::+")
	requireSameMatShape("mat::+", a, b)
	out := NewMat2f(a.rows, a.cols)
	for i := range out.data {
		out.data[i] = a.data[i] + b.data[i]
	}
	ds.Push(out)
}

func runMatSub(ds *Stack[any]) {
	b := popMat2f(ds, "mat::-")
	a := popMat2f(ds, "mat::-")
	requireSameMatShape("mat::-", a, b)
	out := NewMat2f(a.rows, a.cols)
	for i := range out.data {
		out.data[i] = a.data[i] - b.data[i]
	}
	ds.Push(out)
}

// --- OpMatUnary ( A -- … ) ---------------------------------------------------
// Wave: transpose, copy

type SCMatUnary = int

const (
	SC_mat_transpose SCMatUnary = iota
	SC_mat_copy

	SCMatUnaryCount
)

type MatUnaryFn func(ds *Stack[any])

var matUnaryFns = [SCMatUnaryCount]MatUnaryFn{
	SC_mat_transpose: runMatTranspose,
	SC_mat_copy:      runMatCopy,
}

func runMatTranspose(ds *Stack[any]) {
	a := popMat2f(ds, "mat::transpose")
	out := NewMat2f(a.cols, a.rows)
	for r := 0; r < a.rows; r++ {
		for c := 0; c < a.cols; c++ {
			out.data[c*out.cols+r] = a.data[r*a.cols+c]
		}
	}
	ds.Push(out)
}

func runMatCopy(ds *Stack[any]) {
	a := popMat2f(ds, "mat::copy")
	ds.Push(newMat2fCopyOf(a))
}

// --- OpMatScalar ( A s:f -- … ) — s on TOS -----------------------------------
// Wave: scale

type SCMatScalar = int

const (
	SC_mat_scale SCMatScalar = iota

	SCMatScalarCount
)

type MatScalarFn func(ds *Stack[any])

var matScalarFns = [SCMatScalarCount]MatScalarFn{
	SC_mat_scale: runMatScale,
}

func runMatScale(ds *Stack[any]) {
	s := float64(popFloat(ds, "mat::scale"))
	a := popMat2f(ds, "mat::scale")
	out := NewMat2f(a.rows, a.cols)
	for i := range out.data {
		out.data[i] = a.data[i] * s
	}
	ds.Push(out)
}

// --- OpMatIndex ( A i:i -- … ) — i on TOS ------------------------------------
// Wave: row, col

type SCMatIndex = int

const (
	SC_mat_row SCMatIndex = iota
	SC_mat_col

	SCMatIndexCount
)

type MatIndexFn func(ds *Stack[any])

var matIndexFns = [SCMatIndexCount]MatIndexFn{
	SC_mat_row: runMatRow,
	SC_mat_col: runMatCol,
}

func runMatRow(ds *Stack[any]) {
	r := popInt(ds, "mat::row")
	a := popMat2f(ds, "mat::row")
	if r < 0 || r >= a.rows {
		panic(fmt.Errorf("mat::row: row %d out of range [0,%d)", r, a.rows))
	}
	out := NewVecNf(a.cols)
	copy(out.data, a.data[r*a.cols:(r+1)*a.cols])
	ds.Push(out)
}

func runMatCol(ds *Stack[any]) {
	c := popInt(ds, "mat::col")
	a := popMat2f(ds, "mat::col")
	if c < 0 || c >= a.cols {
		panic(fmt.Errorf("mat::col: col %d out of range [0,%d)", c, a.cols))
	}
	out := NewVecNf(a.rows)
	for r := 0; r < a.rows; r++ {
		out.data[r] = a.data[r*a.cols+c]
	}
	ds.Push(out)
}

// --- OpMatFactory ------------------------------------------------------------
// Wave: identity ( n -- I )

type SCMatFactory = int

const (
	SC_mat_identity SCMatFactory = iota

	SCMatFactoryCount
)

type MatFactoryFn func(ds *Stack[any])

var matFactoryFns = [SCMatFactoryCount]MatFactoryFn{
	SC_mat_identity: runMatIdentity,
}

func runMatIdentity(ds *Stack[any]) {
	n := popInt(ds, "mat::identity")
	if n <= 0 {
		panic(fmt.Errorf("mat::identity: n must be > 0, got %d", n))
	}
	out := NewMat2f(n, n)
	for i := 0; i < n; i++ {
		out.data[i*n+i] = 1
	}
	ds.Push(out)
}

func registerMatOps(d *[]*Word) {
	// OpMatVec — wave: matvec, solve
	regSubcodeAxiom(d, "mat::matvec", OpMatVec, SC_mat_matvec, WordFlagAxiom)
	regSubcodeAxiom(d, "mat::solve", OpMatVec, SC_mat_solve, WordFlagAxiom)
	// OpMatMat — wave: *, +, -
	regSubcodeAxiom(d, "mat::*", OpMatMat, SC_mat_mul, WordFlagAxiom)
	regSubcodeAxiom(d, "mat::+", OpMatMat, SC_mat_add, WordFlagAxiom)
	regSubcodeAxiom(d, "mat::-", OpMatMat, SC_mat_sub, WordFlagAxiom)
	// OpMatUnary — wave: transpose, copy
	regSubcodeAxiom(d, "mat::transpose", OpMatUnary, SC_mat_transpose, WordFlagAxiom)
	regSubcodeAxiom(d, "mat::copy", OpMatUnary, SC_mat_copy, WordFlagAxiom)
	// OpMatScalar — wave: scale
	regSubcodeAxiom(d, "mat::scale", OpMatScalar, SC_mat_scale, WordFlagAxiom)
	// OpMatIndex — wave: row, col
	regSubcodeAxiom(d, "mat::row", OpMatIndex, SC_mat_row, WordFlagAxiom)
	regSubcodeAxiom(d, "mat::col", OpMatIndex, SC_mat_col, WordFlagAxiom)
	// OpMatFactory — wave: identity
	regSubcodeAxiom(d, "mat::identity", OpMatFactory, SC_mat_identity, WordFlagAxiom)
}