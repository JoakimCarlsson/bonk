# Installation

## Requirements

- Go 1.24 or later
- Chrome or Chromium installed (bonk auto-detects the binary)

## Install

```bash
go get github.com/joakimcarlsson/bonk
```

## Chrome Discovery

bonk looks for Chrome in this order:

1. `CHROME_PATH` environment variable
2. Common installation paths for your OS
3. `PATH` lookup

You can also specify the binary explicitly:

```go
b, err := bonk.Launch(bonk.ChromePath("/usr/bin/chromium"))
```

## Verify

```go
package main

import (
	"fmt"
	"log"

	"github.com/joakimcarlsson/bonk"
)

func main() {
	b, err := bonk.Launch()
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	fmt.Println("bonk is working")
}
```
