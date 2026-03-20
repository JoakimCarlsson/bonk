package bonk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

// EventType identifies a page event.
type EventType string

const (
	ConsoleEvent  EventType = "console"
	DialogEvent   EventType = "dialog"
	DownloadEvent EventType = "download"
)

// ConsoleMessage represents a console API call.
type ConsoleMessage struct {
	Type string
	Text string
	Args []any
}

// Dialog represents a JavaScript dialog (alert, confirm, prompt).
type Dialog struct {
	page          *Page
	Type          string
	Message       string
	DefaultPrompt string
}

// Accept accepts the dialog, optionally providing text for prompts.
func (d *Dialog) Accept(text ...string) error {
	params := proto.PageHandleJavaScriptDialog(true)
	if len(text) > 0 {
		params = params.WithPromptText(text[0])
	}
	return params.Do(d.page.execCtx)
}

// Dismiss dismisses the dialog.
func (d *Dialog) Dismiss() error {
	return proto.PageHandleJavaScriptDialog(false).Do(d.page.execCtx)
}

// Download represents a file download.
type Download struct {
	page              *Page
	guid              string
	downloadDir       string
	url               string
	suggestedFilename string
	done              chan struct{}
	savedPath         string
	mu                sync.Mutex
}

// URL returns the download URL.
func (d *Download) URL() string {
	return d.url
}

// SuggestedFilename returns the browser-suggested filename.
func (d *Download) SuggestedFilename() string {
	return d.suggestedFilename
}

// SaveAs copies the downloaded file to the given path.
// Blocks until the download completes.
func (d *Download) SaveAs(path string) error {
	return d.SaveAsContext(context.Background(), path)
}

// SaveAsContext copies the downloaded file to the given path,
// respecting the context for cancellation and deadlines.
func (d *Download) SaveAsContext(ctx context.Context, path string) error {
	select {
	case <-d.done:
	case <-ctx.Done():
		return ctx.Err()
	}

	d.mu.Lock()
	src := d.savedPath
	d.mu.Unlock()

	if src == "" {
		return fmt.Errorf("bonk: download failed or path unknown")
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// On registers a typed event handler. The handler type must match the event:
//   - Console  → func(*ConsoleMessage)
//   - Dialog   → func(*Dialog)
//   - Download → func(*Download)
func (p *Page) On(event EventType, handler any) func() {
	switch event {
	case ConsoleEvent:
		fn, ok := handler.(func(*ConsoleMessage))
		if !ok {
			panic("bonk: Console handler must be func(*ConsoleMessage)")
		}
		return p.onConsole(fn)

	case DialogEvent:
		fn, ok := handler.(func(*Dialog))
		if !ok {
			panic("bonk: Dialog handler must be func(*Dialog)")
		}
		return p.onDialog(fn)

	case DownloadEvent:
		fn, ok := handler.(func(*Download))
		if !ok {
			panic("bonk: Download handler must be func(*Download)")
		}
		return p.onDownload(fn)

	default:
		panic(fmt.Sprintf("bonk: unknown event type %q", event))
	}
}

// OnConsole registers a handler for console messages.
func (p *Page) OnConsole(fn func(*ConsoleMessage)) func() {
	return p.onConsole(fn)
}

// OnDialog registers a handler for JavaScript dialogs.
func (p *Page) OnDialog(fn func(*Dialog)) func() {
	return p.onDialog(fn)
}

func (p *Page) onConsole(fn func(*ConsoleMessage)) func() {
	return p.session.Subscribe(
		proto.RuntimeEventConsoleAPICalledMethod,
		func(params json.RawMessage) {
			var ev proto.RuntimeEventConsoleAPICalled
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}

			var args []any
			var texts []string
			for _, arg := range ev.Args {
				if len(arg.Value) > 0 {
					var val any
					json.Unmarshal(arg.Value, &val)
					args = append(args, val)
					texts = append(texts, fmt.Sprint(val))
				} else if arg.Description != "" {
					args = append(args, arg.Description)
					texts = append(texts, arg.Description)
				}
			}

			fn(&ConsoleMessage{
				Type: string(ev.Type),
				Text: strings.Join(texts, " "),
				Args: args,
			})
		},
	)
}

func (p *Page) onDialog(fn func(*Dialog)) func() {
	return p.session.Subscribe(
		proto.PageEventJavascriptDialogOpeningMethod,
		func(params json.RawMessage) {
			var ev proto.PageEventJavascriptDialogOpening
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			fn(&Dialog{
				page:          p,
				Type:          string(ev.Type),
				Message:       ev.Message,
				DefaultPrompt: ev.DefaultPrompt,
			})
		},
	)
}

func (p *Page) onDownload(fn func(*Download)) func() {
	downloadDir, _ := os.MkdirTemp("", "bonk-downloads-*")

	proto.PageSetDownloadBehavior(proto.PageSetDownloadBehaviorBehaviorAllow).
		WithDownloadPath(downloadDir).
		Do(p.execCtx)

	downloads := &sync.Map{}

	u1 := p.session.Subscribe(
		proto.PageEventDownloadWillBeginMethod,
		func(params json.RawMessage) {
			var ev proto.PageEventDownloadWillBegin
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			d := &Download{
				page:              p,
				guid:              ev.Guid,
				downloadDir:       downloadDir,
				url:               ev.URL,
				suggestedFilename: ev.SuggestedFilename,
				done:              make(chan struct{}),
			}
			downloads.Store(ev.Guid, d)
			fn(d)
		},
	)

	u2 := p.session.Subscribe(
		proto.PageEventDownloadProgressMethod,
		func(params json.RawMessage) {
			var ev proto.PageEventDownloadProgress
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			if ev.State != proto.PageDownloadProgressStateCompleted {
				return
			}
			if v, ok := downloads.Load(ev.Guid); ok {
				d := v.(*Download)
				d.mu.Lock()
				d.savedPath = filepath.Join(d.downloadDir, d.guid)
				d.mu.Unlock()
				close(d.done)
			}
		},
	)

	return func() {
		u1()
		u2()
		os.RemoveAll(downloadDir)
	}
}
