package kube

import (
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
