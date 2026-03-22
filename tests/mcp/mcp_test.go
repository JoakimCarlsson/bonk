package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	bonkmcp "github.com/joakimcarlsson/bonk/mcp"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func setupServer(
	t *testing.T,
) (*client.Client, *bonkmcp.Session) {
	t.Helper()

	sess := bonkmcp.NewSession(
		bonkmcp.WithHeadless(true),
		bonkmcp.WithStealth(false),
	)
	t.Cleanup(func() { sess.Close() })

	s := bonkmcp.NewServer(sess)

	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })

	ctx := context.Background()
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "bonkmcp-test",
		Version: "0.1.0",
	}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		t.Fatal(err)
	}

	return c, sess
}

func callTool(
	t *testing.T,
	c *client.Client,
	name string,
	args map[string]any,
) *mcp.CallToolResult {
	t.Helper()

	ctx := context.Background()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args

	result, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return result
}

func resultText(
	t *testing.T,
	result *mcp.CallToolResult,
) string {
	t.Helper()
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatal("no text content in result")
	return ""
}

func TestBrowserLaunchAndClose(t *testing.T) {
	c, _ := setupServer(t)

	result := callTool(t, c, "browser_launch", nil)
	text := resultText(t, result)
	if text != "Browser launched" {
		t.Errorf("got %q, want %q", text, "Browser launched")
	}

	result = callTool(t, c, "browser_launch", nil)
	text = resultText(t, result)
	if text != "Browser already running" {
		t.Errorf(
			"got %q, want %q",
			text,
			"Browser already running",
		)
	}

	result = callTool(t, c, "browser_close", nil)
	text = resultText(t, result)
	if text != "Browser closed" {
		t.Errorf("got %q, want %q", text, "Browser closed")
	}
}

func TestNavigateAndGetContent(t *testing.T) {
	c, _ := setupServer(t)

	result := callTool(t, c, "navigate", map[string]any{
		"url": "https://example.com",
	})
	text := resultText(t, result)
	if !strings.Contains(text, "example.com") {
		t.Errorf("navigate result should contain URL: %s", text)
	}
	if !strings.Contains(text, "Example Domain") {
		t.Errorf(
			"navigate result should contain title: %s",
			text,
		)
	}

	result = callTool(t, c, "get_content", nil)
	text = resultText(t, result)
	if !strings.Contains(text, "Example Domain") {
		t.Errorf("content should contain title: %s", text)
	}
}

func TestScreenshot(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{
		"url": "https://example.com",
	})

	result := callTool(t, c, "screenshot", nil)
	if result.IsError {
		t.Fatalf("screenshot returned error: %v", result.Content)
	}

	hasImage := false
	for _, content := range result.Content {
		if img, ok := content.(mcp.ImageContent); ok {
			hasImage = true
			if img.MIMEType != "image/jpeg" {
				t.Errorf(
					"mime = %q, want image/jpeg",
					img.MIMEType,
				)
			}
			if len(img.Data) < 100 {
				t.Error("image data too small")
			}
		}
	}
	if !hasImage {
		t.Error("screenshot should return image content")
	}
}

func TestClickAndQuery(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><button id="btn" onclick="this.textContent=\'clicked\'">Click me</button></body></html>');document.close()`,
	})

	result := callTool(t, c, "query", map[string]any{
		"selector": "#btn",
	})
	text := resultText(t, result)
	if !strings.Contains(text, "Click me") {
		t.Errorf("query should find button text: %s", text)
	}

	callTool(t, c, "click", map[string]any{
		"selector": "#btn",
	})

	result = callTool(t, c, "query", map[string]any{
		"selector": "#btn",
	})
	text = resultText(t, result)
	if !strings.Contains(text, "clicked") {
		t.Errorf(
			"button text should be 'clicked' after click: %s",
			text,
		)
	}
}

func TestFill(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><input id="input" type="text"></body></html>');document.close()`,
	})

	callTool(t, c, "fill", map[string]any{
		"selector": "#input",
		"text":     "hello world",
	})

	result := callTool(t, c, "evaluate", map[string]any{
		"expression": `document.querySelector('#input').value`,
	})
	text := resultText(t, result)
	if !strings.Contains(text, "hello world") {
		t.Errorf(
			"input value should be 'hello world': %s",
			text,
		)
	}
}

func TestQueryAll(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><ul><li>one</li><li>two</li><li>three</li></ul></body></html>');document.close()`,
	})

	result := callTool(t, c, "query_all", map[string]any{
		"selector": "li",
	})
	text := resultText(t, result)
	if !strings.Contains(text, "Found 3 element(s)") {
		t.Errorf("should find 3 list items: %s", text)
	}
	if !strings.Contains(text, "one") ||
		!strings.Contains(text, "two") ||
		!strings.Contains(text, "three") {
		t.Errorf("should contain all item texts: %s", text)
	}
}

