package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerContextTools(s *server.MCPServer, sess *Session) {
	s.AddTool(
		mcp.NewTool("set_default_timeout",
			mcp.WithDescription(
				"Set the default timeout for all wait/query operations. "+
					"Applies to the browser context and all pages within it.",
			),
			mcp.WithNumber("timeout_ms",
				mcp.Required(),
				mcp.Description("Timeout in milliseconds"),
			),
		),
		sess.handleSetDefaultTimeout,
	)

	s.AddTool(
		mcp.NewTool("set_default_navigation_timeout",
			mcp.WithDescription(
				"Set the default timeout for navigation operations. "+
					"Applies to navigate, reload, go_back, and go_forward.",
			),
			mcp.WithNumber("timeout_ms",
				mcp.Required(),
				mcp.Description("Timeout in milliseconds"),
			),
		),
		sess.handleSetDefaultNavigationTimeout,
	)

	s.AddTool(
		mcp.NewTool("grant_permissions",
			mcp.WithDescription(
				"Grant browser permissions such as geolocation, "+
					"notifications, camera, microphone, etc.",
			),
			mcp.WithArray("permissions",
				mcp.Required(),
				mcp.Description(
					"List of permissions to grant (e.g. geolocation, "+
						"notifications, audioCapture, videoCapture)",
				),
			),
			mcp.WithString("origin",
				mcp.Description("Origin to scope permissions to"),
			),
		),
		sess.handleGrantPermissions,
	)

	s.AddTool(
		mcp.NewTool("clear_permissions",
			mcp.WithDescription(
				"Reset all permission overrides for the browser context.",
			),
		),
		sess.handleClearPermissions,
	)
}

func (s *Session) handleSetDefaultTimeout(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ms, err := req.RequireFloat("timeout_ms")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	d := time.Duration(ms) * time.Millisecond
	s.ctx.SetDefaultTimeout(d)
	return mcp.NewToolResultText(
		fmt.Sprintf("Default timeout set to %s", d),
	), nil
}

func (s *Session) handleSetDefaultNavigationTimeout(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ms, err := req.RequireFloat("timeout_ms")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	d := time.Duration(ms) * time.Millisecond
	s.ctx.SetDefaultNavigationTimeout(d)
	return mcp.NewToolResultText(
		fmt.Sprintf("Default navigation timeout set to %s", d),
	), nil
}

func (s *Session) handleGrantPermissions(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	raw := req.GetArguments()
	permsRaw, ok := raw["permissions"].([]any)
	if !ok || len(permsRaw) == 0 {
		return mcp.NewToolResultError("permissions is required"), nil
	}

	perms := make([]string, len(permsRaw))
	for i, v := range permsRaw {
		s, ok := v.(string)
		if !ok {
			return mcp.NewToolResultError(
				fmt.Sprintf("permission at index %d is not a string", i),
			), nil
		}
		perms[i] = s
	}

	var opts []bonk.PermissionOption
	origin := req.GetString("origin", "")
	if origin != "" {
		opts = append(opts, bonk.PermissionOrigin(origin))
	}

	if err := s.ctx.GrantPermissions(perms, opts...); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Granted permissions: %v", perms),
	), nil
}

func (s *Session) handleClearPermissions(
	_ context.Context,
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureBrowser(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := s.ctx.ClearPermissions(); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("Permissions cleared"), nil
}
