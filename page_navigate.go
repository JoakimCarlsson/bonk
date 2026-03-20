package bonk

import (
	"encoding/json"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// NavigateOption configures navigation behavior.
type NavigateOption func(*navigateConfig)

type navigateConfig struct {
	timeout time.Duration
}

func defaultNavigateConfig() *navigateConfig {
	return &navigateConfig{
		timeout: 30 * time.Second,
	}
}

// WithTimeout sets the navigation timeout.
func WithTimeout(d time.Duration) NavigateOption {
	return func(c *navigateConfig) {
		c.timeout = d
	}
}

// Navigate navigates the page to the given URL and waits for the load event.
func (p *Page) Navigate(url string, opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	unsub := p.session.Subscribe(
		proto.PageEventLoadEventFiredMethod,
		func(_ json.RawMessage) {
			select {
			case done <- struct{}{}:
			default:
			}
		},
	)
	defer unsub()

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

// Reload reloads the current page.
func (p *Page) Reload() error {
	return proto.PageReload().Do(p.execCtx)
}

// GoBack navigates back in history.
func (p *Page) GoBack() error {
	history, err := proto.PageGetNavigationHistory().Do(p.execCtx)
	if err != nil {
		return err
	}
	if history.CurrentIndex <= 0 {
		return nil
	}
	entry := history.Entries[history.CurrentIndex-1]
	return proto.PageNavigateToHistoryEntry(entry.ID).Do(p.execCtx)
}

// GoForward navigates forward in history.
func (p *Page) GoForward() error {
	history, err := proto.PageGetNavigationHistory().Do(p.execCtx)
	if err != nil {
		return err
	}
	if int(history.CurrentIndex) >= len(history.Entries)-1 {
		return nil
	}
	entry := history.Entries[history.CurrentIndex+1]
	return proto.PageNavigateToHistoryEntry(entry.ID).Do(p.execCtx)
}

// WaitNavigation waits for the next navigation to complete.
func (p *Page) WaitNavigation(opts ...NavigateOption) error {
	cfg := defaultNavigateConfig()
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	unsub := p.session.Subscribe(
		proto.PageEventLoadEventFiredMethod,
		func(_ json.RawMessage) {
			select {
			case done <- struct{}{}:
			default:
			}
		},
	)
	defer unsub()

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{Operation: "navigation"}
	}
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
