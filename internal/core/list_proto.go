package core

// ProtoList is the prototype for all List types.
// At runtime, it holds Final methods for things like
// ._renderJson and other such methods that work
// on any List-like instance with an Iterator.
func CreateRegisterListProto() {
	t := &Type{
		name: "ProtoList",
		getZero: func() any {
			return 0 // dummy, no instance value
		},
	}
	// dummy method, to put documentation on
	t.AddMethod("sort", &MethodMetadata{name: "sort"})
	GetTypeRegistrar().RegisterCurrentType(t)
	listProtoType = t
}

var listProtoType *Type

func init() {
//moved to list.go	CreateRegisterListProto()
}
