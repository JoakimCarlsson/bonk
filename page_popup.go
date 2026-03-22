package bonk

import (
	"encoding/json"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// PopupOption configures WaitForPopup behavior.
type PopupOption func(*popupConfig)

type popupConfig struct {
	timeout time.Duration
}

// PopupTimeout sets the maximum time to wait for a popup.
func PopupTimeout(d time.Duration) PopupOption {
	return func(c *popupConfig) {
		c.timeout = d
	}
}

// WaitForPopup waits for a new page (tab/popup) to be opened by this page
// and returns it. The caller should trigger the popup after calling this method.
func (p *Page) WaitForPopup(opts ...PopupOption) (*Page, error) {
	cfg := &popupConfig{timeout: p.resolveTimeout()}
	for _, o := range opts {
		o(cfg)
	}

	browserExecCtx := p.browserCtx.browser.execCtx()

	if err := proto.TargetSetDiscoverTargets(true).Do(browserExecCtx); err != nil {
		return nil, err
	}

	result := make(chan *Page, 1)
	errCh := make(chan error, 1)

	unsub := p.browserCtx.browser.conn.Subscribe(
		proto.TargetEventTargetCreatedMethod,
		func(params json.RawMessage) {
			var ev proto.TargetEventTargetCreated
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}

			if ev.TargetInfo.OpenerID != p.targetID {
				return
			}
			if ev.TargetInfo.Type != "page" {
				return
			}

			popup, err := attachToTarget(p.browserCtx, ev.TargetInfo.TargetID)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			select {
			case result <- popup:
			default:
			}
		},
	)
	defer unsub()

	select {
	case popup := <-result:
		return popup, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(cfg.timeout):
		return nil, &TimeoutError{Operation: "waiting for popup"}
	}
}
