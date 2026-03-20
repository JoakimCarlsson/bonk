package bonk

import (
	"encoding/json"
	"fmt"

	"github.com/joakimcarlsson/bonk/proto"
)

// Text returns the text content of the element.
func (e *Element) Text() (string, error) {
	return e.callForString("function(){return this.textContent}")
}

// HTML returns the outer HTML of the element.
func (e *Element) HTML() (string, error) {
	return e.callForString("function(){return this.outerHTML}")
}

// Attribute returns the value of the named attribute.
func (e *Element) Attribute(name string) (string, error) {
	return e.callForString(
		fmt.Sprintf("function(){return this.getAttribute(%s)}", jsString(name)),
	)
}

// IsVisible reports whether the element is visible.
func (e *Element) IsVisible() (bool, error) {
	res, err := e.callForValue(
		"function(){" +
			"var s=window.getComputedStyle(this);" +
			"return s.display!=='none'&&s.visibility!=='hidden'&&s.opacity!=='0'" +
			"}",
	)
	if err != nil {
		return false, err
	}
	b, _ := res.(bool)
	return b, nil
}

// BoundingBox returns the element's bounding rectangle.
type Box struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// BoundingBox returns the element's bounding box in viewport coordinates.
func (e *Element) BoundingBox() (*Box, error) {
	res, err := e.callForValue(
		"function(){var r=this.getBoundingClientRect();" +
			"return {x:r.x,y:r.y,width:r.width,height:r.height}}",
	)
	if err != nil {
		return nil, err
	}
	m, ok := res.(map[string]any)
	if !ok {
		return nil, nil
	}
	return &Box{
		X:      toFloat(m["x"]),
		Y:      toFloat(m["y"]),
		Width:  toFloat(m["width"]),
		Height: toFloat(m["height"]),
	}, nil
}

func (e *Element) scrollIntoView() error {
	_, err := proto.RuntimeCallFunctionOn(
		"function(){this.scrollIntoViewIfNeeded(true)}",
	).WithObjectID(e.objectID).
		WithReturnByValue(true).
		Do(e.page.execCtx)
	return err
}

func (e *Element) callForString(fn string) (string, error) {
	val, err := e.callForValue(fn)
	if err != nil {
		return "", err
	}
	s, _ := val.(string)
	return s, nil
}

func (e *Element) callForValue(fn string) (any, error) {
	res, err := proto.RuntimeCallFunctionOn(fn).
		WithObjectID(e.objectID).
		WithReturnByValue(true).
		Do(e.page.execCtx)
	if err != nil {
		return nil, err
	}
	if res.ExceptionDetails.ExceptionID != 0 {
		return nil, fmt.Errorf("bonk: js error: %s", res.ExceptionDetails.Text)
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

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}
