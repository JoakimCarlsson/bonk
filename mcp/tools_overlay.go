package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerOverlayTools(
	s *server.MCPServer,
	sess *Session,
) {
	s.AddTool(
		mcp.NewTool("add_locator_handler",
			mcp.WithDescription(
				"Register an auto-dismiss handler. When "+
					"the locator selector is visible before "+
					"an action, the click_selector is "+
					"clicked to dismiss it. Useful for "+
					"cookie banners, notification popups, "+
					"and overlay dialogs.",
			),
			mcp.WithString("locator",
				mcp.Required(),
				mcp.Description(
					"CSS selector that detects the overlay",
				),
			),
			mcp.WithString("click_selector",
				mcp.Required(),
				mcp.Description(
					"CSS selector of the element to click "+
						"to dismiss the overlay",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleAddLocatorHandler,
	)

	s.AddTool(
		mcp.NewTool("remove_locator_handler",
			mcp.WithDescription(
				"Remove a previously registered overlay "+
					"auto-dismiss handler.",
			),
			mcp.WithString("locator",
				mcp.Required(),
				mcp.Description(
					"CSS selector used when registering "+
						"the handler",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleRemoveLocatorHandler,
	)
}

func (s *Session) handleAddLocatorHandler(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	locatorSel, err := req.RequireString("locator")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	clickSel, err := req.RequireString("click_selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, pageID, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	loc := page.Locator(locatorSel)
	page.AddLocatorHandler(loc, func() {
		page.Click(clickSel)
	})

	key := pageID + ":overlay:" + locatorSel
	s.overlays[key] = loc

	return mcp.NewToolResultText(
		fmt.Sprintf(
			"Handler registered: when %q is visible, "+
				"click %q",
			locatorSel, clickSel,
		),
	), nil
}

func (s *Session) handleRemoveLocatorHandler(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	locatorSel, err := req.RequireString("locator")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	page, pageID, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	key := pageID + ":overlay:" + locatorSel
	loc, ok := s.overlays[key]
	if !ok {
		return mcp.NewToolResultError(
			fmt.Sprintf(
				"no handler registered for %q",
				locatorSel,
			),
		), nil
	}

	page.RemoveLocatorHandler(loc)
	delete(s.overlays, key)

	return mcp.NewToolResultText(
		fmt.Sprintf("Handler for %q removed", locatorSel),
	), nil
}
