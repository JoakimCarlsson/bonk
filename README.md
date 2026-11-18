# bonk 🔨

[![Go Reference](https://pkg.go.dev/badge/github.com/joakimcarlsson/bonk.svg)](https://pkg.go.dev/github.com/joakimcarlsson/bonk)

> *Point at a website. Bonk it. Walk away whistling.*

bonk is a fast, sneaky browser automation library for Go. It talks to Chrome over raw DevTools Protocol on a WebSocket: no WebDriver, no relay, no middlemen, and no "are you a robot?" side-eye from the sites you visit.

📚 [Read the docs](https://joakimcarlsson.github.io/bonk)

## Why bonk?

- 🥷 **Stealth by default.** Skips `Runtime.enable`, patches navigator, plugins and WebGL, and scrubs the headless tells. Bot detectors see a perfectly boring human.
- ⚡ **Direct CDP.** A straight WebSocket to Chrome. Nothing in between to slow you down or snitch on you.
- 🤖 **Every CDP domain, generated.** All 55 domains, built from the upstream `.pdl` files, so you never hit a wall.
- ⏱️ **Context-aware.** `page.WithContext(ctx)` and `page.Timeout(d)` for deadlines and cancellation, the Go way.
- 🎯 **Auto-waiting elements.** Waits until things are visible before poking them, and quietly retries stale references.
- 🕸️ **Network interception.** Intercept, rewrite, mock or flat-out block requests and responses.
- 📱 **Device emulation.** Pretend to be an iPhone 15, Pixel 8, iPad Pro, Galaxy S23 and friends.
- 🍪 **Session persistence.** Save cookies and localStorage, come back tomorrow, still logged in.
- 🧳 **Isolated contexts.** Separate cookie jars, proxies, viewports and locales, all in one browser.
- 🪶 **One dependency.** Just [`coder/websocket`](https://github.com/coder/websocket). That's the whole list.

## Get it

```bash
go get github.com/joakimcarlsson/bonk
```

## Bonk your first page

```go
package main

import (
	"fmt"
	"log"

	"github.com/joakimcarlsson/bonk"
)

func main() {
	b, err := bonk.Launch()
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	ctx, err := b.NewContext()
	if err != nil {
		log.Fatal(err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	if err := page.Navigate("https://example.com"); err != nil {
		log.Fatal(err)
	}

	title, _ := page.Title()
	fmt.Println(title)
}
```

That's it. Chrome opens, the page loads, the title gets printed, nobody notices a thing. 🎉

## More bonking

### Click things like a human would 🖱️

```go
page.Fill("input[name=email]", "me@example.com")
page.GetByRole("button", bonk.WithName("Sign in")).Click()
```

### Mess with the network 🕸️

```go
page.OnRequest(func(r *bonk.Request) {
	if strings.HasSuffix(r.URL, ".css") {
		r.Abort()
		return
	}
	r.Continue()
})

page.Route("**/api/user", func(r *bonk.Route) {
	r.Fulfill(200, map[string]string{"Content-Type": "application/json"}, `{"name":"bonk"}`)
})
```

### Become a phone 📱

```go
page.Emulate(bonk.IPhone15)
page.Screenshot("tiny.png")
```

## 🤖 Hand the browser to your AI

bonk ships an [MCP](https://modelcontextprotocol.io/) server, so Claude, Cursor and friends can drive a real, stealthy Chrome for you.

```bash
go install github.com/joakimcarlsson/bonk/cmd/bonkmcp@latest
```

```json
{
  "mcpServers": {
    "bonk": { "command": "bonkmcp", "args": ["--headless"] }
  }
}
```

More in the [MCP docs](https://joakimcarlsson.github.io/bonk/mcp/overview/).

## Contributing

Found a bug or have an idea? Open an issue or a PR. Run `go test ./...` first; the tests launch a real Chrome, so make sure you have one.

## License

See [LICENSE](LICENSE). Go forth and bonk responsibly. 🔨
