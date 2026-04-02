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

func TestLocatorIsChecked(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="on" type="checkbox" checked>
		<input id="off" type="checkbox">
	</body></html>`)

	checked, err := page.Locator("#on").IsChecked()
	if err != nil {
		t.Fatal(err)
	}
	if !checked {
		t.Error("expected checked=true for #on")
	}

	checked, err = page.Locator("#off").IsChecked()
	if err != nil {
		t.Fatal(err)
	}
	if checked {
		t.Error("expected checked=false for #off")
	}
}

func TestLocatorIsDisabled(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="dis" type="text" disabled>
		<input id="en" type="text">
	</body></html>`)

	disabled, err := page.Locator("#dis").IsDisabled()
	if err != nil {
		t.Fatal(err)
	}
	if !disabled {
		t.Error("expected disabled=true")
	}

	disabled, err = page.Locator("#en").IsDisabled()
	if err != nil {
		t.Fatal(err)
	}
	if disabled {
		t.Error("expected disabled=false")
	}
}

func TestLocatorIsEditable(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="editable" type="text">
		<input id="disabled-input" type="text" disabled>
		<div id="plain">text</div>
	</body></html>`)

	editable, err := page.Locator("#editable").IsEditable()
	if err != nil {
		t.Fatal(err)
	}
	if !editable {
		t.Error("expected editable=true for #editable")
	}

	editable, err = page.Locator("#disabled-input").IsEditable()
	if err != nil {
		t.Fatal(err)
	}
	if editable {
		t.Error("expected editable=false for #disabled-input")
	}

	editable, err = page.Locator("#plain").IsEditable()
	if err != nil {
		t.Fatal(err)
	}
	if editable {
		t.Error("expected editable=false for #plain")
	}
}

func TestLocatorFocus(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	if err := page.Locator("#input").Focus(); err != nil {
		t.Fatal(err)
	}

	tag, _ := page.Evaluate("document.activeElement.tagName")
	if tag != "INPUT" {
		t.Errorf("activeElement = %v, want INPUT", tag)
	}
}

func TestLocatorBlur(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	loc := page.Locator("#input")
	loc.Focus()

	if err := loc.Blur(); err != nil {
		t.Fatal(err)
	}

	tag, _ := page.Evaluate("document.activeElement.tagName")
	if tag != "BODY" {
		t.Errorf("after blur activeElement = %v, want BODY", tag)
	}
}

func TestLocatorClear(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	loc := page.Locator("#input")
	loc.Fill("hello world")

	if err := loc.Clear(); err != nil {
		t.Fatal(err)
	}

	val, _ := page.Evaluate(`document.querySelector("#input").value`)
	if val != "" {
		t.Errorf("after clear value = %v, want empty", val)
	}
}

func TestLocatorAllInnerTexts(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<ul>
			<li>Alpha</li>
			<li>Bravo</li>
			<li>Charlie</li>
		</ul>
	</body></html>`)

	texts, err := page.Locator("li").AllInnerTexts()
	if err != nil {
		t.Fatal(err)
	}
	if len(texts) != 3 {
		t.Fatalf("got %d texts, want 3", len(texts))
	}
	want := []string{"Alpha", "Bravo", "Charlie"}
	for i, w := range want {
		if texts[i] != w {
			t.Errorf("texts[%d] = %q, want %q", i, texts[i], w)
		}
	}
}

func TestLocatorAllTextContents(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<ul>
			<li>Alpha</li>
			<li>Bravo</li>
			<li>Charlie</li>
		</ul>
	</body></html>`)

	texts, err := page.Locator("li").AllTextContents()
	if err != nil {
		t.Fatal(err)
	}
	if len(texts) != 3 {
		t.Fatalf("got %d texts, want 3", len(texts))
	}
	want := []string{"Alpha", "Bravo", "Charlie"}
	for i, w := range want {
		if texts[i] != w {
			t.Errorf("texts[%d] = %q, want %q", i, texts[i], w)
		}
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
