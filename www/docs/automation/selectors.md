# Selectors

bonk uses CSS selectors to find elements on a page. All selector methods work on both `Page` and `Frame`.

For semantic selectors based on ARIA roles, labels, and text content, see [Accessibility Selectors](accessibility.md).

## Query

Find the first matching element. Returns `nil` if not found:

```go
el, err := page.Query("h1")
if el == nil {
    fmt.Println("not found")
}
```

## QueryAll

Find all matching elements:

```go
items, err := page.QueryAll(".product-card")
for _, item := range items {
    name, _ := item.Text()
    fmt.Println(name)
}
```

## WaitSelector

Poll until an element appears in the DOM:

```go
el, err := page.WaitSelector("#dynamic-content")
```

### Wait Options

| Option | Default | Description |
|--------|---------|-------------|
| `WaitTimeout(time.Duration)` | 30s | Maximum time to wait |
| `WaitInterval(time.Duration)` | 50ms | Initial polling interval |
| `WaitVisibleOption()` | — | Wait until element is visible |
| `WaitHiddenOption()` | — | Wait until element is hidden or removed |

```go
el, err := page.WaitSelector(".loaded",
    bonk.WaitTimeout(10*time.Second),
    bonk.WaitInterval(100*time.Millisecond),
)
```

Polling uses exponential backoff starting from the interval, capped at 1 second.

### Wait for Visibility

Wait until the element exists AND is visible:

```go
el, err := page.WaitSelector(".modal", bonk.WaitVisibleOption())
```

### Wait for Hidden

Wait until the element is hidden or removed from the DOM:

```go
_, err := page.WaitSelector(".spinner", bonk.WaitHiddenOption())
```

This is useful for waiting for loading indicators to disappear.

## Auto-Wait for Visibility

When you call interaction methods like `Click()`, `Fill()`, or `Type()` on an element, bonk automatically waits for the element to become visible before acting. This prevents errors from interacting with hidden or transitioning elements.

An element is considered visible when:

- `display` is not `none`
- `visibility` is not `hidden`
- `opacity` is not `0`

## Stale Element Retry

If an element's backing DOM node is garbage collected (e.g. after a page re-render), bonk detects the stale reference and automatically:

1. Re-queries the page using the element's original CSS selector
2. Updates the element's internal reference
3. Retries the operation once

Elements without a selector (created via `EvaluateHandle`) cannot be re-resolved and fail with `ErrStaleElement`.

## Selector Escaping

Selectors containing quotes are automatically escaped:

```go
page.Query("[data-name='it\\'s fine']")
```
