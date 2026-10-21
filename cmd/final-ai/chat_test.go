package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleChat_badMethod(t *testing.T) {
	s := &server{store: newChatStore(), grok: newGrokClient("k", "m")}
	req := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	w := httptest.NewRecorder()
	s.handleChat(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestHandleChat_emptyMessage(t *testing.T) {
	s := &server{store: newChatStore(), grok: newGrokClient("k", "m")}
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	s.handleChat(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestHandleChat_setsSessionCookie(t *testing.T) {
	s := &server{store: newChatStore(), grok: newGrokClient("k", "m")}
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"message":"hello"}`))
	w := httptest.NewRecorder()
	s.ensureSessionCookie(w, req)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
		t.Fatalf("cookies = %v", cookies)
	}
}

func TestNewServer_requiresAPIKey(t *testing.T) {
	_, err := newServer(&Config{Provider: "grok", Model: "grok-3"})
	if err == nil {
		t.Fatal("expected error for missing api key")
	}
}

func TestNewServer_rejectsUnknownProvider(t *testing.T) {
	_, err := newServer(&Config{ApiKey: "k", Provider: "chatgpt", Model: "grok-3"})
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSONError(w, http.StatusBadGateway, "upstream failed")
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "upstream failed" {
		t.Fatalf("body = %v", body)
	}
}
