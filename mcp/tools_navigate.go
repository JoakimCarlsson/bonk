package mcp

import (
	"context"
	"fmt"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerNavigateTools(
	s *server.MCPServer,
	sess *Session,
) {
	s.AddTool(
		mcp.NewTool("navigate",
			mcp.WithDescription(
				"Navigate a browser page to a URL. "+
					"Returns the page title and final URL.",
			),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("The URL to navigate to"),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
			mcp.WithString("wait_until",
				mcp.Description(
					"When to consider navigation complete",
				),
				mcp.Enum(
					"load",
					"domcontentloaded",
					"networkidle",
				),
			),
		),
		sess.handleNavigate,
	)

	s.AddTool(
		mcp.NewTool("go_back",
			mcp.WithDescription(
				"Navigate back in browser history.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleGoBack,
	)

	s.AddTool(
		mcp.NewTool("go_forward",
			mcp.WithDescription(
				"Navigate forward in browser history.",
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleGoForward,
	)

	s.AddTool(
		mcp.NewTool("reload",
			mcp.WithDescription("Reload the current page."),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleReload,
	)
}

func parseWaitUntil(
	val string,
) []bonk.NavigateOption {
	switch val {
	case "domcontentloaded":
		return []bonk.NavigateOption{
			bonk.WithWaitUntil(bonk.WaitDOMContentLoaded),
		}
	case "networkidle":
		return []bonk.NavigateOption{
			bonk.WithWaitUntil(bonk.WaitNetworkIdle),
		}
	default:
		return nil
	}
}

func (s *Session) handleNavigate(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	opts := parseWaitUntil(
		req.GetString("wait_until", ""),
	)

	if err := page.Navigate(url, opts...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	title, _ := page.Title()
	finalURL, _ := page.URL()
	return mcp.NewToolResultText(
		fmt.Sprintf("Navigated to %s\nTitle: %s", finalURL, title),
	), nil
}

func (s *Session) handleGoBack(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := page.GoBack(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	url, _ := page.URL()
	return mcp.NewToolResultText(
		fmt.Sprintf("Navigated back to %s", url),
	), nil
}

func (s *Session) handleGoForward(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := page.GoForward(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	url, _ := page.URL()
	return mcp.NewToolResultText(
		fmt.Sprintf("Navigated forward to %s", url),
	), nil
}

func (s *Session) handleReload(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := page.Reload(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Page reloaded"), nil
}
