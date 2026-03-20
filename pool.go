package bonk

import "sync"

// Pool manages a pool of reusable pages for concurrent automation.
type Pool struct {
	ctx   *BrowserContext
	pages chan *Page
	all   []*Page
	mu    sync.Mutex
}

// NewPool creates a pool of the given size, pre-creating pages.
func NewPool(ctx *BrowserContext, size int) (*Pool, error) {
	p := &Pool{
		ctx:   ctx,
		pages: make(chan *Page, size),
		all:   make([]*Page, 0, size),
	}

	for range size {
		page, err := ctx.NewPage()
		if err != nil {
			p.Close()
			return nil, err
		}
		p.all = append(p.all, page)
		p.pages <- page
	}

	return p, nil
}

// Do checks out a page, runs fn, and returns it to the pool.
// Blocks if all pages are in use.
func (p *Pool) Do(fn func(*Page) error) error {
	page := <-p.pages
	err := fn(page)
	p.pages <- page
	return err
}

// Close closes all pages in the pool.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, page := range p.all {
		page.Close()
	}
	p.all = nil
}
