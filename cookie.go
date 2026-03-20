package bonk

import (
	"github.com/joakimcarlsson/bonk/proto"
)

// Cookie represents a browser cookie.
type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

// Cookies returns all cookies in the browser context.
// Requires at least one page to be open.
func (c *BrowserContext) Cookies() ([]Cookie, error) {
	pages := c.Pages()
	if len(pages) == 0 {
		return nil, nil
	}

	res, err := proto.StorageGetCookies().Do(pages[0].execCtx)
	if err != nil {
		return nil, err
	}

	var cookies []Cookie
	for _, nc := range res.Cookies {
		cookies = append(cookies, Cookie{
			Name:     nc.Name,
			Value:    nc.Value,
			Domain:   nc.Domain,
			Path:     nc.Path,
			Expires:  float64(nc.Expires),
			HTTPOnly: nc.HTTPOnly,
			Secure:   nc.Secure,
			SameSite: string(nc.SameSite),
		})
	}
	return cookies, nil
}

// SetCookies sets cookies in the browser context.
func (c *BrowserContext) SetCookies(cookies ...Cookie) error {
	pages := c.Pages()
	if len(pages) == 0 {
		return nil
	}

	var params []proto.NetworkCookieParam
	for _, ck := range cookies {
		params = append(params, proto.NetworkCookieParam{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     ck.Path,
			Expires:  proto.TimeSinceEpoch(ck.Expires),
			HTTPOnly: ck.HTTPOnly,
			Secure:   ck.Secure,
		})
	}

	return proto.StorageSetCookies(params).Do(pages[0].execCtx)
}

// ClearCookies clears all cookies in the browser context.
func (c *BrowserContext) ClearCookies() error {
	pages := c.Pages()
	if len(pages) == 0 {
		return nil
	}
	return proto.StorageClearCookies().Do(pages[0].execCtx)
}
