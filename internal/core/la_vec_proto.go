package core

// ProtoVec is the prototype for vector types (VecNf, later Vec2f…).
// Catalog Type only — not instantiable. Holds shared Final methods
// (e.g. ._renderJson) for any Iterable float vector.
//
// Inheritors (registered from this file's init after the proto):
//   - VecNf
func CreateRegisterVecProto() {
	t := &Type{
		name: "ProtoVec",
		getZero: func() any {
			return 0 // dummy; no instance value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	vecProtoType = t
}

var vecProtoType *Type

func init() {
	CreateRegisterVecProto()
	CreateRegisterVecNfType() // parent = ProtoVec (SetParent; no deferred resolution)
}
