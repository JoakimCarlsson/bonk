package integration

import (
	"os"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestQuery(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	el, err := page.Query("h1")
	if err != nil {
		t.Fatal(err)
	}
	if el == nil {
		t.Fatal("Query returned nil")
	}

	text, _ := el.Text()
	if text != "Example Domain" {
		t.Errorf("text = %q, want %q", text, "Example Domain")
	}
}

func TestQueryNotFound(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	el, err := page.Query("#nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if el != nil {
		t.Error("expected nil for nonexistent selector")
	}
}

func TestQueryAll(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	els, err := page.QueryAll("p")
	if err != nil {
		t.Fatal(err)
	}
	if len(els) < 1 {
		t.Fatal("expected at least 1 paragraph")
	}
}

func TestWaitSelector(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	el, err := page.WaitSelector("h1", bonk.WaitTimeout(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	text, _ := el.Text()
	if text != "Example Domain" {
		t.Errorf("text = %q", text)
	}
}

func TestWaitSelectorTimeout(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	_, err := page.WaitSelector("#nonexistent", bonk.WaitTimeout(1*time.Second))
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var te *bonk.TimeoutError
	if !isTimeoutError(err, &te) {
		t.Errorf("expected TimeoutError, got %T: %v", err, err)
	}
}

func TestElementText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("h1")

	text, err := el.Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Example Domain" {
		t.Errorf("text = %q", text)
	}
}

func TestElementHTML(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("h1")

	html, err := el.HTML()
	if err != nil {
		t.Fatal(err)
	}
	if html != "<h1>Example Domain</h1>" {
		t.Errorf("html = %q", html)
	}
}

func TestElementAttribute(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("a")

	href, err := el.Attribute("href")
	if err != nil {
		t.Fatal(err)
	}
	if href != "https://iana.org/domains/example" {
		t.Errorf("href = %q", href)
	}
}

func TestElementIsVisible(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("h1")

	visible, err := el.IsVisible()
	if err != nil {
		t.Fatal(err)
	}
	if !visible {
		t.Error("h1 should be visible")
	}
}

func TestElementBoundingBox(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("h1")

	box, err := el.BoundingBox()
	if err != nil {
		t.Fatal(err)
	}
	if box == nil {
		t.Fatal("box is nil")
	}
	if box.Width <= 0 || box.Height <= 0 {
		t.Errorf("box = %+v, want positive dimensions", box)
	}
}

func TestElementClick(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	err := page.Click("a")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	url, _ := page.URL()
	if url == "https://example.com/" {
		t.Error("URL didn't change after click")
	}
}

func TestElementScreenshot(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	el, _ := page.Query("h1")

	path := t.TempDir() + "/h1.png"
	if err := el.Screenshot(path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("screenshot file is empty")
	}
}

func TestPageScreenshot(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	path := t.TempDir() + "/page.png"
	if err := page.Screenshot(path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("screenshot file is empty")
	}
}

func TestElementIsChecked(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="on" type="checkbox" checked>
		<input id="off" type="checkbox">
	</body></html>`)

	on, _ := page.Query("#on")
	checked, err := on.IsChecked()
	if err != nil {
		t.Fatal(err)
	}
	if !checked {
		t.Error("expected checked=true for #on")
	}

	off, _ := page.Query("#off")
	checked, err = off.IsChecked()
	if err != nil {
		t.Fatal(err)
	}
	if checked {
		t.Error("expected checked=false for #off")
	}
}

func TestElementIsDisabled(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="dis" type="text" disabled>
		<input id="en" type="text">
	</body></html>`)

	dis, _ := page.Query("#dis")
	disabled, err := dis.IsDisabled()
	if err != nil {
		t.Fatal(err)
	}
	if !disabled {
		t.Error("expected disabled=true for #dis")
	}

	en, _ := page.Query("#en")
	disabled, err = en.IsDisabled()
	if err != nil {
		t.Fatal(err)
	}
	if disabled {
		t.Error("expected disabled=false for #en")
	}
}

func TestElementIsEditable(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body>
		<input id="editable" type="text">
		<input id="disabled-input" type="text" disabled>
		<input id="readonly-input" type="text" readonly>
		<div id="plain">text</div>
		<div id="ce" contenteditable="true">edit me</div>
	</body></html>`)

	cases := []struct {
		sel  string
		want bool
	}{
		{"#editable", true},
		{"#disabled-input", false},
		{"#readonly-input", false},
		{"#plain", false},
		{"#ce", true},
	}
	for _, tc := range cases {
		el, _ := page.Query(tc.sel)
		got, err := el.IsEditable()
		if err != nil {
			t.Fatalf("%s: %v", tc.sel, err)
		}
		if got != tc.want {
			t.Errorf("%s: editable=%t, want %t", tc.sel, got, tc.want)
		}
	}
}

func TestElementBlur(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	el, _ := page.Query("#input")
	el.Focus()

	tag, _ := page.Evaluate("document.activeElement.tagName")
	if tag != "INPUT" {
		t.Fatalf("activeElement = %v, want INPUT", tag)
	}

	if err := el.Blur(); err != nil {
		t.Fatal(err)
	}

	tag, _ = page.Evaluate("document.activeElement.tagName")
	if tag != "BODY" {
		t.Errorf("after blur activeElement = %v, want BODY", tag)
	}
}

func TestElementClear(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	el, _ := page.Query("#input")
	el.Fill("hello world")

	val, _ := page.Evaluate(`document.querySelector("#input").value`)
	if val != "hello world" {
		t.Fatalf("value = %v, want 'hello world'", val)
	}

	if err := el.Clear(); err != nil {
		t.Fatal(err)
	}

	val, _ = page.Evaluate(`document.querySelector("#input").value`)
	if val != "" {
		t.Errorf("after clear value = %v, want empty", val)
	}
}

func isTimeoutError(err error, target **bonk.TimeoutError) bool {
	te, ok := err.(*bonk.TimeoutError)
	if ok {
		*target = te
	}
	return ok
}
