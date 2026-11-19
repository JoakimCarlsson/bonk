package bonk

import (
	"github.com/joakimcarlsson/bonk/proto"
)

// Keyboard provides low-level keyboard control on a page.
type Keyboard struct {
	page *Page
}

// Keyboard returns the page's Keyboard controller.
func (p *Page) Keyboard() *Keyboard {
	return &Keyboard{page: p}
}

// Press sends a key down followed by key up event for a key name such as
// "Enter" or a single character, with its real code and key code. It
// returns the first CDP error.
func (k *Keyboard) Press(key string) error {
	return k.page.pressKey(key)
}

// Down sends a key down event for a key name or character. It returns the
// CDP error.
func (k *Keyboard) Down(key string) error {
	return k.page.dispatchKey(true, lookupKey(key))
}

// Up sends a key up event for a key name or character. It returns the CDP
// error.
func (k *Keyboard) Up(key string) error {
	return k.page.dispatchKey(false, lookupKey(key))
}

// Type types text character by character with key events, pausing like a
// person between keys on a stealth page unless WithDelay sets the gap. It
// returns the first CDP error.
func (k *Keyboard) Type(text string, opts ...TypeOption) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}
	return k.page.typeText(text, cfg.delay)
}

// InsertText inserts text without key events (instant).
func (k *Keyboard) InsertText(text string) error {
	return proto.InputInsertText(text).Do(k.page.execCtx)
}
