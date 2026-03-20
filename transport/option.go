package transport

import "log/slog"

// DialOption configures a WebSocket dial.
type DialOption func(*dialConfig)

type dialConfig struct {
	readLimit int64
	logger    *slog.Logger
}

func defaultConfig() *dialConfig {
	return &dialConfig{
		readLimit: 100 * 1024 * 1024,
	}
}

// WithReadLimit sets the maximum message size in bytes.
func WithReadLimit(n int64) DialOption {
	return func(c *dialConfig) {
		c.readLimit = n
	}
}

// WithLogger sets a structured logger for transport operations.
func WithLogger(l *slog.Logger) DialOption {
	return func(c *dialConfig) {
		c.logger = l
	}
}
