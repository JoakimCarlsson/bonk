# Page

A `Page` represents a single browser tab. It handles navigation, JavaScript evaluation, screenshots, and serves as the entry point for element interaction.

## Create a Page

```go
page, err := ctx.NewPage()
if err != nil {
    log.Fatal(err)
}
defer page.Close()
```

## Navigation

```go
err = page.Navigate("https://example.com")

err = page.Navigate("https://example.com",
    bonk.WithTimeout(30*time.Second),
    bonk.WithWaitUntil(bonk.WaitNetworkIdle),
)

err = page.Reload()
err = page.GoBack()
err = page.GoForward()
```

All navigation methods wait for page load by default. Configure with `NavigateOption`:

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(time.Duration)` | 30s | Maximum wait time |
| `WithWaitUntil(NavigateWait)` | `WaitLoad` | When navigation is considered complete |

`NavigateWait` values:

| Value | Description |
|-------|-------------|
| `WaitLoad` | Wait for the `load` event |
| `WaitDOMContentLoaded` | Wait for `DOMContentLoaded` |
| `WaitNetworkIdle` | Wait until no network requests for 500ms |

## Wait for Navigation

Wait for a navigation triggered by something else (e.g. a click):

```go
go func() {
    page.Click("#link")
}()
err = page.WaitNavigation()
```

## Wait for URL

Wait until the page URL matches a glob pattern:

```go
err = page.WaitForURL("**/dashboard*")
```

## Page Info

```go
url, err := page.URL()
title, err := page.Title()
```

## JavaScript

```go
result, err := page.Evaluate("document.title")

// wait for condition
err = page.WaitForFunction("document.readyState === 'complete'")

// get element handle from JS
el, err := page.EvaluateHandle("document.querySelector('#app')")

// call function on element
val, err := page.EvaluateOn(el, "function(){ return this.textContent }")
```

## Content

```go
html, err := page.Content()

err = page.SetContent("<h1>Hello</h1>")
```

## Screenshots and PDF

```go
err = page.Screenshot("page.png")
err = page.Screenshot("full.png", bonk.FullPage())
err = page.Screenshot("photo.jpg", bonk.ScreenshotQuality(80))

err = page.PDF("page.pdf")
```

| Option | Description |
|--------|-------------|
| `FullPage()` | Capture the full scrollable page |
| `ScreenshotQuality(int)` | JPEG/WebP quality (0-100) |

Format is determined by file extension: `.png`, `.jpg`/`.jpeg`, `.webp`.

## Viewport

```go
err = page.SetViewport(1920, 1080)
```

## Device Emulation

```go
err = page.Emulate(bonk.IPhone15)
```

See [Device Emulation](../automation/device-emulation.md) for all presets.

## Context Control

```go
// set deadline
p := page.Timeout(5 * time.Second)
p.Navigate("https://slow-site.com") // fails after 5s

// use custom context
p = page.WithContext(ctx)
```

See [Context Control](../advanced/context-control.md) for details.
