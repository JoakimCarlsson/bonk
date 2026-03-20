package bonk

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// NavigateWait specifies when navigation is considered complete.
type NavigateWait int

const (
	WaitLoad NavigateWait = iota
	WaitDOMContentLoaded
	WaitNetworkIdle
)

// NavigateOption configures navigation behavior.
type NavigateOption func(*navigateConfig)

type navigateConfig struct {
	timeout   time.Duration
	waitUntil NavigateWait
}

func defaultNavigateConfig() *navigateConfig {
	return &navigateConfig{
		timeout:   30 * time.Second,
		waitUntil: WaitLoad,
	}
}

// WithTimeout sets the navigation timeout.
func WithTimeout(d time.Duration) NavigateOption {
	return func(c *navigateConfig) {
		c.timeout = d
	}
}

// WithWaitUntil sets when navigation is considered complete.
func WithWaitUntil(w NavigateWait) NavigateOption {
	return func(c *navigateConfig) {
		c.waitUntil = w
	}
}

// Navigate navigates the page to the given URL and waits for completion.
func (p *Page) Navigate(url string, opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	signal := func(_ json.RawMessage) {
		select {
		case done <- struct{}{}:
		default:
		}
	}

	var unsubs []func()
	switch cfg.waitUntil {
	case WaitDOMContentLoaded:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventDOMContentEventFiredMethod, signal,
		))
	case WaitNetworkIdle:
		unsubs = append(unsubs, p.subscribeNetworkIdle(done))
	default:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventLoadEventFiredMethod, signal,
		))
	}
	defer func() {
		for _, u := range unsubs {
			u()
		}
	}()

	res, err := proto.PageNavigate(url).Do(p.execCtx)
	if err != nil {
		return err
	}
	if res.ErrorText != "" {
		return &NavigationError{URL: url, Message: res.ErrorText}
	}

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{Operation: "navigation to " + url}
	}
}

// Reload reloads the current page and waits for completion.
func (p *Page) Reload(opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	signal := func(_ json.RawMessage) {
		select {
		case done <- struct{}{}:
		default:
		}
	}

	var unsubs []func()
	switch cfg.waitUntil {
	case WaitDOMContentLoaded:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventDOMContentEventFiredMethod, signal,
		))
	case WaitNetworkIdle:
		unsubs = append(unsubs, p.subscribeNetworkIdle(done))
	default:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventLoadEventFiredMethod, signal,
		))
	}
	defer func() {
		for _, u := range unsubs {
			u()
		}
	}()

	if err := proto.PageReload().Do(p.execCtx); err != nil {
		return err
	}

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{Operation: "reload"}
	}
}

// GoBack navigates back in history and waits for completion.
func (p *Page) GoBack(opts ...NavigateOption) error {
	history, err := proto.PageGetNavigationHistory().Do(p.execCtx)
	if err != nil {
		return err
	}
	if history.CurrentIndex <= 0 {
		return nil
	}
	entry := history.Entries[history.CurrentIndex-1]
	return p.navigateToEntry(entry.ID, opts...)
}

// GoForward navigates forward in history and waits for completion.
func (p *Page) GoForward(opts ...NavigateOption) error {
	history, err := proto.PageGetNavigationHistory().Do(p.execCtx)
	if err != nil {
		return err
	}
	if int(history.CurrentIndex) >= len(history.Entries)-1 {
		return nil
	}
	entry := history.Entries[history.CurrentIndex+1]
	return p.navigateToEntry(entry.ID, opts...)
}

func (p *Page) navigateToEntry(entryID int64, opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	signal := func(_ json.RawMessage) {
		select {
		case done <- struct{}{}:
		default:
		}
	}

	var unsubs []func()
	switch cfg.waitUntil {
	case WaitDOMContentLoaded:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventDOMContentEventFiredMethod, signal,
		))
	case WaitNetworkIdle:
		unsubs = append(unsubs, p.subscribeNetworkIdle(done))
	default:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventLoadEventFiredMethod, signal,
		))
	}
	defer func() {
		for _, u := range unsubs {
			u()
		}
	}()

	if err := proto.PageNavigateToHistoryEntry(entryID).Do(p.execCtx); err != nil {
		return err
	}

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{Operation: "history navigation"}
	}
}

// WaitNavigation waits for the next navigation to complete.
func (p *Page) WaitNavigation(opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	signal := func(_ json.RawMessage) {
		select {
		case done <- struct{}{}:
		default:
		}
	}

	var unsubs []func()
	switch cfg.waitUntil {
	case WaitDOMContentLoaded:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventDOMContentEventFiredMethod, signal,
		))
	case WaitNetworkIdle:
		unsubs = append(unsubs, p.subscribeNetworkIdle(done))
	default:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventLoadEventFiredMethod, signal,
		))
	}
	defer func() {
		for _, u := range unsubs {
			u()
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{Operation: "navigation"}
	}
}

// SetContent replaces the page's document HTML.
func (p *Page) SetContent(html string) error {
	return proto.PageSetDocumentContent(p.frameID, html).Do(p.execCtx)
}

// WaitForURL waits until the page URL matches the given glob pattern.
func (p *Page) WaitForURL(pattern string, opts ...WaitOption) error {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	result, err := poll(p.execCtx, cfg, func() (any, error) {
		u, err := p.URL()
		if err != nil {
			return nil, err
		}
		if globMatch(pattern, u) {
			return true, nil
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	if result == nil {
		return &TimeoutError{Operation: "waiting for URL " + pattern}
	}
	return nil
}

// URL returns the current page URL.
func (p *Page) URL() (string, error) {
	val, err := p.Evaluate("location.href")
	if err != nil {
		return "", err
	}
	s, _ := val.(string)
	return s, nil
}

// Title returns the current page title.
func (p *Page) Title() (string, error) {
	val, err := p.Evaluate("document.title")
	if err != nil {
		return "", err
	}
	s, _ := val.(string)
	return s, nil
}

func (p *Page) subscribeNetworkIdle(done chan struct{}) func() {
	var inflight atomic.Int64
	var mu sync.Mutex
	var timer *time.Timer
	idleDelay := 500 * time.Millisecond

	checkIdle := func() {
		if inflight.Load() == 0 {
			mu.Lock()
			if timer == nil {
				timer = time.AfterFunc(idleDelay, func() {
					select {
					case done <- struct{}{}:
					default:
					}
				})
			}
			mu.Unlock()
		}
	}

	cancelTimer := func() {
		mu.Lock()
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		mu.Unlock()
	}

	u1 := p.session.Subscribe(
		proto.NetworkEventRequestWillBeSentMethod,
		func(_ json.RawMessage) {
			inflight.Add(1)
			cancelTimer()
		},
	)

	u2 := p.session.Subscribe(
		proto.NetworkEventLoadingFinishedMethod,
		func(_ json.RawMessage) {
			inflight.Add(-1)
			checkIdle()
		},
	)

	u3 := p.session.Subscribe(
		proto.NetworkEventLoadingFailedMethod,
		func(_ json.RawMessage) {
			inflight.Add(-1)
			checkIdle()
		},
	)

	return func() {
		u1()
		u2()
		u3()
		cancelTimer()
	}
}

// SetOffline enables or disables offline emulation.
func (p *Page) SetOffline(offline bool) error {
	return proto.NetworkEmulateNetworkConditions(offline, 0, -1, -1).
		Do(p.execCtx)
}
