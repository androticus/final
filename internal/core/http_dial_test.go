package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testPrepareRequest(t *testing.T, mutate func(*HttpReq)) *GoPreparedHttpReq {
	t.Helper()
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	if mutate != nil {
		mutate(req)
	}
	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	return gpr
}

func assertURLString(t *testing.T, gpr *GoPreparedHttpReq, want string) {
	t.Helper()
	got := gpr.url.String()
	if got != want {
		t.Fatalf("url.String() = %q, want %q", got, want)
	}
}

func assertNoExtraSlashInURL(t *testing.T, gpr *GoPreparedHttpReq) {
	t.Helper()
	s := gpr.url.String()
	if strings.Contains(s, "://") {
		afterScheme := strings.SplitN(s, "://", 2)[1]
		if strings.Contains(afterScheme, "//") {
			t.Fatalf("url contains extra slash: %q", s)
		}
	}
}

func TestPrepareRequest_urlDefaults(t *testing.T) {
	gpr := testPrepareRequest(t, nil)

	if gpr.url.Scheme != "https" {
		t.Fatalf("url.Scheme = %q, want https", gpr.url.Scheme)
	}
	if gpr.url.Host != "example.com" {
		t.Fatalf("url.Host = %q, want example.com", gpr.url.Host)
	}
	if gpr.url.Path != "" {
		t.Fatalf("url.Path = %q, want empty", gpr.url.Path)
	}
	if gpr.url.RawQuery != "" {
		t.Fatalf("url.RawQuery = %q, want empty", gpr.url.RawQuery)
	}
	assertURLString(t, gpr, "https://example.com")
	if gpr.method != "GET" {
		t.Fatalf("method = %q, want GET", gpr.method)
	}
}

func TestPrepareRequest_urlSchemeOverride(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.hostAddress.scheme = "http"
	})

	if gpr.url.Scheme != "http" {
		t.Fatalf("url.Scheme = %q, want http", gpr.url.Scheme)
	}
	assertURLString(t, gpr, "http://example.com")
}

func TestPrepareRequest_urlPortOverride(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.hostAddress.port = 8080
	})

	if gpr.url.Host != "example.com:8080" {
		t.Fatalf("url.Host = %q, want example.com:8080", gpr.url.Host)
	}
	assertURLString(t, gpr, "https://example.com:8080")
}

func TestPrepareRequest_urlSchemeAndPortOverride(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.hostAddress.scheme = "http"
		req.hostAddress.port = 3000
	})

	assertURLString(t, gpr, "http://example.com:3000")
}

func TestPrepareRequest_urlNoPath(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.path = ""
	})

	if gpr.url.Path != "" {
		t.Fatalf("url.Path = %q, want empty", gpr.url.Path)
	}
	assertURLString(t, gpr, "https://example.com")
}

func TestPrepareRequest_urlPathWithOrWithoutLeadingSlash(t *testing.T) {
	for _, path := range []string{"v1/users", "/v1/users"} {
		path := path
		t.Run(path, func(t *testing.T) {
			gpr := testPrepareRequest(t, func(req *HttpReq) {
				req.path = path
			})
			if gpr.url.Path != "/v1/users" {
				t.Fatalf("url.Path = %q, want /v1/users", gpr.url.Path)
			}
			assertURLString(t, gpr, "https://example.com/v1/users")
			assertNoExtraSlashInURL(t, gpr)
		})
	}
}

func TestPrepareRequest_urlNoQueryParams(t *testing.T) {
	gpr := testPrepareRequest(t, nil)

	if gpr.url.RawQuery != "" {
		t.Fatalf("url.RawQuery = %q, want empty", gpr.url.RawQuery)
	}
	if len(gpr.url.Query()) != 0 {
		t.Fatalf("url.Query() = %v, want empty", gpr.url.Query())
	}
}

