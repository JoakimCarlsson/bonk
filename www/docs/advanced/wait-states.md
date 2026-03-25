# Wait States

`WaitForLoadState` waits for the page to reach a specific load state as a standalone call. Unlike the wait built into `Navigate`, this is useful after actions that trigger navigation without calling `Navigate()` directly — for example, form submissions, `history.pushState`, or SPA route changes.

## Usage

```go
page.Click("#submit-form")
err := page.WaitForLoadState(bonk.WaitLoad)
```

## Load States

| State | Description |
|-------|-------------|
| `WaitLoad` | Wait for the `load` event (default) |
| `WaitDOMContentLoaded` | Wait for the `DOMContentLoaded` event |
| `WaitNetworkIdle` | Wait until no network requests for 500ms |

## Options

`WaitForLoadState` accepts `NavigateOption` values for configuring timeout:

```go
page.WaitForLoadState(bonk.WaitNetworkIdle,
    bonk.WithTimeout(10*time.Second),
)
```

## When to Use

Use `WaitForLoadState` when:

- A form submission triggers a full page navigation
- A click navigates to a new page via `window.location`
- An SPA framework changes routes after a user action
- You need to wait for network idle after dynamic content loads

For navigations you initiate directly, prefer using `Navigate` with `WithWaitUntil` instead — it subscribes to events before sending the navigation command.
