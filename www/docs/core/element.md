# Element

An `Element` represents a DOM element on a page. Elements are obtained through queries and support interaction, property reading, and screenshots.

## Query

```go
el, err := page.Query("#login")          // first match, nil if not found
els, err := page.QueryAll(".item")       // all matches
el, err := page.WaitSelector("#login")   // poll until found
```

`WaitSelector` accepts wait options:

```go
el, err := page.WaitSelector("#login",
    bonk.WaitTimeout(10*time.Second),
    bonk.WaitInterval(100*time.Millisecond),
)
```

## Interaction

All interaction methods auto-wait for the element to become visible before acting:

```go
err = el.Click()
err = el.DoubleClick()
err = el.Hover()
err = el.Fill("hello@example.com")
err = el.Type("search query", bonk.WithDelay(100*time.Millisecond))
err = el.Press("Enter")
err = el.SelectOption("option-value")
err = el.Check()
err = el.Uncheck()
err = el.Upload("./file.pdf")
err = el.Focus()
err = el.ScrollIntoView()
```

## Read Properties

```go
text, err := el.Text()           // textContent
inner, err := el.InnerText()     // innerText (rendered text)
html, err := el.HTML()           // outerHTML
val, err := el.Attribute("href")
visible, err := el.IsVisible()
box, err := el.BoundingBox()     // *Box{X, Y, Width, Height}
```

## Element Screenshot

```go
err = el.Screenshot("element.png")
```

## Stale Element Retry

If an element's DOM node is garbage collected (e.g. the page re-rendered), bonk automatically re-queries using the original CSS selector and retries the operation once. Elements created from `EvaluateHandle` (no selector) fail immediately with `ErrStaleElement`.

## Shorthand Methods

Page-level methods combine `WaitSelector` + element interaction:

```go
page.Click("#submit")
page.Fill("#email", "test@test.com")
page.Type("#search", "query")
page.Press("#input", "Enter")
```
