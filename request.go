package bonk

// Request represents an intercepted network request.
type Request struct {
	URL      string
	Method   string
	Headers  map[string]string
	PostData string
}
