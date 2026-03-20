package rpc

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

type sessionMap struct {
	mu       sync.RWMutex
	sessions map[proto.SessionID]*Session
}

// Session represents a CDP session attached to a specific target.
type Session struct {
	id      proto.SessionID
	conn    *Conn
	pending *pendingMap
	subs    *subscriptions
}

func newSessionMap() *sessionMap {
	return &sessionMap{
		sessions: make(map[proto.SessionID]*Session),
	}
}

// Attach creates or returns a session for the given ID.
func (sm *sessionMap) Attach(id proto.SessionID, conn *Conn) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if s, ok := sm.sessions[id]; ok {
		return s
	}

	s := &Session{
		id:      id,
		conn:    conn,
		pending: newPendingMap(),
		subs:    newSubscriptions(),
	}
	sm.sessions[id] = s
	return s
}

// Detach removes a session.
func (sm *sessionMap) Detach(id proto.SessionID) {
	sm.mu.Lock()
	s, ok := sm.sessions[id]
	if ok {
		delete(sm.sessions, id)
	}
	sm.mu.Unlock()

	if ok {
		s.pending.RejectAll(ErrConnectionClosed)
	}
}

func (sm *sessionMap) get(id proto.SessionID) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

func (sm *sessionMap) rejectAll(err error) {
	sm.mu.Lock()
	sessions := sm.sessions
	sm.sessions = make(map[proto.SessionID]*Session)
	sm.mu.Unlock()

	for _, s := range sessions {
		s.pending.RejectAll(err)
	}
}

// Execute sends a CDP command through this session.
func (s *Session) Execute(
	ctx context.Context,
	method proto.Method,
	params, result any,
) error {
	return s.conn.CallOn(ctx, s.id, method, params, result)
}

// Subscribe registers a handler for events on this session.
func (s *Session) Subscribe(
	method proto.Method,
	handler func(json.RawMessage),
) func() {
	return s.subs.Add(method, handler)
}
