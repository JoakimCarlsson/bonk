package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

const testTimeout = 30 * time.Second

func launchBrowser(t *testing.T) *bonk.Browser {
	t.Helper()
	deadline := time.After(testTimeout)
	done := make(chan struct{})
	var b *bonk.Browser
	var err error
	go func() {
		b, err = bonk.Launch()
		close(done)
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out launching browser")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })
	return b
}

func launchBrowserNoStealth(t *testing.T) *bonk.Browser {
	t.Helper()
	deadline := time.After(testTimeout)
	done := make(chan struct{})
	var b *bonk.Browser
	var err error
	go func() {
		b, err = bonk.Launch(bonk.Stealth(false))
		close(done)
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out launching browser")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })
	return b
}

func newPage(t *testing.T, b *bonk.Browser) *bonk.Page {
	t.Helper()
	deadline := time.After(testTimeout)
	done := make(chan struct{})
	var ctx *bonk.BrowserContext
	var page *bonk.Page
	var err error
	go func() {
		ctx, err = b.NewContext()
		if err == nil {
			page, err = ctx.NewPage()
		}
		close(done)
	}()
	select {
	case <-done:
	case <-deadline:
		t.Fatal("timed out creating page")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })
	return page
}
