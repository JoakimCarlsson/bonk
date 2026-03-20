package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestLocatorText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	loc := page.Locator("h1")
	text, err := loc.Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Example Domain" {
		t.Errorf("text = %q, want %q", text, "Example Domain")
	}
}

func TestLocatorCount(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	count, err := page.Locator("p").Count()
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Error("expected at least one <p> element")
	}
}

func TestLocatorNth(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	first := page.Locator("p").First()
	text, err := first.Text()
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Error("first <p> text is empty")
	}
}

func TestLocatorNeverStale(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><div id="target">original</div></body></html>`)

	loc := page.Locator("#target")

	text1, _ := loc.Text()
	if text1 != "original" {
		t.Fatalf("initial text = %q, want original", text1)
	}

	page.Evaluate(`document.getElementById('target').textContent = 'updated'`)

	text2, _ := loc.Text()
	if text2 != "updated" {
		t.Errorf("after update text = %q, want updated", text2)
	}
}

func TestLocatorWaitFor(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")

	go func() {
		time.Sleep(200 * time.Millisecond)
		page.Evaluate(`
			var el = document.createElement('div');
			el.id = 'delayed';
			el.textContent = 'appeared';
			document.body.appendChild(el);
		`)
	}()

	loc := page.Locator("#delayed")
	err := loc.WaitFor(bonk.WaitTimeout(5 * time.Second))
	if err != nil {
		t.Fatal(err)
	}

	text, _ := loc.Text()
	if text != "appeared" {
		t.Errorf("text = %q, want appeared", text)
	}
}
