# Frames

bonk supports interacting with iframes. Each `Frame` has its own isolated JavaScript execution context.

## List Frames

```go
frames, err := page.Frames()
for _, f := range frames {
    fmt.Printf("%s: %s\n", f.Name(), f.URL())
}
```

## Find a Frame

Find by name or ID:

```go
frame, err := page.Frame("iframe-name")
frame, err := page.Frame("frame-id")
```

## Query Elements in a Frame

```go
el, err := frame.Query("#button")
els, err := frame.QueryAll(".item")
el, err := frame.WaitSelector("#loaded")
```

## Interact Within a Frame

```go
frame.Click("#submit")
frame.Fill("#email", "test@test.com")
```

## Execute JavaScript in a Frame

```go
result, err := frame.Evaluate("document.title")
```

JavaScript runs in an isolated world context, so it won't interfere with the frame's page scripts.

## Frame Properties

```go
name := frame.Name()          // name attribute
url := frame.URL()            // document URL
id := frame.ID()              // unique frame ID
```
