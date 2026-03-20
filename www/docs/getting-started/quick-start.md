# Quick Start

This guide walks through launching a browser, navigating to a page, interacting with elements, and taking a screenshot.

## Launch and Navigate

```go
b, err := bonk.Launch()
if err != nil {
    log.Fatal(err)
}
defer b.Close()

ctx, err := b.NewContext()
if err != nil {
    log.Fatal(err)
}

page, err := ctx.NewPage()
if err != nil {
    log.Fatal(err)
}

err = page.Navigate("https://example.com")
```

## Query Elements

```go
el, err := page.Query("h1")
if err != nil {
    log.Fatal(err)
}

text, _ := el.Text()
fmt.Println(text) // "Example Domain"
```

## Wait for Elements

```go
el, err := page.WaitSelector(".dynamic-content", bonk.WaitTimeout(10*time.Second))
if err != nil {
    log.Fatal(err)
}
```

## Interact

```go
page.Fill("#email", "user@example.com")
page.Fill("#password", "secret")
page.Click("#login")
```

## Screenshot

```go
page.Screenshot("result.png")
page.Screenshot("full.png", bonk.FullPage())
```

## Run JavaScript

```go
title, err := page.Evaluate("document.title")
fmt.Println(title)
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/joakimcarlsson/bonk"
)

func main() {
    b, err := bonk.Launch()
    if err != nil {
        log.Fatal(err)
    }
    defer b.Close()

    ctx, err := b.NewContext(bonk.WithViewport(1920, 1080))
    if err != nil {
        log.Fatal(err)
    }

    page, err := ctx.NewPage()
    if err != nil {
        log.Fatal(err)
    }

    if err := page.Navigate("https://news.ycombinator.com"); err != nil {
        log.Fatal(err)
    }

    items, err := page.QueryAll(".titleline > a")
    if err != nil {
        log.Fatal(err)
    }

    for i, item := range items {
        if i >= 5 {
            break
        }
        text, _ := item.Text()
        href, _ := item.Attribute("href")
        fmt.Printf("%d. %s (%s)\n", i+1, text, href)
    }

    page.Screenshot("hackernews.png")
}
```
