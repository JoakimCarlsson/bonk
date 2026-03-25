package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestNavigate(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	if err := page.Navigate("https://example.com"); err != nil {
		t.Fatal(err)
	}

	title, err := page.Title()
	if err != nil {
		t.Fatal(err)
	}
	if title != "Example Domain" {
		t.Errorf("title = %q, want %q", title, "Example Domain")
	}

	url, err := page.URL()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(url, "example.com") {
		t.Errorf("url = %q, want containing example.com", url)
	}
}

func TestNavigateDOMContentLoaded(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	err := page.Navigate(
		"https://example.com",
		bonk.WithWaitUntil(bonk.WaitDOMContentLoaded),
	)
	if err != nil {
		t.Fatal(err)
	}

	title, _ := page.Title()
	if title != "Example Domain" {
		t.Errorf("title = %q, want %q", title, "Example Domain")
	}
}

func TestReload(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	if err := page.Reload(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	title, _ := page.Title()
	if title != "Example Domain" {
		t.Errorf("title = %q after reload", title)
	}
}

func TestGoBackForward(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	if err := page.Navigate("https://example.com"); err != nil {
		t.Fatal(err)
	}
	if err := page.Click("a"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	url1, _ := page.URL()
	if !strings.Contains(url1, "iana.org") {
		t.Fatalf("after click url = %q, want iana.org", url1)
	}

	if err := page.GoBack(); err != nil {
		t.Fatal(err)
	}

	url2, _ := page.URL()
	if !strings.Contains(url2, "example.com") {
		t.Errorf("after GoBack url = %q, want example.com", url2)
	}

	if err := page.GoForward(); err != nil {
		t.Fatal(err)
	}

	url3, _ := page.URL()
	if !strings.Contains(url3, "iana.org") {
		t.Errorf("after GoForward url = %q, want iana.org", url3)
	}
}

func TestEvaluate(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	result, err := page.Evaluate("1 + 2 + 3")
	if err != nil {
		t.Fatal(err)
	}
	if result != float64(6) {
		t.Errorf("1+2+3 = %v, want 6", result)
	}
}

func TestEvaluateHandle(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	el, err := page.EvaluateHandle("document.querySelector('h1')")
	if err != nil {
		t.Fatal(err)
	}
	if el == nil {
		t.Fatal("EvaluateHandle returned nil")
	}

	text, _ := el.Text()
	if text != "Example Domain" {
		t.Errorf("text = %q, want %q", text, "Example Domain")
	}
}

func TestEvaluateOn(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	el, _ := page.Query("h1")
	result, err := page.EvaluateOn(el, "function(){return this.tagName}")
	if err != nil {
		t.Fatal(err)
	}
	if result != "H1" {
		t.Errorf("tagName = %v, want H1", result)
	}
}
