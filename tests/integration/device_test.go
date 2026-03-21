package integration

import (
	"strings"
	"testing"

	"github.com/joakimcarlsson/bonk"
)

func TestEmulateIPhone15(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	if err := page.Emulate(bonk.IPhone15); err != nil {
		t.Fatal(err)
	}

	page.Navigate("https://example.com")

	ua, err := page.Evaluate("navigator.userAgent")
	if err != nil {
		t.Fatal(err)
	}
	uaStr, _ := ua.(string)
	if !strings.Contains(uaStr, "iPhone") {
		t.Errorf("userAgent = %q, want containing iPhone", uaStr)
	}

	width, err := page.Evaluate("window.innerWidth")
	if err != nil {
		t.Fatal(err)
	}
	w, _ := width.(float64)
	if w != float64(bonk.IPhone15.ViewportWidth) {
		t.Errorf(
			"innerWidth = %v, want %d",
			w,
			bonk.IPhone15.ViewportWidth,
		)
	}
}

func TestEmulateCustomDevice(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	custom := bonk.Device{
		Name:              "Test Device",
		UserAgent:         "TestAgent/1.0",
		ViewportWidth:     500,
		ViewportHeight:    800,
		DeviceScaleFactor: 2,
		IsMobile:          true,
		HasTouch:          true,
	}

	if err := page.Emulate(custom); err != nil {
		t.Fatal(err)
	}

	page.Navigate("https://example.com")

	ua, _ := page.Evaluate("navigator.userAgent")
	if ua != "TestAgent/1.0" {
		t.Errorf("userAgent = %v, want TestAgent/1.0", ua)
	}

	width, _ := page.Evaluate("window.innerWidth")
	if w, _ := width.(float64); w != 500 {
		t.Errorf("innerWidth = %v, want 500", w)
	}
}
