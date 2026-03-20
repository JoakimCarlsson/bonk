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
