package bonk

import (
	"context"
	"time"
)

// WaitOption configures wait behavior.
type WaitOption func(*waitConfig)

// WaitVisibility specifies what state to wait for.
type WaitVisibility int

// WaitAttached, WaitUntilVisible, and WaitUntilHidden select which element state to wait for.
const (
	WaitAttached WaitVisibility = iota
	WaitUntilVisible
	WaitUntilHidden
)

type waitConfig struct {
	timeout    time.Duration
	interval   time.Duration
	visibility WaitVisibility
}

// WaitTimeout sets the maximum time to wait.
func WaitTimeout(d time.Duration) WaitOption {
	return func(c *waitConfig) {
		c.timeout = d
	}
}

// WaitInterval sets the initial polling interval.
func WaitInterval(d time.Duration) WaitOption {
	return func(c *waitConfig) {
		c.interval = d
	}
}

// WaitVisibleOption sets the visibility condition for WaitSelector.
func WaitVisibleOption() WaitOption {
	return func(c *waitConfig) {
		c.visibility = WaitUntilVisible
	}
}

// WaitHiddenOption sets WaitSelector to wait until the element is hidden or removed.
func WaitHiddenOption() WaitOption {
	return func(c *waitConfig) {
		c.visibility = WaitUntilHidden
	}
}

func defaultWaitConfigFor(p *Page) *waitConfig {
	return &waitConfig{
		timeout:  p.resolveTimeout(),
		interval: 50 * time.Millisecond,
	}
}

func poll(
	ctx context.Context,
	cfg *waitConfig,
	check func() (any, error),
) (any, error) {
	deadline := time.Now().Add(cfg.timeout)
	interval := cfg.interval
	maxInterval := time.Second

	for {
		result, err := check()
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		if time.Now().After(deadline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		interval = min(interval*2, maxInterval)
	}
}
