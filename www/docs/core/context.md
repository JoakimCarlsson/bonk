# Context

A `BrowserContext` is an isolated browser profile with its own cookies, cache, and storage. Each context is independent — like a separate incognito window.

## Create a Context

```go
ctx, err := b.NewContext()
if err != nil {
    log.Fatal(err)
}
defer ctx.Close()
```

## Context Options

```go
ctx, err := b.NewContext(
    bonk.WithProxy("http://proxy:8080"),
    bonk.WithProxyBypass("localhost,127.0.0.1"),
    bonk.WithViewport(1920, 1080),
    bonk.WithUserAgent("Custom UA"),
    bonk.WithLocale("en-US"),
    bonk.WithTimezone("America/New_York"),
    bonk.WithGeolocation(40.7128, -74.0060),
    bonk.WithState("./session.dat"),
)
```

| Option | Description |
|--------|-------------|
| `WithProxy(string)` | Proxy server URL |
| `WithProxyBypass(string)` | Comma-separated list of hosts to bypass proxy |
| `WithViewport(w, h int)` | Default viewport for new pages |
| `WithUserAgent(string)` | User agent override |
| `WithLocale(string)` | Browser locale (e.g. `"en-US"`) |
| `WithTimezone(string)` | Timezone override (e.g. `"America/New_York"`) |
| `WithGeolocation(lat, lon float64)` | Geolocation override |
| `WithState(string)` | Load saved cookies from file |

## Pages

```go
page, err := ctx.NewPage()

pages := ctx.Pages() // all open pages
```

## Cookies

```go
cookies, err := ctx.Cookies()

err = ctx.SetCookies(bonk.Cookie{
    Name:   "session",
    Value:  "abc123",
    Domain: ".example.com",
    Path:   "/",
})

err = ctx.ClearCookies()
```

## State Persistence

Save cookies to disk and restore them later:

```go
// save
err = ctx.SaveState("./session.dat")

// restore in a new context
ctx2, err := b.NewContext(bonk.WithState("./session.dat"))
```

See [Session Persistence](../advanced/session-persistence.md) for details.

## Close

Closes all pages and disposes the browser context:

```go
ctx.Close()
```
