VERSION := $(shell cat VERSION)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build ui dev-ui run clean

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
