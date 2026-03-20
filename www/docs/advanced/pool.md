# Page Pool

The `Pool` manages a set of reusable pages for concurrent automation. It pre-creates pages and checks them out to workers, avoiding the overhead of creating and destroying pages per task.

## Create a Pool

```go
pool, err := bonk.NewPool(ctx, 5) // 5 concurrent pages
if err != nil {
    log.Fatal(err)
}
defer pool.Close()
```

## Use Pages

`Do` checks out a page, runs your function, and returns the page to the pool. It blocks if all pages are in use:

```go
err := pool.Do(func(page *bonk.Page) error {
    if err := page.Navigate("https://example.com"); err != nil {
        return err
    }
    title, _ := page.Title()
    fmt.Println(title)
    return nil
})
```

## Concurrent Scraping

```go
pool, _ := bonk.NewPool(ctx, 10)
defer pool.Close()

var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        pool.Do(func(page *bonk.Page) error {
            page.Navigate(u)
            // ... process page ...
            return nil
        })
    }(url)
}
wg.Wait()
```

## How It Works

- `NewPool` pre-creates `size` pages in the given browser context
- Pages are stored in a buffered channel
- `Do` receives from the channel (blocks if empty), runs the function, then sends the page back
- `Close` closes all pages in the pool
