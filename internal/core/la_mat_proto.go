package core

// ProtoMat is the prototype for matrix types (Mat2f, Mat2u8, Mat2i1, …).
// Catalog Type only — not instantiable. Holds shared Final methods
// for any matrix instance.
func CreateRegisterMatProto() {
	t := &Type{
		name: "ProtoMat",
		getZero: func() any {
			return 0 // dummy; no instance value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	matProtoType = t
}

var matProtoType *Type

func init() {
	CreateRegisterMatProto()
	CreateRegisterMat2Types() // parent = ProtoMat (SetParent; no deferred resolution)
	CreateRegisterMat2i1Type()
	CreateRegisterMat2i4Type()
}
