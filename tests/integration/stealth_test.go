package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestStealthNavigatorWebdriver(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate("navigator.webdriver")
	if err != nil {
		t.Fatal(err)
	}
	if result == true {
		t.Error("navigator.webdriver = true, want false or undefined")
	}
}

func TestStealthNavigatorPlugins(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate("navigator.plugins.length")
	if err != nil {
		t.Fatal(err)
	}
	count, _ := result.(float64)
	if count < 1 {
		t.Errorf("navigator.plugins.length = %v, want > 0", result)
	}
}

func TestStealthChromeRuntime(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate("typeof window.chrome")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.(string)
	if s != "object" {
		t.Errorf("typeof window.chrome = %q, want object", s)
	}
}

func TestStealthUserAgent(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	result, err := page.Evaluate("navigator.userAgent")
	if err != nil {
		t.Fatal(err)
	}
	ua, _ := result.(string)
	if strings.Contains(ua, "Headless") {
		t.Errorf("UA contains 'Headless': %s", ua)
	}
}

func TestStealthLanguages(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate("JSON.stringify(navigator.languages)")
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.(string)
	if !strings.Contains(s, "en-US") {
		t.Errorf("languages = %s, want containing en-US", s)
	}
}

func TestStealthPermissionsConsistency(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate(`
		(async () => {
			var perm = await navigator.permissions.query({name: 'notifications'});
			return perm.state === Notification.permission ||
			       (perm.state === 'prompt' && Notification.permission === 'default');
		})()
	`)
	if err != nil {
		t.Fatal(err)
	}
	if result != true {
		t.Error("permissions API inconsistency detected")
	}
}

func TestStealthWebGLVendor(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate(`
		(function() {
			var canvas = document.createElement('canvas');
			var gl = canvas.getContext('webgl');
			if (!gl) return 'no webgl';
			var ext = gl.getExtension('WEBGL_debug_renderer_info');
			if (!ext) return 'no ext';
			return gl.getParameter(ext.UNMASKED_VENDOR_WEBGL);
		})()
	`)
	if err != nil {
		t.Fatal(err)
	}
	vendor, _ := result.(string)
	if vendor == "" {
		t.Error("WebGL vendor should not be empty")
	}
	t.Logf("WebGL vendor = %q", vendor)
}

func TestStealthHardwareConcurrency(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")
	time.Sleep(200 * time.Millisecond)

	result, err := page.Evaluate("navigator.hardwareConcurrency")
	if err != nil {
		t.Fatal(err)
	}
	cores, _ := result.(float64)
	if cores != 8 {
		t.Errorf("hardwareConcurrency = %v, want 8", result)
	}
}

func TestStealthBotSannysoft(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://bot.sannysoft.com/")
	time.Sleep(3 * time.Second)

	result, err := page.Evaluate(`
		(function() {
			var rows = document.querySelectorAll('table tr');
			var fails = [];
			for (var i = 0; i < rows.length; i++) {
				var cells = rows[i].querySelectorAll('td');
				if (cells.length >= 2) {
					var status = cells[cells.length - 1].textContent.trim().toLowerCase();
					var test = cells[0].textContent.trim();
					if (status.indexOf('fail') !== -1 || status.indexOf('lie') !== -1) {
						fails.push(test + ': ' + status);
					}
				}
			}
			return fails.join('; ');
		})()
	`)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := result.(string)
	if s != "" {
		t.Logf("bot.sannysoft.com failures: %s", s)
	} else {
		t.Log("bot.sannysoft.com: all green")
	}
}

func TestStealthDisabled(t *testing.T) {
	b, err := bonk.Launch(bonk.Stealth(false))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })

	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatal(err)
	}

	page.Navigate("https://example.com")

	result, err := page.Evaluate("navigator.plugins.length")
	if err != nil {
		t.Fatal(err)
	}
	count, _ := result.(float64)
	t.Logf("stealth=false: navigator.plugins.length = %v", count)
}
