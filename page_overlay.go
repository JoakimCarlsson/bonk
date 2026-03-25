package bonk

type locatorHandler struct {
	locator *Locator
	handler func()
}

// AddLocatorHandler registers a handler that runs automatically
// before page actions whenever the given locator is visible.
// This is useful for dismissing cookie banners, notification
// popups, and other overlay dialogs that interfere with
// automation.
func (p *Page) AddLocatorHandler(
	locator *Locator,
	handler func(),
) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.locatorHandlers = append(
		p.locatorHandlers,
		locatorHandler{locator: locator, handler: handler},
	)
}

// RemoveLocatorHandler removes a previously registered handler
// for the given locator. Matching is by locator pointer identity.
func (p *Page) RemoveLocatorHandler(locator *Locator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, h := range p.locatorHandlers {
		if h.locator == locator {
			p.locatorHandlers = append(
				p.locatorHandlers[:i],
				p.locatorHandlers[i+1:]...,
			)
			return
		}
	}
}

const maxOverlayRounds = 10

func (p *Page) runLocatorHandlers() {
	p.mu.Lock()
	handlers := make([]locatorHandler, len(p.locatorHandlers))
	copy(handlers, p.locatorHandlers)
	p.mu.Unlock()

	if len(handlers) == 0 {
		return
	}

	for round := 0; round < maxOverlayRounds; round++ {
		fired := false
		for _, h := range handlers {
			visible, err := h.locator.IsVisible()
			if err != nil || !visible {
				continue
			}
			h.handler()
			fired = true
		}
		if !fired {
			return
		}
	}
}
