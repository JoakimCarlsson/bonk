package integration

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/joakimcarlsson/bonk"
)

func TestPoolDo(t *testing.T) {
	b := launchBrowser(t)
	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })

	pool, err := bonk.NewPool(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })

	err = pool.Do(func(page *bonk.Page) error {
		return page.Navigate("https://example.com")
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPoolConcurrent(t *testing.T) {
	b := launchBrowser(t)
	ctx, err := b.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })

	pool, err := bonk.NewPool(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })

	var count atomic.Int64
	var wg sync.WaitGroup

	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Do(func(page *bonk.Page) error {
				page.Navigate("about:blank")
				count.Add(1)
				return nil
			})
		}()
	}
	wg.Wait()

	if count.Load() != 6 {
		t.Errorf("count = %d, want 6", count.Load())
	}
}
