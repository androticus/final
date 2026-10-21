package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HttpInReq is an inbound HTTP request filled from *http.Request.
// No public app new — created by the listen bridge.
type HttpInReq struct {
	method      string
	path        string         // path only (no query), leading /
	pathParams  *DynamicObject // ServeMux {name} wildcards; filled by listener after match
	queryParams *DynamicObject
	headers     *StringKeyValueList
	host        string
	remoteAddr  string
	gobody      []byte
}

func NewHttpInReqFromHTTPRequest(r *http.Request) (*HttpInReq, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("HttpInReq: read body: %w", err)
	}
	_ = r.Body.Close()

	query := NewDynamicObject()
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			query.Set(k, String(vs[0]))
		}
	}

	return &HttpInReq{
		method:      r.Method,
		path:        r.URL.Path,
		pathParams:  NewDynamicObject(),
		queryParams: query,
		headers:     newStringKeyValueListFromHTTPHeader(r.Header),
		host:        r.Host,
		remoteAddr:  r.RemoteAddr,
		gobody:      body,
	}, nil
}

func (req *HttpInReq) FinalType() *Type {
	return httpInReqType
}

func (req *HttpInReq) GetBodyAsBytes() *ByteList {
	bl := NewByteList()
	bl.items = req.gobody
	return bl
}

func (req *HttpInReq) GetBodyAsString() string {
	return string(req.gobody)
}

// GetFormParams parses application/x-www-form-urlencoded body into a StringKeyValueList
// (first value per key). Returns (form, "", true) or (nil, errMsg, false).
func (req *HttpInReq) GetFormParams() (*StringKeyValueList, string, bool) {
	ct := headerFirst(req.headers, "Content-Type")
	if ct != "" && !strings.Contains(strings.ToLower(ct), "application/x-www-form-urlencoded") {
		return nil, "HttpInReq.getFormParams: Content-Type is not application/x-www-form-urlencoded", false
	}
	vals, err := url.ParseQuery(string(req.gobody))
	if err != nil {
		return nil, err.Error(), false
	}
	form := NewStringKeyValueList()
	for k, vs := range vals {
		if len(vs) > 0 {
			form.Add(k, vs[0])
		}
	}
	return form, "", true
}

func headerFirst(headers *StringKeyValueList, name string) string {
	if headers == nil {
		return ""
	}
	for i := 0; i < headers.Len(); i++ {
		kv := headers.At(i)
		if strings.EqualFold(kv.Key, name) {
			return kv.Value
		}
	}
	return ""
}

func registerHttpInReqMembers(t *Type) {
	t.registerProperty("method", "String",
		func(inst Any) Any { return String(inst.(*HttpInReq).method) },
		nil,
	)
	t.registerProperty("path", "String",
		func(inst Any) Any { return String(inst.(*HttpInReq).path) },
		nil,
	)
	t.registerProperty("pathParams", "DynamicObject",
		func(inst Any) Any { return inst.(*HttpInReq).pathParams },
		nil,
	)
	t.registerProperty("queryParams", "DynamicObject",
		func(inst Any) Any { return inst.(*HttpInReq).queryParams },
		nil,
	)
	t.registerProperty("headers", "StringKeyValueList",
		func(inst Any) Any { return inst.(*HttpInReq).headers },
		nil,
	)
	t.registerProperty("host", "String",
		func(inst Any) Any { return String(inst.(*HttpInReq).host) },
		nil,
	)
	t.registerProperty("remoteAddr", "String",
		func(inst Any) Any { return String(inst.(*HttpInReq).remoteAddr) },
		nil,
	)
	t.registerNativeMethod("getBodyAsBytes", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*HttpInReq).GetBodyAsBytes())
	})
	t.registerNativeMethod("getBodyAsString", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*HttpInReq).GetBodyAsString()))
	})
	t.registerNativeMethod("getFormParams", func(f *RunFrame, ths *TypeObjPair) {
		form, errMsg, ok := ths.Obj.(*HttpInReq).GetFormParams()
		ds := f.RP.DataStack
		if ok {
			ds.Push(form)
			ds.Push(true)
		} else {
			ds.Push(String(errMsg))
			ds.Push(false)
		}
	})
}

func CreateRegisterHttpInReqType() {
	t := &Type{
		name:            "HttpInReq",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpInReq]("HttpInReq"),
		restrict:        SR_HttpListen,
	}
	registerHttpInReqMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpInReqType = t
}

var httpInReqType *Type

func init() {
	CreateRegisterHttpInReqType()
}

var _ Typed = (*HttpInReq)(nil)
