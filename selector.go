package bonk

import (
	"fmt"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

// Query finds the first element matching the CSS selector.
// Returns nil if no element matches.
func (p *Page) Query(selector string) (*Element, error) {
	res, err := proto.RuntimeEvaluate(queryJS(selector)).
		WithReturnByValue(false).
		Do(p.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Result.ObjectID == "" {
		return nil, nil
	}
	return &Element{
		page:     p,
		objectID: res.Result.ObjectID,
		selector: selector,
	}, nil
}

// QueryAll finds all elements matching the CSS selector.
func (p *Page) QueryAll(selector string) ([]*Element, error) {
	res, err := proto.RuntimeEvaluate(queryAllJS(selector)).
		WithReturnByValue(false).
		Do(p.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Result.ObjectID == "" {
		return nil, nil
	}

	props, err := proto.RuntimeGetProperties(res.Result.ObjectID).
		WithOwnProperties(true).
		Do(p.execCtx)
	if err != nil {
		return nil, err
	}

	var elements []*Element
	for _, prop := range props.Result {
		if prop.Value.ObjectID == "" {
			continue
		}
		if prop.Name == "length" || prop.Name == "__proto__" {
			continue
		}
		elements = append(elements, &Element{
			page:     p,
			objectID: prop.Value.ObjectID,
			selector: selector,
		})
	}
	return elements, nil
}

// WaitSelector waits for an element matching the selector to appear.
// By default waits for the element to be attached to the DOM.
// Use WaitVisibleOption() to also require visibility,
// or WaitHiddenOption() to wait until the element is hidden or removed.
func (p *Page) WaitSelector(
	selector string,
	opts ...WaitOption,
) (*Element, error) {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	switch cfg.visibility {
	case WaitUntilHidden:
		result, err := poll(p.execCtx, cfg, func() (any, error) {
			el, err := p.Query(selector)
			if err != nil {
				return nil, err
			}
			if el == nil {
				return true, nil
			}
			visible, err := el.IsVisible()
			if err != nil {
				return nil, err
			}
			if !visible {
				return true, nil
			}
			return nil, nil
		})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, &TimeoutError{Selector: selector}
		}
		return nil, nil

	case WaitUntilVisible:
		result, err := poll(p.execCtx, cfg, func() (any, error) {
			el, err := p.Query(selector)
			if err != nil {
				return nil, err
			}
			if el == nil {
				return nil, nil
			}
			visible, err := el.IsVisible()
			if err != nil {
				return nil, err
			}
			if !visible {
				return nil, nil
			}
			return el, nil
		})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, &TimeoutError{Selector: selector}
		}
		return result.(*Element), nil

	default:
		result, err := poll(p.execCtx, cfg, func() (any, error) {
			el, err := p.Query(selector)
			if el == nil {
				return nil, err
			}
			return el, err
		})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, &TimeoutError{Selector: selector}
		}
		return result.(*Element), nil
	}
}

func queryJS(selector string) string {
	return fmt.Sprintf("document.querySelector(%s)", jsString(selector))
}

func queryAllJS(selector string) string {
	return fmt.Sprintf("document.querySelectorAll(%s)", jsString(selector))
}

func jsString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return "'" + s + "'"
}
