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
					"element. Returns a PNG image.",
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

	selector := req.GetString("selector", "")
	if selector != "" {
		el, err := page.Query(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if el == nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("element %q not found", selector),
			), nil
		}
		data, err := el.ScreenshotBytes()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return mcp.NewToolResultImage(
			"", encoded, "image/png",
		), nil
	}

	var opts []bonk.ScreenshotOption
	if req.GetBool("full_page", false) {
		opts = append(opts, bonk.FullPage())
	}

	data, err := page.ScreenshotBytes(opts...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return mcp.NewToolResultImage(
		"", encoded, "image/png",
	), nil
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
