package bonk

import (
	"github.com/joakimcarlsson/bonk/proto"
)

// Route represents a pattern-matched intercepted request.
type Route struct {
	page      *Page
	requestID proto.RequestID
	Request   *Request
}

// Fulfill responds to the request with a custom response.
func (r *Route) Fulfill(
	status int,
	headers map[string]string,
	body string,
) error {
	return r.Request.Fulfill(status, headers, body)
}

// Continue allows the request to proceed.
func (r *Route) Continue() error {
	return r.Request.Continue()
}

// Abort blocks the request.
func (r *Route) Abort() error {
	return r.Request.Abort()
}

func (r *Route) autoRespond() {
	r.Request.autoRespond()
}
