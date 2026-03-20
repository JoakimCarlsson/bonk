package integration

import (
	"testing"

	"github.com/joakimcarlsson/bonk"
)

func launchBrowser(t *testing.T) *bonk.Browser {
	t.Helper()
	b, err := bonk.Launch()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })
	return b
}

func newPage(t *testing.T, b *bonk.Browser) *bonk.Page {
	t.Helper()
	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatal(err)
	}
	return page
}
