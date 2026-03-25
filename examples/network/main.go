// Package main demonstrates network interception with bonk.
package main

import (
	"fmt"
	"log"
	"strings"

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
		log.Println(err)
		return
	}
	defer ctx.Close()

	page, err := ctx.NewPage()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("=== OnRequest: observe + modify ===")
	unsubReq := page.OnRequest(func(r *bonk.Request) {
		if strings.HasSuffix(r.URL, ".css") {
			fmt.Println("Blocking CSS:", r.URL)
			r.Abort()
			return
		}
		r.SetHeader("X-Bonk", "true")
		r.Continue()
	})

	fmt.Println("=== OnResponse: observe ===")
	unsubResp := page.OnResponse(func(r *bonk.Response) {
		fmt.Printf("[%d] %s (headers=%d)\n", r.Status, r.URL, len(r.Headers))
		r.Continue()
	})

	if err := page.Navigate("https://example.com"); err != nil {
		log.Println(err)
		return
	}

	title, _ := page.Title()
	fmt.Println("Title:", title)

	unsubReq()
	unsubResp()

	fmt.Println("\n=== Route: mock response ===")
	unsubRoute := page.Route("**/api/test", func(r *bonk.Route) {
		fmt.Println("Mocking:", r.Request.URL)
		r.Fulfill(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"mocked": true}`)
	})

	result, err := page.Evaluate(`
		fetch('/api/test').then(r => r.json())
	`)
	if err != nil {
		fmt.Println("Fetch error (expected on example.com):", err)
	} else {
		fmt.Println("Mock result:", result)
	}

	unsubRoute()
	fmt.Println("\nDone")
}
