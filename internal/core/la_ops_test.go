package core

import (
	"math"
	"testing"
)

func TestVecOpsAddSubDotScaleNormCopy(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()

	u := NewVecNf(3)
	v := NewVecNf(3)
	u.data[0], u.data[1], u.data[2] = 1, 2, 3
	v.data[0], v.data[1], v.data[2] = 4, 5, 6

	rp.DataStack.Push(u)
	rp.DataStack.Push(v)
	RunOpsWithSubCode(task, rp, SC_vec_add, OpVecVec, OpStop)
	sum := rp.DataStack.Pop().(*VecNf)
	if sum.data[0] != 5 || sum.data[1] != 7 || sum.data[2] != 9 {
		t.Fatalf("vec::+ = %v", sum.data)
	}

	rp.DataStack.Push(u)
	rp.DataStack.Push(v)
	RunOpsWithSubCode(task, rp, SC_vec_sub, OpVecVec, OpStop)
	diff := rp.DataStack.Pop().(*VecNf)
	if diff.data[0] != -3 || diff.data[1] != -3 || diff.data[2] != -3 {
		t.Fatalf("vec::- = %v", diff.data)
	}

	rp.DataStack.Push(u)
	rp.DataStack.Push(v)
	RunOpsWithSubCode(task, rp, SC_vec_dot, OpVecVec, OpStop)
	dot := rp.DataStack.Pop().(Float)
	if dot != 32 { // 1*4+2*5+3*6
		t.Fatalf("vec::dot = %v, want 32", dot)
	}

	rp.DataStack.Push(u)
	rp.DataStack.Push(Float(2))
	RunOpsWithSubCode(task, rp, SC_vec_scale, OpVecScalar, OpStop)
	scaled := rp.DataStack.Pop().(*VecNf)
	if scaled.data[0] != 2 || scaled.data[1] != 4 || scaled.data[2] != 6 {
		t.Fatalf("vec::scale = %v", scaled.data)
	}

	rp.DataStack.Push(u)
	RunOpsWithSubCode(task, rp, SC_vec_norm, OpVecUnary, OpStop)
	n := float64(rp.DataStack.Pop().(Float))
	want := math.Sqrt(1 + 4 + 9)
	if math.Abs(n-want) > 1e-12 {
		t.Fatalf("vec::norm = %v, want %v", n, want)
	}

	rp.DataStack.Push(u)
	RunOpsWithSubCode(task, rp, SC_vec_copy, OpVecUnary, OpStop)
	cp := rp.DataStack.Pop().(*VecNf)
	cp.data[0] = 99
	if u.data[0] != 1 {
		t.Fatal("vec::copy must be deep")
	}
}

func TestVecOpsWordRegistration(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{"vec::+", "vec::-", "vec::dot", "vec::scale", "vec::norm", "vec::copy"} {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("missing word %q", name)
		}
	}
}

func TestMatOpsMatvecMulTransposeSolve(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()

	// A = [[1,2],[3,4]], v = [1,1] → [3,7]
	a := NewMat2f(2, 2)
	a.data = []float64{1, 2, 3, 4}
	v := NewVecNf(2)
	v.data[0], v.data[1] = 1, 1
	rp.DataStack.Push(a)
	rp.DataStack.Push(v)
	RunOpsWithSubCode(task, rp, SC_mat_matvec, OpMatVec, OpStop)
	y := rp.DataStack.Pop().(*VecNf)
	if y.data[0] != 3 || y.data[1] != 7 {
		t.Fatalf("mat::matvec = %v, want [3 7]", y.data)
	}

	b := NewMat2f(2, 2)
	b.data = []float64{0, 1, 0, 0} // [[0,1],[0,0]]
	rp.DataStack.Push(a)
	rp.DataStack.Push(b)
	RunOpsWithSubCode(task, rp, SC_mat_mul, OpMatMat, OpStop)
	prod := rp.DataStack.Pop().(*Mat2f)
	// [[1,2],[3,4]]*[[0,1],[0,0]] = [[0,1],[0,3]]
	if prod.data[0] != 0 || prod.data[1] != 1 || prod.data[2] != 0 || prod.data[3] != 3 {
		t.Fatalf("mat::* = %v", prod.data)
	}

	rp.DataStack.Push(a)
	RunOpsWithSubCode(task, rp, SC_mat_transpose, OpMatUnary, OpStop)
	at := rp.DataStack.Pop().(*Mat2f)
	if at.rows != 2 || at.cols != 2 || at.data[0] != 1 || at.data[1] != 3 || at.data[2] != 2 || at.data[3] != 4 {
		t.Fatalf("mat::transpose = %v shape %d×%d", at.data, at.rows, at.cols)
	}

	// Solve [[2,0],[0,2]] x = [4,6] → [2,3]
	s := NewMat2f(2, 2)
	s.data = []float64{2, 0, 0, 2}
	rhs := NewVecNf(2)
	rhs.data[0], rhs.data[1] = 4, 6
	rp.DataStack.Push(s)
	rp.DataStack.Push(rhs)
	RunOpsWithSubCode(task, rp, SC_mat_solve, OpMatVec, OpStop)
	x := rp.DataStack.Pop().(*VecNf)
	if math.Abs(x.data[0]-2) > 1e-12 || math.Abs(x.data[1]-3) > 1e-12 {
		t.Fatalf("mat::solve = %v, want [2 3]", x.data)
	}
}

