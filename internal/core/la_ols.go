package core

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// OLSModel is a fitted ordinary least-squares model (Final type OLSModel).
// X is n×p (usually with intercept column); y length n; β length p.
type OLSModel struct {
	beta *VecNf
	yHat *VecNf
	sse  float64
	sst  float64
	n    int
	p    int
}

func (m *OLSModel) FinalType() *Type { return olsModelType }

func (m *OLSModel) Beta() *VecNf { return m.beta }
func (m *OLSModel) YHat() *VecNf { return m.yHat }
func (m *OLSModel) SSE() float64 { return m.sse }
func (m *OLSModel) SST() float64 { return m.sst }
func (m *OLSModel) N() int       { return m.n }
func (m *OLSModel) P() int       { return m.p }

func (m *OLSModel) DF() int {
	return m.n - m.p
}

func (m *OLSModel) R2() float64 {
	if m.sst == 0 {
		return 0
	}
	r2 := 1 - m.sse/m.sst
	if r2 > -1e-12 && r2 < 0 {
		return 0
	}
	return r2
}

func (m *OLSModel) RMSE() float64 {
	df := m.DF()
	if df <= 0 {
		return math.NaN()
	}
	return math.Sqrt(m.sse / float64(df))
}

// FitOLS solves y ≈ X β (least squares via gonum SolveVec) and returns fitted state.
func FitOLS(y *VecNf, x *Mat2f) *OLSModel {
	if y == nil || x == nil {
		panic(fmt.Errorf("OLS.fit: nil y or X"))
	}
	n, p := x.rows, x.cols
	if y.Len() != n {
		panic(fmt.Errorf("OLS.fit: y len %d != X rows %d", y.Len(), n))
	}
	if n < p {
		panic(fmt.Errorf("OLS.fit: need n >= p, got %d×%d", n, p))
	}
	xc := newMat2fCopyOf(x)
	yc := newVecNfCopyOf(y)
	var beta mat.VecDense
	if err := beta.SolveVec(matAsDense(xc), vecAsVecDense(yc)); err != nil {
		panic(fmt.Errorf("OLS.fit: %w", err))
	}
	b := newVecNfFromVecDense(&beta)

	var yHat mat.VecDense
	yHat.MulVec(matAsDense(x), &beta)
	yh := newVecNfFromVecDense(&yHat)

	var sse float64
	var sum float64
	for i := 0; i < n; i++ {
		r := y.data[i] - yh.data[i]
		sse += r * r
		sum += y.data[i]
	}
	mean := sum / float64(n)
	var sst float64
	for i := 0; i < n; i++ {
		d := y.data[i] - mean
		sst += d * d
	}
	return &OLSModel{beta: b, yHat: yh, sse: sse, sst: sst, n: n, p: p}
}

// Predict applies β to Xnew (rows × p); returns length-rows yHat.
func (m *OLSModel) Predict(xNew *Mat2f) *VecNf {
	if xNew.cols != m.p {
		panic(fmt.Errorf("OLS.predict: X cols %d != p %d", xNew.cols, m.p))
	}
	var yHat mat.VecDense
	yHat.MulVec(matAsDense(xNew), vecAsVecDense(m.beta))
	return newVecNfFromVecDense(&yHat)
}

// ModelCompareResult is compact vs augmented OLS comparison (Judd PRE/F).
type ModelCompareResult struct {
	sseC float64
	sseA float64
	pre  float64
	f    float64
	df1  int
	df2  int
	n    int
	pC   int
	pA   int
}

func (r *ModelCompareResult) FinalType() *Type { return modelCompareResultType }

// CompareOLS computes PRE and F for nested models on the same y (pA > pC, same n).
func CompareOLS(compact, augmented *OLSModel) *ModelCompareResult {
	if compact == nil || augmented == nil {
		panic(fmt.Errorf("OLS.compare: nil model"))
	}
	if compact.n != augmented.n {
		panic(fmt.Errorf("OLS.compare: n mismatch %d vs %d", compact.n, augmented.n))
	}
	if augmented.p <= compact.p {
		panic(fmt.Errorf("OLS.compare: need pA > pC, got %d vs %d", augmented.p, compact.p))
	}
	df1 := augmented.p - compact.p
	df2 := augmented.n - augmented.p
	if df2 <= 0 {
		panic(fmt.Errorf("OLS.compare: df2=%d", df2))
	}
	sseC, sseA := compact.sse, augmented.sse
	var pre float64
	if sseC == 0 {
		pre = 0
	} else {
		pre = (sseC - sseA) / sseC
	}
	var fStat float64
	if pre >= 1 {
		fStat = math.Inf(1)
	} else if df1 == 0 {
		fStat = 0
	} else {
		fStat = (pre / float64(df1)) / ((1 - pre) / float64(df2))
	}
	return &ModelCompareResult{
		sseC: sseC, sseA: sseA, pre: pre, f: fStat,
		df1: df1, df2: df2, n: compact.n, pC: compact.p, pA: augmented.p,
	}
}

