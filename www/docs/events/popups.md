# Popups & New Tabs

Detect and interact with new pages opened by `window.open` or `target="_blank"` links.

## Wait for a Popup

```go
// start waiting before triggering the popup
popupCh := make(chan *bonk.Page, 1)
go func() {
    popup, err := page.WaitForPopup()
    if err != nil {
        log.Fatal(err)
    }
    popupCh <- popup
}()

// trigger the popup
page.Click("a[target=_blank]")

// use the popup page
popup := <-popupCh
defer popup.Close()
popup.WaitSelector("#content")
```

## With window.open

```go
go func() {
    popup, _ := page.WaitForPopup()
    popupCh <- popup
}()

page.Evaluate("window.open('https://accounts.google.com/login', '_blank')")
popup := <-popupCh
```

## Options

| Option | Description |
|--------|-------------|
| `PopupTimeout(d time.Duration)` | Maximum time to wait (default: 30s) |

## Custom Timeout

```go
popup, err := page.WaitForPopup(bonk.PopupTimeout(5 * time.Second))
if err != nil {
    log.Fatal("no popup appeared within 5s")
}
```

## Common Use Cases

- **OAuth flows** — click "Login with Google", wait for the OAuth popup, fill credentials, then return to the main page
- **Payment pages** — handle payment provider popups
- **External links** — verify that `target="_blank"` links open correctly
