# Overlay Handlers

Cookie consent banners, notification popups, and modal dialogs constantly break automation. `AddLocatorHandler` handles them declaratively — register a handler once, and it runs automatically before every action when the overlay is visible.

## AddLocatorHandler

```go
banner := page.Locator(".cookie-banner")
page.AddLocatorHandler(banner, func() {
    page.Click(".cookie-banner .accept")
})
```

After registration, every `Click`, `Fill`, `Type`, and `Press` call on both the page and any locator will check if `.cookie-banner` is visible. If it is, the handler clicks the accept button before proceeding with the original action.

## How It Works

1. Before each action (`Click`, `Fill`, `Type`, `Press`), all registered handlers are checked
2. For each handler, `locator.IsVisible()` is called
3. If visible, the handler function runs
4. Steps 2-3 repeat (up to 10 rounds) to handle cascading overlays
5. The original action proceeds

## RemoveLocatorHandler

Remove a handler when it's no longer needed:

```go
banner := page.Locator(".cookie-banner")
page.AddLocatorHandler(banner, func() {
    page.Click(".cookie-banner .accept")
})

page.RemoveLocatorHandler(banner)
```

Matching is by locator pointer identity — pass the same `*Locator` that was used during registration.

## Multiple Handlers

Register handlers for different overlays:

```go
page.AddLocatorHandler(page.Locator(".cookie-banner"), func() {
    page.Click(".cookie-banner .accept")
})

page.AddLocatorHandler(page.Locator(".notification-popup"), func() {
    page.Click(".notification-popup .close")
})

page.AddLocatorHandler(page.Locator("#survey-modal"), func() {
    page.Click("#survey-modal .dismiss")
})
```

All handlers are checked before every action. Multiple overlays can be dismissed in a single round.

## Locator Handlers with Page.WithContext

Handlers are copied when creating a page copy via `WithContext` or `Timeout`:

```go
page.AddLocatorHandler(page.Locator(".banner"), func() {
    page.Click(".banner .close")
})

p := page.Timeout(5 * time.Second)
p.Click("#button") // handler still runs
```
