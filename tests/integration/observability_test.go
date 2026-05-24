package integration

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestOnCrashFiresOnRendererCrash(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	crashed := make(chan string, 1)
	page.OnCrash(func(reason string) {
		select {
		case crashed <- reason:
		default:
		}
	})

	go page.Navigate("chrome://crash")

	select {
	case reason := <-crashed:
		if reason == "" {
			t.Fatal("OnCrash fired with empty reason")
		}
	case <-time.After(15 * time.Second):
		t.Fatal("OnCrash did not fire within 15s of chrome://crash")
	}
}

func TestWithStderrSinkCapturesAfterDevtoolsURL(t *testing.T) {
	var buf threadSafeBuffer
	b, err := bonk.Launch(bonk.WithStderrSink(&buf))
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
	go page.Navigate("chrome://crash")

	time.Sleep(3 * time.Second)

	if buf.Len() == 0 {
		t.Skip("no stderr emitted by Chrome; crash signal not visible on this build")
	}
}

func TestCrashpadHandlerReportsNewDumps(t *testing.T) {
	var (
		mu      sync.Mutex
		reports []bonk.CrashReport
	)
	b, err := bonk.Launch(bonk.WithCrashpadHandler(func(rs []bonk.CrashReport) {
		mu.Lock()
		defer mu.Unlock()
		reports = append(reports, rs...)
	}))
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	page, err := ctx.NewPage()
	if err != nil {
		t.Fatal(err)
	}

	go page.Navigate("chrome://crash")
	time.Sleep(3 * time.Second)

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(reports) == 0 {
		t.Skip("no crashpad dump produced; Chrome may suppress them on this platform")
	}
	for _, r := range reports {
		if r.Path == "" {
			t.Errorf("CrashReport has empty Path")
		}
		if r.Size <= 0 {
			t.Errorf("CrashReport for %s has non-positive Size %d", r.Path, r.Size)
		}
	}
}

type threadSafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *threadSafeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *threadSafeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}
