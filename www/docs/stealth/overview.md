# Stealth

Stealth mode is enabled by default. It makes Chrome automation undetectable by modern anti-bot systems (Cloudflare, DataDome, Kasada) through five layers of protection.

## Enable/Disable

```go
// stealth is on by default
b, _ := bonk.Launch()

// explicitly disable
b, _ := bonk.Launch(bonk.Stealth(false))
```

## Layer 1: Chrome Flags

Launch arguments that prevent detection at the browser level:

- `--disable-blink-features=AutomationControlled` — prevents Blink from setting `navigator.webdriver`
- `--disable-features=IsolateOrigins,site-per-process`
- `--disable-infobars` — hides "Chrome is being controlled" bar
- `--disable-session-crashed-bubble`
- `--disable-search-engine-choice-screen`
- `--window-size=1920,1080` in headless mode

## Layer 2: CDP Domain Control

Advanced bot detection (Cloudflare Turnstile, DataDome) can fingerprint which CDP domains are active. In stealth mode, bonk defers all domain activation until a feature actually needs it:

- **`Runtime.enable`** — never called. Detectors use `Error.prepareStackTrace` to check if something is listening to console events.
- **`Page.enable`** + **`Page.setLifecycleEventsEnabled`** — deferred until first `Navigate()`, `Reload()`, `GoBack()`, `GoForward()`, `WaitNavigation()`, `OnDialog()`, or download handler call.
- **`Network.enable`** — deferred until first `SetExtraHTTPHeaders()`, `SetOffline()`, or `WaitNetworkIdle` navigation.

A page that only uses `Evaluate()` and `QuerySelector()` never activates any domain, keeping the CDP footprint at zero — matching the behavior of [nodriver](https://github.com/ultrafunkamsterdam/nodriver).

CDP **commands** (`Page.navigate`, `Page.captureScreenshot`, `Runtime.evaluate`, etc.) work without their domain being enabled. Only **event subscriptions** require it.

| Works without domain enabling | Requires lazy activation |
|-------------------------------|--------------------------|
| `Evaluate`, `QuerySelector` | `Navigate`, `Reload` (Page events) |
| `Screenshot`, `PDF` | `OnDialog`, download handler (Page events) |
| `AddInitScript`, `BringToFront` | `WaitNetworkIdle` (Network events) |
| `SetViewport`, `SetContent` | `SetExtraHTTPHeaders`, `SetOffline` |
| `OnRequest`, `OnResponse`, `Route` (Fetch domain) | `OnConsole` (Runtime — never enabled) |

## Layer 3: JavaScript Patches

Injected via `Page.addScriptToEvaluateOnNewDocument` before any page script runs:

| Patch | What It Does |
|-------|-------------|
| `navigator.webdriver` | Returns `undefined` instead of `true` |
| `navigator.plugins` | Returns array with PDF Viewer, Chrome PDF Plugin, Native Client |
| `navigator.languages` | Returns `['en-US', 'en']` |
| `window.chrome.runtime` | Adds fake `connect()` and `sendMessage()` |
| `navigator.permissions.query` | Returns consistent state for notifications |
| WebGL parameters | Returns Intel GPU vendor and renderer |
| `navigator.hardwareConcurrency` | Returns `8` |
| `navigator.deviceMemory` | Returns `8` |
| `Function.prototype.toString` | Proxied so patched functions return `[native code]` |

## Layer 4: Network Headers

- Strips "Headless" from the User-Agent string
- Sets User-Agent Client Hints (`sec-ch-ua`) with correct Chrome brand versions
- Ensures `Accept-Language` matches `navigator.languages`
- Detects platform (macOS/Windows/Linux) and architecture (arm/x86) for metadata consistency

## Trade-offs

Stealth mode disables `OnConsole`. If you need console events for debugging, disable stealth:

```go
b, _ := bonk.Launch(bonk.Stealth(false))
```

## Testing

Sites for verifying stealth effectiveness:

- [bot.sannysoft.com](https://bot.sannysoft.com/) — basic detection tests
- [abrahamjuliot.github.io/creepjs](https://abrahamjuliot.github.io/creepjs/) — fingerprinting analysis
- [nowsecure.nl](https://nowsecure.nl/) — Cloudflare challenge test
