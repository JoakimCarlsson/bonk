package integration

import (
	"testing"
)

func TestMouseClick(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button id="btn" onclick="document.title='clicked'" style="width:100px;height:50px;position:absolute;left:10px;top:10px">Click</button>
	</body></html>`)

	mouse := page.Mouse()
	if err := mouse.Click(60, 35); err != nil {
		t.Fatal(err)
	}

	title, _ := page.Title()
	if title != "clicked" {
		t.Errorf("title = %q, want clicked", title)
	}
}

func TestMouseDragTo(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<div id="log"></div>
		<script>
			document.addEventListener('mousedown', () => document.getElementById('log').textContent += 'down,');
			document.addEventListener('mouseup', () => document.getElementById('log').textContent += 'up');
		</script>
	</body></html>`)

	mouse := page.Mouse()
	if err := mouse.DragTo(10, 10, 100, 100); err != nil {
		t.Fatal(err)
	}

	val, _ := page.Evaluate("document.getElementById('log').textContent")
	s, _ := val.(string)
	if s != "down,up" {
		t.Errorf("log = %q, want down,up", s)
	}
}
