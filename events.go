package bonk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
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

// OnConsole registers a handler for console messages.
func (p *Page) OnConsole(fn func(*ConsoleMessage)) func() {
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

// OnDialog registers a handler for JavaScript dialogs.
func (p *Page) OnDialog(fn func(*Dialog)) func() {
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
