package bonk

import (
	"encoding/json"
	"sync"

	"github.com/joakimcarlsson/bonk/proto"
)

type fetchManager struct {
	page            *Page
	mu              sync.Mutex
	requestHandler  func(*Request)
	responseHandler func(*Response)
	routes          []registeredRoute
	unsub           func()
	enabled         bool
}

type registeredRoute struct {
	pattern string
	handler func(*Route)
}

func newFetchManager(page *Page) *fetchManager {
	return &fetchManager{page: page}
}

func (fm *fetchManager) setRequestHandler(fn func(*Request)) func() {
	fm.mu.Lock()
	fm.requestHandler = fn
	fm.mu.Unlock()
	fm.sync()

	return func() {
		fm.mu.Lock()
		fm.requestHandler = nil
		fm.mu.Unlock()
		fm.sync()
	}
}

func (fm *fetchManager) setResponseHandler(fn func(*Response)) func() {
	fm.mu.Lock()
	fm.responseHandler = fn
	fm.mu.Unlock()
	fm.sync()

	return func() {
		fm.mu.Lock()
		fm.responseHandler = nil
		fm.mu.Unlock()
		fm.sync()
	}
}

func (fm *fetchManager) addRoute(pattern string, handler func(*Route)) func() {
	fm.mu.Lock()
	fm.routes = append(fm.routes, registeredRoute{pattern, handler})
	fm.mu.Unlock()
	fm.sync()

	return func() {
		fm.mu.Lock()
		for i, r := range fm.routes {
			if r.pattern == pattern {
				fm.routes = append(fm.routes[:i], fm.routes[i+1:]...)
				break
			}
		}
		fm.mu.Unlock()
		fm.sync()
	}
}

func (fm *fetchManager) sync() {
	fm.mu.Lock()
	hasRequest := fm.requestHandler != nil
	hasResponse := fm.responseHandler != nil
	hasRoutes := len(fm.routes) > 0
	needsFetch := hasRequest || hasResponse || hasRoutes
	fm.mu.Unlock()

	if !needsFetch {
		if fm.unsub != nil {
			fm.unsub()
			fm.unsub = nil
		}
		if fm.enabled {
			proto.FetchDisable().Do(fm.page.execCtx)
			fm.enabled = false
		}
		return
	}

	var patterns []proto.FetchRequestPattern

	if hasRequest || hasRoutes {
		patterns = append(patterns, proto.FetchRequestPattern{
			RequestStage: proto.FetchRequestStageRequest,
		})
	}
	if hasResponse {
		patterns = append(patterns, proto.FetchRequestPattern{
			RequestStage: proto.FetchRequestStageResponse,
		})
	}

	fm.mu.Lock()
	for _, r := range fm.routes {
		patterns = append(patterns, proto.FetchRequestPattern{
			URLPattern:   r.pattern,
			RequestStage: proto.FetchRequestStageRequest,
		})
	}
	fm.mu.Unlock()

	if fm.unsub != nil {
		fm.unsub()
	}

	fm.unsub = fm.page.session.Subscribe(
		proto.FetchEventRequestPausedMethod,
		fm.handleEvent,
	)

	proto.FetchEnable().WithPatterns(patterns).Do(fm.page.execCtx)
	fm.enabled = true
}

func (fm *fetchManager) handleEvent(params json.RawMessage) {
	var ev proto.FetchEventRequestPaused
	if err := json.Unmarshal(params, &ev); err != nil {
		return
	}

	headers := extractHeaders(params, "request", "headers")
	isResponseStage := ev.ResponseStatusCode != 0

	if isResponseStage {
		fm.mu.Lock()
		handler := fm.responseHandler
		fm.mu.Unlock()

		if handler != nil {
			resp := &Response{
				page:      fm.page,
				requestID: ev.RequestID,
				URL:       ev.Request.URL,
				Status:    int64(ev.ResponseStatusCode),
				Headers:   extractResponseHeaders(ev.ResponseHeaders),
			}
			handler(resp)
			resp.autoRespond()
		} else {
			proto.FetchContinueResponse(ev.RequestID).Do(fm.page.execCtx)
		}
		return
	}

	fm.mu.Lock()
	var matchedRoute *registeredRoute
	for i := range fm.routes {
		if matchPattern(fm.routes[i].pattern, ev.Request.URL) {
			matchedRoute = &fm.routes[i]
			break
		}
	}
	requestHandler := fm.requestHandler
	fm.mu.Unlock()

	if matchedRoute != nil {
		req := &Request{
			page:         fm.page,
			requestID:    ev.RequestID,
			URL:          ev.Request.URL,
			Method:       ev.Request.Method,
			Headers:      headers,
			PostData:     ev.Request.PostData,
			ResourceType: string(ev.ResourceType),
		}
		route := &Route{
			page:      fm.page,
			requestID: ev.RequestID,
			Request:   req,
		}
		matchedRoute.handler(route)
		route.autoRespond()
		return
	}

	if requestHandler != nil {
		req := &Request{
			page:         fm.page,
			requestID:    ev.RequestID,
			URL:          ev.Request.URL,
			Method:       ev.Request.Method,
			Headers:      headers,
			PostData:     ev.Request.PostData,
			ResourceType: string(ev.ResourceType),
		}
		requestHandler(req)
		req.autoRespond()
		return
	}

	proto.FetchContinueRequest(ev.RequestID).Do(fm.page.execCtx)
}

// OnRequest registers a handler for intercepted requests.
// The handler can observe, modify headers, abort, or fulfill the request.
// If the handler doesn't call Continue, Abort, or Fulfill, the request
// is automatically continued.
func (p *Page) OnRequest(fn func(*Request)) func() {
	return p.fetch.setRequestHandler(fn)
}

// OnResponse registers a handler for intercepted responses.
// The handler can read the response body before it reaches the page.
// If the handler doesn't call Continue, the response is automatically continued.
func (p *Page) OnResponse(fn func(*Response)) func() {
	return p.fetch.setResponseHandler(fn)
}

// Route intercepts requests matching the URL pattern.
// The handler receives a Route that can be fulfilled, continued, or aborted.
func (p *Page) Route(pattern string, handler func(*Route)) func() {
	return p.fetch.addRoute(pattern, handler)
}

// SetExtraHTTPHeaders sets headers that will be sent with every request.
func (p *Page) SetExtraHTTPHeaders(headers map[string]string) error {
	params := map[string]any{"headers": headers}
	return proto.Execute(
		p.execCtx,
		"Network.setExtraHTTPHeaders",
		params,
		nil,
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

func extractResponseHeaders(
	entries []proto.FetchHeaderEntry,
) map[string]string {
	if len(entries) == 0 {
		return nil
	}
	headers := make(map[string]string, len(entries))
	for _, e := range entries {
		headers[e.Name] = e.Value
	}
	return headers
}

func matchPattern(pattern, url string) bool {
	return globMatch(pattern, url)
}
