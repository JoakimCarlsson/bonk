package bonk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// Screenshot captures a screenshot of the element and saves it to the given path.
func (e *Element) Screenshot(path string, opts ...ScreenshotOption) error {
	if err := e.scrollIntoView(); err != nil {
		return err
	}
	box, err := e.BoundingBox()
	if err != nil {
		return err
	}
	if box == nil {
		return &ElementNotFoundError{Selector: "(detached element)"}
	}

	cfg := &screenshotConfig{}
	for _, o := range opts {
		o(cfg)
	}

	clip := proto.PageViewport{
		X:      box.X,
		Y:      box.Y,
		Width:  box.Width,
		Height: box.Height,
		Scale:  1,
	}

	params := proto.PageCaptureScreenshot().WithClip(clip)

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		params = params.WithFormat(proto.PageCaptureScreenshotFormatJpeg)
		if cfg.quality > 0 {
			params = params.WithQuality(cfg.quality)
		}
	case ".webp":
		params = params.WithFormat(proto.PageCaptureScreenshotFormatWebp)
		if cfg.quality > 0 {
			params = params.WithQuality(cfg.quality)
		}
	default:
		params = params.WithFormat(proto.PageCaptureScreenshotFormatPng)
	}

	res, err := params.Do(e.page.execCtx)
	if err != nil {
		return err
	}

	return os.WriteFile(path, res.Data, 0o644)
}

// ScrollIntoView scrolls the element into view.
func (e *Element) ScrollIntoView() error { return e.scrollIntoView() }

// InnerText returns the rendered text content of the element.
func (e *Element) InnerText() (string, error) {
	return e.callForString("function(){return this.innerText}")
}

func (e *Element) scrollIntoView() error {
	_, err := e.callForValue("function(){this.scrollIntoViewIfNeeded(true)}")
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

func (e *Element) callForValue(fn string, args ...any) (any, error) {
	val, err := e.doCallForValue(fn, args...)
	if err != nil && isStaleError(err) && e.selector != "" {
		if rerr := e.reResolve(); rerr != nil {
			return nil, ErrStaleElement
		}
		return e.doCallForValue(fn, args...)
	}
	return val, err
}

func (e *Element) doCallForValue(fn string, args ...any) (any, error) {
	params := proto.RuntimeCallFunctionOn(fn).
		WithObjectID(e.objectID).
		WithReturnByValue(true)

	if len(args) > 0 {
		var callArgs []proto.RuntimeCallArgument
		for _, arg := range args {
			raw, err := json.Marshal(arg)
			if err != nil {
				return nil, fmt.Errorf("bonk: marshal arg: %w", err)
			}
			callArgs = append(callArgs, proto.RuntimeCallArgument{
				Value: json.RawMessage(raw),
			})
		}
		params = params.WithArguments(callArgs)
	}

	res, err := params.Do(e.page.execCtx)
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

func (e *Element) reResolve() error {
	el, err := e.page.Query(e.selector)
	if err != nil {
		return err
	}
	if el == nil {
		return ErrStaleElement
	}
	e.objectID = el.objectID
	return nil
}

func isStaleError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Could not find object") ||
		strings.Contains(msg, "Cannot find context") ||
		strings.Contains(msg, "Object reference chain")
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
