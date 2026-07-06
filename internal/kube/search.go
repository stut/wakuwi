package kube

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SearchResult struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Status    string `json:"status,omitempty"`
}

func Search(ctx context.Context, contextName, query string, kinds []string) ([]SearchResult, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(query))

	var mu sync.Mutex
	var results []SearchResult
	var wg sync.WaitGroup

	for _, kind := range kinds {
		wg.Add(1)
		go func(kind string) {
			defer wg.Done()
			items := searchKind(ctx, c, kind, q)
			mu.Lock()
			results = append(results, items...)
			mu.Unlock()
		}(kind)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Kind != results[j].Kind {
			return results[i].Kind < results[j].Kind
		}
		if results[i].Namespace != results[j].Namespace {
			return results[i].Namespace < results[j].Namespace
		}
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func match(name, q string) bool {
	return strings.Contains(strings.ToLower(name), q)
}

func searchKind(ctx context.Context, c *kubernetes.Clientset, kind, q string) []SearchResult {
	var results []SearchResult
	switch kind {
	case "pods":
		list, err := c.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, p := range list.Items {
			if !match(p.Name, q) {
				continue
			}
			status := string(p.Status.Phase)
			if p.DeletionTimestamp != nil {
				status = "Terminating"
			} else {
				for _, cs := range p.Status.ContainerStatuses {
					if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
						status = cs.State.Waiting.Reason
						break
					}
				}
			}
			results = append(results, SearchResult{
				Kind: kind, Name: p.Name, Namespace: p.Namespace,
				Age: podAge(p.CreationTimestamp.Time), Status: status,
			})
		}

	case "deployments":
		list, err := c.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, d := range list.Items {
			if !match(d.Name, q) {
				continue
			}
			var desired int32 = 1
			if d.Spec.Replicas != nil {
				desired = *d.Spec.Replicas
			}
			results = append(results, SearchResult{
				Kind: kind, Name: d.Name, Namespace: d.Namespace,
				Age:    podAge(d.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired),
			})
		}

	case "statefulsets":
		list, err := c.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, s := range list.Items {
			if !match(s.Name, q) {
				continue
			}
			var desired int32 = 1
			if s.Spec.Replicas != nil {
				desired = *s.Spec.Replicas
			}
			results = append(results, SearchResult{
				Kind: kind, Name: s.Name, Namespace: s.Namespace,
				Age:    podAge(s.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", s.Status.ReadyReplicas, desired),
			})
		}

	case "services":
		list, err := c.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, svc := range list.Items {
			if !match(svc.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: svc.Name, Namespace: svc.Namespace,
				Age: podAge(svc.CreationTimestamp.Time), Status: string(svc.Spec.Type),
			})
		}

	case "ingresses":
		list, err := c.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, ing := range list.Items {
			if !match(ing.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: ing.Name, Namespace: ing.Namespace,
				Age: podAge(ing.CreationTimestamp.Time),
			})
		}

	case "configmaps":
		list, err := c.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, cm := range list.Items {
			if !match(cm.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: cm.Name, Namespace: cm.Namespace,
				Age: podAge(cm.CreationTimestamp.Time),
			})
		}

	case "secrets":
		list, err := c.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, s := range list.Items {
			if !match(s.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: s.Name, Namespace: s.Namespace,
				Age: podAge(s.CreationTimestamp.Time), Status: string(s.Type),
			})
		}

	case "daemonsets":
		list, err := c.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, d := range list.Items {
			if !match(d.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: d.Name, Namespace: d.Namespace,
				Age:    podAge(d.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", d.Status.NumberReady, d.Status.DesiredNumberScheduled),
			})
		}

	case "jobs":
		list, err := c.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, j := range list.Items {
			if !match(j.Name, q) {
				continue
			}
			status := "Running"
			for _, cond := range j.Status.Conditions {
				if string(cond.Type) == "Complete" && string(cond.Status) == "True" {
					status = "Complete"
					break
				}
				if string(cond.Type) == "Failed" && string(cond.Status) == "True" {
					status = "Failed"
					break
				}
			}
			results = append(results, SearchResult{
				Kind: kind, Name: j.Name, Namespace: j.Namespace,
				Age: podAge(j.CreationTimestamp.Time), Status: status,
			})
		}

	case "cronjobs":
		list, err := c.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}
		for _, cj := range list.Items {
			if !match(cj.Name, q) {
				continue
			}
			results = append(results, SearchResult{
				Kind: kind, Name: cj.Name, Namespace: cj.Namespace,
				Age: podAge(cj.CreationTimestamp.Time), Status: cj.Spec.Schedule,
			})
		}
	}

	return results
}
