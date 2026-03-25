# Accessibility Selectors

Accessibility selectors find elements by their semantic attributes — ARIA roles, labels, text content, and other user-facing properties. They are more resilient than CSS selectors because they target what the user sees, not implementation details.

All methods are available on both `Page` and `Frame`, and return a `*Locator` for chaining.

## GetByRole

Find elements by their ARIA role. Matches both explicit `role` attributes and implicit roles from HTML semantics.

```go
page.GetByRole("button").Click()
page.GetByRole("link").First().Text()
page.GetByRole("heading").Count()
```

### Implicit Role Mapping

Common HTML elements are matched by their implicit ARIA role:

| Role | HTML Elements |
|------|---------------|
| `button` | `<button>`, `<input type="button\|submit\|reset">` |
| `link` | `<a href="...">` |
| `heading` | `<h1>` – `<h6>` |
| `textbox` | `<input type="text\|email\|password\|search\|tel\|url">`, `<input>` (no type), `<textarea>` |
| `checkbox` | `<input type="checkbox">` |
| `radio` | `<input type="radio">` |
| `img` | `<img alt="...">` |
| `list` | `<ul>`, `<ol>` |
| `listitem` | `<li>` |
| `navigation` | `<nav>` |
| `main` | `<main>` |
| `complementary` | `<aside>` |
| `table` | `<table>` |
| `row` | `<tr>` |
| `cell` | `<td>` |
| `columnheader` | `<th>` |

### Filter by Name

Use `WithName` to filter by the element's accessible name (text content, `aria-label`, or `aria-labelledby`):

```go
page.GetByRole("button", bonk.WithName("Submit")).Click()
```

For an exact name match, pass `Exact()`:

```go
page.GetByRole("button", bonk.WithName("Submit", bonk.Exact())).Click()
```

Without `Exact()`, name matching is substring-based — `"Submit"` matches both `"Submit"` and `"Submit Form"`.

## GetByLabel

Find form controls by their associated label. Checks `<label>` elements (both `for` attribute and nested inputs), `aria-label`, and `aria-labelledby`.

```go
page.GetByLabel("Email").Fill("user@example.com")
page.GetByLabel("Password").Fill("secret")
page.GetByLabel("Remember me").Click()
```

Works with all label association patterns:

```go
// <label for="email">Email</label><input id="email">
page.GetByLabel("Email")

// <label>Username <input type="text"></label>
page.GetByLabel("Username")

// <input aria-label="Search">
page.GetByLabel("Search")
```

## GetByText

Find elements by their visible text content. Matches elements that directly own the text (not ancestors that contain it deep in their subtree).

```go
page.GetByText("Welcome back").Text()
page.GetByText("Sign in").Click()
```

Default matching is substring. Use `Exact()` for exact match:

```go
page.GetByText("Hello", bonk.Exact())
```

## GetByPlaceholder

Find elements by their `placeholder` attribute:

```go
page.GetByPlaceholder("Search...").Fill("bonk go")
page.GetByPlaceholder("Enter email", bonk.Exact()).Fill("a@b.com")
```

## GetByTestID

Find elements by their `data-testid` attribute. Always uses exact matching:

```go
page.GetByTestID("submit-button").Click()
page.GetByTestID("user-avatar").Screenshot("avatar.png")
```

## GetByAltText

Find elements by their `alt` attribute:

```go
page.GetByAltText("Company Logo").Screenshot("logo.png")
page.GetByAltText("Logo").Attribute("src")  // substring match
```

## GetByTitle

Find elements by their `title` attribute:

```go
page.GetByTitle("Close dialog").Click()
page.GetByTitle("Close", bonk.Exact()).Click()
```

## Text Matching

All text-based selectors (except `GetByTestID`) default to **substring** matching. Pass `bonk.Exact()` to require an exact match (after trimming whitespace):

| Call | Matches `"Submit Form"` |
|------|------------------------|
| `GetByText("Submit")` | Yes (substring) |
| `GetByText("Submit", bonk.Exact())` | No |
| `GetByText("Submit Form", bonk.Exact())` | Yes |

## Chaining

All methods return a `*Locator`, so the full Locator API is available:

```go
page.GetByRole("listitem").Count()
page.GetByRole("listitem").First().Text()
page.GetByRole("listitem").Nth(2).Click()
page.GetByRole("button", bonk.WithName("Save")).WaitFor()
```

## Frames

All methods work on frames too:

```go
frame, _ := page.Frame("content")
frame.GetByRole("button").Click()
frame.GetByLabel("Search").Fill("query")
```

## When to Use Which

| Selector | Best For |
|----------|----------|
| `GetByRole` | Buttons, links, headings, form controls — anything with ARIA semantics |
| `GetByLabel` | Form inputs associated with a label |
| `GetByText` | Static text content, paragraphs, list items |
| `GetByPlaceholder` | Inputs identified by placeholder text |
| `GetByTestID` | Elements with explicit test IDs — most stable but least semantic |
| `GetByAltText` | Images |
| `GetByTitle` | Elements with title tooltips |
| `page.Locator(css)` | When you need a CSS selector (class, ID, complex structure) |
