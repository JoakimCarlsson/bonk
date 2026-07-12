package integration

import (
	"fmt"
	"testing"

	"github.com/joakimcarlsson/bonk"
)

func TestAccessibilitySnapshotClickByNode(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("about:blank")
	page.SetContent(`<html><body>
		<button onclick="location.hash='b1'">Go</button>
		<button onclick="location.hash='b2'">Go</button>
		<button onclick="location.hash='b3'">Go</button>
	</body></html>`)

	nodes, err := page.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}

	_, refs := bonk.FormatAccessibilityTreeIndexed(nodes)
	var buttons []*bonk.AXNode
	for _, n := range refs {
		if n.Role == "button" {
			buttons = append(buttons, n)
		}
	}
	if len(buttons) != 3 {
		t.Fatalf("button nodes = %d, want 3", len(buttons))
	}
	if buttons[1].BackendNodeID == 0 {
		t.Fatal("node has no backend node id")
	}

	if err := page.ClickNode(buttons[1]); err != nil {
		t.Fatal(err)
	}

	hash, err := page.Evaluate("location.hash")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(hash); got != "#b2" {
		t.Errorf("hash = %q, want %q (clicked the wrong same-named button)", got, "#b2")
	}
}
