package bonk

import (
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// Click scrolls the element into view and clicks it.
func (e *Element) Click() error {
	box, err := e.prepare()
	if err != nil {
		return err
	}
	cx := box.X + box.Width/2
	cy := box.Y + box.Height/2

	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMouseMoved, cx, cy, 0); err != nil {
		return err
	}
	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMousePressed, cx, cy, 1); err != nil {
		return err
	}
	return e.mouseEvent(
		proto.InputDispatchMouseEventTypeMouseReleased,
		cx,
		cy,
		1,
	)
}

// DoubleClick scrolls the element into view and double-clicks it.
func (e *Element) DoubleClick() error {
	box, err := e.prepare()
	if err != nil {
		return err
	}
	cx := box.X + box.Width/2
	cy := box.Y + box.Height/2

	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMouseMoved, cx, cy, 0); err != nil {
		return err
	}
	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMousePressed, cx, cy, 1); err != nil {
		return err
	}
	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMouseReleased, cx, cy, 1); err != nil {
		return err
	}
	if err := e.mouseEvent(proto.InputDispatchMouseEventTypeMousePressed, cx, cy, 2); err != nil {
		return err
	}
	return e.mouseEvent(
		proto.InputDispatchMouseEventTypeMouseReleased,
		cx,
		cy,
		2,
	)
}

// Hover scrolls the element into view and moves the mouse to it.
func (e *Element) Hover() error {
	box, err := e.prepare()
	if err != nil {
		return err
	}
	cx := box.X + box.Width/2
	cy := box.Y + box.Height/2
	return e.mouseEvent(proto.InputDispatchMouseEventTypeMouseMoved, cx, cy, 0)
}

// Fill clears the input field and inserts the given text.
func (e *Element) Fill(text string) error {
	if err := e.focus(); err != nil {
		return err
	}
	if _, err := e.callForValue("function(){this.select()}"); err != nil {
		return err
	}
	return proto.InputInsertText(text).Do(e.page.execCtx)
}

// Type types the given text character by character with key events.
func (e *Element) Type(text string, opts ...TypeOption) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}

	if err := e.focus(); err != nil {
		return err
	}

	for _, ch := range text {
		s := string(ch)
		if err := proto.InputDispatchKeyEvent(proto.InputDispatchKeyEventTypeKeyDown).
			WithText(s).
			WithKey(s).
			Do(e.page.execCtx); err != nil {
			return err
		}
		if err := proto.InputDispatchKeyEvent(proto.InputDispatchKeyEventTypeKeyUp).
			WithKey(s).
			Do(e.page.execCtx); err != nil {
			return err
		}
		if cfg.delay > 0 {
			time.Sleep(cfg.delay)
		}
	}
	return nil
}

// Press sends a single key press event.
func (e *Element) Press(key string) error {
	if err := e.focus(); err != nil {
		return err
	}
	if err := proto.InputDispatchKeyEvent(proto.InputDispatchKeyEventTypeKeyDown).
		WithKey(key).
		Do(e.page.execCtx); err != nil {
		return err
	}
	return proto.InputDispatchKeyEvent(proto.InputDispatchKeyEventTypeKeyUp).
		WithKey(key).
		Do(e.page.execCtx)
}

// SelectOption selects an option in a <select> element by value.
func (e *Element) SelectOption(value string) error {
	_, err := e.callForValue(
		`function(v){`+
			`this.value=v;`+
			`this.dispatchEvent(new Event('input',{bubbles:true}));`+
			`this.dispatchEvent(new Event('change',{bubbles:true}))`+
			`}`,
		value,
	)
	return err
}

// Check checks a checkbox or radio button if not already checked.
func (e *Element) Check() error {
	checked, err := e.callForValue("function(){return this.checked}")
	if err != nil {
		return err
	}
	if b, _ := checked.(bool); !b {
		return e.Click()
	}
	return nil
}

// Uncheck unchecks a checkbox if currently checked.
func (e *Element) Uncheck() error {
	checked, err := e.callForValue("function(){return this.checked}")
	if err != nil {
		return err
	}
	if b, _ := checked.(bool); b {
		return e.Click()
	}
	return nil
}

// Upload sets files on a file input element.
func (e *Element) Upload(paths ...string) error {
	res, err := proto.DOMDescribeNode().
		WithObjectID(e.objectID).
		Do(e.page.execCtx)
	if err != nil {
		return err
	}
	return proto.DOMSetFileInputFiles(paths).
		WithBackendNodeID(res.Node.BackendNodeID).
		Do(e.page.execCtx)
}

// TypeOption configures typing behavior.
type TypeOption func(*typeConfig)

type typeConfig struct {
	delay    time.Duration
	waitOpts []WaitOption
}

func defaultTypeConfig() *typeConfig {
	return &typeConfig{}
}

// WithDelay sets the delay between keystrokes.
func WithDelay(d time.Duration) TypeOption {
	return func(c *typeConfig) {
		c.delay = d
	}
}

// WaitFor adds wait options that apply when page.Type() waits for the selector.
func WaitFor(opts ...WaitOption) TypeOption {
	return func(c *typeConfig) {
		c.waitOpts = append(c.waitOpts, opts...)
	}
}

func (e *Element) prepare() (*Box, error) {
	if err := e.waitVisible(); err != nil {
		return nil, err
	}
	if err := e.scrollIntoView(); err != nil {
		return nil, err
	}
	box, err := e.BoundingBox()
	if err != nil {
		return nil, err
	}
	if box == nil {
		return nil, &ElementNotFoundError{Selector: "(detached element)"}
	}
	return box, nil
}

// WaitForVisible waits until the element becomes visible.
func (e *Element) WaitForVisible(opts ...WaitOption) error {
	cfg := &waitConfig{
		timeout:  e.page.defaultTimeout,
		interval: 50 * time.Millisecond,
	}
	for _, o := range opts {
		o(cfg)
	}
	_, err := poll(e.page.execCtx, cfg, func() (any, error) {
		visible, err := e.IsVisible()
		if err != nil {
			return nil, err
		}
		if !visible {
			return nil, nil
		}
		return true, nil
	})
	return err
}

// WaitForHidden waits until the element becomes hidden.
func (e *Element) WaitForHidden(opts ...WaitOption) error {
	cfg := &waitConfig{
		timeout:  e.page.defaultTimeout,
		interval: 50 * time.Millisecond,
	}
	for _, o := range opts {
		o(cfg)
	}
	_, err := poll(e.page.execCtx, cfg, func() (any, error) {
		visible, err := e.IsVisible()
		if err != nil {
			return nil, err
		}
		if visible {
			return nil, nil
		}
		return true, nil
	})
	return err
}

func (e *Element) waitVisible() error {
	return e.WaitForVisible()
}

// Focus focuses the element.
func (e *Element) Focus() error { return e.focus() }

func (e *Element) focus() error {
	_, err := e.callForValue("function(){this.focus()}")
	return err
}

func (e *Element) mouseEvent(
	typ proto.InputDispatchMouseEventType,
	x, y float64,
	clickCount int64,
) error {
	p := proto.InputDispatchMouseEvent(typ, x, y).
		WithButton(proto.InputMouseButtonLeft).
		WithClickCount(clickCount)
	return p.Do(e.page.execCtx)
}
