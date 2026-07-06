<div align="center">
  <img src="logo.svg" width="80" height="80" alt="wakuwi logo" />
  <h1>wakuwi</h1>
</div>

A lightweight, read-only Kubernetes UI that runs locally and serves a web interface for exploring your clusters. Uses your existing kubeconfig — no additional configuration required.

## Features

- Browse all kubeconfig contexts and namespaces
- List and inspect pods and other Kubernetes resources
- View pod logs in real time
- View resource manifests and descriptions
- Port-forward to pods
- Full-text search across resources
- Detect and surface common resource issues
- Manage and inspect locally-running processes

## Installation

### Homebrew (macOS)

```sh
brew install stut/tools/wakuwi
```

### Build from source

Requirements: Go 1.22+, Node.js

```sh
make build        # builds UI then compiles the binary to ./wakuwi
```

## Run

```sh
wakuwi                  # listens on :8080
wakuwi --port 9090      # custom port
```

Open [http://localhost:9753](http://localhost:9753) in your browser.

The UI is embedded in the binary — no separate assets needed at runtime.

## Development

Run the Go server and the Vite dev server separately for hot-reload:

```sh
# terminal 1
go run ./cmd/wakuwi

# terminal 2
make dev-ui
```

## Why Another KUbernetes Web Interface?

[Stut](https://stut.me/) was bored. [Claude](https://claude.ai) was also there.

## License

[MIT](LICENSE)
