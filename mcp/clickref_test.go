package mcp_test

import (
	"strings"
	"testing"
)

func TestClickRef(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{"url": "about:blank"})
	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.body.innerHTML = '<button onclick="window.clicked=1">Go</button><button onclick="window.clicked=2">Go</button>'`,
	})

	snap := resultText(t, callTool(t, c, "snapshot", nil))
	if !strings.Contains(snap, "[1]") || !strings.Contains(snap, "[2]") {
		t.Fatalf("snapshot missing indexed buttons:\n%s", snap)
	}

	res := callTool(t, c, "click_ref", map[string]any{"ref": 2})
	if res.IsError {
		t.Fatalf("click_ref failed: %s", resultText(t, res))
	}

	got := resultText(t, callTool(t, c, "evaluate", map[string]any{
		"expression": "window.clicked",
	}))
	if !strings.Contains(got, "2") {
		t.Errorf("window.clicked = %q, want 2 (clicked the wrong button)", got)
	}
}

func TestClickRefWithoutSnapshot(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{"url": "about:blank"})

	res := callTool(t, c, "click_ref", map[string]any{"ref": 1})
	if !res.IsError {
		t.Fatal("click_ref without a prior snapshot should error")
	}
}
