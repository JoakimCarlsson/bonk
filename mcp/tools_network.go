package mcp

import (
	"context"
	"fmt"

	"github.com/joakimcarlsson/bonk"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerNetworkTools(
	s *server.MCPServer,
	sess *Session,
) {
	s.AddTool(
		mcp.NewTool("set_extra_headers",
			mcp.WithDescription(
				"Set extra HTTP headers to send with "+
					"every request from this page.",
			),
			mcp.WithObject("headers",
				mcp.Required(),
				mcp.Description(
					"Map of header name to value",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleSetExtraHeaders,
	)

	s.AddTool(
		mcp.NewTool("block_urls",
			mcp.WithDescription(
				"Block requests matching URL patterns. "+
					"Useful for blocking ads, trackers, or images.",
			),
			mcp.WithArray("patterns",
				mcp.Required(),
				mcp.Description(
					"URL patterns to block "+
						"(supports * wildcards)",
				),
			),
			mcp.WithString("page_id",
				mcp.Description(
					"ID of the page. "+
						"Omit to use the default page.",
				),
			),
		),
		sess.handleBlockURLs,
	)
}

func (s *Session) handleSetExtraHeaders(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	raw := req.GetArguments()
	headersRaw, ok := raw["headers"]
	if !ok {
		return mcp.NewToolResultError(
			"headers is required",
		), nil
	}

	headersMap, ok := headersRaw.(map[string]any)
	if !ok {
		return mcp.NewToolResultError(
			"headers must be an object",
		), nil
	}

	headers := make(
		map[string]string,
		len(headersMap),
	)
	for k, v := range headersMap {
		headers[k] = fmt.Sprintf("%v", v)
	}

	if err := page.SetExtraHTTPHeaders(headers); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(
		fmt.Sprintf("Set %d header(s)", len(headers)),
	), nil
}

func (s *Session) handleBlockURLs(
	_ context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	patterns := req.GetStringSlice("patterns", nil)
	if len(patterns) == 0 {
		return mcp.NewToolResultError(
			"patterns is required and must not be empty",
		), nil
	}

	page, _, err := s.pageFromRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	for _, pattern := range patterns {
		p := pattern
		unsub := page.Route(p, func(r *bonk.Route) {
			r.Abort()
		})
		s.routes[p] = unsub
	}

	return mcp.NewToolResultText(
		fmt.Sprintf(
			"Blocking %d URL pattern(s)",
			len(patterns),
		),
	), nil
}
