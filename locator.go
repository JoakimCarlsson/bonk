package bonk

import "fmt"

// Locator is a Playwright-style selector handle that re-queries on every
// action. Locators never go stale because they don't cache a DOM reference.
type Locator struct {
	page      *Page
	frame     *Frame
	selector  string
	jsExpr    string
	jsAllExpr string
	desc      string
	nth       int
}

// Locator returns a Locator for the given CSS selector on the page.
func (p *Page) Locator(selector string) *Locator {
	return &Locator{page: p, selector: selector, nth: -1}
}

// Locator returns a Locator for the given CSS selector within the frame.
func (f *Frame) Locator(selector string) *Locator {
	return &Locator{page: f.page, frame: f, selector: selector, nth: -1}
}

// First returns a Locator that resolves to the first match.
func (l *Locator) First() *Locator {
	return l.Nth(0)
}

// Nth returns a Locator that resolves to the nth match (zero-based).
func (l *Locator) Nth(n int) *Locator {
	return &Locator{
		page:      l.page,
		frame:     l.frame,
		selector:  l.selector,
		jsExpr:    l.jsExpr,
		jsAllExpr: l.jsAllExpr,
		desc:      l.desc,
		nth:       n,
	}
}

// Click waits for the element and clicks it.
func (l *Locator) Click(opts ...WaitOption) error {
	el, err := l.resolve(opts...)
	if err != nil {
		return err
	}
	return el.Click()
}

// Fill waits for the element and fills it with text.
func (l *Locator) Fill(text string, opts ...WaitOption) error {
	el, err := l.resolve(opts...)
	if err != nil {
		return err
	}
	return el.Fill(text)
}

// Type waits for the element and types text character by character.
func (l *Locator) Type(text string, opts ...TypeOption) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}
	el, err := l.resolve(cfg.waitOpts...)
	if err != nil {
		return err
	}
	return el.Type(text, opts...)
}

// Press waits for the element and presses a key.
func (l *Locator) Press(key string, opts ...WaitOption) error {
	el, err := l.resolve(opts...)
	if err != nil {
		return err
	}
	return el.Press(key)
}

// Text returns the text content of the element.
func (l *Locator) Text() (string, error) {
	el, err := l.resolve()
	if err != nil {
		return "", err
	}
	return el.Text()
}

// InnerText returns the rendered text content of the element.
func (l *Locator) InnerText() (string, error) {
	el, err := l.resolve()
	if err != nil {
		return "", err
	}
	return el.InnerText()
}

// HTML returns the outer HTML of the element.
func (l *Locator) HTML() (string, error) {
	el, err := l.resolve()
	if err != nil {
		return "", err
	}
	return el.HTML()
}

// Attribute returns the value of the named attribute.
func (l *Locator) Attribute(name string) (string, error) {
	el, err := l.resolve()
	if err != nil {
		return "", err
	}
	return el.Attribute(name)
}

// IsVisible reports whether the element is visible.
func (l *Locator) IsVisible() (bool, error) {
	el, err := l.queryDirect()
	if err != nil {
		return false, err
	}
	if el == nil {
		return false, nil
	}
	return el.IsVisible()
}

// BoundingBox returns the element's bounding box.
func (l *Locator) BoundingBox() (*Box, error) {
	el, err := l.resolve()
	if err != nil {
		return nil, err
	}
	return el.BoundingBox()
}

// Screenshot captures a screenshot of the element.
func (l *Locator) Screenshot(path string, opts ...ScreenshotOption) error {
	el, err := l.resolve()
	if err != nil {
		return err
	}
	return el.Screenshot(path, opts...)
}

// WaitFor waits until the element is attached to the DOM.
func (l *Locator) WaitFor(opts ...WaitOption) error {
	_, err := l.resolve(opts...)
	return err
}

// Count returns the number of elements matching the selector.
func (l *Locator) Count() (int, error) {
	if l.jsAllExpr != "" {
		return l.countJS()
	}
	if l.frame != nil {
		els, err := l.frame.QueryAll(l.selector)
		if err != nil {
			return 0, err
		}
		return len(els), nil
	}
	els, err := l.page.QueryAll(l.selector)
	if err != nil {
		return 0, err
	}
	return len(els), nil
}

func (l *Locator) description() string {
	if l.desc != "" {
		return l.desc
	}
	return l.selector
}

func (l *Locator) resolve(opts ...WaitOption) (*Element, error) {
	if l.nth >= 0 {
		return l.resolveNth(opts...)
	}
	if l.jsExpr != "" {
		return l.resolveJS(opts...)
	}
	if l.frame != nil {
		return l.frame.WaitSelector(l.selector, opts...)
	}
	return l.page.WaitSelector(l.selector, opts...)
}

func (l *Locator) resolveJS(opts ...WaitOption) (*Element, error) {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	result, err := poll(l.page.execCtx, cfg, func() (any, error) {
		var el *Element
		var qerr error
		if l.frame != nil {
			el, qerr = l.frame.queryJSHandle(l.jsExpr)
		} else {
			el, qerr = l.page.queryJSHandle(l.jsExpr)
		}
		if el == nil {
			return nil, qerr
		}
		return el, qerr
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, &TimeoutError{Selector: l.description()}
	}
	return result.(*Element), nil
}

func (l *Locator) resolveNth(opts ...WaitOption) (*Element, error) {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	result, err := poll(l.page.execCtx, cfg, func() (any, error) {
		els, err := l.queryAllElements()
		if err != nil {
			return nil, err
		}
		if l.nth >= len(els) {
			return nil, nil
		}
		return els[l.nth], nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, &TimeoutError{
			Selector: fmt.Sprintf("%s >> nth=%d", l.description(), l.nth),
		}
	}
	return result.(*Element), nil
}

func (l *Locator) queryDirect() (*Element, error) {
	if l.nth >= 0 {
		els, err := l.queryAllElements()
		if err != nil {
			return nil, err
		}
		if l.nth >= len(els) {
			return nil, nil
		}
		return els[l.nth], nil
	}
	if l.jsExpr != "" {
		if l.frame != nil {
			return l.frame.queryJSHandle(l.jsExpr)
		}
		return l.page.queryJSHandle(l.jsExpr)
	}
	if l.frame != nil {
		return l.frame.Query(l.selector)
	}
	return l.page.Query(l.selector)
}

func (l *Locator) queryAllElements() ([]*Element, error) {
	if l.jsAllExpr != "" {
		if l.frame != nil {
			return l.frame.queryAllJSHandles(l.jsAllExpr)
		}
		return l.page.queryAllJSHandles(l.jsAllExpr)
	}
	if l.frame != nil {
		return l.frame.QueryAll(l.selector)
	}
	return l.page.QueryAll(l.selector)
}

func (l *Locator) countJS() (int, error) {
	if l.frame != nil {
		els, err := l.frame.queryAllJSHandles(l.jsAllExpr)
		if err != nil {
			return 0, err
		}
		return len(els), nil
	}
	els, err := l.page.queryAllJSHandles(l.jsAllExpr)
	if err != nil {
		return 0, err
	}
	return len(els), nil
}