func TestPrepareRequest_urlThreeQueryParams(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.path = "/search"
		req.queryParams.Add("alpha", "1")
		req.queryParams.Add("beta", "two")
		req.queryParams.Add("gamma", "3")
	})

	q := gpr.url.Query()
	for key, want := range map[string]string{
		"alpha": "1",
		"beta":  "two",
		"gamma": "3",
	} {
		if got := q.Get(key); got != want {
			t.Fatalf("query[%q] = %q, want %q", key, got, want)
		}
	}
	if len(q) != 3 {
		t.Fatalf("query param count = %d, want 3", len(q))
	}

	parsed, err := url.Parse(gpr.url.String())
	if err != nil {
		t.Fatalf("parse prepared url: %v", err)
	}
	for key, want := range map[string]string{
		"alpha": "1",
		"beta":  "two",
		"gamma": "3",
	} {
		if got := parsed.Query().Get(key); got != want {
			t.Fatalf("parsed query[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestPrepareRequest_reqHeadersNone(t *testing.T) {
	gpr := testPrepareRequest(t, nil)

	if len(gpr.req.Header) != 0 {
		t.Fatalf("req.Header = %v, want empty", gpr.req.Header)
	}
}

func TestPrepareRequest_reqHeadersThree(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.headers.Add("Content-Type", "application/json")
		req.headers.Add("Authorization", "Bearer token")
		req.headers.Add("X-Custom", "value")
	})

	want := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token",
		"X-Custom":      "value",
	}
	for name, value := range want {
		if got := gpr.req.Header.Get(name); got != value {
			t.Fatalf("req.Header.Get(%q) = %q, want %q", name, got, value)
		}
	}
	if len(gpr.req.Header) != len(want) {
		t.Fatalf("req.Header len = %d, want %d", len(gpr.req.Header), len(want))
	}
}

func TestPrepareRequest_clientSet(t *testing.T) {
	gpr := testPrepareRequest(t, nil)

	if gpr.client == nil {
		t.Fatal("client is nil, want non-nil *http.Client")
	}
}

func TestPrepareRequest_clientTimeoutFromSeconds(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.timeoutS = 30
	})

	want := 30 * time.Second
	if gpr.client.Timeout != want {
		t.Fatalf("client.Timeout = %v, want %v", gpr.client.Timeout, want)
	}
}

func TestPrepareRequest_clientNoTimeoutWhenZero(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.timeoutS = 0
	})

	if gpr.client.Timeout != 0 {
		t.Fatalf("client.Timeout = %v, want 0", gpr.client.Timeout)
	}
}

func TestPrepareRequest_bodyNilWhenReqBodyNil(t *testing.T) {
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.gobody = nil

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	if gpr.body != nil {
		t.Fatalf("body = %v, want nil", gpr.body)
	}
}

func TestPrepareRequest_bodySameReaderWhenReqBodySet(t *testing.T) {
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.SetBodyFromString("payload")

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	if gpr.body != req.gobody {
		t.Fatal("body is not the same io.Reader as req.gobody")
	}
	if gpr.body == nil {
		t.Fatal("body is nil, want req.gobody reader")
	}
	if gpr.method != "POST" {
		t.Fatalf("method = %q, want POST when body is set", gpr.method)
	}
}

func TestPrepareRequest_methodOverride(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.method = "PUT"
		req.SetBodyFromString("payload")
	})

	if gpr.method != "PUT" {
		t.Fatalf("method = %q, want PUT", gpr.method)
	}
}

func TestPrepareRequest_clearBodyRestoresGETDefault(t *testing.T) {
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.SetBodyFromString("payload")
	req.ClearBody()

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	if gpr.body != nil {
		t.Fatalf("body = %v, want nil after ClearBody", gpr.body)
	}
	if gpr.method != "GET" {
		t.Fatalf("method = %q, want GET after ClearBody", gpr.method)
	}
}

func TestPrepareRequest_queryParamsPutUpsert(t *testing.T) {
	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.queryParams.Put(String("foo"), String("bar"))
		req.queryParams.Put(String("foo"), String("baz"))
	})

	if got := gpr.url.Query().Get("foo"); got != "baz" {
		t.Fatalf("query[foo] = %q, want baz", got)
	}
	if len(gpr.url.Query()["foo"]) != 1 {
		t.Fatalf("query[foo] count = %d, want 1", len(gpr.url.Query()["foo"]))
	}
}

