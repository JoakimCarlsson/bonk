# Dialog Events

Handle JavaScript dialogs — `alert()`, `confirm()`, and `prompt()`.

## OnDialog

```go
unsub := page.OnDialog(func(d *bonk.Dialog) {
    fmt.Printf("Dialog [%s]: %s\n", d.Type, d.Message)
    d.Accept()
})
defer unsub()
```

## Dialog Properties

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | `"alert"`, `"confirm"`, `"prompt"`, `"beforeunload"` |
| `Message` | `string` | The dialog message text |
| `DefaultPrompt` | `string` | Default value for prompt dialogs |

## Accept

Accept the dialog, optionally providing text for prompt dialogs:

```go
d.Accept()            // accept with no text
d.Accept("my answer") // accept prompt with text
```

## Dismiss

Dismiss (cancel) the dialog:

```go
d.Dismiss()
```

## Generic Event Handler

```go
page.On(bonk.DialogEvent, func(d *bonk.Dialog) {
    if d.Type == "confirm" {
        d.Accept()
    } else {
        d.Dismiss()
    }
})
```