func TestMultiplePages(t *testing.T) {
	c, _ := setupServer(t)

	result := callTool(t, c, "new_page", map[string]any{
		"url": "https://example.com",
	})
	text := resultText(t, result)
	if !strings.Contains(text, "page_") {
		t.Errorf("new_page should return page ID: %s", text)
	}

	result = callTool(t, c, "list_pages", nil)
	text = resultText(t, result)

	var pages map[string]string
	if err := json.Unmarshal([]byte(text), &pages); err != nil {
		t.Fatalf("list_pages should return valid JSON: %v", err)
	}
	if len(pages) < 1 {
		t.Error("should have at least 1 page")
	}
}

func TestEvaluate(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{
		"url": "https://example.com",
	})

	result := callTool(t, c, "evaluate", map[string]any{
		"expression": "1 + 2",
	})
	text := resultText(t, result)
	if !strings.Contains(text, "3") {
		t.Errorf("evaluate 1+2 should return 3: %s", text)
	}
}

func TestCheckUncheck(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><input id="cb" type="checkbox"></body></html>');document.close()`,
	})

	callTool(t, c, "check", map[string]any{
		"selector": "#cb",
	})

	result := callTool(t, c, "evaluate", map[string]any{
		"expression": `document.querySelector('#cb').checked`,
	})
	text := resultText(t, result)
	if !strings.Contains(text, "true") {
		t.Errorf("checkbox should be checked: %s", text)
	}

	callTool(t, c, "uncheck", map[string]any{
		"selector": "#cb",
	})

	result = callTool(t, c, "evaluate", map[string]any{
		"expression": `document.querySelector('#cb').checked`,
	})
	text = resultText(t, result)
	if !strings.Contains(text, "false") {
		t.Errorf("checkbox should be unchecked: %s", text)
	}
}

func TestSelectOption(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><select id="sel"><option value="a">A</option><option value="b">B</option></select></body></html>');document.close()`,
	})

	callTool(t, c, "select_option", map[string]any{
		"selector": "#sel",
		"value":    "b",
	})

	result := callTool(t, c, "evaluate", map[string]any{
		"expression": `document.querySelector('#sel').value`,
	})
	text := resultText(t, result)
	if !strings.Contains(text, "b") {
		t.Errorf("selected value should be 'b': %s", text)
	}
}

func TestSetExtraHeaders(t *testing.T) {
	c, _ := setupServer(t)

	result := callTool(
		t,
		c,
		"set_extra_headers",
		map[string]any{
			"headers": map[string]any{
				"X-Custom": "test-value",
			},
		},
	)
	text := resultText(t, result)
	if !strings.Contains(text, "1 header(s)") {
		t.Errorf("should confirm 1 header set: %s", text)
	}
}

func TestCookies(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{
		"url": "https://example.com",
	})

	callTool(t, c, "set_cookies", map[string]any{
		"cookies": []any{
			map[string]any{
				"name":   "test",
				"value":  "abc",
				"domain": "example.com",
				"path":   "/",
			},
		},
	})

	result := callTool(t, c, "get_cookies", nil)
	text := resultText(t, result)
	if !strings.Contains(text, "test") ||
		!strings.Contains(text, "abc") {
		t.Errorf(
			"cookies should contain test=abc: %s",
			text,
		)
	}

	callTool(t, c, "clear_cookies", nil)

	result = callTool(t, c, "get_cookies", nil)
	text = resultText(t, result)
	if strings.Contains(text, "test") {
		t.Errorf(
			"cookies should be cleared: %s",
			text,
		)
	}
}

func TestErrorOnMissingElement(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "navigate", map[string]any{
		"url": "https://example.com",
	})

	result := callTool(t, c, "click", map[string]any{
		"selector": "#nonexistent",
	})
	if !result.IsError {
		t.Error("click on nonexistent element should return error")
	}
}

func TestWaitForSelector(t *testing.T) {
	c, _ := setupServer(t)

	callTool(t, c, "evaluate", map[string]any{
		"expression": `document.open();document.write('<html><body><div id="target">here</div></body></html>');document.close()`,
	})

	result := callTool(
		t,
		c,
		"wait_for_selector",
		map[string]any{
			"selector":   "#target",
			"timeout_ms": 5000,
		},
	)
	if result.IsError {
		t.Errorf(
			"wait_for_selector should succeed: %v",
			result.Content,
		)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "here") {
		t.Errorf("should contain element text: %s", text)
	}
}
