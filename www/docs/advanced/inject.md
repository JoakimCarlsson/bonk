# Script & Style Injection

`AddScriptTag` and `AddStyleTag` inject `<script>` and `<style>`/`<link>` tags into the page's `<head>`. Useful for loading polyfills, test helpers, or custom styles.

## AddScriptTag

### By URL

Loads an external script and waits for it to finish loading:

```go
err := page.AddScriptTag(bonk.ScriptTagURL("https://cdn.example.com/lib.js"))
```

### By Content

Injects inline JavaScript:

```go
err := page.AddScriptTag(bonk.ScriptTagContent(`
    window.testHelper = function() { return 42; }
`))
```

### ES Module

```go
err := page.AddScriptTag(
    bonk.ScriptTagURL("https://cdn.example.com/module.js"),
    bonk.ScriptTagType("module"),
)
```

### Options

| Option | Description |
|--------|-------------|
| `ScriptTagURL(url)` | Set the script `src` URL |
| `ScriptTagContent(js)` | Set inline script content |
| `ScriptTagType(t)` | Set the `type` attribute (e.g. `"module"`) |

## AddStyleTag

### By URL

Loads an external stylesheet and waits for it to finish loading:

```go
err := page.AddStyleTag(bonk.StyleTagURL("https://cdn.example.com/style.css"))
```

### By Content

Injects inline CSS:

```go
err := page.AddStyleTag(bonk.StyleTagContent(`
    body { background: red; }
    .hidden { display: none; }
`))
```

### Options

| Option | Description |
|--------|-------------|
| `StyleTagURL(url)` | Set the stylesheet `href` URL |
| `StyleTagContent(css)` | Set inline CSS content |

## AddScriptTag vs AddInitScript

| | `AddScriptTag` | `AddInitScript` |
|---|---|---|
| **When** | Runs once on the current document | Runs on every new document (including navigations) |
| **Waits for load** | Yes (for URL scripts) | No |
| **Use case** | Inject a library into the current page | Override APIs before any page scripts run |
