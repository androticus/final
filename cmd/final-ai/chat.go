package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	aiopenai "github.com/androticus/final/cmd/final-ai/openai"
	"github.com/androticus/final/finalgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

const (
	sessionCookieName = "final_ai_session"
	grokBaseURL       = "https://api.x.ai/v1"
)

type chatRequest struct {
	Message string `json:"message"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

type chatSession struct {
	messages []openai.ChatCompletionMessageParamUnion
}

type chatStore struct {
	mu       sync.Mutex
	sessions map[string]*chatSession
}

func newChatStore() *chatStore {
	return &chatStore{sessions: make(map[string]*chatSession)}
}

func (s *chatStore) getOrCreate(id string) *chatSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		return sess
	}
	sess := &chatSession{}
	s.sessions[id] = sess
	return sess
}

type grokClient struct {
	apiKey string
	client openai.Client
	model  shared.ChatModel
}

func newGrokClient(apiKey, model string) *grokClient {
	apiKey = strings.TrimSpace(apiKey)
	return &grokClient{
		apiKey: apiKey,
		client: openai.NewClient(
			option.WithBaseURL(grokBaseURL),
			option.WithAPIKey(apiKey),
			option.WithHeader("Authorization", "Bearer "+apiKey),
		),
		model: shared.ChatModel(model),
	}
}

type server struct {
	cfg        *Config
	store      *chatStore
	grok       *grokClient
	finalStore *finalStore
}

func newServer(cfg *Config) (*server, error) {
	if strings.TrimSpace(cfg.ApiKey) == "" {
		return nil, fmt.Errorf("api key is required (-apikey or FINAL_AI_API_KEY)")
	}
	if cfg.Provider != "" && cfg.Provider != "grok" {
		return nil, fmt.Errorf("unsupported provider %q (only grok in v0)", cfg.Provider)
	}
	return &server{
		cfg:   cfg,
		store: newChatStore(),
		grok:  newGrokClient(cfg.ApiKey, cfg.Model),
		finalStore: func() *finalStore {
			if cfg.Dev {
				return newFinalStore()
			}
			return nil
		}(),
	}, nil
}

func (s *server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "message is required")
		return
	}

	sessionID, err := s.ensureSessionCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sess := s.store.getOrCreate(sessionID)
	prior := append([]openai.ChatCompletionMessageParamUnion(nil), sess.messages...)

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var reply string
	var newHistory []openai.ChatCompletionMessageParamUnion

	if s.finalStore != nil {
		finalSess, fsErr := s.finalStore.getOrCreate(sessionID)
		if fsErr != nil {
			writeJSONError(w, http.StatusInternalServerError, fsErr.Error())
			return
		}
		snd := finalSess.GetLockedSender()
		defer snd.Unlock()
		ft := finalgo.NewFinalTool()
		toolEntry, toolErr := aiopenai.ToolParam(ft.Name(), ft.Description(), ft.ParametersJSON())
		if toolErr != nil {
			writeJSONError(w, http.StatusInternalServerError, toolErr.Error())
			return
		}
		tools := []openai.ChatCompletionToolParam{toolEntry}
		reply, newHistory, err = runChatTurn(ctx, s.grok.client, s.grok.model, tools, ft, snd, prior, req.Message, s.cfg.Spy)
	} else {
		reply, newHistory, err = s.grok.completeChat(ctx, prior, req.Message)
	}

	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	sess.messages = newHistory

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(chatResponse{Reply: reply})
}

func (s *server) ensureSessionCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
		return c.Value, nil
	}
	id := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
	return id, nil
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