func TestPrepareRequest_appendPathFromList(t *testing.T) {
	segments := NewAnyList()
	segments.AppendFromAny(String("users"))
	segments.AppendFromAny(Int(42))
	segments.AppendFromAny(Uint(7))
	segments.AppendFromAny(String("reviews"))

	gpr := testPrepareRequest(t, func(req *HttpReq) {
		req.path = "api/v1"
		req.AppendPathFromList(segments)
	})

	if gpr.url.Path != "/api/v1/users/42/7/reviews" {
		t.Fatalf("url.Path = %q, want /api/v1/users/42/7/reviews", gpr.url.Path)
	}
}

func appendPathSegments(req *HttpReq, root string, segments ...any) {
	req.path = root
	list := NewAnyList()
	for _, seg := range segments {
		list.AppendFromAny(seg)
	}
	req.AppendPathFromList(list)
}

func TestAppendPathFromList_slashySegmentsCanonicalReqPath(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		segments []any
		wantPath string
	}{
		{
			name:     "leading slash on segment",
			root:     "api/v1",
			segments: []any{String("/users"), Int(42)},
			wantPath: "api/v1/users/42",
		},
		{
			name:     "trailing slash on segment",
			root:     "api/v1",
			segments: []any{String("users/"), String("reviews/")},
			wantPath: "api/v1/users/reviews",
		},
		{
			name:     "both slashes on segments",
			root:     "api/v1",
			segments: []any{String("/users/"), String("/42/"), String("/reviews/")},
			wantPath: "api/v1/users/42/reviews",
		},
		{
			name:     "slashy root and segments",
			root:     "/api/v1/",
			segments: []any{String("/users/"), Uint(7)},
			wantPath: "/api/v1/users/7",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := NewHttpReq()
			appendPathSegments(req, tc.root, tc.segments...)
			if req.path != tc.wantPath {
				t.Fatalf("req.path = %q, want %q", req.path, tc.wantPath)
			}
			trimmed := strings.TrimLeft(req.path, "/")
			if strings.Contains(trimmed, "//") {
				t.Fatalf("req.path contains extra slash: %q", req.path)
			}
		})
	}
}

func TestPrepareRequest_appendPathFromList_slashySegmentsCanonicalURL(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		segments []any
		wantURL string
	}{
		{
			name:     "leading slash on segment",
			root:     "api/v1",
			segments: []any{String("/users"), Int(42)},
			wantURL:  "/api/v1/users/42",
		},
		{
			name:     "trailing slash on segment",
			root:     "api/v1",
			segments: []any{String("users/"), String("reviews/")},
			wantURL:  "/api/v1/users/reviews",
		},
		{
			name:     "both slashes on every segment",
			root:     "api/v1",
			segments: []any{String("/users/"), String("/42/"), String("/reviews/")},
			wantURL:  "/api/v1/users/42/reviews",
		},
		{
			name:     "slashy root and slashy segments",
			root:     "/api/v1/",
			segments: []any{String("/users/"), Uint(7), String("/extra/")},
			wantURL:  "/api/v1/users/7/extra",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gpr := testPrepareRequest(t, func(req *HttpReq) {
				appendPathSegments(req, tc.root, tc.segments...)
			})
			if gpr.url.Path != tc.wantURL {
				t.Fatalf("url.Path = %q, want %q", gpr.url.Path, tc.wantURL)
			}
			assertNoExtraSlashInURL(t, gpr)
		})
	}
}

func TestCoercePathSegment_stringIntUint(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"String", String("users"), "users"},
		{"Int", Int(42), "42"},
		{"Uint", Uint(7), "7"},
		{"IntZero", Int(0), "0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := coercePathSegment(tc.in); got != tc.want {
				t.Fatalf("coercePathSegment(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCoercePathSegment_invalidTypePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("coercePathSegment: expected panic for invalid type")
		}
	}()
	coercePathSegment(true)
}

func TestAppendPathFromList_invalidSegmentPanics(t *testing.T) {
	segments := NewAnyList()
	segments.AppendFromAny(String("ok"))
	segments.AppendFromAny(true)

	req := NewHttpReq()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("AppendPathFromList: expected panic for invalid segment type")
		}
	}()
	req.AppendPathFromList(segments)
}

