.PHONY: build ui dev-ui run clean

build: ui
	go build -o wakuwi ./cmd/wakuwi

ui:
	npm --prefix ui install
	npm --prefix ui run build

run: ui
	go run ./cmd/wakuwi

dev-ui:
	npm --prefix ui run dev

clean:
	rm -f wakuwi
	rm -rf ui/dist ui/node_modules
