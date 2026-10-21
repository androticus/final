package core

// HttpAddress represents the target server address.
type HttpAddress struct {
	scheme string // "http" or "https". If empty, defaults to "https".
	host   string
	port   int // 0 means use the scheme's default port (80 for http, 443 for https)
}

func NewHttpAddress() *HttpAddress {
	return &HttpAddress{}
}

func (a *HttpAddress) FinalType() *Type {
	return httpAddressType
}

func registerHttpAddressMembers(t *Type) {
	t.registerProperty("scheme", "String",
		func(inst Any) Any { return String(inst.(*HttpAddress).scheme) },
		func(inst Any, value Any) { inst.(*HttpAddress).scheme = string(value.(String)) },
	)
	t.registerProperty("host", "String",
		func(inst Any) Any { return String(inst.(*HttpAddress).host) },
		func(inst Any, value Any) { inst.(*HttpAddress).host = string(value.(String)) },
	)
	t.registerProperty("port", "Int",
		func(inst Any) Any { return Int(inst.(*HttpAddress).port) },
		func(inst Any, value Any) { inst.(*HttpAddress).port = int(value.(Int)) },
	)
}

func CreateRegisterHttpAddressType() {
	t := &Type{
		name:            "HttpAddress",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpAddress]("HttpAddress"),
	}
	registerHttpAddressMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewHttpAddress())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	httpAddressType = t
}

var httpAddressType *Type

func init() {
	CreateRegisterHttpAddressType()
}

var _ Typed = (*HttpAddress)(nil)
