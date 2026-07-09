package process

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/stut/wakuwi/internal/kube"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardParams struct {
	Context    string
	Namespace  string
	PodName    string
	LocalPort  int
	RemotePort int
}

func (m *Manager) StartPortForward(params PortForwardParams) (string, error) {
	name := fmt.Sprintf("%s %d→%d", params.PodName, params.LocalPort, params.RemotePort)

	p, logFile, err := m.register(
		KindPortForward, name,
		params.Context, params.Namespace, params.PodName,
		params.LocalPort, params.RemotePort,
	)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(m.ctx)
	m.mu.Lock()
	p.cancel = cancel
	m.mu.Unlock()

	go func() {
		defer logFile.Close()
		err := runPortForward(ctx, params, logFile)
		m.markDone(p, err)
		writeLog(logFile, "process ended")
	}()

	return p.ID, nil
}

func runPortForward(ctx context.Context, params PortForwardParams, logFile *os.File) error {
	restCfg, err := kube.RESTConfig(params.Context)
	if err != nil {
		writeLog(logFile, "error: "+err.Error())
		return err
	}

	serverURL, err := url.Parse(restCfg.Host)
	if err != nil {
		writeLog(logFile, "error: "+err.Error())
		return err
	}
	serverURL.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		params.Namespace, params.PodName)

	roundTripper, upgrader, err := spdy.RoundTripperFor(restCfg)
	if err != nil {
		writeLog(logFile, "error: "+err.Error())
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	ports := []string{fmt.Sprintf("%d:%d", params.LocalPort, params.RemotePort)}

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, logFile, logFile)
	if err != nil {
		writeLog(logFile, "error: "+err.Error())
		return err
	}

	go func() {
		<-ctx.Done()
		close(stopChan)
	}()

	go func() {
		<-readyChan
		writeLog(logFile, fmt.Sprintf("forwarding 127.0.0.1:%d -> %d", params.LocalPort, params.RemotePort))
	}()

	writeLog(logFile, fmt.Sprintf("starting port-forward %d -> %d", params.LocalPort, params.RemotePort))
	return fw.ForwardPorts()
}

func writeLog(f *os.File, msg string) {
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	fmt.Fprintf(f, "%s %s\n", ts, msg) //nolint:errcheck
}
