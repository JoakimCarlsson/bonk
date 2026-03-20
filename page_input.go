package bonk

// Click waits for the selector and clicks the element.
func (p *Page) Click(selector string, opts ...WaitOption) error {
	el, err := p.WaitSelector(selector, opts...)
	if err != nil {
		return err
	}
	return el.Click()
}

// Fill waits for the selector and fills the input with text.
func (p *Page) Fill(selector, text string, opts ...WaitOption) error {
	el, err := p.WaitSelector(selector, opts...)
	if err != nil {
		return err
	}
	return el.Fill(text)
}

// Type waits for the selector and types text character by character.
func (p *Page) Type(
	selector, text string,
	opts ...TypeOption,
) error {
	cfg := defaultTypeConfig()
	for _, o := range opts {
		o(cfg)
	}
	el, err := p.WaitSelector(selector, cfg.waitOpts...)
	if err != nil {
		return err
	}
	return el.Type(text, opts...)
}

// Press waits for the selector and presses a key.
func (p *Page) Press(selector, key string, opts ...WaitOption) error {
	el, err := p.WaitSelector(selector, opts...)
	if err != nil {
		return err
	}
	return el.Press(key)
}
