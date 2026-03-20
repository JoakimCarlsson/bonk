package bonk

import (
	"encoding/json"
	"fmt"

	"github.com/joakimcarlsson/bonk/proto"
)

// Evaluate executes a JavaScript expression and returns the result as a Go value.
func (p *Page) Evaluate(expression string) (any, error) {
	res, err := proto.RuntimeEvaluate(expression).
		WithReturnByValue(true).
		WithAwaitPromise(true).
		Do(p.execCtx)
	if err != nil {
		return nil, err
	}
	if res.ExceptionDetails.ExceptionID != 0 {
		return nil, fmt.Errorf(
			"bonk: js error: %s",
			res.ExceptionDetails.Text,
		)
	}
	if len(res.Result.Value) == 0 {
		return nil, nil
	}
	var val any
	if err := json.Unmarshal(res.Result.Value, &val); err != nil {
		return nil, err
	}
	return val, nil
}

// Content returns the full HTML content of the page.
func (p *Page) Content() (string, error) {
	val, err := p.Evaluate("document.documentElement.outerHTML")
	if err != nil {
		return "", err
	}
	s, _ := val.(string)
	return s, nil
}

// WaitForFunction waits until the given JavaScript expression evaluates to a truthy value.
func (p *Page) WaitForFunction(expression string, opts ...WaitOption) error {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	result, err := poll(p.execCtx, cfg, func() (any, error) {
		val, err := p.Evaluate(expression)
		if err != nil {
			return nil, err
		}
		if isTruthy(val) {
			return val, nil
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	if result == nil {
		return &TimeoutError{Operation: "waiting for function"}
	}
	return nil
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}

// EvaluateHandle executes a JavaScript expression and returns the result
// as an Element handle (for non-primitive return values like DOM nodes).
func (p *Page) EvaluateHandle(expression string) (*Element, error) {
	res, err := proto.RuntimeEvaluate(expression).
		WithReturnByValue(false).
		WithAwaitPromise(true).
		Do(p.execCtx)
	if err != nil {
		return nil, err
	}
	if res.ExceptionDetails.ExceptionID != 0 {
		return nil, fmt.Errorf(
			"bonk: js error: %s",
			res.ExceptionDetails.Text,
		)
	}
	if res.Result.ObjectID == "" {
		return nil, nil
	}
	return &Element{page: p, objectID: res.Result.ObjectID}, nil
}

// EvaluateOn executes a JavaScript function with the element as `this`.
// The function should be in the form "function(arg1, arg2) { ... }".
func (p *Page) EvaluateOn(el *Element, fn string, args ...any) (any, error) {
	return el.callForValue(fn, args...)
}
