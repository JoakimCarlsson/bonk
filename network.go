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
				URL:    ev.Request.URL,
				Method: ev.Request.Method,
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
					URL:    ev.Request.URL,
					Method: ev.Request.Method,
				},
			}
			handler(route)
		},
	)
}
