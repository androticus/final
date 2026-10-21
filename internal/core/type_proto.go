package core

import "fmt"

// protoTypeType is the default prototype root for all sealed types (ProtoType).
var protoTypeType *Type

// CreateRegisterProtoType registers ProtoType and its system class methods.
func CreateRegisterProtoType() {
	t := &Type{
		name: "ProtoType",
		getZero: func() any {
			panic("ProtoType: no zero value")
		},
		checkCoerceType: func(value any) any {
			typ, ok := value.(*Type)
			if !ok {
				panic(fmt.Errorf("ProtoType.checkCoerceType: expected *Type, got %T", value))
			}
			return typ
		},
	}
	// ( provider:a -- iter:Iterator true ) || ( -- provider false )
	t.registerNativeMethod("_tryGetIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(false)
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	t.parent = t
	protoTypeType = t
}

func init() {
	CreateRegisterProtoType()
}
