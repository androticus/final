package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleFinalRun_notInDevMode(t *testing.T) {
	s := &server{store: newChatStore(), grok: newGrokClient("k", "m")}
	req := httptest.NewRequest(http.MethodPost, "/api/final/run", strings.NewReader(`{"script":"1 2 +"}`))
	w := httptest.NewRecorder()
	s.handleFinalRun(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestHandleFinalRun_emptyScript(t *testing.T) {
	s := &server{
		store:      newChatStore(),
		grok:       newGrokClient("k", "m"),
		finalStore: newFinalStore(),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/final/run", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	s.handleFinalRun(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestHandleFinalRun_happyPath(t *testing.T) {
	s := &server{
		store:      newChatStore(),
		grok:       newGrokClient("k", "m"),
		finalStore: newFinalStore(),
	}
	body := strings.NewReader(`{"script":"1 2 ds$","timeoutS":0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/final/run", body)
	w := httptest.NewRecorder()
	s.ensureSessionCookie(w, req)
	s.handleFinalRun(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var resp finalRunResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Response != "1 2 " {
		t.Fatalf("response = %q", resp.Response)
	}
	if resp.Error != "" {
		t.Fatalf("error = %q", resp.Error)
	}
}

func TestHandleFinalRun_dictionaryPersists(t *testing.T) {
	s := &server{
		store:      newChatStore(),
		grok:       newGrokClient("k", "m"),
		finalStore: newFinalStore(),
	}

	var sessionCookie *http.Cookie
	run := func(script string) finalRunResponse {
		req := httptest.NewRequest(http.MethodPost, "/api/final/run", strings.NewReader(`{"script":`+jsonString(script)+`}`))
		if sessionCookie != nil {
			req.AddCookie(sessionCookie)
		}
		w := httptest.NewRecorder()
		s.handleFinalRun(w, req)
		if sessionCookie == nil {
			for _, c := range w.Result().Cookies() {
				if c.Name == sessionCookieName {
					sessionCookie = c
					break
				}
			}
		}
		if w.Code != http.StatusOK {
			t.Fatalf("script %q: status = %d %s", script, w.Code, w.Body.String())
		}
		var resp finalRunResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		return resp
	}

	if resp := run(": bar 99 ;"); resp.Error != "" {
		t.Fatalf("define: %s", resp.Error)
	}
	if resp := run("bar"); resp.Error != "" {
		t.Fatalf("run bar: %s", resp.Error)
	}
}

func TestHandleFinalRun_waitsForUnlock(t *testing.T) {
	s := &server{
		store:      newChatStore(),
		grok:       newGrokClient("k", "m"),
		finalStore: newFinalStore(),
	}
	const id = "sess-lock-test"
	sess, err := s.finalStore.getOrCreate(id)
	if err != nil {
		t.Fatal(err)
	}
	held := sess.GetLockedSender()
	defer func() { _ = held.Unlock() }()

	started := make(chan struct{})
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodPost, "/api/final/run", strings.NewReader(`{"script":"1 2 ds$","timeoutS":0}`))
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: id})
		w := httptest.NewRecorder()
		close(started)
		s.handleFinalRun(w, req)
		done <- w
	}()
	<-started
	select {
	case w := <-done:
		t.Fatalf("handleFinalRun returned while lock held: %s", w.Body.String())
	case <-time.After(200 * time.Millisecond):
	}
	if err := held.Unlock(); err != nil {
		t.Fatal(err)
	}
	select {
	case w := <-done:
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
		}
		var resp finalRunResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Response != "1 2 " || resp.Error != "" {
			t.Fatalf("resp = %+v", resp)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("handleFinalRun did not finish after Unlock")
	}
}

func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}
