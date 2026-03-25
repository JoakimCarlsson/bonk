# Pause

`Pause` blocks script execution so you can manually inspect the browser in headed mode. Execution resumes when you press Enter on stdin.

## Usage

```go
page.Navigate("https://example.com")

page.Pause()

page.Click("#button")
```

When `Pause()` is called, the message `bonk: paused — press Enter to continue...` is printed to stderr and the program blocks until you press Enter.

## When to Use

- Debugging automation scripts in headed mode
- Inspecting page state at a specific point in the flow
- Manually verifying element positions before an automated click
- Pausing between steps to observe visual behavior

!!! tip
    `Pause` is most useful when the browser is running in headed mode (`bonk.Headless(false)`). In headless mode the browser window isn't visible, so there is nothing to inspect.
