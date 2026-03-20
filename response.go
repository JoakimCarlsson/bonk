package bonk

import (
	"encoding/base64"

	"github.com/joakimcarlsson/bonk/proto"
)

// Response represents a network response.
type Response struct {
	page      *Page
	requestID proto.RequestID
	URL       string
	Status    int64
	Headers   map[string]string
}

// Body returns the response body bytes.
func (r *Response) Body() ([]byte, error) {
	res, err := proto.NetworkGetResponseBody(r.requestID).Do(r.page.execCtx)
	if err != nil {
		return nil, err
	}
	if res.Base64Encoded {
		return base64.StdEncoding.DecodeString(string(res.Body))
	}
	return []byte(res.Body), nil
}
