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

	page.OnResponse(func(r *bonk.Response) {
		fmt.Printf("[%d] %s\n", r.Status, r.URL)
	})

	if err := page.Navigate("https://example.com"); err != nil {
		log.Fatal(err)
	}

	title, _ := page.Title()
	fmt.Println("Done:", title)
}
