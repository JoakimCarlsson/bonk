# Stealth

Stealth mode is enabled by default. It makes Chrome automation undetectable by modern anti-bot systems (Cloudflare, DataDome, Kasada) through four layers of protection.

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

The primary detection signal in 2025 is `Runtime.enable`. When called, Chrome dispatches console events — detectors use `Error.prepareStackTrace` to check if something is listening.

In stealth mode, bonk skips `Runtime.enable` entirely. This means:

| Works | Doesn't Work |
|-------|--------------|
| `Runtime.evaluate` | `Runtime.consoleAPICalled` events |
| `Runtime.callFunctionOn` | `Runtime.executionContextCreated` events |
| All element interaction | `OnConsole` handler |
| Navigation | `Runtime.exceptionThrown` events |
| Screenshots, PDF | — |

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
