package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerStateTools(
	s *server.MCPServer,
	sess *Session,
) {
	s.AddTool(
		mcp.NewTool("get_cookies",
			mcp.WithDescription(
				"Get all cookies from the browser context.",
			),
		),
		sess.handleGetCookies,
	)

	s.AddTool(
		mcp.NewTool("set_cookies",
			mcp.WithDescription("Set cookies in the browser context."),
			mcp.WithArray("cookies",
				mcp.Required(),
				mcp.Description(
					"Array of cookie objects with name, value, "+
						"domain, path, etc.",
				),
			),
		),
		sess.handleSetCookies,
	)

	s.AddTool(
		mcp.NewTool("clear_cookies",
			mcp.WithDescription(
				"Clear all cookies from the browser context.",
			),
		),
		sess.handleClearCookies,
	)

	s.AddTool(
		mcp.NewTool("save_state",
			mcp.WithDescription(
				"Save browser state (cookies, localStorage) "+
					"to a file.",
			),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("File path to save state to"),
			),
		),
		sess.handleSaveState,
	)

	s.AddTool(
		mcp.NewTool("load_state",
			mcp.WithDescription(
				"Load browser state (cookies, localStorage) "+
					"from a file.",
			),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description(
					"File path to load state from",
				),
			),
		),
		sess.handleLoadState,
	)
}

func (s *Session) handleGetCookies(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cookies, err := s.ctx.Cookies()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.MarshalIndent(cookies, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *Session) handleSetCookies(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	raw := req.GetArguments()
	cookiesRaw, ok := raw["cookies"]
	if !ok {
		return mcp.NewToolResultError(
			"cookies is required",
		), nil
	}

	data, err := json.Marshal(cookiesRaw)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("invalid cookies: %s", err),
		), nil
	}

	var cookies []bonk.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("invalid cookies: %s", err),
		), nil
	}

	if err := s.ctx.SetCookies(cookies...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Set %d cookie(s)", len(cookies)),
	), nil
}

func (s *Session) handleClearCookies(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ctx.ClearCookies(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Cookies cleared"), nil
}

func (s *Session) handleSaveState(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ctx.SaveState(path); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("State saved to %s", path),
	), nil
}

func (s *Session) handleLoadState(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ctx.LoadState(path); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("State loaded from %s", path),
	), nil
}
