package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stut/wakuwi/internal/kube"
	"github.com/stut/wakuwi/internal/process"
)

// Options controls which features the server exposes.
type Options struct {
	// InCluster disables features that make no sense inside a pod
	// (port-forward, local process management).
	InCluster bool
	// ShowSecrets exposes the Secret resource kind through the API. Off by
	// default; RBAC is the real security boundary — this is belt-and-braces.
	ShowSecrets bool
}

type Server struct {
	mux     *http.ServeMux
	static  fs.FS
	pm      *process.Manager
	version string
	opts    Options
}

func New(files fs.FS, pm *process.Manager, version string, opts Options) *Server {
	static, err := fs.Sub(files, "ui/dist")
	if err != nil {
		panic(err)
	}

	s := &Server{
		mux:     http.NewServeMux(),
		static:  static,
		pm:      pm,
		version: version,
		opts:    opts,
	}
	s.routes()
	return s
}

// processesEnabled reports whether the local process manager (port-forward
// et al) is available: disabled in-cluster and when no manager was supplied.
func (s *Server) processesEnabled() bool {
	return !s.opts.InCluster && s.pm != nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := &statusRecorder{ResponseWriter: w, status: 200}
	start := time.Now()
	s.mux.ServeHTTP(rec, r)
	log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/version", s.handleVersion)
	s.mux.HandleFunc("/api/contexts", s.handleContexts)
	s.mux.HandleFunc("/api/namespaces", s.handleNamespaces)
	s.mux.HandleFunc("/api/pods", s.handlePodList)
	s.mux.HandleFunc("/api/pods/", s.handlePodDetail)
	s.mux.HandleFunc("/api/resources", s.handleResourceList)
	s.mux.HandleFunc("/api/resources/", s.handleResourceDetail)
	s.mux.HandleFunc("/api/manifest", s.handleManifest)
	s.mux.HandleFunc("/api/issues", s.handleIssues)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/livereload", s.handleLivereload)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/processes", s.handleProcesses)
	s.mux.HandleFunc("/api/processes/", s.handleProcess)
	s.mux.Handle("/", s.spaHandler())
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok")) //nolint:errcheck
}

// Capabilities tells the UI which features to show.
type Capabilities struct {
	InCluster bool `json:"inCluster"`
	Processes bool `json:"processes"` // port-forward + process manager
	Secrets   bool `json:"secrets"`
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct { //nolint:errcheck
		Version      string       `json:"version"`
		Capabilities Capabilities `json:"capabilities"`
	}{
		Version: s.version,
		Capabilities: Capabilities{
			InCluster: s.opts.InCluster,
			Processes: s.processesEnabled(),
			Secrets:   s.opts.ShowSecrets,
		},
	})
}

func (s *Server) handleContexts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contexts, err := kube.Contexts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reason lets the UI surface an informational notice (not an error) when
	// there is nothing to select: either no kubeconfig exists at all, or a
	// kubeconfig is present but defines no contexts.
	resp := struct {
		Contexts []kube.Context `json:"contexts"`
		Reason   string         `json:"reason,omitempty"`
	}{Contexts: contexts}
	if len(contexts) == 0 {
		if kube.HasKubeconfig() {
			resp.Reason = "no-contexts"
		} else {
			resp.Reason = "no-kubeconfig"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contextName := r.URL.Query().Get("context")
	if contextName == "" {
		http.Error(w, "context query param required", http.StatusBadRequest)
		return
	}

	namespaces, err := kube.Namespaces(r.Context(), contextName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(namespaces) //nolint:errcheck
}

func (s *Server) handlePodList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	if contextName == "" || namespace == "" {
		http.Error(w, "context and namespace query params required", http.StatusBadRequest)
		return
	}
	pods, err := kube.ListPods(r.Context(), contextName, namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pods) //nolint:errcheck
}

func (s *Server) handlePodDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/pods/")
	if name == "" {
		http.Error(w, "pod name required", http.StatusBadRequest)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	if contextName == "" || namespace == "" {
		http.Error(w, "context and namespace query params required", http.StatusBadRequest)
		return
	}
	pod, err := kube.GetPod(r.Context(), contextName, namespace, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pod) //nolint:errcheck
}

func (s *Server) handleResourceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	if contextName == "" || namespace == "" || kind == "" {
		http.Error(w, "context, namespace and kind query params required", http.StatusBadRequest)
		return
	}
	if s.kindHidden(kind) {
		http.Error(w, "secrets are disabled; start wakuwi with --show-secrets", http.StatusForbidden)
		return
	}
	resources, err := kube.ListResources(r.Context(), contextName, namespace, kind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources) //nolint:errcheck
}

func (s *Server) handleResourceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/resources/")
	if name == "" {
		http.Error(w, "resource name required", http.StatusBadRequest)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	if contextName == "" || namespace == "" || kind == "" {
		http.Error(w, "context, namespace and kind query params required", http.StatusBadRequest)
		return
	}
	if s.kindHidden(kind) {
		http.Error(w, "secrets are disabled; start wakuwi with --show-secrets", http.StatusForbidden)
		return
	}
	detail, err := kube.GetResource(r.Context(), contextName, namespace, kind, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail) //nolint:errcheck
}

