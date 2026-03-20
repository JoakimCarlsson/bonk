package bonk

import (
	"encoding/json"

	"github.com/joakimcarlsson/bonk/proto"
)

// OnRequest registers a handler for outgoing network requests.
// Returns an unsubscribe function.
func (p *Page) OnRequest(fn func(*Request)) func() {
	return p.session.Subscribe(
		proto.NetworkEventRequestWillBeSentMethod,
		func(params json.RawMessage) {
			var ev proto.NetworkEventRequestWillBeSent
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			fn(&Request{
				URL:      ev.Request.URL,
				Method:   ev.Request.Method,
				Headers:  extractHeaders(params, "request", "headers"),
				PostData: ev.Request.PostData,
			})
		},
	)
}

// OnResponse registers a handler for network responses.
// Returns an unsubscribe function.
func (p *Page) OnResponse(fn func(*Response)) func() {
	return p.session.Subscribe(
		proto.NetworkEventResponseReceivedMethod,
		func(params json.RawMessage) {
			var ev proto.NetworkEventResponseReceived
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			fn(&Response{
				page:      p,
				requestID: ev.RequestID,
				URL:       ev.Response.URL,
				Status:    int64(ev.Response.Status),
				Headers:   extractHeaders(params, "response", "headers"),
			})
		},
	)
}

// Route intercepts requests matching the URL pattern.
// The handler receives a Route that must be continued, fulfilled, or aborted.
// Returns an unsubscribe function.
func (p *Page) Route(pattern string, handler func(*Route)) func() {
	proto.FetchEnable().
		WithPatterns([]proto.FetchRequestPattern{
			{URLPattern: pattern},
		}).
		Do(p.execCtx)

	return p.session.Subscribe(
		proto.FetchEventRequestPausedMethod,
		func(params json.RawMessage) {
			var ev proto.FetchEventRequestPaused
			if err := json.Unmarshal(params, &ev); err != nil {
				return
			}
			route := &Route{
				page:      p,
				requestID: ev.RequestID,
				Request: &Request{
					URL:     ev.Request.URL,
					Method:  ev.Request.Method,
					Headers: extractHeaders(params, "request", "headers"),
				},
			}
			handler(route)
		},
	)
}

func extractHeaders(raw json.RawMessage, path ...string) map[string]string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}

	current := obj
	for _, key := range path[:len(path)-1] {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(current[key], &nested); err != nil {
			return nil
		}
		current = nested
	}

	lastKey := path[len(path)-1]
	headerRaw, ok := current[lastKey]
	if !ok {
		return nil
	}

	var headers map[string]string
	if err := json.Unmarshal(headerRaw, &headers); err != nil {
		return nil
	}
	return headers
}
