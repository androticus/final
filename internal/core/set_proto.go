package core

// ProtoSet is the prototype for all Set types.
// Catalog Type only — not instantiable. Holds shared Final methods
// (e.g. ._renderJson) for any Set-like instance.
//
// Inheritors (registered from set.go init after the proto):
//   - Set
//   - StringSet
func CreateRegisterProtoSet() {
	t := &Type{
		name: "ProtoSet",
		getZero: func() any {
			return 0 // dummy; no instance value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	protoSetType = t
}

var protoSetType *Type
