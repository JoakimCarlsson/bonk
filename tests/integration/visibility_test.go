package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestWaitSelectorVisible(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="target" style="display:none">hidden</div>
	</body></html>`)

	go func() {
		time.Sleep(300 * time.Millisecond)
		page.Evaluate(
			`document.getElementById('target').style.display = 'block'`,
		)
	}()

	el, err := page.WaitSelector("#target",
		bonk.WaitVisibleOption(),
		bonk.WaitTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	text, _ := el.Text()
	if text != "hidden" {
		t.Errorf("text = %q, want hidden", text)
	}
}

func TestWaitSelectorHidden(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="spinner">Loading...</div>
	</body></html>`)

	go func() {
		time.Sleep(300 * time.Millisecond)
		page.Evaluate(
			`document.getElementById('spinner').style.display = 'none'`,
		)
	}()

	_, err := page.WaitSelector("#spinner",
		bonk.WaitHiddenOption(),
		bonk.WaitTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestElementWaitForVisible(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="target" style="opacity:0">invisible</div>
	</body></html>`)

	el, err := page.Query("#target")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(300 * time.Millisecond)
		page.Evaluate(`document.getElementById('target').style.opacity = '1'`)
	}()

	if err := el.WaitForVisible(bonk.WaitTimeout(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
}

func TestElementWaitForHidden(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="target">visible</div>
	</body></html>`)

	el, err := page.Query("#target")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(300 * time.Millisecond)
		page.Evaluate(
			`document.getElementById('target').style.display = 'none'`,
		)
	}()

	if err := el.WaitForHidden(bonk.WaitTimeout(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
}

func TestStaleElementRetry(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><div id="target">original</div></body></html>`)

	el, err := page.Query("#target")
	if err != nil {
		t.Fatal(err)
	}

	page.Evaluate(`
		document.body.innerHTML = '<div id="target">replaced</div>';
	`)

	text, err := el.Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "replaced" {
		t.Errorf("text = %q, want replaced", text)
	}
}
