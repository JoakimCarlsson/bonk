package bonk

import "fmt"

// GetByRole returns a Locator matching elements by ARIA role.
// Matches both explicit role attributes and implicit roles from HTML semantics.
func (p *Page) GetByRole(role string, opts ...GetByRoleOption) *Locator {
	cfg := defaultGetByRoleConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByRoleJSExpr(role, cfg.name, cfg.hasName, cfg.exact)
	desc := fmt.Sprintf("getByRole(%s)", jsString(role))
	if cfg.hasName {
		desc = fmt.Sprintf(
			"getByRole(%s, name=%s)",
			jsString(role),
			jsString(cfg.name),
		)
	}
	return &Locator{
		page:      p,
		jsExpr:    jsExpr,
		jsAllExpr: jsAllExpr,
		desc:      desc,
		nth:       -1,
	}
}

// GetByLabel returns a Locator matching form controls by their associated label text.
// Checks <label> elements, aria-label, and aria-labelledby attributes.
func (p *Page) GetByLabel(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByLabelJSExpr(text, cfg.exact)
	return &Locator{
		page: p, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByLabel(%s)", jsString(text)), nth: -1,
	}
}

// GetByText returns a Locator matching elements by their visible text content.
func (p *Page) GetByText(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByTextJSExpr(text, cfg.exact)
	return &Locator{
		page: p, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByText(%s)", jsString(text)), nth: -1,
	}
}

// GetByPlaceholder returns a Locator matching elements by their placeholder attribute.
func (p *Page) GetByPlaceholder(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("placeholder", text, cfg.exact)
	return &Locator{
		page: p, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByPlaceholder(%s)", jsString(text)), nth: -1,
	}
}

// GetByTestID returns a Locator matching elements by their data-testid attribute.
func (p *Page) GetByTestID(id string) *Locator {
	sel := fmt.Sprintf("[data-testid=%s]", jsString(id))
	return &Locator{
		page: p, selector: sel,
		desc: fmt.Sprintf("getByTestID(%s)", jsString(id)), nth: -1,
	}
}

// GetByAltText returns a Locator matching elements by their alt attribute.
func (p *Page) GetByAltText(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("alt", text, cfg.exact)
	return &Locator{
		page: p, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByAltText(%s)", jsString(text)), nth: -1,
	}
}

// GetByTitle returns a Locator matching elements by their title attribute.
func (p *Page) GetByTitle(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("title", text, cfg.exact)
	return &Locator{
		page: p, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByTitle(%s)", jsString(text)), nth: -1,
	}
}

// GetByRole returns a Locator matching elements by ARIA role within the frame.
func (f *Frame) GetByRole(role string, opts ...GetByRoleOption) *Locator {
	cfg := defaultGetByRoleConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByRoleJSExpr(role, cfg.name, cfg.hasName, cfg.exact)
	desc := fmt.Sprintf("getByRole(%s)", jsString(role))
	if cfg.hasName {
		desc = fmt.Sprintf(
			"getByRole(%s, name=%s)",
			jsString(role),
			jsString(cfg.name),
		)
	}
	return &Locator{
		page:      f.page,
		frame:     f,
		jsExpr:    jsExpr,
		jsAllExpr: jsAllExpr,
		desc:      desc,
		nth:       -1,
	}
}

// GetByLabel returns a Locator matching form controls by their associated label text within the frame.
func (f *Frame) GetByLabel(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByLabelJSExpr(text, cfg.exact)
	return &Locator{
		page: f.page, frame: f, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByLabel(%s)", jsString(text)), nth: -1,
	}
}

// GetByText returns a Locator matching elements by their visible text content within the frame.
func (f *Frame) GetByText(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByTextJSExpr(text, cfg.exact)
	return &Locator{
		page: f.page, frame: f, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByText(%s)", jsString(text)), nth: -1,
	}
}

// GetByPlaceholder returns a Locator matching elements by their placeholder attribute within the frame.
func (f *Frame) GetByPlaceholder(
	text string,
	opts ...TextMatchOption,
) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("placeholder", text, cfg.exact)
	return &Locator{
		page: f.page, frame: f, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByPlaceholder(%s)", jsString(text)), nth: -1,
	}
}

// GetByTestID returns a Locator matching elements by their data-testid attribute within the frame.
func (f *Frame) GetByTestID(id string) *Locator {
	sel := fmt.Sprintf("[data-testid=%s]", jsString(id))
	return &Locator{
		page: f.page, frame: f, selector: sel,
		desc: fmt.Sprintf("getByTestID(%s)", jsString(id)), nth: -1,
	}
}

// GetByAltText returns a Locator matching elements by their alt attribute within the frame.
func (f *Frame) GetByAltText(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("alt", text, cfg.exact)
	return &Locator{
		page: f.page, frame: f, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByAltText(%s)", jsString(text)), nth: -1,
	}
}

// GetByTitle returns a Locator matching elements by their title attribute within the frame.
func (f *Frame) GetByTitle(text string, opts ...TextMatchOption) *Locator {
	cfg := defaultTextMatchConfig()
	for _, o := range opts {
		o(cfg)
	}
	jsExpr, jsAllExpr := getByAttrJSExpr("title", text, cfg.exact)
	return &Locator{
		page: f.page, frame: f, jsExpr: jsExpr, jsAllExpr: jsAllExpr,
		desc: fmt.Sprintf("getByTitle(%s)", jsString(text)), nth: -1,
	}
}
