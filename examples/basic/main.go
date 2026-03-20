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

	ctx, err := b.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	if err := page.Navigate("https://example.com"); err != nil {
		log.Fatal(err)
	}

	title, err := page.Title()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Title:", title)

	if err := page.Screenshot("example.png"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Screenshot saved to example.png")
}
