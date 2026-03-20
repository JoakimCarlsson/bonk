package bonk

import (
	"errors"
	"fmt"
)

var (
	ErrBrowserClosed  = errors.New("bonk: browser closed")
	ErrContextClosed  = errors.New("bonk: context closed")
	ErrPageClosed     = errors.New("bonk: page closed")
	ErrSessionClosed  = errors.New("bonk: session closed")
	ErrChromeNotFound = errors.New("bonk: chrome binary not found")
	ErrChromeStartup  = errors.New("bonk: chrome failed to start")
)

// TimeoutError is returned when an operation exceeds its deadline.
type TimeoutError struct {
	Operation string
	Selector  string
}

func (e *TimeoutError) Error() string {
	if e.Selector != "" {
		return fmt.Sprintf("bonk: timeout waiting for selector %q", e.Selector)
	}
	return fmt.Sprintf("bonk: timeout during %s", e.Operation)
}

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
}

func (e *NavigationError) Error() string {
	return fmt.Sprintf("bonk: navigation to %s failed: %s", e.URL, e.Message)
}
