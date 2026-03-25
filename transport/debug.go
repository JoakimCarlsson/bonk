// Package transport provides WebSocket transport for Chrome DevTools Protocol.
package transport

import (
	"context"
	"log/slog"
)

// Debug wraps a Transport with logging.
type Debug struct {
	Inner  Transport
	Logger *slog.Logger
}

// Send logs and delegates to the inner transport.
func (d *Debug) Send(ctx context.Context, data []byte) error {
	d.Logger.Debug("send", "data", string(data))
	return d.Inner.Send(ctx, data)
}

// Recv delegates to the inner transport and logs.
func (d *Debug) Recv(ctx context.Context) ([]byte, error) {
	data, err := d.Inner.Recv(ctx)
	if err != nil {
		d.Logger.Debug("recv error", "err", err)
		return nil, err
	}
	d.Logger.Debug("recv", "data", string(data))
	return data, nil
}

// Close delegates to the inner transport.
func (d *Debug) Close() error {
	return d.Inner.Close()
}
