package kube

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
)

func StreamLogs(ctx context.Context, contextName, namespace, podName, container string) (io.ReadCloser, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	tail := int64(200)
	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &tail,
	}

	return c.CoreV1().Pods(namespace).GetLogs(podName, opts).Stream(ctx)
}
