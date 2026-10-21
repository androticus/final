package core

import (
	"sync"
	"sync/atomic"
)

// SessionStats holds metrics for one session (no *Session pointer — safe for central registry).
type SessionStats struct {
	SessionID uint64
	RunCount  int
}

var (
	sessionStatsMu sync.Mutex
	sessionStats   = map[uint64]*SessionStats{}
	nextSessionID  atomic.Uint64
)

func allocSessionStatsID() uint64 {
	id := nextSessionID.Add(1)
	sessionStatsMu.Lock()
	sessionStats[id] = &SessionStats{SessionID: id}
	sessionStatsMu.Unlock()
	return id
}

func recordSessionRun(id uint64) {
	if id == 0 {
		return
	}
	sessionStatsMu.Lock()
	if st, ok := sessionStats[id]; ok {
		st.RunCount++
	}
	sessionStatsMu.Unlock()
}

func unregisterSessionStats(id uint64) {
	if id == 0 {
		return
	}
	sessionStatsMu.Lock()
	delete(sessionStats, id)
	sessionStatsMu.Unlock()
}

// SessionStatsSnapshot returns a copy of stats for id, or nil if unknown.
func SessionStatsSnapshot(id uint64) *SessionStats {
	sessionStatsMu.Lock()
	defer sessionStatsMu.Unlock()
	st, ok := sessionStats[id]
	if !ok {
		return nil
	}
	cp := *st
	return &cp
}
