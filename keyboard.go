package bonk

import (
	"time"

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

// Press sends a key down followed by key up event.
func (k *Keyboard) Press(key string) error {
	if err := k.Down(key); err != nil {
		return err
	}
	return k.Up(key)
}

// Down sends a key down event.
func (k *Keyboard) Down(key string) error {
	return proto.InputDispatchKeyEvent(
		proto.InputDispatchKeyEventTypeKeyDown,
	).WithKey(key).Do(k.page.execCtx)
}

// Up sends a key up event.
func (k *Keyboard) Up(key string) error {
	return proto.InputDispatchKeyEvent(
		proto.InputDispatchKeyEventTypeKeyUp,
	).WithKey(key).Do(k.page.execCtx)
}

// Type types text character by character with key events.
func (k *Keyboard) Type(text string, opts ...TypeOption) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}

	for _, ch := range text {
		s := string(ch)
		if err := proto.InputDispatchKeyEvent(
			proto.InputDispatchKeyEventTypeKeyDown,
		).WithText(s).WithKey(s).Do(k.page.execCtx); err != nil {
			return err
		}
		if err := proto.InputDispatchKeyEvent(
			proto.InputDispatchKeyEventTypeKeyUp,
		).WithKey(s).Do(k.page.execCtx); err != nil {
			return err
		}
		if cfg.delay > 0 {
			time.Sleep(cfg.delay)
		}
	}
	return nil
}

// InsertText inserts text without key events (instant).
func (k *Keyboard) InsertText(text string) error {
	return proto.InputInsertText(text).Do(k.page.execCtx)
}
