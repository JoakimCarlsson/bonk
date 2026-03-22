# MCP Server

bonkmcp is an [MCP](https://modelcontextprotocol.io/) server that exposes bonk's browser automation as tools for AI agents. Any MCP-compatible client (Claude Code, Claude Desktop, Cursor, etc.) can launch a browser, navigate pages, take screenshots, click elements, and more.

## Install

```bash
go install github.com/joakimcarlsson/bonk/cmd/bonkmcp@latest
```

## Usage with Claude Code

Add to your MCP configuration:

```json
{
  "mcpServers": {
    "bonk": {
      "command": "bonkmcp",
      "args": ["--headless"]
    }
  }
}
```

The AI agent can then call tools like `navigate`, `screenshot`, `click`, and `fill` to drive a browser.

## Usage with SSE

For remote or multi-client access, run in SSE mode:

```bash
bonkmcp -transport sse -port 8080
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-transport` | `stdio` | Transport mode: `stdio` or `sse` |
| `-port` | `8080` | Port for SSE transport |
| `-headless` | `true` | Run Chrome in headless mode |
| `-stealth` | `true` | Enable stealth mode |
| `-chrome` | *(auto)* | Path to Chrome binary |

## Session Behavior

bonkmcp manages a single browser instance across all tool calls:

- The browser launches automatically on the first tool that needs it (e.g. `navigate`), or explicitly via `browser_launch`.
- A single browser context is shared across all pages. Cookies and state are shared.
- Pages are tracked by ID (`page_1`, `page_2`, etc.). Most tools accept an optional `page_id` parameter. When omitted, the default (first) page is used.
- Call `new_page` to open additional tabs, `close_page` to close them, and `list_pages` to see what's open.

## Available Tools

### Browser Lifecycle

| Tool | Description |
|------|-------------|
| `browser_launch` | Launch the browser explicitly |
| `browser_close` | Close the browser and all pages |
| `list_pages` | List open pages with IDs and URLs |
| `new_page` | Open a new tab, optionally navigate |
| `close_page` | Close a specific tab |

### Navigation

| Tool | Description |
|------|-------------|
| `navigate` | Navigate to a URL |
| `go_back` | Go back in history |
| `go_forward` | Go forward in history |
| `reload` | Reload the page |

### Page Content

| Tool | Description |
|------|-------------|
| `screenshot` | Take a screenshot (returns PNG image) |
| `pdf` | Save page as PDF (returns base64) |
| `get_content` | Get the page HTML |
| `evaluate` | Execute JavaScript |

### Element Interaction

| Tool | Description |
|------|-------------|
| `click` | Click an element |
| `fill` | Fill an input field |
| `type_text` | Type text character by character |
| `select_option` | Select a dropdown option |
| `check` / `uncheck` | Toggle checkboxes |
| `hover` | Hover over an element |
| `upload` | Upload files to a file input |

### Element Inspection

| Tool | Description |
|------|-------------|
| `query` | Find an element, return text and visibility |
| `query_all` | Find all matching elements |
| `wait_for_selector` | Wait for an element to appear |

### Cookies & State

| Tool | Description |
|------|-------------|
| `get_cookies` | List all cookies |
| `set_cookies` | Set cookies |
| `clear_cookies` | Clear all cookies |
| `save_state` | Save session to a file |
| `load_state` | Restore session from a file |

### Network

| Tool | Description |
|------|-------------|
| `set_extra_headers` | Set HTTP headers on all requests |
| `block_urls` | Block requests matching patterns |

## Example Workflow

A typical AI agent session might look like:

1. `navigate` to `https://example.com`
2. `screenshot` to see the page
3. `query_all` with `a` to find links
4. `click` on a specific link
5. `fill` a search box with text
6. `screenshot` to verify the result
7. `get_content` to extract data

For detailed parameter documentation for each tool, see the [Tools Reference](tools-reference.md).
