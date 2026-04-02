package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerElementTools(
	s *server.MCPServer,
	sess *Session,
) {
	s.AddTool(
		mcp.NewTool("click",
			mcp.WithDescription("Click an element on the page."),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the element to click"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleClick,
	)

	s.AddTool(
		mcp.NewTool("fill",
			mcp.WithDescription(
				"Clear an input field and fill it with text.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the input"),
			),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Text to fill into the input"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleFill,
	)

	s.AddTool(
		mcp.NewTool("type_text",
			mcp.WithDescription(
				"Type text into an element character by character. "+
					"Unlike fill, this fires individual key events.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the element"),
			),
			mcp.WithString("text",
				mcp.Required(),
				mcp.Description("Text to type"),
			),
			mcp.WithNumber("delay_ms",
				mcp.Description(
					"Delay between keystrokes in milliseconds",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleTypeText,
	)

	s.AddTool(
		mcp.NewTool("select_option",
			mcp.WithDescription("Select a dropdown option by value."),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the select element"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("Value of the option to select"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleSelectOption,
	)

	s.AddTool(
		mcp.NewTool("check",
			mcp.WithDescription("Check a checkbox."),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the checkbox"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleCheck,
	)

	s.AddTool(
		mcp.NewTool("uncheck",
			mcp.WithDescription("Uncheck a checkbox."),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the checkbox"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleUncheck,
	)

	s.AddTool(
		mcp.NewTool("hover",
			mcp.WithDescription("Hover over an element."),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector of the element"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleHover,
	)

	s.AddTool(
		mcp.NewTool("upload",
			mcp.WithDescription(
				"Upload files to a file input element.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the file input",
				),
			),
			mcp.WithArray("paths",
				mcp.Required(),
				mcp.Description("File paths to upload"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleUpload,
	)

	s.AddTool(
		mcp.NewTool("is_checked",
			mcp.WithDescription(
				"Check if a checkbox or radio button is checked.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the element",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleIsChecked,
	)

	s.AddTool(
		mcp.NewTool("is_disabled",
			mcp.WithDescription(
				"Check if an element is disabled.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the element",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleIsDisabled,
	)

	s.AddTool(
		mcp.NewTool("is_editable",
			mcp.WithDescription(
				"Check if an element is editable "+
					"(input, textarea, select, or "+
					"contenteditable that is not "+
					"disabled or readonly).",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the element",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleIsEditable,
	)

	s.AddTool(
		mcp.NewTool("query",
			mcp.WithDescription(
				"Find an element and return its text, "+
					"visibility, and attributes.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector to query"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleQuery,
	)

	s.AddTool(
		mcp.NewTool("query_all",
			mcp.WithDescription(
				"Find all elements matching a selector. "+
					"Returns count and text content of each.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector to query"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleQueryAll,
	)

	s.AddTool(
		mcp.NewTool("wait_for_selector",
			mcp.WithDescription(
				"Wait for an element matching a selector "+
					"to appear in the page.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description("CSS selector to wait for"),
			),
			mcp.WithBoolean("visible",
				mcp.Description(
					"Wait for the element to be visible",
				),
			),
			mcp.WithBoolean("hidden",
				mcp.Description(
					"Wait for the element to be hidden",
				),
			),
			mcp.WithNumber("timeout_ms",
				mcp.Description(
					"Timeout in milliseconds (default 30000)",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleWaitForSelector,
	)

	s.AddTool(
		mcp.NewTool("dispatch_event",
			mcp.WithDescription(
				"Fire a DOM event on an element "+
					"programmatically. Useful when "+
					"simulated clicks don't trigger "+
					"framework-level handlers.",
			),
			mcp.WithString("selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the element",
				),
			),
			mcp.WithString("event_type",
				mcp.Required(),
				mcp.Description(
					"DOM event type (e.g. click, input, "+
						"change, submit)",
				),
			),
			mcp.WithObject("event_init",
				mcp.Description(
					"Event init options (e.g. "+
						"{\"bubbles\":true}). "+
						"Defaults to {bubbles: true}.",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleDispatchEvent,
	)
}

func (s *Session) handleClick(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Click(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Clicked %q", selector),
	), nil
}

func (s *Session) handleFill(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Fill(text); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Filled %q", selector),
	), nil
}

func (s *Session) handleTypeText(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	var opts []bonk.TypeOption
	delayMs := req.GetFloat("delay_ms", 0)
	if delayMs > 0 {
		opts = append(opts, bonk.WithDelay(
			time.Duration(delayMs)*time.Millisecond,
		))
	}

	if err := el.Type(text, opts...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Typed into %q", selector),
	), nil
}

func (s *Session) handleSelectOption(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.SelectOption(value); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(
			"Selected %q in %q", value, selector,
		),
	), nil
}

func (s *Session) handleCheck(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Check(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Checked %q", selector),
	), nil
}

func (s *Session) handleUncheck(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Uncheck(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Unchecked %q", selector),
	), nil
}

func (s *Session) handleIsChecked(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	checked, err := el.IsChecked()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("%t", checked),
	), nil
}

func (s *Session) handleIsDisabled(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	disabled, err := el.IsDisabled()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("%t", disabled),
	), nil
}

func (s *Session) handleIsEditable(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	editable, err := el.IsEditable()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("%t", editable),
	), nil
}

func (s *Session) handleHover(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Hover(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Hovered over %q", selector),
	), nil
}

func (s *Session) handleUpload(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	paths := req.GetStringSlice("paths", nil)
	if len(paths) == 0 {
		return mcp.NewToolResultError(
			"paths is required and must not be empty",
		), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	if err := el.Upload(paths...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(
			"Uploaded %d file(s) to %q",
			len(paths),
			selector,
		),
	), nil
}

type queryResult struct {
	Text    string `json:"text"`
	Visible bool   `json:"visible"`
}

func (s *Session) handleQuery(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	el, err := page.Query(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if el == nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("element %q not found", selector),
		), nil
	}

	text, _ := el.Text()
	visible, _ := el.IsVisible()

	data, _ := json.MarshalIndent(queryResult{
		Text:    text,
		Visible: visible,
	}, "", "  ")

	return mcp.NewToolResultText(string(data)), nil
}

type queryAllItem struct {
	Index   int    `json:"index"`
	Text    string `json:"text"`
	Visible bool   `json:"visible"`
}

func (s *Session) handleQueryAll(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	elements, err := page.QueryAll(selector)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	items := make([]queryAllItem, 0, len(elements))
	for i, el := range elements {
		text, _ := el.Text()
		visible, _ := el.IsVisible()
		items = append(items, queryAllItem{
			Index:   i,
			Text:    text,
			Visible: visible,
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return mcp.NewToolResultText(
		fmt.Sprintf("Found %d element(s):\n%s", len(items), data),
	), nil
}

func (s *Session) handleWaitForSelector(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var opts []bonk.WaitOption
	timeoutMs := req.GetFloat("timeout_ms", 0)
	if timeoutMs > 0 {
		opts = append(opts, bonk.WaitTimeout(
			time.Duration(timeoutMs)*time.Millisecond,
		))
	}
	if req.GetBool("visible", false) {
		opts = append(opts, bonk.WaitVisibleOption())
	}
	if req.GetBool("hidden", false) {
		opts = append(opts, bonk.WaitHiddenOption())
	}

	el, err := page.WaitSelector(selector, opts...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, _ := el.Text()
	visible, _ := el.IsVisible()

	data, _ := json.MarshalIndent(queryResult{
		Text:    text,
		Visible: visible,
	}, "", "  ")

	return mcp.NewToolResultText(
		fmt.Sprintf("Element %q found:\n%s", selector, data),
	), nil
}

func (s *Session) handleDispatchEvent(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	eventType, err := req.RequireString("event_type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var initArg []map[string]any
	args := req.GetArguments()
	if raw, ok := args["event_init"]; ok {
		if m, ok := raw.(map[string]any); ok {
			initArg = append(initArg, m)
		}
	}

	if err := page.DispatchEvent(
		selector, eventType, initArg...,
	); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(
			"Dispatched %q on %q", eventType, selector,
		),
	), nil
}
