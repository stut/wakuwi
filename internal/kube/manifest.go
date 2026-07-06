package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

var gvrMap = map[string]schema.GroupVersionResource{
	"pods":                   {Group: "", Version: "v1", Resource: "pods"},
	"deployments":            {Group: "apps", Version: "v1", Resource: "deployments"},
	"statefulsets":           {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"daemonsets":             {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"jobs":                   {Group: "batch", Version: "v1", Resource: "jobs"},
	"cronjobs":               {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"services":               {Group: "", Version: "v1", Resource: "services"},
	"ingresses":              {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"configmaps":             {Group: "", Version: "v1", Resource: "configmaps"},
	"secrets":                {Group: "", Version: "v1", Resource: "secrets"},
	"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
}

func GetManifest(ctx context.Context, contextName, namespace, kind, name string) ([]byte, error) {
	gvr, ok := gvrMap[kind]
	if !ok {
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	dc, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	obj, err := dc.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// strip managed fields — too noisy
	obj.SetManagedFields(nil)

	j, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(j)
}