func TestMatOpsAddSubScaleRowColIdentityCopy(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()

	a := NewMat2f(2, 2)
	a.data = []float64{1, 2, 3, 4}
	b := NewMat2f(2, 2)
	b.data = []float64{1, 1, 1, 1}

	rp.DataStack.Push(a)
	rp.DataStack.Push(b)
	RunOpsWithSubCode(task, rp, SC_mat_add, OpMatMat, OpStop)
	sum := rp.DataStack.Pop().(*Mat2f)
	if sum.data[0] != 2 || sum.data[3] != 5 {
		t.Fatalf("mat::+ = %v", sum.data)
	}

	rp.DataStack.Push(a)
	rp.DataStack.Push(b)
	RunOpsWithSubCode(task, rp, SC_mat_sub, OpMatMat, OpStop)
	diff := rp.DataStack.Pop().(*Mat2f)
	if diff.data[0] != 0 || diff.data[1] != 1 {
		t.Fatalf("mat::- = %v", diff.data)
	}

	rp.DataStack.Push(a)
	rp.DataStack.Push(Float(0.5))
	RunOpsWithSubCode(task, rp, SC_mat_scale, OpMatScalar, OpStop)
	sc := rp.DataStack.Pop().(*Mat2f)
	if sc.data[0] != 0.5 || sc.data[3] != 2 {
		t.Fatalf("mat::scale = %v", sc.data)
	}

	rp.DataStack.Push(a)
	rp.DataStack.Push(Int(1))
	RunOpsWithSubCode(task, rp, SC_mat_row, OpMatIndex, OpStop)
	row := rp.DataStack.Pop().(*VecNf)
	if row.Len() != 2 || row.data[0] != 3 || row.data[1] != 4 {
		t.Fatalf("mat::row = %v", row.data)
	}

	rp.DataStack.Push(a)
	rp.DataStack.Push(Int(0))
	RunOpsWithSubCode(task, rp, SC_mat_col, OpMatIndex, OpStop)
	col := rp.DataStack.Pop().(*VecNf)
	if col.Len() != 2 || col.data[0] != 1 || col.data[1] != 3 {
		t.Fatalf("mat::col = %v", col.data)
	}

	rp.DataStack.Push(Int(3))
	RunOpsWithSubCode(task, rp, SC_mat_identity, OpMatFactory, OpStop)
	id := rp.DataStack.Pop().(*Mat2f)
	if id.rows != 3 || id.data[0] != 1 || id.data[4] != 1 || id.data[8] != 1 || id.data[1] != 0 {
		t.Fatalf("mat::identity = %v", id.data)
	}

	rp.DataStack.Push(a)
	RunOpsWithSubCode(task, rp, SC_mat_copy, OpMatUnary, OpStop)
	cp := rp.DataStack.Pop().(*Mat2f)
	cp.data[0] = 99
	if a.data[0] != 1 {
		t.Fatal("mat::copy must be deep")
	}
}

func TestMatOpsWordRegistration(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{
		"mat::matvec", "mat::solve", "mat::*", "mat::+", "mat::-",
		"mat::transpose", "mat::copy", "mat::scale", "mat::row", "mat::col", "mat::identity",
	} {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("missing word %q", name)
		}
	}
}

func TestVecNfParentIsProtoVec(t *testing.T) {
	ResetTestRuntime()
	if vecNfType.GetParent() != vecProtoType {
		t.Fatalf("VecNf.parent = %v, want ProtoVec", vecNfType.GetParent())
	}
}

func TestMat2fParentIsProtoMat(t *testing.T) {
	ResetTestRuntime()
	if mat2fType.GetParent() != matProtoType {
		t.Fatalf("Mat2f.parent = %v, want ProtoMat", mat2fType.GetParent())
	}
}
