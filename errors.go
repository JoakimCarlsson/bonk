package bonk

import (
	"errors"
	"fmt"
)

// ErrBrowserClosed is returned when the browser has been closed.
// The other vars indicate a closed context, page, or session; missing or failing Chrome; or a stale element.
var (
	ErrBrowserClosed  = errors.New("bonk: browser closed")
	ErrContextClosed  = errors.New("bonk: context closed")
	ErrPageClosed     = errors.New("bonk: page closed")
	ErrSessionClosed  = errors.New("bonk: session closed")
	ErrChromeNotFound = errors.New("bonk: chrome binary not found")
	ErrChromeStartup  = errors.New("bonk: chrome failed to start")
	ErrStaleElement   = errors.New("bonk: stale element reference")
)

// TimeoutError is returned when an operation exceeds its deadline.
type TimeoutError struct {
	Operation string
	Selector  string
	Cause     error
}

func (e *TimeoutError) Error() string {
	if e.Selector != "" {
		return fmt.Sprintf("bonk: timeout waiting for selector %q", e.Selector)
	}
	return fmt.Sprintf("bonk: timeout during %s", e.Operation)
}

// Unwrap returns the underlying cause.
func (e *TimeoutError) Unwrap() error { return e.Cause }

// ElementNotFoundError is returned when a selector matches no elements.
type ElementNotFoundError struct {
	Selector string
}

func (e *ElementNotFoundError) Error() string {
	return fmt.Sprintf("bonk: element not found: %s", e.Selector)
}

// NavigationError is returned when page navigation fails.
type NavigationError struct {
	URL     string
	Message string
	Cause   error
}

func (e *NavigationError) Error() string {
	return fmt.Sprintf("bonk: navigation to %s failed: %s", e.URL, e.Message)
}

// Unwrap returns the underlying cause.
func (e *NavigationError) Unwrap() error { return e.Cause }
