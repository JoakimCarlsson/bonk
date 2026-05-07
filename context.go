package bonk

import (
	"sync"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// BrowserContext represents an isolated browser context with its own
// cookies, cache, and storage.
type BrowserContext struct {
	browser *Browser
	id      proto.BrowserContextID
	cfg     *contextConfig

	mu                       sync.Mutex
	pages                    []*Page
	closed                   bool
	defaultTimeout           time.Duration
	defaultNavigationTimeout time.Duration
}

// NewPage creates a new page (tab) in this browser context.
func (c *BrowserContext) NewPage() (*Page, error) {
	return newPage(c)
}

// Pages returns all open pages in this context.
func (c *BrowserContext) Pages() []*Page {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]*Page, len(c.pages))
	copy(result, c.pages)
	return result
}

// Browser returns the parent browser.
func (c *BrowserContext) Browser() *Browser {
	return c.browser
}

// Close disposes the browser context and all its pages.
func (c *BrowserContext) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	pages := c.pages
	c.pages = nil
	c.mu.Unlock()

	for _, p := range pages {
		p.Close()
	}

	if c.id == "" {
		return nil
	}
	return proto.TargetDisposeBrowserContext(c.id).Do(c.browser.execCtx())
}

// SetDefaultTimeout sets the default timeout for all wait/query operations
// in this context. Zero means no context-level default (falls back to 30s).
func (c *BrowserContext) SetDefaultTimeout(d time.Duration) {
	c.mu.Lock()
	c.defaultTimeout = d
	c.mu.Unlock()
}

// SetDefaultNavigationTimeout sets the default timeout for navigation
// operations in this context. Zero means no context-level default.
func (c *BrowserContext) SetDefaultNavigationTimeout(d time.Duration) {
	c.mu.Lock()
	c.defaultNavigationTimeout = d
	c.mu.Unlock()
}

func (c *BrowserContext) getDefaultTimeout() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.defaultTimeout
}

func (c *BrowserContext) getDefaultNavigationTimeout() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.defaultNavigationTimeout
}

// PermissionOption configures permission granting behavior.
type PermissionOption func(*permissionConfig)

type permissionConfig struct {
	origin string
}

// PermissionOrigin scopes the granted permissions to a specific origin.
func PermissionOrigin(origin string) PermissionOption {
	return func(c *permissionConfig) {
		c.origin = origin
	}
}

// GrantPermissions grants the specified browser permissions for this context.
func (c *BrowserContext) GrantPermissions(
	permissions []string,
	opts ...PermissionOption,
) error {
	cfg := &permissionConfig{}
	for _, o := range opts {
		o(cfg)
	}

	types := make([]proto.BrowserPermissionType, len(permissions))
	for i, p := range permissions {
		types[i] = proto.BrowserPermissionType(p)
	}

	cmd := proto.BrowserGrantPermissions(types).
		WithBrowserContextID(c.id)
	if cfg.origin != "" {
		cmd = cmd.WithOrigin(cfg.origin)
	}
	return cmd.Do(c.browser.execCtx())
}

// ClearPermissions resets all permission overrides for this context.
func (c *BrowserContext) ClearPermissions() error {
	return proto.BrowserResetPermissions().
		WithBrowserContextID(c.id).
		Do(c.browser.execCtx())
}

func (c *BrowserContext) addPage(p *Page) {
	c.mu.Lock()
	c.pages = append(c.pages, p)
	c.mu.Unlock()
}

func (c *BrowserContext) removePage(p *Page) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, page := range c.pages {
		if page == p {
			c.pages = append(c.pages[:i], c.pages[i+1:]...)
			return
		}
	}
}
