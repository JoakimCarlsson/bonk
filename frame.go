package bonk

import (
	"encoding/json"
	"fmt"

	"github.com/joakimcarlsson/bonk/proto"
)

// Frame represents an iframe within a page.
type Frame struct {
	page      *Page
	id        proto.FrameID
	name      string
	url       string
	contextID proto.RuntimeExecutionContextID
}

// Frames returns all frames in the page's frame tree.
func (p *Page) Frames() ([]*Frame, error) {
	res, err := proto.PageGetFrameTree().Do(p.execCtx)
	if err != nil {
		return nil, err
	}

	var frames []*Frame
	walkFrameTree(&res.FrameTree, p, &frames)
	return frames, nil
}

// Frame finds a frame by name or ID.
func (p *Page) Frame(nameOrID string) (*Frame, error) {
	frames, err := p.Frames()
	if err != nil {
		return nil, err
	}
	for _, f := range frames {
		if f.name == nameOrID || string(f.id) == nameOrID {
			return f, nil
		}
	}
	return nil, fmt.Errorf("bonk: frame %q not found", nameOrID)
}

// Query finds the first element matching the CSS selector within this frame.
func (f *Frame) Query(selector string) (*Element, error) {
	ctx, err := f.ensureContext()
	if err != nil {
		return nil, err
	}

	res, err := proto.RuntimeEvaluate(queryJS(selector)).
		WithContextID(ctx).
		WithReturnByValue(false).
		Do(f.page.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Result.ObjectID == "" {
		return nil, nil
	}
	return &Element{page: f.page, objectID: res.Result.ObjectID}, nil
}

// QueryAll finds all elements matching the CSS selector within this frame.
func (f *Frame) QueryAll(selector string) ([]*Element, error) {
	ctx, err := f.ensureContext()
	if err != nil {
		return nil, err
	}

	res, err := proto.RuntimeEvaluate(queryAllJS(selector)).
		WithContextID(ctx).
		WithReturnByValue(false).
		Do(f.page.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Result.ObjectID == "" {
		return nil, nil
	}

	props, err := proto.RuntimeGetProperties(res.Result.ObjectID).
		WithOwnProperties(true).
		Do(f.page.execCtx)
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
			page:     f.page,
			objectID: prop.Value.ObjectID,
		})
	}
	return elements, nil
}

// WaitSelector waits for an element matching the selector within this frame.
func (f *Frame) WaitSelector(
	selector string,
	opts ...WaitOption,
) (*Element, error) {
	cfg := defaultWaitConfig()
	for _, o := range opts {
		o(cfg)
	}

	result, err := poll(f.page.execCtx, cfg, func() (any, error) {
		el, err := f.Query(selector)
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

// Evaluate executes JavaScript in this frame's context.
func (f *Frame) Evaluate(expression string) (any, error) {
	ctx, err := f.ensureContext()
	if err != nil {
		return nil, err
	}

	res, err := proto.RuntimeEvaluate(expression).
		WithContextID(ctx).
		WithReturnByValue(true).
		WithAwaitPromise(true).
		Do(f.page.execCtx)
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

// Click waits for the selector in this frame and clicks the element.
func (f *Frame) Click(selector string, opts ...WaitOption) error {
	el, err := f.WaitSelector(selector, opts...)
	if err != nil {
		return err
	}
	return el.Click()
}

// Fill waits for the selector in this frame and fills the input.
func (f *Frame) Fill(selector, text string, opts ...WaitOption) error {
	el, err := f.WaitSelector(selector, opts...)
	if err != nil {
		return err
	}
	return el.Fill(text)
}

// Name returns the frame's name attribute.
func (f *Frame) Name() string {
	return f.name
}

// URL returns the frame's document URL.
func (f *Frame) URL() string {
	return f.url
}

// ID returns the frame's unique identifier.
func (f *Frame) ID() proto.FrameID {
	return f.id
}

func (f *Frame) ensureContext() (proto.RuntimeExecutionContextID, error) {
	if f.contextID != 0 {
		return f.contextID, nil
	}

	res, err := proto.PageCreateIsolatedWorld(f.id).
		WithGrantUniveralAccess(true).
		Do(f.page.execCtx)
	if err != nil {
		return 0, err
	}

	f.contextID = res.ExecutionContextID
	return f.contextID, nil
}

func walkFrameTree(
	tree *proto.PageFrameTree,
	page *Page,
	out *[]*Frame,
) {
	*out = append(*out, &Frame{
		page: page,
		id:   tree.Frame.ID,
		name: tree.Frame.Name,
		url:  tree.Frame.URL,
	})
	for i := range tree.ChildFrames {
		walkFrameTree(&tree.ChildFrames[i], page, out)
	}
}
