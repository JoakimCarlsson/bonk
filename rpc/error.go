package rpc

import (
	"errors"
	"fmt"
)

// ProtocolError represents an error returned by the CDP protocol.
type ProtocolError struct {
	Code    int64
	Message string
	Data    string
}

func (e *ProtocolError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("cdp: %s (code %d: %s)", e.Message, e.Code, e.Data)
	}
	return fmt.Sprintf("cdp: %s (code %d)", e.Message, e.Code)
}

const (
	ErrCodeServerError    = -32000
	ErrCodeInvalidParams  = -32602
	ErrCodeMethodNotFound = -32601
	ErrCodeInternalError  = -32603
)

var (
	ErrConnectionClosed = errors.New("bonk: connection closed")
	ErrResponseTimeout  = errors.New("bonk: response timeout")
)

// IsMethodNotFound reports whether the error is a method-not-found CDP error.
func IsMethodNotFound(err error) bool {
	var pe *ProtocolError
	return errors.As(err, &pe) && pe.Code == ErrCodeMethodNotFound
}

// IsInvalidParams reports whether the error is an invalid-params CDP error.
func IsInvalidParams(err error) bool {
	var pe *ProtocolError
	return errors.As(err, &pe) && pe.Code == ErrCodeInvalidParams
}
