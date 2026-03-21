package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestWaitForPopup(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	popupCh := make(chan *bonk.Page, 1)
	errCh := make(chan error, 1)
	go func() {
		popup, err := page.WaitForPopup()
		if err != nil {
			errCh <- err
			return
		}
		popupCh <- popup
	}()

	time.Sleep(100 * time.Millisecond)
	page.Evaluate("window.open('https://example.com', '_blank')")

	select {
	case popup := <-popupCh:
		defer popup.Close()
		title, err := popup.Title()
		if err != nil {
			t.Fatalf("popup title: %v", err)
		}
		if title == "" {
			t.Error("popup title is empty")
		}
	case err := <-errCh:
		t.Fatalf("WaitForPopup: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for popup")
	}
}

func TestWaitForPopupFromClick(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.SetContent(
		`<a id="link" href="https://example.com" target="_blank">Open</a>`,
	)

	popupCh := make(chan *bonk.Page, 1)
	errCh := make(chan error, 1)
	go func() {
		popup, err := page.WaitForPopup()
		if err != nil {
			errCh <- err
			return
		}
		popupCh <- popup
	}()

	time.Sleep(100 * time.Millisecond)
	page.Click("#link")

	select {
	case popup := <-popupCh:
		defer popup.Close()
		url, err := popup.URL()
		if err != nil {
			t.Fatalf("popup URL: %v", err)
		}
		if url == "" {
			t.Error("popup URL is empty")
		}
	case err := <-errCh:
		t.Fatalf("WaitForPopup: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for popup")
	}
}

func TestWaitForPopupTimeout(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	_, err := page.WaitForPopup(bonk.PopupTimeout(500 * time.Millisecond))
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if _, ok := err.(*bonk.TimeoutError); !ok {
		t.Fatalf("expected TimeoutError, got %T: %v", err, err)
	}
}
