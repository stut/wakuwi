package process

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusError   Status = "error"
)

type Kind string

const (
	KindPortForward Kind = "portforward"
)

type Process struct {
	ID         string `json:"id"`
	Kind       Kind   `json:"kind"`
	Name       string `json:"name"`
	Context    string `json:"context"`
	Namespace  string `json:"namespace"`
	Resource   string `json:"resource"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
	Status     Status `json:"status"`
	StartedAt  string `json:"startedAt"`
	StoppedAt  string `json:"stoppedAt,omitempty"`
	LogFile    string `json:"logFile"`
	cancel     context.CancelFunc
}

type Manager struct {
	mu        sync.RWMutex
	processes map[string]*Process
	ctx       context.Context
	logDir    string
}

func NewManager(ctx context.Context) (*Manager, error) {
	logDir := filepath.Join(os.TempDir(), "wakuwi")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	return &Manager{
		processes: make(map[string]*Process),
		ctx:       ctx,
		logDir:    logDir,
	}, nil
}

func (m *Manager) newID() string {
	b := make([]byte, 4)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

func (m *Manager) LogPath(id string) string {
	return filepath.Join(m.logDir, id+".log")
}

func (m *Manager) List() []*Process {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Process, 0, len(m.processes))
	for _, p := range m.processes {
		out = append(out, p)
	}
	return out
}

func (m *Manager) Get(id string) (*Process, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.processes[id]
	return p, ok
}

func (m *Manager) DismissAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, p := range m.processes {
		if p.Status != StatusRunning {
			os.Remove(p.LogFile) //nolint:errcheck
			delete(m.processes, id)
		}
	}
}

func (m *Manager) Kill(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.processes[id]
	if !ok {
		return fmt.Errorf("process %s not found", id)
	}
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (m *Manager) Dismiss(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.processes[id]
	if !ok {
		return fmt.Errorf("process %s not found", id)
	}
	if p.Status == StatusRunning {
		return fmt.Errorf("process %s is still running", id)
	}
	os.Remove(p.LogFile) //nolint:errcheck
	delete(m.processes, id)
	return nil
}

func (m *Manager) register(kind Kind, name, contextName, namespace, resource string, local, remote int) (*Process, *os.File, error) {
	id := m.newID()
	logPath := m.LogPath(id)
	f, err := os.Create(logPath)
	if err != nil {
		return nil, nil, fmt.Errorf("create log file: %w", err)
	}

	ctx, cancel := context.WithCancel(m.ctx)
	p := &Process{
		ID:         id,
		Kind:       kind,
		Name:       name,
		Context:    contextName,
		Namespace:  namespace,
		Resource:   resource,
		LocalPort:  local,
		RemotePort: remote,
		Status:     StatusRunning,
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
		LogFile:    logPath,
		cancel:     cancel,
	}

	m.mu.Lock()
	m.processes[id] = p
	m.mu.Unlock()

	// Watch parent context
	go func() {
		<-ctx.Done()
	}()

	return p, f, nil
}

func (m *Manager) markDone(p *Process, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.StoppedAt = time.Now().UTC().Format(time.RFC3339)
	if err != nil && err != context.Canceled {
		p.Status = StatusError
	} else {
		p.Status = StatusStopped
	}
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}
