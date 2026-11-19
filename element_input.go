package bonk

import (
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// Click scrolls the element into view and clicks it, moving the pointer
// there along a human path on a stealth page. It fails when the element
// cannot be found or shown, or a CDP command fails.
func (e *Element) Click() error {
	return e.clickN(1)
}

// DoubleClick scrolls the element into view and double-clicks it. It fails
// like Click.
func (e *Element) DoubleClick() error {
	return e.clickN(2)
}

// clickN scrolls the element into view and clicks it count times at a
// point inside it. It fails like Click.
func (e *Element) clickN(count int64) error {
	box, err := e.prepare()
	if err != nil {
		return err
	}
	x, y := e.page.targetPoint(box)
	return e.page.clickAt(x, y, count)
}

// Hover scrolls the element into view and moves the mouse to it. It fails
// like Click.
func (e *Element) Hover() error {
	box, err := e.prepare()
	if err != nil {
		return err
	}
	x, y := e.page.targetPoint(box)
	return e.page.moveCursor(x, y)
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

// Clear clears the input field contents.
func (e *Element) Clear() error {
	return e.Fill("")
}

// Type focuses the element and types the given text character by character
// with full key events, pausing like a person between keys on a stealth
// page unless WithDelay sets the gap. It returns the first error.
func (e *Element) Type(text string, opts ...TypeOption) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}
	if err := e.focus(); err != nil {
		return err
	}
	return e.page.typeText(text, cfg.delay)
}

// Press focuses the element and presses one key, given by a name such as
// "Enter" or a single character. It returns the first error.
func (e *Element) Press(key string) error {
	if err := e.focus(); err != nil {
		return err
	}
	return e.page.pressKey(key)
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

// DispatchEvent fires a DOM event on the element programmatically.
// The eventInit map is forwarded to the Event constructor; if empty,
// {bubbles: true} is used as the default.
func (e *Element) DispatchEvent(
	eventType string,
	eventInit ...map[string]any,
) error {
	var init map[string]any
	if len(eventInit) > 0 && eventInit[0] != nil {
		init = eventInit[0]
	} else {
		init = map[string]any{"bubbles": true}
	}
	_, err := e.callForValue(
		`function(t,i){this.dispatchEvent(new Event(t,i))}`,
		eventType,
		init,
	)
	return err
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
		timeout:  e.page.resolveTimeout(),
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
		timeout:  e.page.resolveTimeout(),
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

// Blur removes focus from the element.
func (e *Element) Blur() error {
	_, err := e.callForValue("function(){this.blur()}")
	return err
}

func (e *Element) focus() error {
	_, err := e.callForValue("function(){this.focus()}")
	return err
}