func registerOLSModelMembers(t *Type) {
	t.registerProperty("beta", "VecNf",
		func(inst Any) Any { return inst.(*OLSModel).beta },
		nil,
	)
	t.registerProperty("yHat", "VecNf",
		func(inst Any) Any { return inst.(*OLSModel).yHat },
		nil,
	)
	t.registerProperty("sse", "Float",
		func(inst Any) Any { return Float(inst.(*OLSModel).sse) },
		nil,
	)
	t.registerProperty("sst", "Float",
		func(inst Any) Any { return Float(inst.(*OLSModel).sst) },
		nil,
	)
	t.registerProperty("r2", "Float",
		func(inst Any) Any { return Float(inst.(*OLSModel).R2()) },
		nil,
	)
	t.registerProperty("rmse", "Float",
		func(inst Any) Any { return Float(inst.(*OLSModel).RMSE()) },
		nil,
	)
	t.registerProperty("n", "Int",
		func(inst Any) Any { return Int(inst.(*OLSModel).n) },
		nil,
	)
	t.registerProperty("p", "Int",
		func(inst Any) Any { return Int(inst.(*OLSModel).p) },
		nil,
	)
	t.registerProperty("df", "Int",
		func(inst Any) Any { return Int(inst.(*OLSModel).DF()) },
		nil,
	)
	// ( y:VecNf X:Mat2f -- model:OLSModel )
	t.registerNativeClassMethod("fit", func(f *RunFrame, _ *TypeObjPair) {
		x := popMat2f(f.RP.DataStack, "OLSModel.fit")
		y := popVecNf(f.RP.DataStack, "OLSModel.fit")
		f.RP.DataStack.Push(FitOLS(y, x))
	})
	// ( c:OLSModel a:OLSModel -- result:ModelCompareResult )
	t.registerNativeClassMethod("compare", func(f *RunFrame, _ *TypeObjPair) {
		a := f.RP.DataStack.Pop().(*OLSModel)
		c := f.RP.DataStack.Pop().(*OLSModel)
		f.RP.DataStack.Push(CompareOLS(c, a))
	})
	// ( Xnew:Mat2f -- yHat:VecNf )
	t.registerNativeMethod("predict", func(f *RunFrame, ths *TypeObjPair) {
		xNew := popMat2f(f.RP.DataStack, "OLSModel.predict")
		f.RP.DataStack.Push(ths.Obj.(*OLSModel).Predict(xNew))
	})
}

func registerModelCompareResultMembers(t *Type) {
	t.registerProperty("sseC", "Float",
		func(inst Any) Any { return Float(inst.(*ModelCompareResult).sseC) },
		nil,
	)
	t.registerProperty("sseA", "Float",
		func(inst Any) Any { return Float(inst.(*ModelCompareResult).sseA) },
		nil,
	)
	t.registerProperty("pre", "Float",
		func(inst Any) Any { return Float(inst.(*ModelCompareResult).pre) },
		nil,
	)
	t.registerProperty("f", "Float",
		func(inst Any) Any { return Float(inst.(*ModelCompareResult).f) },
		nil,
	)
	t.registerProperty("df1", "Int",
		func(inst Any) Any { return Int(inst.(*ModelCompareResult).df1) },
		nil,
	)
	t.registerProperty("df2", "Int",
		func(inst Any) Any { return Int(inst.(*ModelCompareResult).df2) },
		nil,
	)
	t.registerProperty("n", "Int",
		func(inst Any) Any { return Int(inst.(*ModelCompareResult).n) },
		nil,
	)
	t.registerProperty("pC", "Int",
		func(inst Any) Any { return Int(inst.(*ModelCompareResult).pC) },
		nil,
	)
	t.registerProperty("pA", "Int",
		func(inst Any) Any { return Int(inst.(*ModelCompareResult).pA) },
		nil,
	)
}

func CreateRegisterOLSModelType() {
	t := &Type{
		name:            "OLSModel",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[OLSModel]("OLSModel"),
	}
	registerOLSModelMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	olsModelType = t
}

func CreateRegisterModelCompareResultType() {
	t := &Type{
		name:            "ModelCompareResult",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[ModelCompareResult]("ModelCompareResult"),
	}
	registerModelCompareResultMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	modelCompareResultType = t
}

var (
	olsModelType           *Type
	modelCompareResultType *Type
)

func init() {
	CreateRegisterOLSModelType()
	CreateRegisterModelCompareResultType()
}

var (
	_ Typed = (*OLSModel)(nil)
	_ Typed = (*ModelCompareResult)(nil)
)
