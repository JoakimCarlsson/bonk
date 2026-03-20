package integration

import (
	"testing"
	"time"

	"github.com/joakimcarlsson/bonk"
)

func TestOnConsole(t *testing.T) {
	b := launchBrowserNoStealth(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	ch := make(chan *bonk.ConsoleMessage, 3)
	unsub := page.On(bonk.ConsoleEvent, func(msg *bonk.ConsoleMessage) {
		ch <- msg
	})
	defer unsub()

	page.Evaluate("console.log('hello')")
	page.Evaluate("console.warn('warning')")
	page.Evaluate("console.error('error')")

	types := map[string]bool{}
	for i := 0; i < 3; i++ {
		select {
		case msg := <-ch:
			types[msg.Type] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for console message %d", i+1)
		}
	}

	if !types["log"] {
		t.Error("missing console.log message")
	}
	if !types["warning"] {
		t.Error("missing console.warn message")
	}
	if !types["error"] {
		t.Error("missing console.error message")
	}
}

func TestOnDialog(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	ch := make(chan *bonk.Dialog, 1)
	unsub := page.On(bonk.DialogEvent, func(d *bonk.Dialog) {
		ch <- d
		d.Accept()
	})
	defer unsub()

	page.Evaluate("alert('test alert')")

	select {
	case d := <-ch:
		if d.Type != "alert" {
			t.Errorf("type = %q, want alert", d.Type)
		}
		if d.Message != "test alert" {
			t.Errorf("message = %q, want 'test alert'", d.Message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for dialog")
	}
}

func TestOnConsoleConvenience(t *testing.T) {
	b := launchBrowserNoStealth(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	ch := make(chan string, 1)
	unsub := page.OnConsole(func(msg *bonk.ConsoleMessage) {
		ch <- msg.Text
	})
	defer unsub()

	page.Evaluate("console.log('convenience')")

	select {
	case text := <-ch:
		if text != "convenience" {
			t.Errorf("text = %q", text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestOnDialogConvenience(t *testing.T) {
	b := launchBrowser(t)
	page := newPage(t, b)

	page.Navigate("https://example.com")

	ch := make(chan string, 1)
	unsub := page.OnDialog(func(d *bonk.Dialog) {
		ch <- d.Message
		d.Dismiss()
	})
	defer unsub()

	page.Evaluate("alert('dismiss me')")

	select {
	case msg := <-ch:
		if msg != "dismiss me" {
			t.Errorf("message = %q", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
