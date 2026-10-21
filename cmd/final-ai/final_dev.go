package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/androticus/final/finalgo"
)

type finalStore struct {
	mu       sync.Mutex
	sessions map[string]*finalgo.LockSession
}

func newFinalStore() *finalStore {
	return &finalStore{sessions: make(map[string]*finalgo.LockSession)}
}

func (fs *finalStore) getOrCreate(id string) (*finalgo.LockSession, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if sess, ok := fs.sessions[id]; ok {
		return sess, nil
	}
	sess, err := finalgo.NewLockSession()
	if err != nil {
		return nil, err
	}
	fs.sessions[id] = sess
	return sess, nil
}

type finalRunRequest struct {
	Script   string `json:"script"`
	TimeoutS int    `json:"timeoutS"`
}

type finalRunResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (s *server) handleFinalRun(w http.ResponseWriter, r *http.Request) {
	if s.finalStore == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req finalRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Script == "" {
		writeJSONError(w, http.StatusBadRequest, "script is required")
		return
	}

	sessionID, err := s.ensureSessionCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	finalSess, err := s.finalStore.getOrCreate(sessionID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	snd := finalSess.GetLockedSender()
	defer snd.Unlock()
	response, runErr := snd.Send(req.Script, req.TimeoutS)
	resp := finalRunResponse{Response: response}
	if runErr != nil {
		resp.Error = runErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
