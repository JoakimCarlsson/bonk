.PHONY: install fmt

install:
	@command -v goimports >/dev/null 2>&1 || go install golang.org/x/tools/cmd/goimports@latest
	@command -v golines >/dev/null 2>&1 || go install github.com/segmentio/golines@latest

fmt:
	$(shell go env GOPATH)/bin/goimports -w .
	$(shell go env GOPATH)/bin/golines -m 80 -w .