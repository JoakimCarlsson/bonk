# Input

bonk provides methods for simulating user input — clicks, typing, key presses, file uploads, and form interactions.

## Click

```go
el.Click()        // single click
el.DoubleClick()  // double click
el.Hover()        // move mouse to element
```

All mouse methods scroll the element into view and wait for visibility first. Clicks target the center of the element's bounding box.

## Text Input

### Fill

Clears the field and inserts text instantly:

```go
el.Fill("hello@example.com")
```

### Type

Types character by character with key events:

```go
el.Type("search query")
el.Type("slow typing", bonk.WithDelay(100*time.Millisecond))
```

Each character dispatches `keyDown` and `keyUp` events.

## Key Press

Send a single key:

```go
el.Press("Enter")
el.Press("Tab")
el.Press("Escape")
el.Press("ArrowDown")
```

## Focus

```go
el.Focus()
```

## Form Controls

### Select

```go
el.SelectOption("option-value")
```

Dispatches `input` and `change` events.

### Checkbox / Radio

```go
el.Check()    // checks if not already checked
el.Uncheck()  // unchecks if currently checked
```

### File Upload

```go
el.Upload("./document.pdf")
el.Upload("./photo1.jpg", "./photo2.jpg")
```

## Page Shorthand

Page-level methods combine `WaitSelector` with element interaction:

```go
page.Click("#submit")
page.Fill("#email", "test@test.com")
page.Type("#search", "query", bonk.WithDelay(50*time.Millisecond))
page.Press("#input", "Enter")
```

`Click`, `Fill`, and `Press` accept `WaitOption` for configuring the selector wait:

```go
page.Click("#submit", bonk.WaitTimeout(5*time.Second))
```

`Type` accepts `TypeOption`, including `WaitFor` for selector wait options:

```go
page.Type("#search", "query",
    bonk.WithDelay(50*time.Millisecond),
    bonk.WaitFor(bonk.WaitTimeout(5*time.Second)),
)
```
