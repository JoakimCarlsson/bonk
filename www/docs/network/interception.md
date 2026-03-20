# Network Interception

bonk intercepts network requests and responses using the CDP Fetch domain. You can observe, modify, block, or mock any request.

## Intercept Requests

```go
unsub := page.OnRequest(func(r *bonk.Request) {
    fmt.Println(r.Method, r.URL)
    r.Continue()
})
defer unsub()
```

!!! warning
    You must call `r.Continue()`, `r.Abort()`, or `r.Fulfill()` on every intercepted request. If the handler returns without calling any of these, `Continue()` is called automatically.

## Request Properties

```go
r.URL          // string
r.Method       // string (GET, POST, etc.)
r.Headers      // map[string]string
r.PostData     // string
r.ResourceType // string (Document, Script, Image, etc.)
```

## Modify Headers

```go
page.OnRequest(func(r *bonk.Request) {
    r.SetHeader("Authorization", "Bearer token123")
    r.SetHeader("X-Custom", "value")
    r.Continue()
})
```

## Block Requests

```go
page.OnRequest(func(r *bonk.Request) {
    if strings.Contains(r.URL, "analytics") {
        r.Abort()
        return
    }
    r.Continue()
})
```

## Mock Responses

Respond to a request with custom data:

```go
page.OnRequest(func(r *bonk.Request) {
    if strings.Contains(r.URL, "/api/user") {
        r.Fulfill(200, map[string]string{
            "Content-Type": "application/json",
        }, `{"name": "test"}`)
        return
    }
    r.Continue()
})
```

## Intercept Responses

```go
page.OnResponse(func(r *bonk.Response) {
    fmt.Printf("%d %s\n", r.Status, r.URL)

    body, err := r.Body()
    if err == nil {
        fmt.Println(len(body), "bytes")
    }

    r.Continue()
})
```

## Response Properties

```go
r.URL     // string
r.Status  // int64
r.Headers // map[string]string
```

## Unsubscribe

All handler registration methods return an unsubscribe function:

```go
unsub := page.OnRequest(handler)
// later:
unsub()
```
