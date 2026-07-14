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

const (
	// searchPageSize bounds how many objects are held in memory per API
	// call; the search pages through the cluster rather than listing
	// everything at once.
	searchPageSize = 250
	// maxResultsPerKind caps the results returned per kind; matches
	// beyond the cap are only counted.
	maxResultsPerKind = 10
	// maxConcurrentKinds bounds how many kinds are listed at the same
	// time so peak memory is a few pages, not one page per kind.
	maxConcurrentKinds = 3
)

type SearchResult struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Status    string `json:"status,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	// More counts matches per kind beyond maxResultsPerKind.
	More map[string]int `json:"more,omitempty"`
}

func Search(ctx context.Context, contextName, query string, kinds []string) (*SearchResponse, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(query))

	var mu sync.Mutex
	results := []SearchResult{}
	more := map[string]int{}
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentKinds)

	for _, kind := range kinds {
		wg.Add(1)
		go func(kind string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			items, extra := searchKind(ctx, c, kind, q)
			mu.Lock()
			results = append(results, items...)
			if extra > 0 {
				more[kind] = extra
			}
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

	return &SearchResponse{Results: results, More: more}, nil
}

func match(name, q string) bool {
	return strings.Contains(strings.ToLower(name), q)
}

// searchKind pages through all objects of kind, returning the first
// maxResultsPerKind matches and a count of matches beyond that.
func searchKind(ctx context.Context, c *kubernetes.Clientset, kind, q string) ([]SearchResult, int) {
	var results []SearchResult
	more := 0
	opts := metav1.ListOptions{Limit: searchPageSize}
	for {
		matches, cont, err := searchKindPage(ctx, c, kind, q, opts)
		if err != nil {
			return results, more
		}
		for _, m := range matches {
			if len(results) < maxResultsPerKind {
				results = append(results, m)
			} else {
				more++
			}
		}
		if cont == "" {
			return results, more
		}
		opts.Continue = cont
	}
}

// searchKindPage lists a single page of objects of kind across all
// namespaces and returns the matches plus the continue token.
func searchKindPage(ctx context.Context, c *kubernetes.Clientset, kind, q string, opts metav1.ListOptions) ([]SearchResult, string, error) {
	var results []SearchResult
	switch kind {
	case "pods":
		list, err := c.CoreV1().Pods("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "deployments":
		list, err := c.AppsV1().Deployments("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "statefulsets":
		list, err := c.AppsV1().StatefulSets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "services":
		list, err := c.CoreV1().Services("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "ingresses":
		list, err := c.NetworkingV1().Ingresses("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "configmaps":
		list, err := c.CoreV1().ConfigMaps("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "secrets":
		list, err := c.CoreV1().Secrets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "daemonsets":
		list, err := c.AppsV1().DaemonSets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "jobs":
		list, err := c.BatchV1().Jobs("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil

	case "cronjobs":
		list, err := c.BatchV1().CronJobs("").List(ctx, opts)
		if err != nil {
			return nil, "", err
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
		return results, list.Continue, nil
	}

	return nil, "", nil
}
