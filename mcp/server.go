// Package mcp provides an MCP server for browser automation.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates an MCP server with all bonk tools registered.
func NewServer(sess *Session) *server.MCPServer {
	s := server.NewMCPServer(
		"bonkmcp",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerBrowserTools(s, sess)
	registerContextTools(s, sess)
	registerNavigateTools(s, sess)
	registerPageTools(s, sess)
	registerElementTools(s, sess)
	registerNetworkTools(s, sess)
	registerStateTools(s, sess)
	registerOverlayTools(s, sess)

	return s
}
