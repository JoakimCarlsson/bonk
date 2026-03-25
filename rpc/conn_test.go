package rpc

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

type mockTransport struct {
	mu     sync.Mutex
	sent   [][]byte
	recvCh chan []byte
	closed bool
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		recvCh: make(chan []byte, 100),
	}
}

func (m *mockTransport) Send(_ context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, data)
	return nil
}

func (m *mockTransport) Recv(ctx context.Context) ([]byte, error) {
	select {
	case data := <-m.recvCh:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *mockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockTransport) feedResponse(id int64, result any) {
	raw, _ := json.Marshal(result)
	msg := proto.Message{
		ID:     id,
		Result: raw,
	}
	data, _ := json.Marshal(msg)
	m.recvCh <- data
}

func (m *mockTransport) feedEvent(method proto.Method, params any) {
	raw, _ := json.Marshal(params)
	msg := proto.Message{
		Method: method,
		Params: raw,
	}
	data, _ := json.Marshal(msg)
	m.recvCh <- data
}

func (m *mockTransport) feedError(id int64, code int64, message string) {
	msg := proto.Message{
		ID: id,
		Error: &proto.Error{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(msg)
	m.recvCh <- data
}

func TestCallAndResponse(t *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go conn.Listen(ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		mt.feedResponse(1, map[string]string{"protocolVersion": "1.3"})
	}()

	var result struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	err := conn.Call(ctx, "Browser.getVersion", nil, &result)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}

	if result.ProtocolVersion != "1.3" {
		t.Errorf("got protocolVersion=%q, want 1.3", result.ProtocolVersion)
	}
}

func TestCallProtocolError(t *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go conn.Listen(ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		mt.feedError(1, ErrCodeMethodNotFound, "method not found")
	}()

	err := conn.Call(ctx, "Fake.method", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsMethodNotFound(err) {
		t.Errorf("expected method not found, got: %v", err)
	}
}

func TestSubscribe(t *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go conn.Listen(ctx)

	received := make(chan json.RawMessage, 1)
	unsub := conn.Subscribe(
		"Page.loadEventFired",
		func(params json.RawMessage) {
			received <- params
		},
	)
	defer unsub()

	mt.feedEvent("Page.loadEventFired", map[string]float64{"timestamp": 1234.5})

	select {
	case params := <-received:
		var ev struct {
			Timestamp float64 `json:"timestamp"`
		}
		json.Unmarshal(params, &ev)
		if ev.Timestamp != 1234.5 {
			t.Errorf("got timestamp=%v, want 1234.5", ev.Timestamp)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestUnsubscribe(t *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go conn.Listen(ctx)

	count := 0
	unsub := conn.Subscribe(
		"Page.loadEventFired",
		func(_ json.RawMessage) {
			count++
		},
	)

	mt.feedEvent("Page.loadEventFired", nil)
	time.Sleep(50 * time.Millisecond)

	unsub()

	mt.feedEvent("Page.loadEventFired", nil)
	time.Sleep(50 * time.Millisecond)

	if count != 1 {
		t.Errorf("got count=%d, want 1", count)
	}
}

func TestExecutorInterface(_ *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	var _ proto.Executor = conn
}

func TestConcurrentCalls(_ *testing.T) {
	mt := newMockTransport()
	conn := New(mt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go conn.Listen(ctx)

	go func() {
		for i := 1; i <= 10; i++ {
			time.Sleep(5 * time.Millisecond)
			mt.feedResponse(int64(i), map[string]string{"ok": "true"})
		}
	}()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var result struct {
				OK string `json:"ok"`
			}
			conn.Call(ctx, "Test.method", nil, &result)
		}()
	}

	wg.Wait()
}
