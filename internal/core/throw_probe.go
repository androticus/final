package core

// ThrowProbe is a class-only exemplar for native Throw (161 Phase 2).
// Call from Final: ThrowProbe.boom^ / ThrowProbe.boomJoin^ inside try…catch.

func CreateRegisterThrowProbeType() {
	t := &Type{
		name: "ThrowProbe",
		getZero: func() any {
			panic("ThrowProbe: no instances")
		},
	}
	registerThrowProbeMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func registerThrowProbeMembers(t *Type) {
	// ( -- ) throws: "native boom"
	t.registerNativeClassMethod("boom", func(f *RunFrame, _ *TypeObjPair) {
		Throw("native boom", f)
	})
	// ( -- ) throws: outer with caused-by inner (nested Throw in one native)
	t.registerNativeClassMethod("boomJoin", func(f *RunFrame, _ *TypeObjPair) {
		Throw("inner detail", f)
		Throw("outer failure", f)
	})
}

func init() {
	CreateRegisterThrowProbeType()
}
