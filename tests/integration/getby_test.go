package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestGetByText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<p>Hello World</p>
		<span>Click me</span>
	</body></html>`)

	text, err := page.GetByText("Hello World").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello World" {
		t.Errorf("text = %q, want %q", text, "Hello World")
	}
}

func TestGetByTextSubstring(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><p>Hello World</p></body></html>`)

	text, err := page.GetByText("Hello").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello World" {
		t.Errorf("text = %q, want %q", text, "Hello World")
	}
}

func TestGetByTextExact(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<p>Hello</p>
		<p>Hello World</p>
	</body></html>`)

	text, err := page.GetByText("Hello", bonk.Exact()).Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello" {
		t.Errorf("text = %q, want %q", text, "Hello")
	}
}

func TestGetByRole(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button>Submit</button>
		<a href="/home">Home</a>
	</body></html>`)

	text, err := page.GetByRole("button").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Submit" {
		t.Errorf("text = %q, want %q", text, "Submit")
	}

	linkText, err := page.GetByRole("link").Text()
	if err != nil {
		t.Fatal(err)
	}
	if linkText != "Home" {
		t.Errorf("link text = %q, want %q", linkText, "Home")
	}
}

func TestGetByRoleImplicitHeading(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<h1>Main Title</h1>
		<h2>Subtitle</h2>
	</body></html>`)

	count, err := page.GetByRole("heading").Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("heading count = %d, want 2", count)
	}

	text, err := page.GetByRole("heading").First().Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Main Title" {
		t.Errorf("text = %q, want %q", text, "Main Title")
	}
}

func TestGetByRoleWithName(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button>Cancel</button>
		<button>Submit</button>
	</body></html>`)

	text, err := page.GetByRole("button", bonk.WithName("Submit")).Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Submit" {
		t.Errorf("text = %q, want %q", text, "Submit")
	}
}

func TestGetByRoleWithNameExact(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button>Submit Form</button>
		<button>Submit</button>
	</body></html>`)

	text, err := page.GetByRole("button", bonk.WithName("Submit", bonk.Exact())).
		Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Submit" {
		t.Errorf("text = %q, want %q", text, "Submit")
	}
}

func TestGetByRoleExplicit(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div role="button">Custom Button</div>
	</body></html>`)

	text, err := page.GetByRole("button").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Custom Button" {
		t.Errorf("text = %q, want %q", text, "Custom Button")
	}
}

func TestGetByLabel(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<label for="email">Email</label>
		<input id="email" type="email" value="test@example.com">
	</body></html>`)

	val, err := page.GetByLabel("Email").Attribute("value")
	if err != nil {
		t.Fatal(err)
	}
	if val != "test@example.com" {
		t.Errorf("value = %q, want %q", val, "test@example.com")
	}
}

func TestGetByLabelNested(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<label>Username <input type="text" value="alice"></label>
	</body></html>`)

	val, err := page.GetByLabel("Username").Attribute("value")
	if err != nil {
		t.Fatal(err)
	}
	if val != "alice" {
		t.Errorf("value = %q, want %q", val, "alice")
	}
}

func TestGetByLabelAriaLabel(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<input aria-label="Search" type="search" value="query">
	</body></html>`)

	val, err := page.GetByLabel("Search").Attribute("value")
	if err != nil {
		t.Fatal(err)
	}
	if val != "query" {
		t.Errorf("value = %q, want %q", val, "query")
	}
}

func TestGetByPlaceholder(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<input placeholder="Enter your name" value="John">
	</body></html>`)

	val, err := page.GetByPlaceholder("Enter your name").Attribute("value")
	if err != nil {
		t.Fatal(err)
	}
	if val != "John" {
		t.Errorf("value = %q, want %q", val, "John")
	}
}

func TestGetByPlaceholderSubstring(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<input placeholder="Enter your name" value="John">
	</body></html>`)

	val, err := page.GetByPlaceholder("your name").Attribute("value")
	if err != nil {
		t.Fatal(err)
	}
	if val != "John" {
		t.Errorf("value = %q, want %q", val, "John")
	}
}

func TestGetByTestID(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div data-testid="greeting">Hello</div>
	</body></html>`)

	text, err := page.GetByTestID("greeting").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Hello" {
		t.Errorf("text = %q, want %q", text, "Hello")
	}
}

func TestGetByAltText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<img alt="Company Logo" src="logo.png">
	</body></html>`)

	src, err := page.GetByAltText("Company Logo").Attribute("src")
	if err != nil {
		t.Fatal(err)
	}
	if src != "logo.png" {
		t.Errorf("src = %q, want %q", src, "logo.png")
	}
}

func TestGetByAltTextSubstring(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<img alt="Company Logo" src="logo.png">
	</body></html>`)

	src, err := page.GetByAltText("Logo").Attribute("src")
	if err != nil {
		t.Fatal(err)
	}
	if src != "logo.png" {
		t.Errorf("src = %q, want %q", src, "logo.png")
	}
}

func TestGetByTitle(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<span title="Close dialog">X</span>
	</body></html>`)

	text, err := page.GetByTitle("Close dialog").Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "X" {
		t.Errorf("text = %q, want %q", text, "X")
	}
}

func TestGetByRoleCount(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
			<li>Item 3</li>
		</ul>
	</body></html>`)

	count, err := page.GetByRole("listitem").Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestGetByRoleNth(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<ul>
			<li>First</li>
			<li>Second</li>
			<li>Third</li>
		</ul>
	</body></html>`)

	text, err := page.GetByRole("listitem").Nth(1).Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Second" {
		t.Errorf("text = %q, want %q", text, "Second")
	}
}

func TestGetByTextTimeout(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><p>Hello</p></body></html>`)

	err := page.GetByText("nonexistent").
		WaitFor(bonk.WaitTimeout(500 * time.Millisecond))
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestGetByRoleWithAriaLabel(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button aria-label="Close">X</button>
		<button>Save</button>
	</body></html>`)

	text, err := page.GetByRole("button", bonk.WithName("Close")).Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "X" {
		t.Errorf("text = %q, want %q", text, "X")
	}
}
