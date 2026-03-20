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

	ctx, err := b.NewContext(bonk.WithViewport(1280, 720))
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Navigate with WaitDOMContentLoaded ===")
	err = page.Navigate(
		"https://example.com",
		bonk.WithWaitUntil(bonk.WaitDOMContentLoaded),
	)
	if err != nil {
		log.Fatal("DOMContentLoaded nav:", err)
	}
	fmt.Println("OK")

	fmt.Println("\n=== Title + URL ===")
	title, _ := page.Title()
	url, _ := page.URL()
	fmt.Println("Title:", title)
	fmt.Println("URL:", url)

	fmt.Println("\n=== Query + Element methods ===")
	h1, err := page.Query("h1")
	if err != nil {
		log.Fatal("Query h1:", err)
	}
	text, _ := h1.Text()
	html, _ := h1.HTML()
	visible, _ := h1.IsVisible()
	box, _ := h1.BoundingBox()
	fmt.Println("Text:", text)
	fmt.Println("HTML:", html)
	fmt.Println("Visible:", visible)
	fmt.Printf(
		"Box: %.0fx%.0f at (%.0f,%.0f)\n",
		box.Width,
		box.Height,
		box.X,
		box.Y,
	)

	fmt.Println("\n=== QueryAll ===")
	paragraphs, err := page.QueryAll("p")
	if err != nil {
		log.Fatal("QueryAll p:", err)
	}
	fmt.Printf("Found %d paragraphs\n", len(paragraphs))
	for i, p := range paragraphs {
		t, _ := p.Text()
		fmt.Printf("  [%d] %s\n", i, t[:min(60, len(t))])
	}

	fmt.Println("\n=== Attribute ===")
	link, _ := page.Query("a")
	href, _ := link.Attribute("href")
	fmt.Println("Link href:", href)

	fmt.Println("\n=== Evaluate ===")
	result, err := page.Evaluate("1 + 2 + 3")
	if err != nil {
		log.Fatal("Evaluate:", err)
	}
	fmt.Println("1+2+3 =", result)

	fmt.Println("\n=== EvaluateHandle ===")
	el, err := page.EvaluateHandle("document.querySelector('h1')")
	if err != nil {
		log.Fatal("EvaluateHandle:", err)
	}
	handleText, _ := el.Text()
	fmt.Println("Handle text:", handleText)

	fmt.Println("\n=== EvaluateOn ===")
	tagResult, err := page.EvaluateOn(h1, "function(){return this.tagName}")
	if err != nil {
		log.Fatal("EvaluateOn:", err)
	}
	fmt.Println("Tag name:", tagResult)

	fmt.Println("\n=== Element.Screenshot ===")
	err = h1.Screenshot("/tmp/bonk_h1.png")
	if err != nil {
		log.Fatal("Element screenshot:", err)
	}
	fmt.Println("H1 screenshot saved to /tmp/bonk_h1.png")

	fmt.Println("\n=== Page.Screenshot ===")
	err = page.Screenshot("/tmp/bonk_page.png")
	if err != nil {
		log.Fatal("Page screenshot:", err)
	}
	fmt.Println("Page screenshot saved")

	fmt.Println("\n=== WaitSelector with timeout ===")
	_, err = page.WaitSelector("#nonexistent", bonk.WaitTimeout(1*time.Second))
	if err != nil {
		fmt.Println("Expected timeout:", err)
	} else {
		log.Fatal("Should have timed out")
	}

	fmt.Println("\n=== WaitSelector success ===")
	found, err := page.WaitSelector("h1", bonk.WaitTimeout(5*time.Second))
	if err != nil {
		log.Fatal("WaitSelector h1:", err)
	}
	foundText, _ := found.Text()
	fmt.Println("Found:", foundText)

	fmt.Println("\n=== Click shorthand ===")
	err = page.Click("a")
	if err != nil {
		log.Fatal("Click a:", err)
	}
	time.Sleep(500 * time.Millisecond)
	newURL, _ := page.URL()
	fmt.Println("After click URL:", newURL)

	fmt.Println("\n=== GoBack ===")
	err = page.GoBack()
	if err != nil {
		log.Fatal("GoBack:", err)
	}
	time.Sleep(500 * time.Millisecond)
	backURL, _ := page.URL()
	fmt.Println("After GoBack URL:", backURL)

	fmt.Println("\n=== Cookies ===")
	cookies, err := ctx.Cookies()
	if err != nil {
		log.Fatal("Cookies:", err)
	}
	fmt.Printf("Cookies: %d\n", len(cookies))

	fmt.Println("\n=== OnConsole ===")
	consoleDone := make(chan string, 1)
	unsub := page.OnConsole(func(msg *bonk.ConsoleMessage) {
		consoleDone <- msg.Text
	})
	page.Evaluate("console.log('hello from bonk')")
	select {
	case msg := <-consoleDone:
		fmt.Println("Console:", msg)
	case <-time.After(2 * time.Second):
		fmt.Println("Console timeout (might be ok)")
	}
	unsub()

	fmt.Println("\n=== Network interception ===")
	blocked := 0
	unroute := page.Route("**/*.css", func(r *bonk.Route) {
		blocked++
		r.Abort()
	})
	page.Navigate("https://example.com")
	fmt.Printf("Blocked %d CSS requests\n", blocked)
	unroute()

	fmt.Println("\n=== Frames ===")
	frames, err := page.Frames()
	if err != nil {
		log.Fatal("Frames:", err)
	}
	fmt.Printf("Frame count: %d\n", len(frames))
	for _, f := range frames {
		fmt.Printf("  Frame: id=%s name=%q url=%s\n", f.ID(), f.Name(), f.URL())
	}

	fmt.Println("\n=== All tests passed ===")
}
