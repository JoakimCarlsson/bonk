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

## Context Configuration

### set_default_timeout

Set the default timeout for all wait/query operations.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `timeout_ms` | number | yes | Timeout in milliseconds |

### set_default_navigation_timeout

Set the default timeout for navigation operations (navigate, reload, go_back, go_forward).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `timeout_ms` | number | yes | Timeout in milliseconds |

### grant_permissions

Grant browser permissions such as geolocation, notifications, camera, microphone.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `permissions` | string[] | yes | List of permissions (e.g. `geolocation`, `notifications`, `audioCapture`, `videoCapture`) |
| `origin` | string | no | Origin to scope permissions to |

### clear_permissions

Reset all permission overrides for the browser context.

*No parameters.*

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

### wait_for_load_state

Wait for the page to reach a specific load state. Useful after actions that trigger navigation without using `navigate` (e.g. form submissions, SPA route changes).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `state` | string | yes | Load state: `load`, `domcontentloaded`, or `networkidle` |
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

### snapshot

Get the accessibility tree of the page as indexed, structured text. Interactive elements get numbered indices (`[N]`). Use it to understand page structure before acting; the numbered elements can be clicked directly with `click_ref`.

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

### pause

Pause execution for manual inspection in headed mode. Resumes when the user presses Enter on stdin.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page_id` | string | no | Target page |

### add_script_tag

Inject a `<script>` tag into the page. Provide either `url` or `content`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | no | URL of the script to load |
| `content` | string | no | Inline script content |
| `type` | string | no | Script type attribute (e.g. `module`) |
| `page_id` | string | no | Target page |

### add_style_tag

Inject a `<style>` or `<link>` stylesheet tag into the page. Provide either `url` or `content`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | no | URL of the stylesheet to load |
| `content` | string | no | Inline CSS content |
| `page_id` | string | no | Target page |

### wait_for_event

Wait for the next occurrence of a page event and return the event payload.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `event` | string | yes | Event type: `console`, `dialog`, or `download` |
| `timeout_ms` | number | no | Timeout in milliseconds (default 30000) |
| `page_id` | string | no | Target page |

Returns the JSON-serialized event payload.

## Element Interaction

### click

Click an element.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

### click_ref

Click the interactive element at the given index from the page's most recent `snapshot` (the number shown as `[N]`). More reliable than a selector for elements read from the snapshot, since it acts on that exact node. Call `snapshot` first.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ref` | number | yes | The `[N]` index of the element from the latest snapshot |
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

### dispatch_event

Fire a DOM event on an element programmatically. Useful when simulated clicks don't trigger framework-level handlers.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `event_type` | string | yes | DOM event type (e.g. `click`, `input`, `change`, `submit`) |
| `event_init` | object | no | Event init options (e.g. `{"bubbles": true}`). Defaults to `{bubbles: true}` |
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

### is_checked

Check if a checkbox or radio button is checked.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

Returns `"true"` or `"false"`.

### is_disabled

Check if an element is disabled.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

Returns `"true"` or `"false"`.

### is_editable

Check if an element is editable (input, textarea, select, or contenteditable that is not disabled or readonly).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `selector` | string | yes | CSS selector of the element |
| `page_id` | string | no | Target page |

Returns `"true"` or `"false"`.

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

## Overlay Handlers

### add_locator_handler

Register an auto-dismiss handler. When the locator selector is visible before an action, the `click_selector` is clicked to dismiss it. Useful for cookie banners, notification popups, and overlay dialogs.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `locator` | string | yes | CSS selector that detects the overlay |
| `click_selector` | string | yes | CSS selector of the element to click to dismiss |
| `page_id` | string | no | Target page |

### remove_locator_handler

Remove a previously registered overlay auto-dismiss handler.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `locator` | string | yes | CSS selector used when registering the handler |
| `page_id` | string | no | Target page |
