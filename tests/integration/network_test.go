package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestOnRequestObserve(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	var urls []string
	unsub := page.OnRequest(func(r *bonk.Request) {
		urls = append(urls, r.URL)
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://example.com")

	if len(urls) == 0 {
		t.Fatal("no requests captured")
	}

	found := false
	for _, u := range urls {
		if strings.Contains(u, "example.com") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected example.com in requests, got %v", urls)
	}
}

func TestOnRequestModifyHeaders(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	unsub := page.OnRequest(func(r *bonk.Request) {
		r.SetHeader("X-Custom-Header", "bonk-test")
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://httpbin.org/headers")
	time.Sleep(500 * time.Millisecond)

	result, err := page.Evaluate("document.body.innerText")
	if err != nil {
		t.Fatal(err)
	}

	s, _ := result.(string)
	if !strings.Contains(s, "X-Custom-Header") {
		t.Error("injected header not found in response")
	}
}

func TestOnRequestAbort(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	blocked := 0
	unsub := page.OnRequest(func(r *bonk.Request) {
		if strings.HasSuffix(r.URL, ".css") {
			blocked++
			r.Abort()
			return
		}
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://example.com")
	t.Logf("blocked %d CSS requests", blocked)
}

func TestOnResponseBody(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	bodyCh := make(chan string, 1)
	unsub := page.OnResponse(func(r *bonk.Response) {
		if strings.Contains(r.URL, "example.com") && r.Status == 200 {
			body, err := r.Body()
			if err == nil && len(body) > 0 {
				select {
				case bodyCh <- string(body[:min(100, len(body))]):
				default:
				}
			}
		}
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://example.com")

	select {
	case body := <-bodyCh:
		if !strings.Contains(body, "Example Domain") {
			t.Errorf("unexpected body: %s", body)
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for response body")
	}
}

func TestRouteFulfill(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	unsub := page.Route("**/api/mock", func(r *bonk.Route) {
		r.Fulfill(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"status":"mocked","value":42}`)
	})
	defer unsub()

	result, err := page.Evaluate(`fetch('/api/mock').then(r => r.json())`)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T: %v", result, result)
	}
	if m["status"] != "mocked" {
		t.Errorf("status = %v, want mocked", m["status"])
	}
	if m["value"] != float64(42) {
		t.Errorf("value = %v, want 42", m["value"])
	}
}

func TestSetExtraHTTPHeaders(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetExtraHTTPHeaders(map[string]string{
		"X-Global": "everywhere",
	})

	sawHeader := false
	unsub := page.OnRequest(func(r *bonk.Request) {
		if r.Headers["X-Global"] == "everywhere" {
			sawHeader = true
		}
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://example.com")

	if !sawHeader {
		t.Error("global header not seen in requests")
	}
}

func TestOnResponseHeaders(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	headersCh := make(chan map[string]string, 1)
	unsub := page.OnResponse(func(r *bonk.Response) {
		if strings.Contains(r.URL, "example.com") && len(r.Headers) > 0 {
			select {
			case headersCh <- r.Headers:
			default:
			}
		}
		r.Continue()
	})
	defer unsub()

	page.Navigate("https://example.com")

	select {
	case headers := <-headersCh:
		if len(headers) == 0 {
			t.Error("response headers empty")
		}
		t.Logf("captured %d response headers", len(headers))
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for response headers")
	}
}

func TestOnRequestAutoContine(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	unsub := page.OnRequest(func(r *bonk.Request) {
		fmt.Println("observed:", r.URL)
	})
	defer unsub()

	err := page.Navigate("https://example.com")
	if err != nil {
		t.Fatal("navigation should succeed with auto-continue:", err)
	}
}
