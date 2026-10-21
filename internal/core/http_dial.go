package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type HttpDialer struct {
}

func NewHttpDialer() *HttpDialer {
	return &HttpDialer{}
}

// PrepareRequest prepares the request for execution.
// It is primarily used for testing and debugging.
func (d *HttpDialer) PrepareRequest(req *HttpReq) (*GoPreparedHttpReq, error) {
	preq := &GoPreparedHttpReq{}
	// --- 1. Build the URL ---
	scheme := req.hostAddress.scheme
	if scheme == "" {
		scheme = "https"
	}

	host := req.hostAddress.host
	if host == "" {
		return nil, fmt.Errorf("hostAddress.host is required")
	}

	portPart := ""
	if req.hostAddress.port != 0 {
		portPart = fmt.Sprintf(":%d", req.hostAddress.port)
	}

	base := fmt.Sprintf("%s://%s%s", scheme, host, portPart)
	var err error
	preq.url, err = url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("invalid hostAddress: %w", err)
	}

	// Append path safely
	if req.path != "" {
		u := preq.url
		p := strings.TrimLeft(req.path, "/")
		if p != "" {
			if u.Path == "" || u.Path == "/" {
				u.Path = "/" + p
			} else {
				u.Path = path.Join(u.Path, p)
			}
		}
	}

	// Query parameters
	if req.queryParams.Len() > 0 {
		u := preq.url
		q := u.Query()
		applyStringKeyValueListToURLValues(q, req.queryParams)
		u.RawQuery = q.Encode()
	}

	// Determine HTTP method ---
	method := req.method
	if method == "" {
		if req.gobody != nil {
			method = "POST"
		} else {
			method = "GET"
		}
	}
	preq.method = method

	// Prepare body reader ---
	if req.gobody == nil {
		preq.body = nil
	} else {
		preq.body = req.gobody
	}

	// Create the *http.Request ---
	httpReq, err := http.NewRequest(method, preq.url.String(), preq.body)
	if err != nil {
		return nil, fmt.Errorf("failed to create *http.Request: %w", err)
	}

	applyStringKeyValueListToHTTPHeader(httpReq.Header, req.headers)
	preq.req = httpReq
	// Configure client with timeout ---
	client := &http.Client{}
	if req.timeoutS > 0 {
		client.Timeout = time.Duration(req.timeoutS) * time.Second
	}
	preq.client = client

	return preq, nil
}

// Send executes the HttpReq by preparing it and sending it.
// Error is returned ONLY for transport-level problems (network, timeout, TLS, etc.).
// HTTP-level status codes (4xx/5xx) are left in res.status — check them yourself.
func (d *HttpDialer) Send(req *HttpReq) (*HttpRes, error) {
	preq, err := d.PrepareRequest(req)
	if err != nil {
		return nil, err
	}

	// Execute the request
	resp, err := preq.client.Do(preq.req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HttpRes{
		status:  resp.StatusCode,
		headers: newStringKeyValueListFromHTTPHeader(resp.Header),
		gobody:  resBody,
	}, nil
}

func applyStringKeyValueListToURLValues(q url.Values, pairs *StringKeyValueList) {
	iter := pairs.GetIter()
	for iter.GetNext() {
		kv := iter.Current().(StringKeyValue)
		q.Add(kv.Key, kv.Value)
	}
}

func applyStringKeyValueListToHTTPHeader(h http.Header, pairs *StringKeyValueList) {
	iter := pairs.GetIter()
	for iter.GetNext() {
		kv := iter.Current().(StringKeyValue)
		h.Add(kv.Key, kv.Value)
	}
}

func newStringKeyValueListFromHTTPHeader(h http.Header) *StringKeyValueList {
	out := NewStringKeyValueList()
	for key, values := range h {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}


func (d *HttpDialer) FinalType() *Type {
	return httpDialerType
}

func registerHttpDialerMembers(t *Type) {
	t.registerNativeMethod("send", func(f *RunFrame, ths *TypeObjPair) {
		dialer := ths.Obj.(*HttpDialer)
		req := f.RP.DataStack.Pop().(*HttpReq)
		if req.timeoutS == 0 {
			req.timeoutS = f.AdvisoryTimeoutS() // advisory run budget; could also be 0
		}
		res, err := dialer.Send(req)
		if err != nil {
			f.RP.DataStack.Push(String(err.Error()))
			f.RP.DataStack.Push(false)
		} else {
			f.RP.DataStack.Push(res)
			f.RP.DataStack.Push(true)
		}
	})
}

func CreateRegisterHttpDialerType() {
	t := &Type{
		name:            "HttpDialer",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpDialer]("HttpDialer"),
		restrict:        SR_HttpDial,
	}
	registerHttpDialerMembers(t)
	t.registerNativeClassMethod("new", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(NewHttpDialer())
	})
	GetTypeRegistrar().RegisterCurrentType(t)
	httpDialerType = t
}

var httpDialerType *Type

func init() {
	CreateRegisterHttpDialerType()
}

var _ Typed = (*HttpDialer)(nil)
