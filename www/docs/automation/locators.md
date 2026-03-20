# Locators

Locators are Playwright-style selector handles that re-query the DOM on every action. Unlike `Element`, a `Locator` never goes stale — it stores the selector, not the DOM reference.

## Create a Locator

```go
loc := page.Locator("#submit")

// from a frame
loc := frame.Locator(".button")
```

## Actions

All actions wait for the element to appear before acting:

```go
loc.Click()
loc.Fill("hello@example.com")
loc.Type("query", bonk.WithDelay(50*time.Millisecond))
loc.Press("Enter")
```

## Read Properties

```go
text, err := loc.Text()
inner, err := loc.InnerText()
html, err := loc.HTML()
val, err := loc.Attribute("href")
visible, err := loc.IsVisible()
box, err := loc.BoundingBox()
```

## Screenshot

```go
loc.Screenshot("element.png")
```

## Wait

Wait for the element to be attached to the DOM:

```go
err := loc.WaitFor()
err := loc.WaitFor(bonk.WaitTimeout(5 * time.Second))
```

## Count

```go
count, err := loc.Count()
fmt.Printf("Found %d items\n", count)
```

## Nth and First

Select a specific match by index:

```go
first := loc.First()          // same as Nth(0)
third := loc.Nth(2)           // zero-based

text, err := third.Text()
```

## Locator vs Element

| | Locator | Element |
|---|---------|---------|
| Stores | CSS selector | DOM object ID |
| Goes stale | Never | Yes (auto-retries once) |
| Re-queries | Every action | Only on stale error |
| Created via | `page.Locator()` | `page.Query()`, `page.WaitSelector()` |

Use `Locator` when the DOM is dynamic and elements may be re-rendered. Use `Element` for performance when the DOM is stable.

## Example

```go
// scrape a dynamic list that re-renders
items := page.Locator(".product-card")

count, _ := items.Count()
for i := range count {
    name, _ := items.Nth(i).Text()
    price, _ := items.Nth(i).Attribute("data-price")
    fmt.Printf("%s: %s\n", name, price)
}
```
