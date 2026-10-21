package core

// ProtoMap is the prototype for all Map types.
// At runtime, it holds Final methods for things like
// ._renderJson and other such methods that work
// on any Map-like instance.
func CreateRegisterMapProto() {
	t := &Type{
		name: "ProtoMap",
		getZero: func() any {
			return 0 // dummy, no instance value
		},
	}
	GetTypeRegistrar().RegisterCurrentType(t)
	mapProtoType = t
}

var mapProtoType *Type

func init() {
	CreateRegisterMapProto()
}
