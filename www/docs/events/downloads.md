# Download Events

Handle file downloads triggered by the page.

## Listen for Downloads

```go
page.On(bonk.DownloadEvent, func(d *bonk.Download) {
    fmt.Printf("Downloading: %s\n", d.SuggestedFilename())
    err := d.SaveAs("./downloads/" + d.SuggestedFilename())
    if err != nil {
        log.Println("download failed:", err)
    }
})
```

## Download Methods

| Method | Description |
|--------|-------------|
| `URL() string` | The download URL |
| `SuggestedFilename() string` | Browser-suggested filename |
| `SaveAs(path string) error` | Save to disk (blocks until download completes) |
| `SaveAsContext(ctx, path) error` | Save with context for cancellation/timeout |

## Save with Timeout

`SaveAs` blocks until the download finishes. Use `SaveAsContext` to set a deadline:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := d.SaveAsContext(ctx, "./file.pdf")
if err != nil {
    log.Println("download timed out or failed:", err)
}
```

## Directory Creation

`SaveAs` automatically creates parent directories if they don't exist.
