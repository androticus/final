package core

func (sess *Session) FinalType() *Type {
	return sessionType
}

// NOTE: initConfig is a Final method
func registerSessionMembers(t *Type) {
	t.registerProperty("configJson", "String",
		func(inst Any) Any { return (inst.(*Session).configJson) },
		func(inst Any, value Any) { inst.(*Session).configJson = value.(string) },
	)
	t.registerProperty("config", "DynamicObject",
		func(inst Any) Any { return (inst.(*Session).config) },
		nil,
	)
	t.registerProperty("logger", "Logger",
		func(inst Any) Any { return (inst.(*Session).logger) },
		nil,
	)

	t.registerNativeMethod("setConfig", func(f *RunFrame, ths *TypeObjPair) {
		config := (f.RP.DataStack.Pop().(*DynamicObject))
		ths.Obj.(*Session).config = config
	})
}

func CreateRegisterSessionType() {
	t := &Type{
		name:            "Session",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Session]("Session"),
	}
	registerSessionMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	sessionType = t
}

var sessionType *Type

func init() {
	CreateRegisterSessionType()
}
