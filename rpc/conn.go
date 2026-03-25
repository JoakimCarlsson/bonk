// Package rpc provides JSON-RPC style CDP calls and events over a [transport.Transport].
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/joakimcarlsson/bonk/proto"
	"github.com/joakimcarlsson/bonk/transport"
)

// Conn manages CDP communication over a transport.
type Conn struct {
	transport transport.Transport
	nextID    atomic.Int64
	pending   *pendingMap
	subs      *subscriptions
	sessions  *sessionMap
	onClose   func(error)

	mu     sync.Mutex
	closed bool
	done   chan struct{}
	err    error
}

// Option configures a Conn.
type Option func(*Conn)

// New creates a new CDP connection over the given transport.
func New(t transport.Transport, opts ...Option) *Conn {
	c := &Conn{
		transport: t,
		pending:   newPendingMap(),
		subs:      newSubscriptions(),
		sessions:  newSessionMap(),
		done:      make(chan struct{}),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Listen starts the read loop. Blocks until the connection is closed or
// an error occurs. Must be called in a goroutine.
func (c *Conn) Listen(ctx context.Context) error {
	for {
		data, err := c.transport.Recv(ctx)
		if err != nil {
			c.closeWithError(err)
			return err
		}

		var msg proto.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.SessionID != "" {
			c.routeSession(&msg)
			continue
		}

		if msg.ID != 0 {
			c.pending.Resolve(msg.ID, &msg)
			continue
		}

		if msg.Method != "" {
			c.subs.Dispatch(msg.Method, msg.Params)
			continue
		}
	}
}

// Call sends a CDP command and waits for the response.
func (c *Conn) Call(
	ctx context.Context,
	method proto.Method,
	params, result any,
) error {
	return c.callOn(ctx, "", method, params, result)
}

// CallOn sends a CDP command targeted at a specific session.
func (c *Conn) CallOn(
	ctx context.Context,
	sessionID proto.SessionID,
	method proto.Method,
	params, result any,
) error {
	return c.callOn(ctx, sessionID, method, params, result)
}

func (c *Conn) callOn(
	ctx context.Context,
	sessionID proto.SessionID,
	method proto.Method,
	params, result any,
) error {
	id := c.nextID.Add(1)

	var rawParams json.RawMessage
	if params != nil {
		var err error
		rawParams, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("bonk: marshal params: %w", err)
		}
	}

	msg := proto.Message{
		ID:        id,
		SessionID: sessionID,
		Method:    method,
		Params:    rawParams,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("bonk: marshal message: %w", err)
	}

	pm := c.pending
	if sessionID != "" {
		s := c.sessions.Attach(sessionID, c)
		pm = s.pending
	}

	respCh := pm.Add(id)

	if err := c.transport.Send(ctx, data); err != nil {
		return fmt.Errorf("bonk: send: %w", err)
	}

	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return &ProtocolError{
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
				Data:    resp.Error.Data,
			}
		}
		if result != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return fmt.Errorf("bonk: unmarshal result: %w", err)
			}
		}
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-c.done:
		return ErrConnectionClosed
	}
}

// Execute implements proto.Executor.
func (c *Conn) Execute(
	ctx context.Context,
	method proto.Method,
	params, result any,
) error {
	return c.Call(ctx, method, params, result)
}

// Subscribe registers a handler for a specific event method.
func (c *Conn) Subscribe(
	method proto.Method,
	handler func(json.RawMessage),
) func() {
	return c.subs.Add(method, handler)
}

// Session returns or creates a session for the given ID.
func (c *Conn) Session(id proto.SessionID) *Session {
	return c.sessions.Attach(id, c)
}

// Close gracefully closes the connection.
func (c *Conn) Close() error {
	c.closeWithError(nil)
	return c.transport.Close()
}

// OnClose registers a callback invoked when the connection closes unexpectedly.
func (c *Conn) OnClose(fn func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onClose = fn
}

// Err returns the first error that caused the connection to close.
func (c *Conn) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *Conn) closeWithError(err error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.err = err
	fn := c.onClose
	close(c.done)

	c.pending.RejectAll(ErrConnectionClosed)
	c.sessions.rejectAll(ErrConnectionClosed)
	c.mu.Unlock()

	if fn != nil && err != nil {
		fn(err)
	}
}

func (c *Conn) routeSession(msg *proto.Message) {
	s := c.sessions.get(msg.SessionID)
	if s == nil {
		return
	}

	if msg.ID != 0 {
		s.pending.Resolve(msg.ID, msg)
		return
	}

	if msg.Method != "" {
		s.subs.Dispatch(msg.Method, msg.Params)
	}
}
