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

	el, err := page.Query("h1")
	if err != nil {
		log.Fatal(err)
	}
	if el != nil {
		text, _ := el.Text()
		fmt.Println("H1:", text)

		visible, _ := el.IsVisible()
		fmt.Println("Visible:", visible)

		box, _ := el.BoundingBox()
		fmt.Printf("BoundingBox: x=%.0f y=%.0f w=%.0f h=%.0f\n",
			box.X, box.Y, box.Width, box.Height)
	}

	links, err := page.QueryAll("a")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d links\n", len(links))

	if err := page.Screenshot("example.png"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Screenshot saved")
}