func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	if contextName == "" || namespace == "" || kind == "" || name == "" {
		http.Error(w, "context, namespace, kind and name required", http.StatusBadRequest)
		return
	}
	if s.kindHidden(kind) {
		http.Error(w, "secrets are disabled; start wakuwi with --show-secrets", http.StatusForbidden)
		return
	}
	data, err := kube.GetManifest(r.Context(), contextName, namespace, kind, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s-%s.yaml"`, kind, name))
	w.Write(data) //nolint:errcheck
}

func (s *Server) handleIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	if contextName == "" {
		http.Error(w, "context required", http.StatusBadRequest)
		return
	}
	issues, err := kube.ListIssues(r.Context(), contextName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues) //nolint:errcheck
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	q := r.URL.Query().Get("q")
	kindsParam := r.URL.Query().Get("kinds")
	if contextName == "" || q == "" || kindsParam == "" {
		http.Error(w, "context, q and kinds required", http.StatusBadRequest)
		return
	}
	kinds := make([]string, 0)
	for _, k := range strings.Split(kindsParam, ",") {
		if !s.kindHidden(k) {
			kinds = append(kinds, k)
		}
	}
	if len(kinds) == 0 {
		http.Error(w, "no permitted kinds requested", http.StatusForbidden)
		return
	}
	results, err := kube.Search(r.Context(), contextName, q, kinds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results) //nolint:errcheck
}

func (s *Server) handleLivereload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	contextName := r.URL.Query().Get("context")
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	if contextName == "" || namespace == "" || pod == "" {
		http.Error(w, "context, namespace and pod required", http.StatusBadRequest)
		return
	}

	stream, err := kube.StreamLogs(r.Context(), contextName, namespace, pod, container)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		select {
		case <-r.Context().Done():
			return
		default:
			fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
			flusher.Flush()
		}
	}
}

// kindHidden reports whether a resource kind is filtered from the API.
// Only secrets are gated, behind the --show-secrets flag.
func (s *Server) kindHidden(kind string) bool {
	return kind == "secrets" && !s.opts.ShowSecrets
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
	if !s.processesEnabled() {
		http.Error(w, "process management is disabled in-cluster", http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list := s.pm.List()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list) //nolint:errcheck

	case http.MethodDelete:
		s.pm.DismissAll()
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPost:
		var body struct {
			Context    string `json:"context"`
			Namespace  string `json:"namespace"`
			PodName    string `json:"podName"`
			LocalPort  int    `json:"localPort"`
			RemotePort int    `json:"remotePort"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := s.pm.StartPortForward(process.PortForwardParams{
			Context:    body.Context,
			Namespace:  body.Namespace,
			PodName:    body.PodName,
			LocalPort:  body.LocalPort,
			RemotePort: body.RemotePort,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": id}) //nolint:errcheck

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProcess(w http.ResponseWriter, r *http.Request) {
	if !s.processesEnabled() {
		http.Error(w, "process management is disabled in-cluster", http.StatusNotFound)
		return
	}
	// Path: /api/processes/:id or /api/processes/:id/logs or /api/processes/:id/dismiss
	rest := strings.TrimPrefix(r.URL.Path, "/api/processes/")
	parts := strings.SplitN(rest, "/", 2)
	id := parts[0]
	sub := ""
	if len(parts) == 2 {
		sub = parts[1]
	}

	if id == "" {
		http.Error(w, "process id required", http.StatusBadRequest)
		return
	}

	switch {
	case sub == "logs" && r.Method == http.MethodGet:
		s.handleProcessLogs(w, r, id)

	case sub == "" && r.Method == http.MethodDelete:
		if err := s.pm.Kill(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case sub == "dismiss" && r.Method == http.MethodDelete:
		if err := s.pm.Dismiss(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (s *Server) handleProcessLogs(w http.ResponseWriter, r *http.Request, id string) {
	p, ok := s.pm.Get(id)
	if !ok {
		http.Error(w, "process not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	f, err := os.Open(p.LogFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	buf := make([]byte, 4096)
	var partial []byte

	sendLine := func(line string) {
		fmt.Fprintf(w, "data: %s\n\n", line)
		flusher.Flush()
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			for {
				n, err := f.Read(buf)
				if n > 0 {
					chunk := append(partial, buf[:n]...)
					partial = nil
					for {
						nl := bytes.IndexByte(chunk, '\n')
						if nl < 0 {
							partial = chunk
							break
						}
						sendLine(string(chunk[:nl]))
						chunk = chunk[nl+1:]
					}
				}
				if err != nil {
					break
				}
			}
			// Stop streaming when process is done and we've sent all output
			proc, exists := s.pm.Get(id)
			if !exists || proc.Status != "running" {
				// Send any remaining partial line
				if len(partial) > 0 {
					sendLine(string(partial))
				}
				return
			}
		}
	}
}


func (s *Server) spaHandler() http.Handler {
	fileServer := http.FileServer(http.FS(s.static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := s.static.Open(path); err != nil {
			// Only fall back to index.html for routes (no extension or .html).
			// Asset requests (.js, .css, etc.) get a real 404 so stale binaries fail visibly.
			if ext := filepath.Ext(path); ext != "" && ext != ".html" {
				http.NotFound(w, r)
				return
			}
			r.URL.Path = "/"
		}
		// embed.FS has zero modtime so http.FileServer generates a stale ETag.
		// Force revalidation for HTML; assets are safe to cache (vite content-hashes their names).
		if path == "index.html" || strings.HasSuffix(path, ".html") {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Clear-Site-Data", `"storage"`)
		}
		fileServer.ServeHTTP(w, r)
	})
}
