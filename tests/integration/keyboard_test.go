package integration

import (
	"testing"
)

func TestKeyboardType(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	page.Click("#input")
	kb := page.Keyboard()
	if err := kb.Type("hello"); err != nil {
		t.Fatal(err)
	}

	val, _ := page.Evaluate("document.getElementById('input').value")
	if val != "hello" {
		t.Errorf("input value = %v, want hello", val)
	}
}

func TestKeyboardPress(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="log"></div>
		<script>
			document.addEventListener('keydown', e => {
				document.getElementById('log').textContent = e.key;
			});
		</script>
	</body></html>`)

	kb := page.Keyboard()
	if err := kb.Press("Enter"); err != nil {
		t.Fatal(err)
	}

	val, _ := page.Evaluate("document.getElementById('log').textContent")
	if val != "Enter" {
		t.Errorf("key = %v, want Enter", val)
	}
}

func TestKeyboardInsertText(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body><input id="input" type="text"></body></html>`)

	page.Click("#input")
	kb := page.Keyboard()
	if err := kb.InsertText("pasted"); err != nil {
		t.Fatal(err)
	}

	val, _ := page.Evaluate("document.getElementById('input').value")
	if val != "pasted" {
		t.Errorf("input value = %v, want pasted", val)
	}
}
