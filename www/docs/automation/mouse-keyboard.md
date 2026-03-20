# Mouse & Keyboard

Standalone `Mouse` and `Keyboard` controllers for low-level input that isn't tied to a specific element. Use these for drag-and-drop, keyboard shortcuts, and coordinate-based interactions.

## Mouse

```go
mouse := page.Mouse()
```

### Move

```go
mouse.Move(100, 200)
```

### Click

Moves to the coordinates and clicks:

```go
mouse.Click(100, 200)
```

### Down / Up

Press and release the mouse button separately:

```go
mouse.Down()
// ... move ...
mouse.Up()
```

### Drag

Drag from one point to another with smooth movement:

```go
mouse.DragTo(100, 100, 300, 300)
```

Performs 10 intermediate move steps with 10ms delay between each.

## Keyboard

```go
kb := page.Keyboard()
```

### Press

Key down + key up:

```go
kb.Press("Enter")
kb.Press("Tab")
kb.Press("Escape")
```

### Down / Up

Press and release keys separately (for key combinations):

```go
kb.Down("Control")
kb.Press("a")
kb.Up("Control")
```

### Type

Type text character by character with key events:

```go
kb.Type("hello world")
kb.Type("slow", bonk.WithDelay(100*time.Millisecond))
```

### InsertText

Insert text instantly without key events:

```go
kb.InsertText("pasted content")
```

## When to Use

| Scenario | Use |
|----------|-----|
| Click a button | `el.Click()` or `page.Click("#btn")` |
| Click at coordinates | `page.Mouse().Click(x, y)` |
| Drag and drop | `page.Mouse().DragTo(...)` |
| Type in a field | `el.Type("text")` or `page.Type("#input", "text")` |
| Keyboard shortcut | `page.Keyboard().Down("Control"); page.Keyboard().Press("c"); page.Keyboard().Up("Control")` |
| Paste text | `page.Keyboard().InsertText("text")` |
