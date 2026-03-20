package bonk

import (
	"encoding/base64"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

// Response represents an intercepted network response.
type Response struct {
	page      *Page
	requestID proto.RequestID
	URL       string
	Status    int64
	Headers   map[string]string

	mu        sync.Mutex
	responded bool
}

// Body returns the response body bytes.
func (r *Response) Body() ([]byte, error) {
	res, err := proto.FetchGetResponseBody(r.requestID).Do(r.page.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Base64Encoded {
		return base64.StdEncoding.DecodeString(res.Body)
	}
	return []byte(res.Body), nil
}

// Continue allows the response to proceed to the page.
func (r *Response) Continue() error {
	r.mu.Lock()
	if r.responded {
		r.mu.Unlock()
		return nil
	}
	r.responded = true
	r.mu.Unlock()

	return proto.FetchContinueResponse(r.requestID).Do(r.page.execCtx)
}

func (r *Response) autoRespond() {
	r.mu.Lock()
	responded := r.responded
	r.mu.Unlock()
	if !responded {
		r.Continue()
	}
}
