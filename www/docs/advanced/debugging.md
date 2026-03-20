# Debugging

bonk provides CDP message logging for debugging automation scripts.

## Enable Logging

Pass an `*slog.Logger` at launch to log all CDP messages sent and received:

```go
import "log/slog"

b, err := bonk.Launch(
    bonk.WithLogger(slog.Default()),
)
```

## Log Output

Every CDP message is logged at `DEBUG` level:

```
DEBUG send data={"id":1,"method":"Target.createBrowserContext",...}
DEBUG recv data={"id":1,"result":{"browserContextId":"..."}}
DEBUG send data={"id":2,"method":"Target.createTarget",...}
DEBUG recv data={"id":2,"result":{"targetId":"..."}}
```

## Custom Logger

Use a custom `slog.Handler` to control output format and destination:

```go
handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
logger := slog.New(handler)

b, err := bonk.Launch(bonk.WithLogger(logger))
```

## Headless Debugging

Run with `Headless(false)` to see what the browser is doing:

```go
b, err := bonk.Launch(
    bonk.Headless(false),
    bonk.WithLogger(slog.Default()),
)
```

## Screenshots for Debugging

Take screenshots at key points in your script to see the page state:

```go
page.Navigate("https://example.com")
page.Screenshot("debug-1-after-navigate.png")

page.Click("#login")
page.Screenshot("debug-2-after-click.png")
```
