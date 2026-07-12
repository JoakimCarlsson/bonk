package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPageTools(s *server.MCPServer, sess *Session) {
	s.AddTool(
		mcp.NewTool("screenshot",
			mcp.WithDescription(
				"Take a screenshot of the page or a specific "+
					"element. Defaults to JPEG at max 1280px "+
					"width for compact responses.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
			mcp.WithBoolean("full_page",
				mcp.Description(
					"Capture the full scrollable page",
				),
			),
			mcp.WithString("selector",
				mcp.Description(
					"CSS selector of a specific element "+
						"to screenshot",
				),
			),
			mcp.WithNumber("max_width",
				mcp.Description(
					"Maximum image width in pixels. "+
						"Defaults to 1280.",
				),
			),
			mcp.WithString("format",
				mcp.Description(
					"Image format: jpeg or png. "+
						"Defaults to jpeg.",
				),
				mcp.Enum("jpeg", "png"),
			),
		),
		sess.handleScreenshot,
	)

	s.AddTool(
		mcp.NewTool("pdf",
			mcp.WithDescription(
				"Save the current page as a PDF. "+
					"Returns base64-encoded PDF data.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handlePDF,
	)

	s.AddTool(
		mcp.NewTool("get_content",
			mcp.WithDescription(
				"Get the full HTML content of the page.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleGetContent,
	)

	s.AddTool(
		mcp.NewTool("evaluate",
			mcp.WithDescription(
				"Execute JavaScript in the page and "+
					"return the result.",
			),
			mcp.WithString("expression",
				mcp.Required(),
				mcp.Description("JavaScript expression to evaluate"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleEvaluate,
	)

	s.AddTool(
		mcp.NewTool("snapshot",
			mcp.WithDescription(
				"Get the accessibility tree of the page. "+
					"Returns an indexed, structured view of "+
					"all elements. Interactive elements get "+
					"numbered indices. Use this to understand "+
					"page structure before clicking or filling. "+
					"Click a numbered element directly with "+
					"click_ref.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleSnapshot,
	)

	s.AddTool(
		mcp.NewTool("click_ref",
			mcp.WithDescription(
				"Click the interactive element at the given "+
					"index from the most recent snapshot of "+
					"the page (the number shown as [N]). More "+
					"reliable than a selector for elements read "+
					"from the snapshot.",
			),
			mcp.WithNumber("ref",
				mcp.Required(),
				mcp.Description(
					"The [N] index of the element from the "+
						"latest snapshot.",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleClickRef,
	)

	s.AddTool(
		mcp.NewTool("pause",
			mcp.WithDescription(
				"Pause execution for manual inspection in "+
					"headed mode. Resumes when the user "+
					"presses Enter on stdin.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handlePause,
	)

	s.AddTool(
		mcp.NewTool("add_script_tag",
			mcp.WithDescription(
				"Inject a <script> tag into the page. "+
					"Provide either url or content.",
			),
			mcp.WithString("url",
				mcp.Description("URL of the script to load"),
			),
			mcp.WithString("content",
				mcp.Description("Inline script content"),
			),
			mcp.WithString("type",
				mcp.Description(
					"Script type attribute (e.g. module)",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleAddScriptTag,
	)

	s.AddTool(
		mcp.NewTool("add_style_tag",
			mcp.WithDescription(
				"Inject a <style> or <link> stylesheet "+
					"tag into the page. Provide either "+
					"url or content.",
			),
			mcp.WithString("url",
				mcp.Description(
					"URL of the stylesheet to load",
				),
			),
			mcp.WithString("content",
				mcp.Description("Inline CSS content"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleAddStyleTag,
	)

	s.AddTool(
		mcp.NewTool("wait_for_event",
			mcp.WithDescription(
				"Wait for the next occurrence of a page "+
					"event. Returns the event payload.",
			),
			mcp.WithString("event",
				mcp.Required(),
				mcp.Description("Event type to wait for"),
				mcp.Enum("console", "dialog", "download"),
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
		sess.handleWaitForEvent,
	)
}

func (s *Session) handleScreenshot(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format := req.GetString("format", "jpeg")
	maxWidth := req.GetInt("max_width", 1280)
	mime := "image/jpeg"

	var opts []bonk.ScreenshotOption
	if format != "png" {
		opts = append(opts, bonk.AsJPEG())
	} else {
		mime = "image/png"
	}
	if maxWidth > 0 {
		opts = append(opts, bonk.MaxWidth(maxWidth))
	}

	selector := req.GetString("selector", "")
	if selector != "" {
		el, err := page.Query(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if el == nil {
			return mcp.NewToolResultError(
				fmt.Sprintf(
					"element %q not found", selector,
				),
			), nil
		}
		data, err := el.ScreenshotBytes(opts...)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return mcp.NewToolResultImage(
			"", encoded, mime,
		), nil
	}

	if req.GetBool("full_page", false) {
		opts = append(opts, bonk.FullPage())
	}

	data, err := page.ScreenshotBytes(opts...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return mcp.NewToolResultImage("", encoded, mime), nil
}

func (s *Session) handlePDF(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := page.PDFBytes()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return mcp.NewToolResultText(
		fmt.Sprintf("base64-encoded PDF (%d bytes):\n%s", len(data), encoded),
	), nil
}

func (s *Session) handleGetContent(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	html, err := page.Content()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(html), nil
}

func (s *Session) handleEvaluate(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expr, err := req.RequireString("expression")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := page.Evaluate(expr)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Session) handleSnapshot(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, id, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	nodes, err := page.AccessibilityTree()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, refs := bonk.FormatAccessibilityTreeIndexed(nodes)
	s.snapshots[id] = refs
	if text == "" {
		text = "(empty accessibility tree)"
	}
	return mcp.NewToolResultText(text), nil
}

func (s *Session) handleClickRef(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ref := int(req.GetFloat("ref", 0))

	page, id, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	refs := s.snapshots[id]
	if len(refs) == 0 {
		return mcp.NewToolResultError(
			"no snapshot for this page; call snapshot first",
		), nil
	}
	if ref < 1 || ref > len(refs) {
		return mcp.NewToolResultError(
			fmt.Sprintf("ref %d out of range 1..%d", ref, len(refs)),
		), nil
	}

	if err := page.ClickNode(refs[ref-1]); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("Clicked [%d]", ref),
	), nil
}

func (s *Session) handlePause(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := page.Pause(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Resumed"), nil
}

func (s *Session) handleAddScriptTag(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var opts []bonk.ScriptTagOption
	if u := req.GetString("url", ""); u != "" {
		opts = append(opts, bonk.ScriptTagURL(u))
	}
	if c := req.GetString("content", ""); c != "" {
		opts = append(opts, bonk.ScriptTagContent(c))
	}
	if t := req.GetString("type", ""); t != "" {
		opts = append(opts, bonk.ScriptTagType(t))
	}

	if err := page.AddScriptTag(opts...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		"Script tag added",
	), nil
}

func (s *Session) handleAddStyleTag(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var opts []bonk.StyleTagOption
	if u := req.GetString("url", ""); u != "" {
		opts = append(opts, bonk.StyleTagURL(u))
	}
	if c := req.GetString("content", ""); c != "" {
		opts = append(opts, bonk.StyleTagContent(c))
	}

	if err := page.AddStyleTag(opts...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		"Style tag added",
	), nil
}

func parseEventType(
	val string,
) bonk.EventType {
	switch val {
	case "dialog":
		return bonk.DialogEvent
	case "download":
		return bonk.DownloadEvent
	default:
		return bonk.ConsoleEvent
	}
}

func (s *Session) handleWaitForEvent(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, err := req.RequireString("event")
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

	result, err := page.WaitForEvent(
		parseEventType(event), opts...,
	)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
