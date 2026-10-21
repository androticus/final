package core

import "fmt"

// VecNf is a dense realized-length float64 vector (Final type VecNf).
// Length is fixed at construction — not a growable List.
type VecNf struct {
	data []float64
}

func (v *VecNf) FinalType() *Type { return vecNfType }

func (v *VecNf) Len() int { return len(v.data) }

// NewVecNf allocates a zero-filled length-n vector. n must be > 0.
func NewVecNf(n int) *VecNf {
	if n <= 0 {
		panic(fmt.Errorf("VecNf.new: length must be > 0, got %d", n))
	}
	return &VecNf{data: make([]float64, n)}
}

func (v *VecNf) At(i int) Float {
	if i < 0 || i >= len(v.data) {
		panic(fmt.Errorf("VecNf: index %d out of range [0,%d)", i, len(v.data)))
	}
	return Float(v.data[i])
}

func (v *VecNf) Put(i int, val Float) {
	if i < 0 || i >= len(v.data) {
		panic(fmt.Errorf("VecNf: index %d out of range [0,%d)", i, len(v.data)))
	}
	v.data[i] = float64(val)
}

// GetIter returns a cell iterator (VecNf is Iterable).
func (v *VecNf) GetIter() Iterator {
	return &VecNfIter{vec: v, i: -1}
}

func registerVecNfMembers(t *Type) {
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*VecNf).Len()) },
		nil,
	)
	// Flat @()/!(); store casts to Float (no extra coerce).
	t.registerItemsData("Float",
		func(inst Any, idx Any) any {
			return inst.(*VecNf).At(int(idx.(Int)))
		},
		func(inst Any, idx Any, value Any) {
			inst.(*VecNf).Put(int(idx.(Int)), value.(Float))
		},
		nil, nil, nil,
	)
	// ( n:i -- v:VecNf )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(NewVecNf(n))
	})
}

func CreateRegisterVecNfType() {
	t := &Type{
		name:            "VecNf",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[VecNf]("VecNf"),
	}
	registerVecNfMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	t.SetParent(vecProtoType)
	vecNfType = t
}

var vecNfType *Type

var (
	_ Typed    = (*VecNf)(nil)
	_ Iterable = (*VecNf)(nil)
)
