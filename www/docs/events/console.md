# Console Events

Listen for `console.log`, `console.error`, and other console API calls from page JavaScript.

## OnConsole

```go
unsub := page.OnConsole(func(msg *bonk.ConsoleMessage) {
    fmt.Printf("[%s] %s\n", msg.Type, msg.Text)
})
defer unsub()
```

## ConsoleMessage

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | `"log"`, `"error"`, `"warn"`, `"info"`, `"debug"` |
| `Text` | `string` | The formatted message text |
| `Args` | `[]any` | Raw arguments passed to the console call |

## Generic Event Handler

You can also use the generic `On` method:

```go
page.On(bonk.ConsoleEvent, func(msg *bonk.ConsoleMessage) {
    fmt.Println(msg.Text)
})
```

!!! warning "Stealth Mode"
    Console events do not work when stealth mode is enabled because `Runtime.enable` is skipped. This is the primary CDP detection signal — enabling it would make the browser detectable.
