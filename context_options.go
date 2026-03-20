package bonk

// ContextOption configures a browser context.
type ContextOption func(*contextConfig)

type contextConfig struct {
	proxyServer     string
	proxyBypassList string
	viewportWidth   int
	viewportHeight  int
	userAgent       string
	locale          string
	timezone        string
	geoLatitude     float64
	geoLongitude    float64
	geoAccuracy     float64
	hasGeo          bool
	statePath       string
}

func defaultContextConfig() *contextConfig {
	return &contextConfig{}
}

// WithProxy sets a proxy server for the context.
func WithProxy(server string) ContextOption {
	return func(c *contextConfig) {
		c.proxyServer = server
	}
}

// WithProxyBypass sets the proxy bypass list.
func WithProxyBypass(list string) ContextOption {
	return func(c *contextConfig) {
		c.proxyBypassList = list
	}
}

// WithViewport sets the default viewport size for new pages.
func WithViewport(width, height int) ContextOption {
	return func(c *contextConfig) {
		c.viewportWidth = width
		c.viewportHeight = height
	}
}

// WithUserAgent sets the user agent string for the context.
func WithUserAgent(ua string) ContextOption {
	return func(c *contextConfig) {
		c.userAgent = ua
	}
}

// WithLocale sets the browser locale.
func WithLocale(locale string) ContextOption {
	return func(c *contextConfig) {
		c.locale = locale
	}
}

// WithTimezone sets the timezone override.
func WithTimezone(tz string) ContextOption {
	return func(c *contextConfig) {
		c.timezone = tz
	}
}

// WithGeolocation sets the geolocation override.
func WithGeolocation(lat, lon float64) ContextOption {
	return func(c *contextConfig) {
		c.geoLatitude = lat
		c.geoLongitude = lon
		c.geoAccuracy = 1
		c.hasGeo = true
	}
}

// WithState loads browser state (cookies, localStorage) from a file.
func WithState(path string) ContextOption {
	return func(c *contextConfig) {
		c.statePath = path
	}
}
