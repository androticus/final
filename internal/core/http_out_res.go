package core

// HttpOutRes is the outbound HTTP response written by a handler pip.
// Flushed to http.ResponseWriter after the reply Pipe signals done.
type HttpOutRes struct {
	status  int // 0 → 200 at flush
	headers *StringKeyValueList
	gobody  []byte
}

func NewHttpOutRes() *HttpOutRes {
	return &HttpOutRes{
		headers: NewStringKeyValueList(),
	}
}

func (res *HttpOutRes) FinalType() *Type {
	return httpOutResType
}

func (res *HttpOutRes) SetBodyFromString(body string) {
	res.gobody = []byte(body)
}

func (res *HttpOutRes) SetBodyFromBytes(body *ByteList) {
	out := make([]byte, len(body.items))
	copy(out, body.items)
	res.gobody = out
}

func (res *HttpOutRes) ClearBody() {
	res.gobody = nil
}

func (res *HttpOutRes) StatusOrDefault() int {
	if res.status == 0 {
		return 200
	}
	return res.status
}

func registerHttpOutResMembers(t *Type) {
	t.registerProperty("status", "Int",
		func(inst Any) Any { return Int(inst.(*HttpOutRes).status) },
		func(inst Any, value Any) { inst.(*HttpOutRes).status = int(value.(Int)) },
	)
	t.registerProperty("headers", "StringKeyValueList",
		func(inst Any) Any { return inst.(*HttpOutRes).headers },
		nil,
	)
	t.registerNativeMethod("setBodyFromString", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*HttpOutRes).SetBodyFromString(s)
	})
	t.registerNativeMethod("setBodyFromBytes", func(f *RunFrame, ths *TypeObjPair) {
		body := f.RP.DataStack.Pop().(*ByteList)
		ths.Obj.(*HttpOutRes).SetBodyFromBytes(body)
	})
	t.registerNativeMethod("clearBody", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*HttpOutRes).ClearBody()
	})
}

func CreateRegisterHttpOutResType() {
	t := &Type{
		name:            "HttpOutRes",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpOutRes]("HttpOutRes"),
		restrict:        SR_HttpListen,
	}
	registerHttpOutResMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpOutResType = t
}

var httpOutResType *Type

func init() {
	CreateRegisterHttpOutResType()
}

var _ Typed = (*HttpOutRes)(nil)
