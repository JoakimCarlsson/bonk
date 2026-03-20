# Context Control

bonk supports Go's `context.Context` for deadlines and cancellation on all page operations.

## WithContext

Returns a shallow copy of the Page that uses the given context for all CDP calls:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

p := page.WithContext(ctx)
err := p.Navigate("https://slow-site.com") // respects 10s deadline
```

The returned Page shares the underlying session and fetch manager — only the context changes.

## Timeout

Shorthand for creating a context with a deadline:

```go
p := page.Timeout(5 * time.Second)
err := p.Navigate("https://slow-site.com")
```

Equivalent to:

```go
ctx, cancel := context.WithTimeout(page.execCtx, 5*time.Second)
defer cancel()
p := page.WithContext(ctx)
```

## Cancellation

Cancel long-running operations:

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(3 * time.Second)
    cancel() // abort after 3 seconds
}()

p := page.WithContext(ctx)
err := p.WaitSelector("#never-appears") // cancelled after 3s
```

## Scope

The context flows through to all operations on the returned Page:

- Navigation (`Navigate`, `Reload`, `GoBack`, `GoForward`)
- JavaScript evaluation (`Evaluate`, `WaitForFunction`)
- Element interaction (via `page.execCtx`)
- Selector waiting (`WaitSelector`, `WaitForURL`)
- Screenshots and PDF
- Network interception

## Element Methods

Element methods use the context from their parent Page. If you called `page.WithContext(ctx)` and then query an element, that element's methods will also respect the context:

```go
p := page.Timeout(5 * time.Second)
el, err := p.WaitSelector("#button") // respects timeout
err = el.Click()                      // also respects timeout
```
