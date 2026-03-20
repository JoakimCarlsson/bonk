package transport

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

// WebSocket implements Transport over a WebSocket connection.
type WebSocket struct {
	conn *websocket.Conn
}

// Dial connects to a CDP WebSocket endpoint.
func Dial(
	ctx context.Context,
	url string,
	opts ...DialOption,
) (*WebSocket, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("bonk: dial %s: %w", url, err)
	}

	conn.SetReadLimit(cfg.readLimit)

	return &WebSocket{conn: conn}, nil
}

// Send sends a raw JSON message over the WebSocket.
func (ws *WebSocket) Send(ctx context.Context, data []byte) error {
	return ws.conn.Write(ctx, websocket.MessageText, data)
}

// Recv receives a raw JSON message from the WebSocket.
func (ws *WebSocket) Recv(ctx context.Context) ([]byte, error) {
	_, data, err := ws.conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Close closes the WebSocket connection.
func (ws *WebSocket) Close() error {
	return ws.conn.Close(websocket.StatusNormalClosure, "")
}
