# Dispatch Event

`DispatchEvent` fires a DOM event on an element programmatically. This is useful when simulated clicks via CDP input events don't trigger framework-level handlers (e.g. React synthetic events or custom Web Components).

## Element.DispatchEvent

```go
el, _ := page.Query("#my-button")
el.DispatchEvent("click")
```

### With Event Options

Pass an event init map to configure bubbling, cancelability, and other properties:

```go
el.DispatchEvent("input", map[string]any{
    "bubbles":    true,
    "cancelable": true,
})
```

If no event init is provided, `{bubbles: true}` is used as the default.

## Page.DispatchEvent

The page-level convenience method waits for the selector before dispatching:

```go
page.DispatchEvent("#my-button", "click")

page.DispatchEvent("#input", "change", map[string]any{
    "bubbles": true,
})
```

## When to Use

- Framework event listeners (React, Vue, Angular) that don't respond to CDP-level input simulation
- Custom `change` or `input` events after programmatically setting a value
- `submit` events on forms
- Custom events for Web Components

## Comparison

| Method | How it works |
|--------|-------------|
| `el.Click()` | CDP `Input.dispatchMouseEvent` — simulates real mouse at element center |
| `el.DispatchEvent("click")` | DOM `dispatchEvent(new Event("click"))` — fires the JS event directly |

Use `Click()` for normal interactions. Use `DispatchEvent` when the framework only listens to DOM-level events.
