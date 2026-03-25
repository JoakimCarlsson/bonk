// Package main demonstrates page interaction with bonk.
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
		log.Println(err)
		return
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		log.Println(err)
		return
	}

	if err := page.Navigate("https://google.com"); err != nil {
		log.Println(err)
		return
	}

	el, err := page.WaitSelector("textarea[name=q]")
	if err != nil {
		log.Println(err)
		return
	}

	if err := el.Type("bonk go cdp", bonk.WithDelay(50*time.Millisecond)); err != nil {
		log.Println(err)
		return
	}

	if err := el.Press("Enter"); err != nil {
		log.Println(err)
		return
	}

	page.WaitSelector("#search")

	url, _ := page.URL()
	fmt.Println("Searched, now at:", url)
}
