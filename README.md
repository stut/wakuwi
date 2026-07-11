<div align="center">
  <img src="logo.svg" width="80" height="80" alt="wakuwi logo" />
  <h1>wakuwi</h1>
</div>

A lightweight, read-only Kubernetes UI that runs locally and serves a web interface for exploring your clusters. Uses your existing kubeconfig — no additional configuration required. Also supports in-cluster deployment.

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

### Homebrew (macOS and Linux)

```sh
brew install stut/tools/wakuwi
```

### Windows

Download `wakuwi-windows-amd64.exe` from the [latest release](https://github.com/stut/wakuwi/releases/latest).

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

### Flags

```sh
wakuwi --port 9090        # custom port (default 9753)
wakuwi --show-secrets     # expose the Secret resource kind (hidden by default)
wakuwi --access-log       # log every HTTP request (off by default)
```

## Running in a cluster

wakuwi can run inside a Kubernetes cluster, authenticating with the pod's
service account instead of a kubeconfig. In-cluster mode is detected
automatically and:

- presents a single `in-cluster` context backed by the service account
- disables port-forward and local process management (they only make sense
  when wakuwi runs on your machine)

Deploy with the bundled manifest (ServiceAccount, read-only ClusterRole,
Deployment and ClusterIP Service):

```sh
kubectl apply -f deploy/wakuwi.yaml
kubectl -n wakuwi port-forward svc/wakuwi 9753:9753
```

Build the image with the bundled `Dockerfile`:

```sh
docker build -t wakuwi:latest .
```

> [!WARNING]
> wakuwi has **no built-in authentication**. The bundled manifest exposes it
> as a ClusterIP service only, reachable via `kubectl port-forward`. If you
> expose it through an Ingress, put an auth layer (e.g.
> [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/)) in front of
> it — anyone who can reach wakuwi can read everything its service account
> can.

### Secrets

Secrets are hidden by default, in both local and in-cluster modes. Two
independent switches control access in-cluster:

1. the `--show-secrets` flag (UI/API layer)
2. the `secrets` RBAC rule, commented out in `deploy/wakuwi.yaml` (the real
   security boundary)

Both must be enabled to browse secrets.

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
