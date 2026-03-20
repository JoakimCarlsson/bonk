package transport

import "context"

// Transport abstracts the bidirectional message transport for CDP.
type Transport interface {
	Send(ctx context.Context, data []byte) error
	Recv(ctx context.Context) ([]byte, error)
	Close() error
}
