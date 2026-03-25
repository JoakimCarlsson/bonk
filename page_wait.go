package bonk

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// WaitForLoadState waits for the page to reach the specified load
// state. Unlike the wait built into Navigate, this is a standalone
// call useful after actions that trigger navigation without using
// Navigate (e.g. form submissions, SPA route changes).
func (p *Page) WaitForLoadState(
	state NavigateWait,
	opts ...NavigateOption,
) error {
	if err := p.ensurePageDomain(); err != nil {
		return err
	}

	cfg := defaultNavigateConfig(p)
	for _, o := range opts {
		o(cfg)
	}

	done := make(chan struct{}, 1)
	signal := func(_ json.RawMessage) {
		select {
		case done <- struct{}{}:
		default:
		}
	}

	var unsubs []func()
	switch state {
	case WaitDOMContentLoaded:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventDOMContentEventFiredMethod,
			signal,
		))
	case WaitNetworkIdle:
		if err := p.ensureNetworkDomain(); err != nil {
			return err
		}
		unsubs = append(
			unsubs,
			p.subscribeNetworkIdle(done),
		)
	default:
		unsubs = append(unsubs, p.session.Subscribe(
			proto.PageEventLoadEventFiredMethod,
			signal,
		))
	}
	defer func() {
		for _, u := range unsubs {
			u()
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(cfg.timeout):
		return &TimeoutError{
			Operation: "waiting for load state",
		}
	}
}

// WaitForEvent waits for the next occurrence of the specified event
// and returns its payload. The returned value is one of
// *ConsoleMessage, *Dialog, or *Download depending on the event type.
func (p *Page) WaitForEvent(
	event EventType,
	opts ...WaitOption,
) (any, error) {
	cfg := defaultWaitConfigFor(p)
	for _, o := range opts {
		o(cfg)
	}

	type result struct {
		value any
	}

	ch := make(chan result, 1)

	var unsub func()
	switch event {
	case ConsoleEvent:
		unsub = p.onConsole(func(msg *ConsoleMessage) {
			select {
			case ch <- result{value: msg}:
			default:
			}
		})
	case DialogEvent:
		unsub = p.onDialog(func(d *Dialog) {
			select {
			case ch <- result{value: d}:
			default:
			}
		})
	case DownloadEvent:
		unsub = p.onDownload(func(d *Download) {
			select {
			case ch <- result{value: d}:
			default:
			}
		})
	default:
		return nil, fmt.Errorf(
			"bonk: unknown event type %q", event,
		)
	}
	defer unsub()

	select {
	case r := <-ch:
		return r.value, nil
	case <-time.After(cfg.timeout):
		return nil, &TimeoutError{
			Operation: "waiting for event " + string(event),
		}
	}
}

// Pause blocks execution and prints a message to stderr so the
// user can inspect the page in headed mode. Execution resumes when
// the user presses Enter on stdin.
func (p *Page) Pause() error {
	_, _ = fmt.Fprintln(
		os.Stderr,
		"bonk: paused — press Enter to continue...",
	)
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}
