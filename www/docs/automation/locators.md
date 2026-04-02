# Locators

Locators are Playwright-style selector handles that re-query the DOM on every action. Unlike `Element`, a `Locator` never goes stale — it stores the selector, not the DOM reference.

## Create a Locator

```go
loc := page.Locator("#submit")

// from a frame
loc := frame.Locator(".button")
```

[Accessibility selectors](accessibility.md) also return Locators:

```go
loc := page.GetByRole("button", bonk.WithName("Submit"))
loc := page.GetByLabel("Email")
loc := page.GetByText("Welcome")
```

## Actions

All actions wait for the element to appear before acting:

```go
loc.Click()
loc.Fill("hello@example.com")
loc.Type("query", bonk.WithDelay(50*time.Millisecond))
loc.Press("Enter")
loc.Focus()
loc.Blur()
loc.Clear()
```

## Read Properties

```go
text, err := loc.Text()
inner, err := loc.InnerText()
html, err := loc.HTML()
val, err := loc.Attribute("href")
visible, err := loc.IsVisible()
checked, err := loc.IsChecked()
disabled, err := loc.IsDisabled()
editable, err := loc.IsEditable()
box, err := loc.BoundingBox()
```

## Bulk Text

Retrieve text from all matching elements at once:

```go
texts, err := loc.AllInnerTexts()      // rendered innerText of each match
contents, err := loc.AllTextContents()  // textContent of each match
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

## Filter

Narrow matches by text content or sub-locator containment. All criteria are AND'd together.

### HasText / HasNotText

```go
rows := page.Locator("tr")
activeRows := rows.Filter(bonk.LocatorFilter{HasText: "Active"})
withoutDraft := rows.Filter(bonk.LocatorFilter{HasNotText: "Draft"})
```

### Has / HasNot

Keep or exclude elements that contain a descendant matching a sub-locator:

```go
cards := page.Locator(".card")

// only cards that contain an <h2>
withHeading := cards.Filter(bonk.LocatorFilter{
    Has: page.Locator("h2"),
})

// cards without a .badge element
withoutBadge := cards.Filter(bonk.LocatorFilter{
    HasNot: page.Locator(".badge"),
})
```

Sub-locators can be CSS or accessibility selectors:

```go
// cards containing the text "Total"
cards.Filter(bonk.LocatorFilter{
    Has: page.GetByText("Total"),
})
```

### Combining criteria

```go
// active cards with a heading
cards.Filter(bonk.LocatorFilter{
    HasText: "Active",
    Has:     page.Locator("h2"),
})
```

### Chaining filters

```go
page.Locator(".card").
    Filter(bonk.LocatorFilter{HasText: "Active"}).
    Filter(bonk.LocatorFilter{Has: page.Locator(".price")})
```

## And

Match elements that satisfy **both** locators (intersection):

```go
loc := page.Locator(".highlight").And(page.Locator(".visible"))

// combine CSS with accessibility selectors
submitBtn := page.Locator("button").And(page.GetByText("Submit"))
```

## Or

Match elements that satisfy **either** locator (union). Results are in document order:

```go
alerts := page.Locator(".error").Or(page.Locator(".warning"))

count, _ := alerts.Count()
fmt.Printf("Found %d alerts\n", count)
```

## Locator vs Element

| | Locator | Element |
|---|---------|---------|
| Stores | Selector or JS expression | DOM object ID |
| Goes stale | Never | Yes (auto-retries once) |
| Re-queries | Every action | Only on stale error |
| Created via | `page.Locator()`, `page.GetByRole()`, etc. | `page.Query()`, `page.WaitSelector()` |

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
