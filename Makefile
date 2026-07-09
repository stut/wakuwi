VERSION := $(shell cat VERSION)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build ui dev-ui run clean ci

build: ui
	go build -ldflags "$(LDFLAGS)" -o wakuwi ./cmd/wakuwi

ui:
	npm --prefix ui install
	npm --prefix ui run build

run: ui
	go run -ldflags "$(LDFLAGS)" ./cmd/wakuwi

dev-ui:
	npm --prefix ui run dev

clean:
	rm -f wakuwi
	rm -rf ui/dist ui/node_modules

# Run every check the CI workflow runs, continuing past failures and
# reporting them all at the end.
ci:
	@failed=""; \
	echo "==> ui build"; \
	(npm --prefix ui install && npm --prefix ui run build) || failed="$$failed ui-build"; \
	echo "==> gofmt"; \
	unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-formatted:"; echo "$$unformatted"; failed="$$failed gofmt"; \
	fi; \
	echo "==> go vet"; \
	go vet ./... || failed="$$failed vet"; \
	echo "==> golangci-lint"; \
	if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./... || failed="$$failed lint"; \
	else \
		echo "golangci-lint not installed (https://golangci-lint.run/docs/welcome/install/)"; \
		failed="$$failed lint"; \
	fi; \
	echo "==> go build"; \
	go build ./... || failed="$$failed build"; \
	echo "==> go test"; \
	go test ./... || failed="$$failed test"; \
	echo; \
	if [ -n "$$failed" ]; then echo "CI FAILED:$$failed"; exit 1; fi; \
	echo "CI passed"
