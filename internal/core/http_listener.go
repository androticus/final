package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HttpListener is a stateful HTTP server. Must new^. Owns ServeMux + route table.
// listen blocks the calling pip; spawn a server pip from main.
type HttpListener struct {
	address       *HttpAddress
	routes        *List[any] // *HttpRoute elements; AnyList avoids typed-list registration
	mux           *http.ServeMux
	server        *http.Server
	readTimeoutS  int
	writeTimeoutS int
}

func NewHttpListener() *HttpListener {
	return &HttpListener{
		address: NewHttpAddress(),
		routes:  NewAnyListWithCapacity(16),
		mux:     http.NewServeMux(),
	}
}

func (l *HttpListener) FinalType() *Type {
	return httpListenerType
}

func (l *HttpListener) routeAt(i int) *HttpRoute {
	return l.routes.AtAsAny(i).(*HttpRoute)
}

// Handle registers a ServeMux pattern bound to requestPipe.
// Stack-oriented Final API: ( pattern:s name:s xt:XT requestPipe:Pipe -- )
func (l *HttpListener) Handle(pattern, name string, spawnXt XT, requestPipe *Pipe) *HttpRoute {
	if requestPipe == nil {
		panic("HttpListener.handle: requestPipe is nil")
	}
	if pattern == "" {
		panic("HttpListener.handle: pattern is empty")
	}
	idx := l.routes.Len()
	route := &HttpRoute{
		index:       idx,
		name:        name,
		pattern:     pattern,
		requestPipe: requestPipe,
		spawnXt:     spawnXt,
		paramNames:  serveMuxPathParamNames(pattern),
	}
	route.goHandler = l.mintHandler(idx)
	l.routes.Append(route)
	l.mux.Handle(pattern, route.goHandler)
	return route
}

func (l *HttpListener) mintHandler(index int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.serveRoute(index, w, r)
	})
}

func (l *HttpListener) serveRoute(index int, w http.ResponseWriter, r *http.Request) {
	route := l.routeAt(index)

	inReq, err := NewHttpInReqFromHTTPRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	outRes := NewHttpOutRes()
	for _, name := range route.paramNames {
		inReq.pathParams.Set(name, String(r.PathValue(name)))
	}
	reply := NewPipe(1)
	job := NewHttpServeJob(inReq, outRes, reply)

	route.requestPipe.Push(job)
	if _, ok := reply.Pull(); !ok {
		http.Error(w, "HttpListener: reply pipe closed", http.StatusInternalServerError)
		return
	}
	flushHttpOutRes(w, outRes)
}

func flushHttpOutRes(w http.ResponseWriter, res *HttpOutRes) {
	applyStringKeyValueListToHTTPHeader(w.Header(), res.headers)
	w.WriteHeader(res.StatusOrDefault())
	if len(res.gobody) > 0 {
		_, _ = w.Write(res.gobody)
	}
}

func (l *HttpListener) listenAddr() (string, error) {
	host := l.address.host
	if host == "" {
		host = "127.0.0.1"
	}
	if l.address.port == 0 {
		return "", fmt.Errorf("HttpListener.listen: address.port is required")
	}
	return fmt.Sprintf("%s:%d", host, l.address.port), nil
}

// Listen binds and serves. Blocks until Shutdown or a fatal accept error.
// Returns (nil) on graceful shutdown; error otherwise.
func (l *HttpListener) Listen() error {
	addr, err := l.listenAddr()
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler: l.mux,
	}
	if l.readTimeoutS > 0 {
		srv.ReadTimeout = time.Duration(l.readTimeoutS) * time.Second
	}
	if l.writeTimeoutS > 0 {
		srv.WriteTimeout = time.Duration(l.writeTimeoutS) * time.Second
	}
	l.server = srv
	err = srv.Serve(ln)
	l.server = nil
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Shutdown gracefully stops a listening server.
func (l *HttpListener) Shutdown() error {
	srv := l.server
	if srv == nil {
		return fmt.Errorf("HttpListener.shutdown: not listening")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func registerHttpListenerMembers(t *Type) {
	t.registerProperty("address", "HttpAddress",
		func(inst Any) Any { return inst.(*HttpListener).address },
		nil,
	)
	t.registerProperty("routes", "List",
		func(inst Any) Any { return inst.(*HttpListener).routes },
		nil,
	)
	t.registerProperty("readTimeoutS", "Int",
		func(inst Any) Any { return Int(inst.(*HttpListener).readTimeoutS) },
		func(inst Any, value Any) { inst.(*HttpListener).readTimeoutS = int(value.(Int)) },
	)
	t.registerProperty("writeTimeoutS", "Int",
		func(inst Any) Any { return Int(inst.(*HttpListener).writeTimeoutS) },
		func(inst Any, value Any) { inst.(*HttpListener).writeTimeoutS = int(value.(Int)) },
	)

	// ( pattern:s name:s xt:XT requestPipe:Pipe -- )
	t.registerNativeMethod("handle", func(f *RunFrame, ths *TypeObjPair) {
		ds := f.RP.DataStack
		pipe := ds.Pop().(*Pipe)
		xt := ds.Pop().(XT)
		name := string(ds.Pop().(String))
		pattern := string(ds.Pop().(String))
		ths.Obj.(*HttpListener).Handle(pattern, name, xt, pipe)
	})
	// ( path:s name:s xt:XT requestPipe:Pipe -- )
	t.registerNativeMethod("get", func(f *RunFrame, ths *TypeObjPair) {
		listenerHandleMethod(f, ths, "GET")
	})
	t.registerNativeMethod("post", func(f *RunFrame, ths *TypeObjPair) {
		listenerHandleMethod(f, ths, "POST")
	})
	t.registerNativeMethod("put", func(f *RunFrame, ths *TypeObjPair) {
		listenerHandleMethod(f, ths, "PUT")
	})
	t.registerNativeMethod("delete", func(f *RunFrame, ths *TypeObjPair) {
		listenerHandleMethod(f, ths, "DELETE")
	})

	t.registerNativeMethod("listen", func(f *RunFrame, ths *TypeObjPair) {
		err := ths.Obj.(*HttpListener).Listen()
		ds := f.RP.DataStack
		if err != nil {
			ds.Push(String(err.Error()))
			ds.Push(false)
		} else {
			ds.Push(true)
		}
	})
	t.registerNativeMethod("shutdown", func(f *RunFrame, ths *TypeObjPair) {
		err := ths.Obj.(*HttpListener).Shutdown()
		ds := f.RP.DataStack
		if err != nil {
			ds.Push(String(err.Error()))
			ds.Push(false)
		} else {
			ds.Push(true)
		}
	})

	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewHttpListener())
	})
}

func listenerHandleMethod(f *RunFrame, ths *TypeObjPair, method string) {
	ds := f.RP.DataStack
	pipe := ds.Pop().(*Pipe)
	xt := ds.Pop().(XT)
	name := string(ds.Pop().(String))
	path := string(ds.Pop().(String))
	pattern := method + " " + path
	ths.Obj.(*HttpListener).Handle(pattern, name, xt, pipe)
}

func CreateRegisterHttpListenerType() {
	t := &Type{
		name:            "HttpListener",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpListener]("HttpListener"),
		restrict:        SR_HttpListen,
	}
	registerHttpListenerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpListenerType = t
}

var httpListenerType *Type

func init() {
	CreateRegisterHttpListenerType()
}

var _ Typed = (*HttpListener)(nil)
