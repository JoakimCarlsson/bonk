# bonk

A fast, stealth-first browser automation library for Go. Direct Chrome DevTools Protocol over WebSocket — no WebDriver, no relay, no detection.

## Features

- **Direct CDP** — WebSocket to Chrome, zero intermediaries, undetectable by default
- **Stealth by default** — skips `Runtime.enable`, patches navigator/plugins/WebGL, strips headless signals
- **Code-generated protocol bindings** — full coverage of all 55 CDP domains from upstream `.pdl` files
- **Context-aware** — `page.WithContext(ctx)` / `page.Timeout(d)` for deadlines and cancellation
- **Auto-wait elements** — polls for visibility before interaction, retries stale references
- **Network interception** — intercept, modify, mock, or block requests and responses
- **Device emulation** — built-in presets for iPhone 15, Pixel 8, iPad Pro, Galaxy S23, and more
- **Session persistence** — save and restore cookies across runs
- **Isolated contexts** — separate cookie jars, proxies, viewports, and locales per context
- **Single dependency** — only [`coder/websocket`](https://github.com/coder/websocket)

## Install

```bash
go get github.com/joakimcarlsson/bonk
```

## Quick Example

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

## Next Steps

- [Installation](getting-started/installation.md) — requirements and setup
- [Quick Start](getting-started/quick-start.md) — your first automation script
- [Browser](core/browser.md) — launching and connecting to Chrome
- [Stealth](stealth/overview.md) — how anti-detection works
