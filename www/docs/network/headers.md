# Headers

Set extra HTTP headers that are sent with every request from the page.

## SetExtraHTTPHeaders

```go
err := page.SetExtraHTTPHeaders(map[string]string{
    "Authorization": "Bearer token123",
    "X-Custom":      "value",
    "Accept-Language": "en-US,en;q=0.9",
})
```

These headers are added to all requests made by the page, including navigations, XHR, fetch, and resource loads.

## Per-Request Headers

To modify headers on specific requests, use request interception instead:

```go
page.OnRequest(func(r *bonk.Request) {
    if strings.Contains(r.URL, "/api/") {
        r.SetHeader("Authorization", "Bearer token123")
    }
    r.Continue()
})
```

See [Network Interception](interception.md) for details.
