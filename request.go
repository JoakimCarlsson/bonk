package bonk

import (
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

// Request represents an intercepted network request.
// Call Continue, Abort, or Fulfill to handle the request.
type Request struct {
	page         *Page
	requestID    proto.RequestID
	URL          string
	Method       string
	Headers      map[string]string
	PostData     string
	ResourceType string

	mu         sync.Mutex
	modHeaders map[string]string
	responded  bool
}

// SetHeader sets or overrides a request header.
// The modified header is applied when Continue is called.
func (r *Request) SetHeader(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.modHeaders == nil {
		r.modHeaders = make(map[string]string)
	}
	r.modHeaders[key] = value
}

// Continue allows the request to proceed, applying any modified headers.
func (r *Request) Continue() error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	headers := r.mergedHeaders()
	r.mu.Unlock()

	params := proto.FetchContinueRequest(r.requestID)
	if len(headers) > 0 {
		var entries []proto.FetchHeaderEntry
		for k, v := range headers {
			entries = append(entries, proto.FetchHeaderEntry{Name: k, Value: v})
		}
		params = params.WithHeaders(entries)
	}
	return params.Do(r.page.execCtx)
}

// Abort blocks the request.
func (r *Request) Abort() error {
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

// Fulfill responds to the request with a custom response.
func (r *Request) Fulfill(
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
		params = params.WithBody([]byte(body))
	}
	return params.Do(r.page.execCtx)
}

func (r *Request) autoRespond() {
	r.mu.Lock()
	responded := r.responded
	r.mu.Unlock()
	if !responded {
		r.Continue()
	}
}

func (r *Request) mergedHeaders() map[string]string {
	if len(r.modHeaders) == 0 {
		return nil
	}
	merged := make(map[string]string)
	for k, v := range r.Headers {
		merged[k] = v
	}
	for k, v := range r.modHeaders {
		merged[k] = v
	}
	return merged
}
