package bonk

import (
	"context"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
	"github.com/joakimcarlsson/bonk/rpc"
)

// Page represents a single browser tab/page.
type Page struct {
	browserCtx *BrowserContext
	targetID   proto.TargetTargetID
	sessionID  proto.SessionID
	session    *rpc.Session
	execCtx    context.Context
	frameID    proto.FrameID
	fetch      *fetchManager
	stealth    bool

	mu     sync.Mutex
	closed bool
}

func newPage(c *BrowserContext) (*Page, error) {
	browserExecCtx := c.browser.execCtx()

	createRes, err := proto.TargetCreateTarget("about:blank").
		WithBrowserContextID(c.id).
		Do(browserExecCtx)
	if err != nil {
		return nil, err
	}

	attachRes, err := proto.TargetAttachToTarget(createRes.TargetID).
		WithFlatten(true).
		Do(browserExecCtx)
	if err != nil {
		return nil, err
	}

	sessionID := proto.SessionID(attachRes.SessionID)
	session := c.browser.conn.Session(sessionID)
	execCtx := proto.WithExecutor(c.browser.ctx, session)

	stealth := c.browser.stealth

	if err := proto.PageEnable().Do(execCtx); err != nil {
		return nil, err
	}
	if err := proto.PageSetLifecycleEventsEnabled(true).Do(execCtx); err != nil {
		return nil, err
	}
	if !stealth {
		if err := proto.RuntimeEnable().Do(execCtx); err != nil {
			return nil, err
		}
	}
	if err := proto.NetworkEnable().Do(execCtx); err != nil {
		return nil, err
	}

	var frameID proto.FrameID
	tree, err := proto.PageGetFrameTree().Do(execCtx)
	if err == nil && tree != nil {
		frameID = tree.FrameTree.Frame.ID
	}

	p := &Page{
		browserCtx: c,
		targetID:   createRes.TargetID,
		sessionID:  sessionID,
		session:    session,
		execCtx:    execCtx,
		frameID:    frameID,
		stealth:    stealth,
	}
	p.fetch = newFetchManager(p)

	if stealth {
		if err := applyStealth(p); err != nil {
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

	if c.cfg.statePath != "" {
		if err := c.LoadState(c.cfg.statePath); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// SetViewport sets the page viewport dimensions.
func (p *Page) SetViewport(width, height int) error {
	return proto.EmulationSetDeviceMetricsOverride(
		int64(width), int64(height), 1, false,
	).Do(p.execCtx)
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

// Context returns the parent browser context.
func (p *Page) Context() *BrowserContext {
	return p.browserCtx
}
