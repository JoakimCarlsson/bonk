package integration

import (
	"context"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestWithContextTimeout(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	p := page.Timeout(1 * time.Millisecond)
	err := p.Navigate("https://example.com")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWithContextCancel(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := page.WithContext(ctx)
	err := p.Navigate("https://example.com")
	if err == nil {
		t.Fatal("expected context canceled error, got nil")
	}
}

func TestIsClosed(t *testing.T) {
	b := launchBrowser(t)
	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatal(err)
	}

	if page.IsClosed() {
		t.Error("page should not be closed")
	}

	page.Close()

	if !page.IsClosed() {
		t.Error("page should be closed")
	}
}

func TestSetOffline(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	if err := page.SetOffline(true); err != nil {
		t.Fatal(err)
	}

	val, err := page.Evaluate("navigator.onLine")
	if err != nil {
		t.Fatal(err)
	}
	if val != false {
		t.Errorf("navigator.onLine = %v while offline, want false", val)
	}

	if err := page.SetOffline(false); err != nil {
		t.Fatal(err)
	}

	val, err = page.Evaluate("navigator.onLine")
	if err != nil {
		t.Fatal(err)
	}
	if val != true {
		t.Errorf("navigator.onLine = %v after going online, want true", val)
	}
}

func TestAddInitScript(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.AddInitScript("window.__bonk_test = 42")
	page.Navigate("https://example.com")

	val, err := page.Evaluate("window.__bonk_test")
	if err != nil {
		t.Fatal(err)
	}
	if val != float64(42) {
		t.Errorf("__bonk_test = %v, want 42", val)
	}
}

func TestContent(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	html, err := page.Content()
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("Content() returned empty string")
	}
}

func TestSetContent(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")

	err := page.SetContent("<html><body><h1>Test</h1></body></html>")
	if err != nil {
		t.Fatal(err)
	}

	val, err := page.Evaluate("document.querySelector('h1').textContent")
	if err != nil {
		t.Fatal(err)
	}
	if val != "Test" {
		t.Errorf("h1 text = %v, want Test", val)
	}
}

func TestWaitForFunction(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")

	go func() {
		time.Sleep(200 * time.Millisecond)
		page.Evaluate("window.__ready = true")
	}()

	err := page.WaitForFunction(
		"window.__ready === true",
		bonk.WaitTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitForURL(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	err := page.WaitForURL("*example.com*", bonk.WaitTimeout(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
}
