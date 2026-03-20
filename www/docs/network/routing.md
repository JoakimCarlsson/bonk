# Routing

Routes match requests by URL pattern and provide a higher-level abstraction over request interception.

## Register a Route

```go
unsub := page.Route("**/api/users", func(r *bonk.Route) {
    r.Fulfill(200, map[string]string{
        "Content-Type": "application/json",
    }, `[{"id": 1, "name": "Alice"}]`)
})
defer unsub()
```

## Pattern Matching

Routes use glob patterns:

| Pattern | Matches |
|---------|---------|
| `**/api/*` | Any URL containing `/api/` followed by one path segment |
| `**/*.js` | Any URL ending in `.js` |
| `https://example.com/**` | Any URL on `example.com` |
| `**/users?page=*` | URLs with `/users` and a `page` query parameter |

Special characters:

| Character | Meaning |
|-----------|---------|
| `*` | Match any characters except `/` |
| `**` | Match any characters including `/` |
| `?` | Match exactly one character |

## Route Actions

### Fulfill

Respond with custom data:

```go
r.Fulfill(200, map[string]string{
    "Content-Type": "text/html",
}, "<h1>Mocked</h1>")
```

### Continue

Let the request proceed normally:

```go
r.Continue()
```

### Abort

Block the request:

```go
r.Abort()
```

## Route Properties

```go
r.Request.URL
r.Request.Method
r.Request.Headers
```

## Multiple Routes

Multiple routes can be registered. If a URL matches multiple patterns, the first registered route handles it:

```go
page.Route("**/api/users", handleUsers)
page.Route("**/api/posts", handlePosts)
page.Route("**/api/**", handleOtherAPI)
```
