package bonk

import (
	"encoding/json"
	"os"
)

// BrowserState holds serializable browser state for persistence.
type BrowserState struct {
	Cookies []Cookie `json:"cookies"`
}

// SaveState saves the browser context state (cookies) to a file.
func (c *BrowserContext) SaveState(path string) error {
	cookies, err := c.Cookies()
	if err != nil {
		return err
	}

	state := BrowserState{Cookies: cookies}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// LoadState restores browser context state from a file.
func (c *BrowserContext) LoadState(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state BrowserState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	if len(state.Cookies) > 0 {
		return c.SetCookies(state.Cookies...)
	}
	return nil
}
