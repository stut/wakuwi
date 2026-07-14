package kube

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	// searchCacheTTL is how long a kind's reduced listing is served
	// from memory before the cluster is scanned again. Successive
	// keystrokes within the TTL filter in-memory instead of re-listing.
	searchCacheTTL = 60 * time.Second
	// searchCacheEvictAfter is how long an unrefreshed listing is kept
	// before being dropped entirely.
	searchCacheEvictAfter = 5 * time.Minute
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

// kindCacheEntry holds the reduced listing (SearchResult per object,
// ~100 bytes each — never full objects or secret/configmap payloads) of
// one kind in one context. mu serializes refreshes so concurrent cache
// misses share a single cluster scan.
type kindCacheEntry struct {
	mu            sync.Mutex
	entries       []SearchResult
	fetchedAtNano atomic.Int64
}

var searchCache = struct {
	sync.Mutex
	m map[string]*kindCacheEntry
}{m: map[string]*kindCacheEntry{}}

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
			items, extra := searchKind(ctx, c, contextName, kind, q)
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

// searchKind filters the (cached) reduced listing of kind, returning the
// first maxResultsPerKind matches and a count of matches beyond that.
func searchKind(ctx context.Context, c *kubernetes.Clientset, contextName, kind, q string) ([]SearchResult, int) {
	var results []SearchResult
	more := 0
	for _, e := range cachedKindEntries(ctx, c, contextName, kind) {
		if !match(e.Name, q) {
			continue
		}
		if len(results) < maxResultsPerKind {
			results = append(results, e)
		} else {
			more++
		}
	}
	return results, more
}

// cachedKindEntries returns the reduced listing of kind in contextName,
// scanning the cluster only when the cached copy is older than
// searchCacheTTL. On a failed refresh the stale copy is served.
func cachedKindEntries(ctx context.Context, c *kubernetes.Clientset, contextName, kind string) []SearchResult {
	key := contextName + "/" + kind

	searchCache.Lock()
	for k, e := range searchCache.m {
		if at := e.fetchedAtNano.Load(); at != 0 && time.Since(time.Unix(0, at)) > searchCacheEvictAfter {
			delete(searchCache.m, k)
		}
	}
	entry := searchCache.m[key]
	if entry == nil {
		entry = &kindCacheEntry{}
		searchCache.m[key] = entry
	}
	searchCache.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()
	if at := entry.fetchedAtNano.Load(); at != 0 && time.Since(time.Unix(0, at)) < searchCacheTTL {
		return entry.entries
	}
	entries, err := listKindEntries(ctx, c, kind)
	if err != nil {
		return entry.entries
	}
	entry.entries = entries
	entry.fetchedAtNano.Store(time.Now().UnixNano())
	return entries
}

// listKindEntries pages through all objects of kind across all
// namespaces, reducing each to a SearchResult.
func listKindEntries(ctx context.Context, c *kubernetes.Clientset, kind string) ([]SearchResult, error) {
	var entries []SearchResult
	opts := metav1.ListOptions{Limit: searchPageSize}
	for {
		page, cont, err := listKindPage(ctx, c, kind, opts)
		if err != nil {
			return nil, err
		}
		entries = append(entries, page...)
		if cont == "" {
			return entries, nil
		}
		opts.Continue = cont
	}
}

// listKindPage lists a single page of objects of kind across all
// namespaces, returning the reduced entries plus the continue token.
func listKindPage(ctx context.Context, c *kubernetes.Clientset, kind string, opts metav1.ListOptions) ([]SearchResult, string, error) {
	var entries []SearchResult
	switch kind {
	case "pods":
		list, err := c.CoreV1().Pods("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, p := range list.Items {
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
			entries = append(entries, SearchResult{
				Kind: kind, Name: p.Name, Namespace: p.Namespace,
				Age: podAge(p.CreationTimestamp.Time), Status: status,
			})
		}
		return entries, list.Continue, nil

	case "deployments":
		list, err := c.AppsV1().Deployments("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, d := range list.Items {
			var desired int32 = 1
			if d.Spec.Replicas != nil {
				desired = *d.Spec.Replicas
			}
			entries = append(entries, SearchResult{
				Kind: kind, Name: d.Name, Namespace: d.Namespace,
				Age:    podAge(d.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired),
			})
		}
		return entries, list.Continue, nil

	case "statefulsets":
		list, err := c.AppsV1().StatefulSets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, s := range list.Items {
			var desired int32 = 1
			if s.Spec.Replicas != nil {
				desired = *s.Spec.Replicas
			}
			entries = append(entries, SearchResult{
				Kind: kind, Name: s.Name, Namespace: s.Namespace,
				Age:    podAge(s.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", s.Status.ReadyReplicas, desired),
			})
		}
		return entries, list.Continue, nil

	case "services":
		list, err := c.CoreV1().Services("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, svc := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: svc.Name, Namespace: svc.Namespace,
				Age: podAge(svc.CreationTimestamp.Time), Status: string(svc.Spec.Type),
			})
		}
		return entries, list.Continue, nil

	case "ingresses":
		list, err := c.NetworkingV1().Ingresses("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, ing := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: ing.Name, Namespace: ing.Namespace,
				Age: podAge(ing.CreationTimestamp.Time),
			})
		}
		return entries, list.Continue, nil

	case "configmaps":
		list, err := c.CoreV1().ConfigMaps("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, cm := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: cm.Name, Namespace: cm.Namespace,
				Age: podAge(cm.CreationTimestamp.Time),
			})
		}
		return entries, list.Continue, nil

	case "secrets":
		list, err := c.CoreV1().Secrets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, s := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: s.Name, Namespace: s.Namespace,
				Age: podAge(s.CreationTimestamp.Time), Status: string(s.Type),
			})
		}
		return entries, list.Continue, nil

	case "daemonsets":
		list, err := c.AppsV1().DaemonSets("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, d := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: d.Name, Namespace: d.Namespace,
				Age:    podAge(d.CreationTimestamp.Time),
				Status: fmt.Sprintf("%d/%d", d.Status.NumberReady, d.Status.DesiredNumberScheduled),
			})
		}
		return entries, list.Continue, nil

	case "jobs":
		list, err := c.BatchV1().Jobs("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, j := range list.Items {
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
			entries = append(entries, SearchResult{
				Kind: kind, Name: j.Name, Namespace: j.Namespace,
				Age: podAge(j.CreationTimestamp.Time), Status: status,
			})
		}
		return entries, list.Continue, nil

	case "cronjobs":
		list, err := c.BatchV1().CronJobs("").List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		for _, cj := range list.Items {
			entries = append(entries, SearchResult{
				Kind: kind, Name: cj.Name, Namespace: cj.Namespace,
				Age: podAge(cj.CreationTimestamp.Time), Status: cj.Spec.Schedule,
			})
		}
		return entries, list.Continue, nil
	}

	return nil, "", nil
}
