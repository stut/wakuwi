package kube

import (
	"context"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Namespaces(ctx context.Context, contextName string) ([]string, error) {
	client, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}
	sort.Strings(names)
	return names, nil
}
