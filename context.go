package bonk

import (
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

// BrowserContext represents an isolated browser context with its own
// cookies, cache, and storage.
type BrowserContext struct {
	browser *Browser
	id      proto.BrowserContextID
	cfg     *contextConfig

	mu     sync.Mutex
	pages  []*Page
	closed bool
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

	return proto.TargetDisposeBrowserContext(c.id).Do(c.browser.execCtx())
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
