# Browser

The `Browser` represents a running Chrome instance. It manages the WebSocket connection, browser contexts, and the Chrome process lifecycle.

## Launch

Start a new Chrome process:

```go
b, err := bonk.Launch()
if err != nil {
    log.Fatal(err)
}
defer b.Close()
```

## Launch Options

```go
b, err := bonk.Launch(
    bonk.Headless(false),              // show browser window
    bonk.Stealth(true),                // anti-detection (default: true)
    bonk.ChromePath("/usr/bin/chromium"), // explicit binary path
    bonk.UserDataDir("./profile"),     // persistent profile directory
    bonk.Args("--proxy-server=socks5://localhost:1080"),
    bonk.Env("DISPLAY=:1"),
    bonk.WithLogger(slog.Default()),   // log all CDP messages
)
```

| Option | Default | Description |
|--------|---------|-------------|
| `Headless(bool)` | `true` | Run Chrome without a visible window |
| `Stealth(bool)` | `true` | Enable anti-detection measures |
| `ChromePath(string)` | auto-detect | Path to Chrome binary |
| `UserDataDir(string)` | temp dir | User data directory (temp dir cleaned up on Close) |
| `Args(...string)` | — | Additional Chrome command-line flags |
| `Env(...string)` | — | Additional environment variables |
| `WithLogger(*slog.Logger)` | — | Enable CDP message logging |

## Connect to Existing Browser

Attach to a Chrome instance that's already running with `--remote-debugging-port`:

```go
b, err := bonk.Connect("ws://localhost:9222/devtools/browser/...")
if err != nil {
    log.Fatal(err)
}
defer b.Close()
```

## Create Contexts

```go
ctx, err := b.NewContext()
```

See [Context](context.md) for context options.

## Close

`Close()` performs a graceful shutdown:

1. Disposes all browser contexts and their pages
2. Sends `Browser.close` CDP command
3. Waits up to 5 seconds for the process to exit
4. Sends SIGTERM, waits another 5 seconds
5. Sends SIGKILL if still running
6. Cleans up temporary directories

```go
b.Close()
```
