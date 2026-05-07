// Package bonk provides a high-level API for controlling Chromium-based browsers.
package bonk

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
	"github.com/joakimcarlsson/bonk/rpc"
	"github.com/joakimcarlsson/bonk/transport"
)

// Browser represents a running Chrome instance.
type Browser struct {
	conn    *rpc.Conn
	process *os.Process
	cmd     *exec.Cmd
	dataDir string
	tempDir string
	wsURL   string
	stealth bool
	cfg     *launchConfig
	ctx     context.Context
	cancel  context.CancelFunc

	mu           sync.Mutex
	contexts     []*BrowserContext
	closed       bool
	onDisconnect func()
}

// NewContext creates a new isolated browser context.
func (b *Browser) NewContext(opts ...ContextOption) (*BrowserContext, error) {
	cfg := defaultContextConfig()
	for _, o := range opts {
		o(cfg)
	}

	execCtx := proto.WithExecutor(b.ctx, b.conn)

	params := proto.TargetCreateBrowserContext().
		WithDisposeOnDetach(true)

	if cfg.proxyServer != "" {
		params = params.WithProxyServer(cfg.proxyServer)
	}
	if cfg.proxyBypassList != "" {
		params = params.WithProxyBypassList(cfg.proxyBypassList)
	}

	res, err := params.Do(execCtx)
	if err != nil {
		return nil, err
	}

	bc := &BrowserContext{
		browser: b,
		id:      res.BrowserContextID,
		cfg:     cfg,
	}

	b.mu.Lock()
	b.contexts = append(b.contexts, bc)
	b.mu.Unlock()

	return bc, nil
}

// NewPage creates a page in the browser's default context.
func (b *Browser) NewPage(opts ...ContextOption) (*Page, error) {
	cfg := defaultContextConfig()
	for _, o := range opts {
		o(cfg)
	}

	bc := &BrowserContext{
		browser: b,
		cfg:     cfg,
	}

	b.mu.Lock()
	b.contexts = append(b.contexts, bc)
	b.mu.Unlock()

	return bc.NewPage()
}

// FirstPage attaches to Chrome's initial page in the default context.
func (b *Browser) FirstPage(opts ...ContextOption) (*Page, error) {
	cfg := defaultContextConfig()
	for _, o := range opts {
		o(cfg)
	}

	execCtx := proto.WithExecutor(b.ctx, b.conn)
	targets, err := proto.TargetGetTargets().Do(execCtx)
	if err != nil {
		return nil, err
	}
	for _, target := range targets.TargetInfos {
		if target.Type != "page" || target.Attached {
			continue
		}
		bc := &BrowserContext{
			browser: b,
			cfg:     cfg,
		}
		b.mu.Lock()
		b.contexts = append(b.contexts, bc)
		b.mu.Unlock()
		return attachToTarget(bc, target.TargetID)
	}
	return b.NewPage(opts...)
}

// Close shuts down the browser and cleans up resources.
func (b *Browser) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	contexts := b.contexts
	b.contexts = nil
	b.mu.Unlock()

	for _, c := range contexts {
		c.Close()
	}

	proto.BrowserClose().Do(b.execCtx())

	b.cancel()
	b.conn.Close()

	if b.process != nil {
		b.waitForExit()
	}

	cleanupDir(b.tempDir)
	return nil
}

func (b *Browser) waitForExit() {
	done := make(chan struct{})
	go func() {
		b.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
	}

	b.process.Signal(syscall.SIGTERM)
	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
	}

	b.process.Kill()
	<-done
}

// OnDisconnect registers a callback invoked when the WebSocket connection drops.
func (b *Browser) OnDisconnect(fn func()) {
	b.mu.Lock()
	b.onDisconnect = fn
	b.mu.Unlock()
}

// Reconnect re-establishes the WebSocket connection to the browser.
// Useful after the connection drops unexpectedly.
func (b *Browser) Reconnect() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrBrowserClosed
	}
	b.mu.Unlock()

	ws, err := transport.Dial(b.ctx, b.wsURL)
	if err != nil {
		return err
	}

	var t transport.Transport = ws
	if b.cfg != nil && b.cfg.logger != nil {
		t = &transport.Debug{Inner: ws, Logger: b.cfg.logger}
	}

	conn := rpc.New(t)
	b.setupOnClose(conn)
	go conn.Listen(b.ctx)

	b.mu.Lock()
	b.conn = conn
	b.mu.Unlock()

	return nil
}

func (b *Browser) setupOnClose(conn *rpc.Conn) {
	conn.OnClose(func(_ error) {
		b.mu.Lock()
		fn := b.onDisconnect
		b.mu.Unlock()
		if fn != nil {
			fn()
		}
	})
}

func (b *Browser) execCtx() context.Context {
	return proto.WithExecutor(b.ctx, b.conn)
}
