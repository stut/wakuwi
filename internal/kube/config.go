package kube

import (
	"os"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

type Context struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
	Current   bool   `json:"current"`
}

func Contexts() ([]Context, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	raw, err := rules.Load()
	if err != nil {
		return nil, err
	}

	contexts := make([]Context, 0, len(raw.Contexts))
	for name, ctx := range raw.Contexts {
		c := Context{
			Name:      name,
			Cluster:   ctx.Cluster,
			Namespace: ctx.Namespace,
			Current:   name == raw.CurrentContext,
		}
		if cluster, ok := raw.Clusters[ctx.Cluster]; ok {
			c.Server = cluster.Server
		}
		contexts = append(contexts, c)
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts, nil
}

// HasKubeconfig reports whether any kubeconfig file was found on disk using the
// same default precedence (KUBECONFIG env, ~/.kube/config) that Contexts uses.
// This lets callers distinguish "no kubeconfig at all" from "kubeconfig present
// but defines no contexts".
func HasKubeconfig() bool {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	paths := rules.Precedence
	if rules.ExplicitPath != "" {
		paths = append([]string{rules.ExplicitPath}, paths...)
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
