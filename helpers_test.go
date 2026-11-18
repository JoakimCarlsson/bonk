package bonk_test

import (
	"net/http"
	"net/http/httptest"
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

// examplePageHTML is the classic example.com document the element, locator
// and page tests were written against, served locally so the tests do not
// depend on the live site's markup.
const examplePageHTML = `<!doctype html>
<html>
<head><title>Example Domain</title></head>
<body>
<div>
<h1>Example Domain</h1>
<p>This domain is for use in illustrative examples in documents.</p>
<p><a href="https://iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>`

// exampleURL starts a local HTTP server serving examplePageHTML for the
// duration of the test and returns its URL.
func exampleURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(examplePageHTML))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/"
}
