package core

import "bytes"

// HttpRes is the result returned by Send.
// It contains the status code, headers, and body.
// Retrieve the body as the type you want with the GetBody* methods.
type HttpRes struct {
	status  int                 // HTTP status code, e.g. 200, 404, 500
	headers *StringKeyValueList // Response headers; one pair per header line.
	gobody  []byte              // Raw response body bytes.
}

func (res *HttpRes) FinalType() *Type {
	return httpResType
}

func (res *HttpRes) GetBodyAsBytes() *ByteList {
	bl := NewByteList()
	// nothing should mutate the response body, so we can safely use it.
	bl.items = res.gobody
	return bl
}

func (res *HttpRes) GetBodyAsString() string {
	return string(res.gobody)
}

// GetBodyAsTextRead returns a TextRead over the response body (UTF-8).
func (res *HttpRes) GetBodyAsTextRead() TextRead {
	rd, err := NewTextReaderFromReader(bytes.NewReader(res.gobody))
	if err != nil {
		panic(err)
	}
	return rd
}

// GetBodyAsByteRead returns a ByteRead over the response body.
func (res *HttpRes) GetBodyAsByteRead() ByteRead {
	rd, err := NewByteReaderFromReader(bytes.NewReader(res.gobody))
	if err != nil {
		panic(err)
	}
	return rd
}

func registerHttpResMembers(t *Type) {
	t.registerProperty("status", "Int",
		func(inst Any) Any { return Int(inst.(*HttpRes).status) },
		nil,
	)
	t.registerProperty("headers", "StringKeyValueList",
		func(inst Any) Any { return inst.(*HttpRes).headers },
		nil,
	)
	t.registerNativeMethod("getBodyAsBytes", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*HttpRes).GetBodyAsBytes())
	})
	t.registerNativeMethod("getBodyAsString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*HttpRes).GetBodyAsString()))
	})
	t.registerNativeMethod("getBodyAsTextRead", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*HttpRes).GetBodyAsTextRead())
	})
	t.registerNativeMethod("getBodyAsByteRead", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*HttpRes).GetBodyAsByteRead())
	})
}

func CreateRegisterHttpResType() {
	t := &Type{
		name:            "HttpRes",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpRes]("HttpRes"),
		restrict:        SR_HttpDial,
	}
	registerHttpResMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpResType = t
}

var httpResType *Type

func init() {
	CreateRegisterHttpResType()
}

var _ Typed = (*HttpRes)(nil)
