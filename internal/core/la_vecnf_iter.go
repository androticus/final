package core

// VecNfIter walks VecNf cells in index order (Current → Float).
type VecNfIter struct {
	vec *VecNf
	i   int // last fetched index; -1 before first GetNext
}

func (it *VecNfIter) FinalType() *Type { return vecNfIterType }

func (it *VecNfIter) GetNext() bool {
	next := it.i + 1
	if next >= len(it.vec.data) {
		return false
	}
	it.i = next
	return true
}

func (it *VecNfIter) Current() any {
	return Float(it.vec.data[it.i])
}

func CreateRegisterVecNfIterType() {
	t := &Type{
		name:            "VecNfIter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[VecNfIter]("VecNfIter"),
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	vecNfIterType = t
}

var vecNfIterType *Type

func init() {
	CreateRegisterVecNfIterType()
}

var (
	_ Typed    = (*VecNfIter)(nil)
	_ Iterator = (*VecNfIter)(nil)
)
