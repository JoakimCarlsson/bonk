package rpc

import (
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/joakimcarlsson/bonk/proto"
)

type subscriptions struct {
	mu     sync.RWMutex
	subs   map[proto.Method][]subscription
	nextID atomic.Uint64
}

type subscription struct {
	id      uint64
	handler func(json.RawMessage)
}

func newSubscriptions() *subscriptions {
	return &subscriptions{
		subs: make(map[proto.Method][]subscription),
	}
}

// Add registers a handler for events with the given method.
// Returns a function that unsubscribes the handler.
func (s *subscriptions) Add(
	method proto.Method,
	handler func(json.RawMessage),
) func() {
	id := s.nextID.Add(1)
	sub := subscription{id: id, handler: handler}

	s.mu.Lock()
	s.subs[method] = append(s.subs[method], sub)
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		subs := s.subs[method]
		for i, existing := range subs {
			if existing.id == id {
				s.subs[method] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

// Dispatch sends the event params to all matching subscribers.
func (s *subscriptions) Dispatch(method proto.Method, params json.RawMessage) {
	s.mu.RLock()
	subs := make([]subscription, len(s.subs[method]))
	copy(subs, s.subs[method])
	s.mu.RUnlock()

	for _, sub := range subs {
		sub.handler(params)
	}
}
