package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBrowserTools(s *server.MCPServer, sess *Session) {
	s.AddTool(
		mcp.NewTool("browser_launch",
			mcp.WithDescription(
				"Launch a browser instance. "+
					"The browser launches automatically on first "+
					"use, but this tool lets you launch it explicitly.",
			),
		),
		sess.handleBrowserLaunch,
	)

	s.AddTool(
		mcp.NewTool("browser_close",
			mcp.WithDescription(
				"Close the browser and all open pages.",
			),
		),
		sess.handleBrowserClose,
	)

	s.AddTool(
		mcp.NewTool("list_pages",
			mcp.WithDescription(
				"List all open pages with their IDs and URLs.",
			),
		),
		sess.handleListPages,
	)

	s.AddTool(
		mcp.NewTool("new_page",
			mcp.WithDescription(
				"Open a new browser tab. "+
					"Optionally navigate to a URL immediately.",
			),
			mcp.WithString("url",
				mcp.Description("URL to navigate to after opening"),
			),
		),
		sess.handleNewPage,
	)

	s.AddTool(
		mcp.NewTool("close_page",
			mcp.WithDescription("Close a specific browser tab."),
			mcp.WithString("page_id",
				mcp.Required(),
				mcp.Description("ID of the page to close"),
			),
		),
		sess.handleClosePage,
	)
}

func (s *Session) handleBrowserLaunch(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.browser != nil {
		return mcp.NewToolResultText(
			"Browser already running",
		), nil
	}

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText("Browser launched"), nil
}

func (s *Session) handleBrowserClose(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	if err := s.Close(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText("Browser closed"), nil
}

func (s *Session) handleListPages(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pages := s.listPages()
	data, _ := json.MarshalIndent(pages, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Session) handleNewPage(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, id, err := s.newPage()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	url := req.GetString("url", "")
	if url != "" {
		if err := page.Navigate(url); err != nil {
			return mcp.NewToolResultError(
				fmt.Sprintf("opened %s but navigate failed: %s", id, err),
			), nil
		}
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Opened page %s", id),
	), nil
}

func (s *Session) handleClosePage(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := req.RequireString("page_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.closePage(id); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("Closed page %s", id),
	), nil
}