func applyTestServerHost(req *HttpReq, srvURL string) {
	u, err := url.Parse(srvURL)
	if err != nil {
		panic(err)
	}
	req.hostAddress.scheme = u.Scheme
	req.hostAddress.host = u.Hostname()
	if portStr := u.Port(); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			panic(err)
		}
		req.hostAddress.port = port
	}
}

func TestSend_httptestGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hello" {
			t.Errorf("request path = %q, want /hello", r.URL.Path)
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	req := NewHttpReq()
	applyTestServerHost(req, srv.URL)
	req.path = "/hello"

	res, err := NewHttpDialer().Send(req)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if res.status != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.status)
	}
	if got := res.GetBodyAsString(); got != "ok" {
		t.Fatalf("body = %q, want ok", got)
	}
}

func TestPrepareRequest_bodyFromTextRead(t *testing.T) {
	tr, err := NewTextReaderFromString("hello stream")
	if err != nil {
		t.Fatal(err)
	}
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.SetBodyFromTextRead(tr)

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	if gpr.method != "POST" {
		t.Fatalf("method = %q, want POST", gpr.method)
	}
	got, err := io.ReadAll(gpr.body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello stream" {
		t.Fatalf("body = %q, want hello stream", got)
	}
}

func TestPrepareRequest_bodyFromByteRead(t *testing.T) {
	list := NewByteList()
	list.items = append(list.items, 0x89, 'P', 'N', 'G')
	br, err := NewByteReaderFromBytes(list)
	if err != nil {
		t.Fatal(err)
	}
	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.SetBodyFromByteRead(br)

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatalf("PrepareRequest: %v", err)
	}
	got, err := io.ReadAll(gpr.body)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x89, 'P', 'N', 'G'}
	if string(got) != string(want) {
		t.Fatalf("body = %v, want %v", got, want)
	}
}

func TestSend_httptestBodyStreams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}
		if string(body) != "ping" {
			t.Errorf("request body = %q, want ping", body)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("pong-α"))
	}))
	defer srv.Close()

	tr, err := NewTextReaderFromString("ping")
	if err != nil {
		t.Fatal(err)
	}
	req := NewHttpReq()
	applyTestServerHost(req, srv.URL)
	req.path = "/echo"
	req.SetBodyFromTextRead(tr)
	req.headers.Add("Content-Type", "text/plain")

	res, err := NewHttpDialer().Send(req)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if res.status != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.status)
	}

	text := res.GetBodyAsTextRead()
	var runes []rune
	for {
		ch, eof := text.GetNextRune()
		if eof {
			break
		}
		runes = append(runes, ch)
	}
	if got := string(runes); got != "pong-α" {
		t.Fatalf("TextRead drain = %q, want pong-α", got)
	}

	bytesR := res.GetBodyAsByteRead()
	gotBytes := bytesR.ReadBytes(100)
	if string(gotBytes) != "pong-α" {
		t.Fatalf("ByteRead = %q, want pong-α", gotBytes)
	}
}

func TestSetBodyFromTextReadNilPanics(t *testing.T) {
	req := NewHttpReq()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	req.SetBodyFromTextRead(nil)
}

func TestSetBodyFromByteReadNilPanics(t *testing.T) {
	req := NewHttpReq()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	req.SetBodyFromByteRead(nil)
}

func TestSend_httptestResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "a", Value: "1"})
		http.SetCookie(w, &http.Cookie{Name: "b", Value: "2"})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req := NewHttpReq()
	applyTestServerHost(req, srv.URL)

	res, err := NewHttpDialer().Send(req)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	var cookieValues []string
	iter := res.headers.GetIter()
	for iter.GetNext() {
		kv := iter.Current().(StringKeyValue)
		if kv.Key == "Set-Cookie" {
			cookieValues = append(cookieValues, kv.Value)
		}
	}
	if len(cookieValues) != 2 {
		t.Fatalf("Set-Cookie pair count = %d, want 2; headers len=%d", len(cookieValues), res.headers.Len())
	}
}
