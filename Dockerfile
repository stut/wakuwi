# Build the UI (on the builder's native arch; output is arch-independent)
FROM --platform=$BUILDPLATFORM node:22-alpine AS ui
WORKDIR /src
COPY ui/package.json ui/package-lock.json ui/
RUN npm --prefix ui ci
COPY ui/ ui/
RUN npm --prefix ui run build

# Build the binary (UI is embedded via go:embed), cross-compiling for the
# target platform so nothing runs under emulation
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /src/ui/dist ui/dist
ARG VERSION=dev
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=${VERSION}" -o /wakuwi ./cmd/wakuwi

# Final image: single static binary, non-root
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /wakuwi /wakuwi
EXPOSE 9753
ENTRYPOINT ["/wakuwi"]
