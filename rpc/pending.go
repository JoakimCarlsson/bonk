package rpc

import (
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

type pendingMap struct {
	mu       sync.Mutex
	requests map[int64]chan *proto.Message
}

func newPendingMap() *pendingMap {
	return &pendingMap{
		requests: make(map[int64]chan *proto.Message),
	}
}

func (pm *pendingMap) Add(id int64) <-chan *proto.Message {
	ch := make(chan *proto.Message, 1)
	pm.mu.Lock()
	pm.requests[id] = ch
	pm.mu.Unlock()
	return ch
}

func (pm *pendingMap) Resolve(id int64, msg *proto.Message) {
	pm.mu.Lock()
	ch, ok := pm.requests[id]
	if ok {
		delete(pm.requests, id)
	}
	pm.mu.Unlock()

	if ok {
		ch <- msg
	}
}

func (pm *pendingMap) RejectAll(err error) {
	pm.mu.Lock()
	pending := pm.requests
	pm.requests = make(map[int64]chan *proto.Message)
	pm.mu.Unlock()

	msg := &proto.Message{
		Error: &proto.Error{
			Code:    -1,
			Message: err.Error(),
		},
	}

	for _, ch := range pending {
		ch <- msg
	}
}
