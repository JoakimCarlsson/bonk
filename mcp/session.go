package mcp

import (
	"fmt"
	"sync"

	"github.com/joakimcarlsson/bonk"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// SessionOption configures a Session.
type SessionOption func(*sessionConfig)

type sessionConfig struct {
	headless   bool
	stealth    bool
	chromePath string
	maxPages   int
}

// WithHeadless sets whether the browser runs headless.
func WithHeadless(v bool) SessionOption {
	return func(c *sessionConfig) { c.headless = v }
}

// WithStealth sets whether stealth mode is enabled.
func WithStealth(v bool) SessionOption {
	return func(c *sessionConfig) { c.stealth = v }
}

// WithChromePath sets the path to the Chrome binary.
func WithChromePath(path string) SessionOption {
	return func(c *sessionConfig) { c.chromePath = path }
}

// WithMaxPages sets the maximum number of concurrent pages.
// When the limit is reached, the least recently used page is
// automatically closed to make room. Defaults to 10.
func WithMaxPages(n int) SessionOption {
	return func(c *sessionConfig) { c.maxPages = n }
}

// Session manages a single browser instance and its pages
// across MCP tool calls.
type Session struct {
	mu        sync.Mutex
	browser   *bonk.Browser
	ctx       *bonk.BrowserContext
	pages     map[string]*bonk.Page
	pageOrder []string
	nextID    int
	cfg       sessionConfig
	routes    map[string]func()
}

// NewSession creates a new session with the given options.
func NewSession(opts ...SessionOption) *Session {
	cfg := sessionConfig{headless: true, stealth: true, maxPages: 10}
	for _, o := range opts {
		o(&cfg)
	}
	return &Session{
		pages:  make(map[string]*bonk.Page),
		routes: make(map[string]func()),
		cfg:    cfg,
	}
}

func (s *Session) ensureBrowser() error {
	if s.browser != nil {
		return nil
	}

	var opts []bonk.LaunchOption
	opts = append(opts, bonk.Headless(s.cfg.headless))
	opts = append(opts, bonk.Stealth(s.cfg.stealth))
	if s.cfg.chromePath != "" {
		opts = append(
			opts,
			bonk.ChromePath(s.cfg.chromePath),
		)
	}

	b, err := bonk.Launch(opts...)
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}

	var ctxOpts []bonk.ContextOption
	if s.cfg.headless && s.cfg.stealth {
		ctxOpts = append(
			ctxOpts,
			bonk.WithViewport(1920, 1080),
		)
	}

	ctx, err := b.NewContext(ctxOpts...)
	if err != nil {
		b.Close()
		return fmt.Errorf("create context: %w", err)
	}

	s.browser = b
	s.ctx = ctx
	return nil
}

func (s *Session) ensurePage() (*bonk.Page, string, error) {
	if err := s.ensureBrowser(); err != nil {
		return nil, "", err
	}

	for id, p := range s.pages {
		if !p.IsClosed() {
			s.touchPage(id)
			return p, id, nil
		}
	}

	return s.newPage()
}

func (s *Session) newPage() (*bonk.Page, string, error) {
	if err := s.ensureBrowser(); err != nil {
		return nil, "", err
	}

	if s.cfg.maxPages > 0 && len(s.pages) >= s.cfg.maxPages {
		s.evictLRU()
	}

	p, err := s.ctx.NewPage()
	if err != nil {
		return nil, "", fmt.Errorf("create page: %w", err)
	}

	s.nextID++
	id := fmt.Sprintf("page_%d", s.nextID)
	s.pages[id] = p
	s.pageOrder = append(s.pageOrder, id)
	return p, id, nil
}

func (s *Session) touchPage(id string) {
	for i, v := range s.pageOrder {
		if v == id {
			s.pageOrder = append(
				s.pageOrder[:i],
				s.pageOrder[i+1:]...,
			)
			break
		}
	}
	s.pageOrder = append(s.pageOrder, id)
}

func (s *Session) evictLRU() {
	for len(s.pageOrder) > 0 {
		victim := s.pageOrder[0]
		s.pageOrder = s.pageOrder[1:]
		if p, ok := s.pages[victim]; ok {
			p.Close()
			delete(s.pages, victim)
			return
		}
	}
}

func (s *Session) getPage(
	id string,
) (*bonk.Page, error) {
	p, ok := s.pages[id]
	if !ok {
		return nil, fmt.Errorf("page %q not found", id)
	}
	if p.IsClosed() {
		delete(s.pages, id)
		return nil, fmt.Errorf("page %q is closed", id)
	}
	return p, nil
}

func (s *Session) pageFromRequest(
	req mcpgo.CallToolRequest,
) (*bonk.Page, string, error) {
	id := req.GetString("page_id", "")
	if id == "" {
		return s.ensurePage()
	}
	p, err := s.getPage(id)
	if err == nil {
		s.touchPage(id)
	}
	return p, id, err
}

func (s *Session) listPages() map[string]string {
	result := make(map[string]string, len(s.pages))
	for id, p := range s.pages {
		if p.IsClosed() {
			delete(s.pages, id)
			continue
		}
		u, _ := p.URL()
		result[id] = u
	}
	return result
}

func (s *Session) closePage(id string) error {
	p, ok := s.pages[id]
	if !ok {
		return fmt.Errorf("page %q not found", id)
	}
	delete(s.pages, id)
	for i, v := range s.pageOrder {
		if v == id {
			s.pageOrder = append(
				s.pageOrder[:i],
				s.pageOrder[i+1:]...,
			)
			break
		}
	}
	return p.Close()
}

// Close shuts down the browser and clears all state.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, unsub := range s.routes {
		unsub()
	}
	s.routes = make(map[string]func())

	if s.browser != nil {
		err := s.browser.Close()
		s.browser = nil
		s.ctx = nil
		s.pages = make(map[string]*bonk.Page)
		return err
	}
	return nil
}
