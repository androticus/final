package core

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestHttpListenerServeHealthRoundTrip(t *testing.T) {
	ResetTestRuntime()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	reqs := NewPipe(4)
	listener := NewHttpListener()
	listener.address.host = "127.0.0.1"
	listener.address.port = port
	listener.Handle("GET /health", "healthLoop", 0, reqs)

	done := make(chan error, 1)
	go func() {
		done <- listener.Listen()
	}()

	// Serial Final-style handler in Go: pull job, write body, signal reply.
	go func() {
		for {
			v, ok := reqs.Pull()
			if !ok {
				return
			}
			job := v.(*HttpServeJob)
			job.res.status = 200
			job.res.headers.Add("Content-Type", "text/plain")
			job.res.SetBodyFromString("ok")
			job.reply.Push(Int(0))
		}
	}()

	deadline := time.Now().Add(3 * time.Second)
	var resp *http.Response
	for {
		resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET /health: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != "ok" {
		t.Fatalf("body = %q, want ok", body)
	}

	if err := listener.Shutdown(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Listen = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Listen did not return after Shutdown")
	}
}

func TestHttpListenerPathParam(t *testing.T) {
	ResetTestRuntime()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	reqs := NewPipe(4)
	listener := NewHttpListener()
	listener.address.host = "127.0.0.1"
	listener.address.port = port
	listener.Handle("GET /api/items/{id}", "getItem", 0, reqs)

	go func() { _ = listener.Listen() }()
	defer listener.Shutdown()

	go func() {
		v, ok := reqs.Pull()
		if !ok {
			return
		}
		job := v.(*HttpServeJob)
		id, _ := job.req.pathParams.Get("id").(String)
		job.res.status = 200
		job.res.SetBodyFromString(string(id))
		job.reply.Push(Int(0))
	}()

	deadline := time.Now().Add(3 * time.Second)
	var resp *http.Response
	for {
		resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/items/42", port))
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "42" {
		t.Fatalf("body = %q, want 42", body)
	}
}

func TestServeMuxPathParamNames(t *testing.T) {
	got := serveMuxPathParamNames("GET /users/{uid}/orders/{oid}")
	if len(got) != 2 || got[0] != "uid" || got[1] != "oid" {
		t.Fatalf("names = %v", got)
	}
}
