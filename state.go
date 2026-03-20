package bonk

import (
	"encoding/json"
	"fmt"
	"os"
)

// BrowserState holds serializable browser state for persistence.
// Includes cookies and localStorage. IndexedDB is not supported.
type BrowserState struct {
	Cookies      []Cookie                     `json:"cookies"`
	LocalStorage map[string]map[string]string `json:"localStorage,omitempty"`
}

// SaveState saves the browser context state (cookies, localStorage) to a file.
func (c *BrowserContext) SaveState(path string) error {
	cookies, err := c.Cookies()
	if err != nil {
		return err
	}

	localStorage := make(map[string]map[string]string)
	for _, page := range c.Pages() {
		origin, err := page.Evaluate("location.origin")
		if err != nil {
			continue
		}
		originStr, ok := origin.(string)
		if !ok || originStr == "" || originStr == "null" {
			continue
		}

		val, err := page.Evaluate(
			"JSON.stringify(Object.entries(localStorage))",
		)
		if err != nil {
			continue
		}
		s, ok := val.(string)
		if !ok || s == "" || s == "[]" {
			continue
		}

		var entries [][]string
		if err := json.Unmarshal([]byte(s), &entries); err != nil {
			continue
		}
		if len(entries) > 0 {
			m := make(map[string]string, len(entries))
			for _, entry := range entries {
				if len(entry) == 2 {
					m[entry[0]] = entry[1]
				}
			}
			localStorage[originStr] = m
		}
	}

	state := BrowserState{
		Cookies:      cookies,
		LocalStorage: localStorage,
	}
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
		if err := c.SetCookies(state.Cookies...); err != nil {
			return err
		}
	}

	if len(state.LocalStorage) > 0 {
		pages := c.Pages()
		if len(pages) == 0 {
			return nil
		}
		page := pages[0]

		for origin, entries := range state.LocalStorage {
			if err := page.Navigate(origin); err != nil {
				continue
			}
			for k, v := range entries {
				js := fmt.Sprintf(
					"localStorage.setItem(%s,%s)",
					jsString(k), jsString(v),
				)
				page.Evaluate(js)
			}
		}
	}

	return nil
}
