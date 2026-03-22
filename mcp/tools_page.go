package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

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
					"page structure before clicking or filling.",
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

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	nodes, err := page.AccessibilityTree()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text := bonk.FormatAccessibilityTree(nodes)
	if text == "" {
		text = "(empty accessibility tree)"
	}
	return mcp.NewToolResultText(text), nil
}
