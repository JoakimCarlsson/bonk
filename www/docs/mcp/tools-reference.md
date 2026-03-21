# Tools Reference

Detailed parameter reference for all bonkmcp tools. Every tool that operates on a page accepts an optional `page_id` parameter. When omitted, the default page is used.

## Browser Lifecycle

### browser_launch

Launch the browser. Called automatically on first use, but can be called explicitly.

*No parameters.*

### browser_close

Close the browser and all open pages.

*No parameters.*

### list_pages

List all open pages with their IDs and current URLs.

*No parameters.*

Returns a JSON object mapping page IDs to URLs.

### new_page

Open a new browser tab.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | no | URL to navigate to after opening |

### close_page

Close a specific browser tab.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | yes | ID of the page to close |

## Navigation

### navigate

Navigate a page to a URL.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | yes | The URL to navigate to |
| `page_id` | string | no | Target page |
| `wait_until` | string | no | When navigation is complete: `load`, `domcontentloaded`, or `networkidle` |

Returns the final URL and page title.

### go_back

Navigate back in browser history.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

### go_forward

Navigate forward in browser history.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

### reload

Reload the current page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

## Page Content

### screenshot

Take a screenshot. Returns a PNG image.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |
| `full_page` | boolean | no | Capture the full scrollable page |
| `selector` | string | no | CSS selector of a specific element to screenshot |

When `selector` is provided, only that element is captured.

### pdf

Save the page as a PDF.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

Returns base64-encoded PDF data.

### get_content

Get the full HTML content of the page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

### evaluate

Execute JavaScript in the page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `expression` | string | yes | JavaScript expression to evaluate |
| `page_id` | string | no | Target page |

Returns the JSON-serialized result.

## Element Interaction

### click

Click an element.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

### fill

Clear an input field and fill it with text.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the input |
| `text` | string | yes | Text to fill |
| `page_id` | string | no | Target page |

### type_text

Type text character by character, firing individual key events. Useful when `fill` doesn't trigger the expected behavior.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `text` | string | yes | Text to type |
| `delay_ms` | number | no | Delay between keystrokes in milliseconds |
| `page_id` | string | no | Target page |

### select_option

Select a dropdown option by value.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the select element |
| `value` | string | yes | Value of the option to select |
| `page_id` | string | no | Target page |

### check

Check a checkbox.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the checkbox |
| `page_id` | string | no | Target page |

### uncheck

Uncheck a checkbox.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the checkbox |
| `page_id` | string | no | Target page |

### hover

Hover over an element.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

### upload

Upload files to a file input element.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the file input |
| `paths` | string[] | yes | File paths to upload |
| `page_id` | string | no | Target page |

## Element Inspection

### query

Find an element and return its text content and visibility.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector |
| `page_id` | string | no | Target page |

Returns JSON:

```json
{
  "text": "element text content",
  "visible": true
}
```

### query_all

Find all elements matching a selector.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector |
| `page_id` | string | no | Target page |

Returns a count and JSON array of element info.

### wait_for_selector

Wait for an element to appear in the page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector to wait for |
| `visible` | boolean | no | Wait for the element to be visible |
| `hidden` | boolean | no | Wait for the element to be hidden |
| `timeout_ms` | number | no | Timeout in milliseconds (default 30000) |
| `page_id` | string | no | Target page |

## Cookies & State

### get_cookies

Get all cookies from the browser context.

*No parameters.*

Returns a JSON array of cookie objects.

### set_cookies

Set cookies in the browser context.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cookies` | object[] | yes | Array of cookie objects with `name`, `value`, `domain`, `path`, etc. |

### clear_cookies

Clear all cookies from the browser context.

*No parameters.*

### save_state

Save browser state (cookies, localStorage) to a file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | yes | File path to save state to |

### load_state

Load browser state from a file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | yes | File path to load state from |

## Network

### set_extra_headers

Set extra HTTP headers sent with every request from the page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `headers` | object | yes | Map of header name to value |
| `page_id` | string | no | Target page |

### block_urls

Block requests matching URL patterns.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `patterns` | string[] | yes | URL patterns to block (supports `*` wildcards) |
| `page_id` | string | no | Target page |
