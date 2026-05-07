package bonk

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
	"github.com/joakimcarlsson/bonk/rpc"
)

type domainState struct {
	pageEnabled    atomic.Bool
	networkEnabled atomic.Bool
}

// Page represents a single browser tab/page.
type Page struct {
	browserCtx               *BrowserContext
	targetID                 proto.TargetTargetID
	sessionID                proto.SessionID
	session                  *rpc.Session
	execCtx                  context.Context
	frameID                  proto.FrameID
	fetch                    *fetchManager
	domains                  *domainState
	stealth                  bool
	defaultTimeout           time.Duration
	defaultNavigationTimeout time.Duration
	cancelTimeout            context.CancelFunc

	mu              sync.Mutex
	closed          bool
	locatorHandlers []locatorHandler
}

func newPage(c *BrowserContext) (*Page, error) {
	browserExecCtx := c.browser.execCtx()

	createRes, err := proto.TargetCreateTarget("about:blank").
		WithBrowserContextID(c.id).
		Do(browserExecCtx)
	if err != nil {
		return nil, err
	}

	p, err := attachToTarget(c, createRes.TargetID)
	if err != nil {
		return nil, err
	}

	if c.cfg.statePath != "" {
		if err := c.LoadState(c.cfg.statePath); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func attachToTarget(
	c *BrowserContext,
	targetID proto.TargetTargetID,
) (*Page, error) {
	browserExecCtx := c.browser.execCtx()

	attachRes, err := proto.TargetAttachToTarget(targetID).
		WithFlatten(true).
		Do(browserExecCtx)
	if err != nil {
		return nil, err
	}

	sessionID := proto.SessionID(attachRes.SessionID)
	session := c.browser.conn.Session(sessionID)
	execCtx := proto.WithExecutor(c.browser.ctx, session)

	stealth := c.browser.stealth
	domains := &domainState{}

	if !stealth {
		if err := proto.PageEnable().Do(execCtx); err != nil {
			return nil, err
		}
		if err := proto.PageSetLifecycleEventsEnabled(true).Do(execCtx); err != nil {
			return nil, err
		}
		if err := proto.RuntimeEnable().Do(execCtx); err != nil {
			return nil, err
		}
		if err := proto.NetworkEnable().Do(execCtx); err != nil {
			return nil, err
		}
		domains.pageEnabled.Store(true)
		domains.networkEnabled.Store(true)
	}

	var frameID proto.FrameID
	tree, err := proto.PageGetFrameTree().Do(execCtx)
	if err == nil && tree != nil {
		frameID = tree.FrameTree.Frame.ID
	}

	p := &Page{
		browserCtx:     c,
		targetID:       targetID,
		sessionID:      sessionID,
		session:        session,
		execCtx:        execCtx,
		frameID:        frameID,
		domains:        domains,
		stealth:        stealth,
		defaultTimeout: 0,
	}
	p.fetch = newFetchManager(p)

	if stealth && c.cfg.userAgent == "" {
		if err := applyStealth(p, c.cfg.locale); err != nil {
			return nil, err
		}
	}

	if c.cfg.viewportWidth > 0 && c.cfg.viewportHeight > 0 {
		if err := p.SetViewport(c.cfg.viewportWidth, c.cfg.viewportHeight); err != nil {
			return nil, err
		}
	}
	if c.cfg.userAgent != "" {
		if err := proto.EmulationSetUserAgentOverride(c.cfg.userAgent).Do(execCtx); err != nil {
			return nil, err
		}
		if err := addStealthScript(p, c.cfg.locale); err != nil {
			return nil, err
		}
	}
	if c.cfg.timezone != "" {
		if err := proto.EmulationSetTimezoneOverride(c.cfg.timezone).Do(execCtx); err != nil {
			return nil, err
		}
	}
	if c.cfg.locale != "" {
		if err := proto.EmulationSetLocaleOverride().WithLocale(c.cfg.locale).Do(execCtx); err != nil {
			return nil, err
		}
	}
	if c.cfg.hasGeo {
		if err := proto.EmulationSetGeolocationOverride().
			WithLatitude(c.cfg.geoLatitude).
			WithLongitude(c.cfg.geoLongitude).
			WithAccuracy(c.cfg.geoAccuracy).
			Do(execCtx); err != nil {
			return nil, err
		}
	}

	c.addPage(p)

	return p, nil
}

// SetViewport sets the page viewport dimensions.
func (p *Page) SetViewport(width, height int) error {
	return proto.EmulationSetDeviceMetricsOverride(
		int64(width), int64(height), 1, false,
	).
		WithScreenWidth(int64(width)).
		WithScreenHeight(int64(height)).
		Do(p.execCtx)
}

// Close closes the page.
func (p *Page) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	p.browserCtx.removePage(p)
	_, err := proto.TargetCloseTarget(p.targetID).
		Do(p.browserCtx.browser.execCtx())
	return err
}

func (p *Page) ensurePageDomain() error {
	if p.domains.pageEnabled.Load() {
		return nil
	}
	if err := proto.PageEnable().Do(p.execCtx); err != nil {
		return err
	}
	if err := proto.PageSetLifecycleEventsEnabled(true).Do(p.execCtx); err != nil {
		return err
	}
	p.domains.pageEnabled.Store(true)
	return nil
}

func (p *Page) ensureNetworkDomain() error {
	if p.domains.networkEnabled.Load() {
		return nil
	}
	if err := proto.NetworkEnable().Do(p.execCtx); err != nil {
		return err
	}
	p.domains.networkEnabled.Store(true)
	return nil
}

// WithContext returns a shallow copy of the Page with the given context.
// All CDP calls made on the returned Page will respect the context's
// deadline and cancellation. The copy shares the underlying session,
// fetch manager, and mutex with the original.
func (p *Page) WithContext(ctx context.Context) *Page {
	p.mu.Lock()
	handlers := make(
		[]locatorHandler, len(p.locatorHandlers),
	)
	copy(handlers, p.locatorHandlers)
	p.mu.Unlock()

	return &Page{
		browserCtx:               p.browserCtx,
		targetID:                 p.targetID,
		sessionID:                p.sessionID,
		session:                  p.session,
		execCtx:                  proto.WithExecutor(ctx, p.session),
		frameID:                  p.frameID,
		fetch:                    p.fetch,
		domains:                  p.domains,
		stealth:                  p.stealth,
		defaultTimeout:           p.defaultTimeout,
		defaultNavigationTimeout: p.defaultNavigationTimeout,
		locatorHandlers:          handlers,
	}
}

// Timeout returns a shallow copy of the Page with a context deadline
// set to the given duration from now.
func (p *Page) Timeout(d time.Duration) *Page {
	ctx, cancel := context.WithTimeout(p.execCtx, d)
	page := p.WithContext(ctx)
	page.cancelTimeout = cancel
	return page
}

// Context returns the parent browser context.
func (p *Page) Context() *BrowserContext {
	return p.browserCtx
}

// AddInitScript adds a script that will be evaluated on every new document
// created in the page, including iframes. Useful for injecting polyfills
// or overriding APIs before page scripts run.
func (p *Page) AddInitScript(script string) error {
	if err := p.ensurePageDomain(); err != nil {
		return err
	}
	_, err := proto.PageAddScriptToEvaluateOnNewDocument(script).Do(p.execCtx)
	return err
}

// BringToFront activates this page (tab).
func (p *Page) BringToFront() error {
	return proto.PageBringToFront().Do(p.execCtx)
}

// IsClosed reports whether the page has been closed.
func (p *Page) IsClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

// SetDefaultTimeout sets the default timeout for wait/query operations on
// this page. Zero clears the page-level override (inherits from context).
func (p *Page) SetDefaultTimeout(d time.Duration) {
	p.mu.Lock()
	p.defaultTimeout = d
	p.mu.Unlock()
}

// SetDefaultNavigationTimeout sets the default timeout for navigation
// operations on this page. Zero clears the page-level override.
func (p *Page) SetDefaultNavigationTimeout(d time.Duration) {
	p.mu.Lock()
	p.defaultNavigationTimeout = d
	p.mu.Unlock()
}

const fallbackTimeout = 30 * time.Second

func (p *Page) resolveTimeout() time.Duration {
	if p.defaultTimeout > 0 {
		return p.defaultTimeout
	}
	if d := p.browserCtx.getDefaultTimeout(); d > 0 {
		return d
	}
	return fallbackTimeout
}

func (p *Page) resolveNavigationTimeout() time.Duration {
	if p.defaultNavigationTimeout > 0 {
		return p.defaultNavigationTimeout
	}
	if d := p.browserCtx.getDefaultNavigationTimeout(); d > 0 {
		return d
	}
	if p.defaultTimeout > 0 {
		return p.defaultTimeout
	}
	if d := p.browserCtx.getDefaultTimeout(); d > 0 {
		return d
	}
	return fallbackTimeout
}
