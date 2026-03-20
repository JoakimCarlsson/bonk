package bonk

import (
	"encoding/base64"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

// Route represents an intercepted request that can be continued, fulfilled, or aborted.
type Route struct {
	page      *Page
	requestID proto.RequestID
	Request   *Request

	mu        sync.Mutex
	responded bool
}

// Continue allows the request to proceed normally.
func (r *Route) Continue() error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	r.mu.Unlock()

	return proto.FetchContinueRequest(r.requestID).Do(r.page.execCtx)
}

// ContinueWith allows the request to proceed with modified URL or headers.
func (r *Route) ContinueWith(url string, headers map[string]string) error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	r.mu.Unlock()

	params := proto.FetchContinueRequest(r.requestID)
	if url != "" {
		params = params.WithURL(url)
	}
	if len(headers) > 0 {
		var entries []proto.FetchHeaderEntry
		for k, v := range headers {
			entries = append(entries, proto.FetchHeaderEntry{Name: k, Value: v})
		}
		params = params.WithHeaders(entries)
	}
	return params.Do(r.page.execCtx)
}

// Fulfill responds to the request with the given status, headers, and body.
func (r *Route) Fulfill(
	status int,
	headers map[string]string,
	body string,
) error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	r.mu.Unlock()

	params := proto.FetchFulfillRequest(r.requestID, int64(status))
	if len(headers) > 0 {
		var entries []proto.FetchHeaderEntry
		for k, v := range headers {
			entries = append(entries, proto.FetchHeaderEntry{Name: k, Value: v})
		}
		params = params.WithResponseHeaders(entries)
	}
	if body != "" {
		params = params.WithBody(
			[]byte(base64.StdEncoding.EncodeToString([]byte(body))),
		)
	}
	return params.Do(r.page.execCtx)
}

// Abort blocks the request with a failure.
func (r *Route) Abort() error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	r.mu.Unlock()

	return proto.FetchFailRequest(
		r.requestID,
		proto.NetworkErrorReasonBlockedByClient,
	).Do(r.page.execCtx)
}
