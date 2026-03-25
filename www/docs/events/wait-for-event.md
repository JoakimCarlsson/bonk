# Wait for Event

`WaitForEvent` is a one-shot wait that blocks until the next occurrence of a page event and returns the event payload.

## Usage

```go
val, err := page.WaitForEvent(bonk.ConsoleEvent)
msg := val.(*bonk.ConsoleMessage)
fmt.Println(msg.Text)
```

## Supported Events

| Event | Return Type |
|-------|-------------|
| `ConsoleEvent` | `*ConsoleMessage` |
| `DialogEvent` | `*Dialog` |
| `DownloadEvent` | `*Download` |

## Options

`WaitForEvent` accepts `WaitOption` values:

```go
val, err := page.WaitForEvent(bonk.DialogEvent,
    bonk.WaitTimeout(5*time.Second),
)
```

## WaitForEvent vs On Handlers

| | `WaitForEvent` | `On` / `OnConsole` / `OnDialog` |
|---|---|---|
| **Blocks** | Yes — returns after the first event | No — handler runs asynchronously |
| **One-shot** | Yes — unsubscribes after first event | No — fires on every event |
| **Returns payload** | Yes | No (payload is passed to callback) |

Use `WaitForEvent` when you need to synchronously wait for a single event. Use `On` handlers when you need to react to every occurrence.

## Example: Wait for Dialog

```go
go func() {
    page.Evaluate(`setTimeout(() => alert("hello"), 100)`)
}()

val, err := page.WaitForEvent(bonk.DialogEvent)
dialog := val.(*bonk.Dialog)
dialog.Accept()
```
