// Package proto contains generated Chrome DevTools Protocol types and commands.
package proto

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Method represents a CDP method name like "Page.navigate" or "DOM.getDocument".
type Method string

// Domain returns the domain portion of the method (e.g. "Page" from "Page.navigate").
func (m Method) Domain() string {
	if i := strings.IndexByte(string(m), '.'); i >= 0 {
		return string(m[:i])
	}
	return string(m)
}

// Command returns the command portion of the method (e.g. "navigate" from "Page.navigate").
func (m Method) Command() string {
	if i := strings.IndexByte(string(m), '.'); i >= 0 {
		return string(m[i+1:])
	}
	return ""
}

// String returns the method as a string.
func (m Method) String() string {
	return string(m)
}

// SessionID uniquely identifies an attached debugging session.
type SessionID string

// String returns the session ID as a string.
func (s SessionID) String() string {
	return string(s)
}

// Message is the wire format for CDP messages over WebSocket.
type Message struct {
	ID        int64           `json:"id,omitempty"`
	SessionID SessionID       `json:"sessionId,omitempty"`
	Method    Method          `json:"method,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *Error          `json:"error,omitempty"`
}

// Error represents a CDP protocol error returned in a message response.
type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("cdp: %s (code %d: %s)", e.Message, e.Code, e.Data)
	}
	return fmt.Sprintf("cdp: %s (code %d)", e.Message, e.Code)
}
