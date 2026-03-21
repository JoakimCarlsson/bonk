package integration

import (
	"strings"
	"testing"

	"github.com/joakimcarlsson/bonk"
)

func TestFilterHasText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="card"><p>Apple</p></div>
		<div class="card"><p>Banana</p></div>
		<div class="card"><p>Cherry</p></div>
	</body></html>`)

	loc := page.Locator(".card").Filter(bonk.LocatorFilter{HasText: "Banana"})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	text, err := loc.First().Text()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Banana") {
		t.Errorf("text = %q, want to contain Banana", text)
	}
}

func TestFilterHasNotText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><table>
		<tr><td>Active</td><td>Alice</td></tr>
		<tr><td>Inactive</td><td>Bob</td></tr>
		<tr><td>Active</td><td>Carol</td></tr>
	</table></body></html>`)

	loc := page.Locator("tr").Filter(bonk.LocatorFilter{HasNotText: "Inactive"})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestFilterHas(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="card"><h2>Title</h2></div>
		<div class="card"><p>No heading here</p></div>
		<div class="card"><h2>Another</h2><p>With text</p></div>
	</body></html>`)

	loc := page.Locator(".card").Filter(bonk.LocatorFilter{
		Has: page.Locator("h2"),
	})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestFilterHasJSLocator(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="row"><span>Total: $50</span></div>
		<div class="row"><span>Subtotal: $30</span></div>
		<div class="row"><span>Total: $100</span></div>
	</body></html>`)

	loc := page.Locator(".row").Filter(bonk.LocatorFilter{
		Has: page.GetByText("Total"),
	})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestFilterHasNot(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="item"><span class="badge">New</span> Item A</div>
		<div class="item">Item B</div>
		<div class="item"><span class="badge">New</span> Item C</div>
	</body></html>`)

	loc := page.Locator(".item").Filter(bonk.LocatorFilter{
		HasNot: page.Locator(".badge"),
	})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	text, err := loc.First().Text()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Item B") {
		t.Errorf("text = %q, want to contain Item B", text)
	}
}

func TestFilterMultipleCriteria(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="card"><h2>Premium</h2><p>Active</p></div>
		<div class="card"><h2>Basic</h2><p>Active</p></div>
		<div class="card"><p>Active but no heading</p></div>
		<div class="card"><h2>Premium</h2><p>Expired</p></div>
	</body></html>`)

	loc := page.Locator(".card").Filter(bonk.LocatorFilter{
		HasText: "Active",
		Has:     page.Locator("h2"),
	})
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestFilterCount(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<li>Alpha</li>
		<li>Beta</li>
		<li>Gamma</li>
		<li>Delta</li>
	</body></html>`)

	count, err := page.Locator("li").
		Filter(bonk.LocatorFilter{HasText: "lph"}).
		Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (Alpha)", count)
	}
}

func TestFilterChained(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="card"><span class="tag">Go</span><p>Active</p></div>
		<div class="card"><span class="tag">Go</span><p>Archived</p></div>
		<div class="card"><span class="tag">Rust</span><p>Active</p></div>
	</body></html>`)

	loc := page.Locator(".card").
		Filter(bonk.LocatorFilter{HasText: "Active"}).
		Filter(bonk.LocatorFilter{HasText: "Go"})

	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestAndCSS(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="highlight visible">A</div>
		<div class="highlight">B</div>
		<div class="visible">C</div>
		<div class="highlight visible">D</div>
	</body></html>`)

	loc := page.Locator(".highlight").And(page.Locator(".visible"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	text, err := loc.First().Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "A" {
		t.Errorf("text = %q, want A", text)
	}
}

func TestAndMixed(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button>Submit</button>
		<button>Cancel</button>
		<a href="#">Submit</a>
	</body></html>`)

	loc := page.Locator("button").And(page.GetByText("Submit"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	text, err := loc.First().Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Submit" {
		t.Errorf("text = %q, want Submit", text)
	}
}

func TestAndDisjoint(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="alpha">A</div>
		<div class="beta">B</div>
	</body></html>`)

	loc := page.Locator(".alpha").And(page.Locator(".beta"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestOrCSS(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="error">E1</div>
		<div class="warning">W1</div>
		<div class="info">I1</div>
		<div class="error">E2</div>
	</body></html>`)

	loc := page.Locator(".error").Or(page.Locator(".warning"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestOrMixed(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button>Click</button>
		<span>Hello World</span>
		<p>Other</p>
	</body></html>`)

	loc := page.Locator("button").Or(page.GetByText("Hello World"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestOrDocumentOrder(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="a" class="second">A</div>
		<div id="b" class="first">B</div>
		<div id="c" class="second">C</div>
	</body></html>`)

	loc := page.Locator(".second").Or(page.Locator(".first"))
	texts := make([]string, 0, 3)
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	for i := range count {
		text, err := loc.Nth(i).Text()
		if err != nil {
			t.Fatal(err)
		}
		texts = append(texts, text)
	}

	want := "A,B,C"
	got := strings.Join(texts, ",")
	if got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}

func TestOrDedup(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div class="a b">Both</div>
		<div class="a">Only A</div>
		<div class="b">Only B</div>
	</body></html>`)

	loc := page.Locator(".a").Or(page.Locator(".b"))
	count, err := loc.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3 (deduped)", count)
	}
}
