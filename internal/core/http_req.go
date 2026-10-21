package core

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
)

// HttpReq contains all the information needed to make an HTTP request.
// Most fields are optional, and have sensible defaults.
// The only required field is the host address.
// The timeout is 0 by default, which means no timeout.
// The method is "POST" by default when the body is non-empty, else "GET".
// The path is an empty string by default.
// The query params and headers are empty lists by default.
// The body is empty by default.
// The body can be set using SetBodyFromString, SetBodyFromBytes,
// SetBodyFromTextRead, or SetBodyFromByteRead.
// In Final, Bytes means a ByteList carrier.
// Use ClearBody before a GET when reusing a req after a POST.
type HttpReq struct {
	timeoutS    int // Request timeout in seconds. 0 = no timeout (use with caution).
	hostAddress *HttpAddress
	method      string // HTTP method: "GET", "POST", "PUT", "DELETE", etc.
	// If empty: defaults to "POST" when body non-empty, else "GET".
	path        string              // URL path, e.g. "/v1/users" or "v1/users" (leading slash optional).
	queryParams *StringKeyValueList // Ordered key-value query parameters.
	headers     *StringKeyValueList // Ordered request headers, e.g. Content-Type, Authorization.
	gobody      io.Reader           // Request body as raw bytes.
}

func NewHttpReq() *HttpReq {
	addr := NewHttpAddress()
	queryParams := NewStringKeyValueList()
	headers := NewStringKeyValueList()
	return &HttpReq{
		timeoutS:    0,
		hostAddress: addr,
		queryParams: queryParams,
		headers:     headers,
	}
}

func (req *HttpReq) FinalType() *Type {
	return httpReqType
}

func (req *HttpReq) SetBodyFromString(body string) {
	req.gobody = strings.NewReader(body)
}

func (req *HttpReq) SetBodyFromBytes(body *ByteList) {
	req.gobody = bytes.NewReader(body.items)
}

// SetBodyFromTextRead sets the request body from a TextRead (UTF-8 stream).
func (req *HttpReq) SetBodyFromTextRead(src TextRead) {
	if src == nil {
		panic(fmt.Errorf("HttpReq.SetBodyFromTextRead: nil TextRead"))
	}
	r, err := NewReaderFromTextRead(src)
	if err != nil {
		panic(err)
	}
	req.gobody = r
}

// SetBodyFromByteRead sets the request body from a ByteRead stream.
func (req *HttpReq) SetBodyFromByteRead(src ByteRead) {
	if src == nil {
		panic(fmt.Errorf("HttpReq.SetBodyFromByteRead: nil ByteRead"))
	}
	r, err := NewReaderFromByteRead(src)
	if err != nil {
		panic(err)
	}
	req.gobody = r
}

func (req *HttpReq) ClearBody() {
	req.gobody = nil
}

// AppendPathFromList joins each list element as a path segment onto req.path via path.Join.
func (req *HttpReq) AppendPathFromList(segments *List[any]) {
	for i := 0; i < segments.Len(); i++ {
		req.path = path.Join(req.path, coercePathSegment(segments.AtAsAny(i)))
	}
}

func coercePathSegment(v any) string {
	switch s := v.(type) {
	case String:
		return string(s)
	case Int, Uint:
		return fmt.Sprint(s)
	default:
		panic(fmt.Errorf("coercePathSegment: expected String, Int, or Uint, got %T", v))
	}
}

func registerHttpReqMembers(t *Type) {
	t.registerProperty("timeoutS", "Int",
		func(inst Any) Any { return Int(inst.(*HttpReq).timeoutS) },
		func(inst Any, value Any) { inst.(*HttpReq).timeoutS = int(value.(Int)) },
	)
	t.registerProperty("hostAddress", "HttpAddress",
		func(inst Any) Any { return inst.(*HttpReq).hostAddress },
		nil,
	)
	t.registerProperty("method", "String",
		func(inst Any) Any { return String(inst.(*HttpReq).method) },
		func(inst Any, value Any) { inst.(*HttpReq).method = string(value.(String)) },
	)
	t.registerProperty("path", "String",
		func(inst Any) Any { return String(inst.(*HttpReq).path) },
		func(inst Any, value Any) { inst.(*HttpReq).path = string(value.(String)) },
	)
	t.registerProperty("queryParams", "StringKeyValueList",
		func(inst Any) Any { return inst.(*HttpReq).queryParams },
		nil,
	)
	t.registerProperty("headers", "StringKeyValueList",
		func(inst Any) Any { return inst.(*HttpReq).headers },
		nil,
	)
	t.registerNativeMethod("setBodyFromString", func(f *RunFrame, ths *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*HttpReq).SetBodyFromString(s)
	})
	t.registerNativeMethod("setBodyFromBytes", func(f *RunFrame, ths *TypeObjPair) {
		body := f.RP.DataStack.Pop().(*ByteList)
		ths.Obj.(*HttpReq).SetBodyFromBytes(body)
	})
	t.registerNativeMethod("setBodyFromTextRead", func(f *RunFrame, ths *TypeObjPair) {
		src := f.RP.DataStack.Pop().(TextRead)
		ths.Obj.(*HttpReq).SetBodyFromTextRead(src)
	})
	t.registerNativeMethod("setBodyFromByteRead", func(f *RunFrame, ths *TypeObjPair) {
		src := f.RP.DataStack.Pop().(ByteRead)
		ths.Obj.(*HttpReq).SetBodyFromByteRead(src)
	})
	t.registerNativeMethod("clearBody", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*HttpReq).ClearBody()
	})
	t.registerNativeMethod("appendPathFromList", func(f *RunFrame, ths *TypeObjPair) {
		segments := f.RP.DataStack.Pop().(*List[any])
		ths.Obj.(*HttpReq).AppendPathFromList(segments)
	})
}

func CreateRegisterHttpReqType() {
	t := &Type{
		name:            "HttpReq",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpReq]("HttpReq"),
		restrict:        SR_HttpDial,
	}
	registerHttpReqMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewHttpReq())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	httpReqType = t
}

var httpReqType *Type

func init() {
	CreateRegisterHttpReqType()
}

var _ Typed = (*HttpReq)(nil)
