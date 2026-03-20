package bonk

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
	"github.com/joakimcarlsson/bonk/rpc"
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
	ctx     context.Context
	cancel  context.CancelFunc

	mu       sync.Mutex
	contexts []*BrowserContext
	closed   bool
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

	b.cancel()
	b.conn.Close()

	if b.process != nil {
		b.process.Signal(os.Interrupt)
		b.cmd.Wait()
	}

	cleanupDir(b.tempDir)
	return nil
}

func (b *Browser) execCtx() context.Context {
	return proto.WithExecutor(b.ctx, b.conn)
}
