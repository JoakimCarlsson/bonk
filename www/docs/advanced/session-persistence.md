# Session Persistence

bonk can save and restore browser state (cookies) between sessions. This is useful for maintaining login sessions across automation runs.

## Save State

```go
err := ctx.SaveState("./session.dat")
```

Serializes cookies to a JSON file.

## Load State

### On Context Creation

```go
ctx, err := b.NewContext(bonk.WithState("./session.dat"))
```

### On Existing Context

```go
err := ctx.LoadState("./session.dat")
```

## What's Saved

Currently, state persistence saves and restores:

- All cookies (name, value, domain, path, expiry, httpOnly, secure, sameSite)

## Cookie Management

For finer-grained control, use the cookie methods directly:

```go
// get all cookies
cookies, err := ctx.Cookies()

// set specific cookies
err = ctx.SetCookies(
    bonk.Cookie{
        Name:     "session",
        Value:    "abc123",
        Domain:   ".example.com",
        Path:     "/",
        Secure:   true,
        HTTPOnly: true,
    },
)

// clear all cookies
err = ctx.ClearCookies()
```

## Use Cases

### Login Once, Reuse Session

```go
// first run: log in and save
page.Navigate("https://app.example.com/login")
page.Fill("#email", "user@example.com")
page.Fill("#password", "secret")
page.Click("#submit")
page.WaitForURL("**/dashboard*")
ctx.SaveState("./session.dat")

// subsequent runs: restore session
ctx, _ := b.NewContext(bonk.WithState("./session.dat"))
page, _ := ctx.NewPage()
page.Navigate("https://app.example.com/dashboard") // already logged in
```

### Resumable Crawlers

```go
// save progress periodically
for i, url := range urls {
    page.Navigate(url)
    // ... process page ...

    if i%10 == 0 {
        ctx.SaveState("./crawler-state.dat")
    }
}
```
