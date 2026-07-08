package kube

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// InClusterContextName is the synthetic context name used when running
// inside a Kubernetes cluster with no kubeconfig.
const InClusterContextName = "in-cluster"

// InCluster reports whether we appear to be running inside a Kubernetes
// pod (the service host env var is injected into every container).
func InCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}

// RESTConfig builds a rest.Config for the given context name. In-cluster
// it uses the pod's service account (token + CA, auto-refreshed by
// client-go); otherwise it loads the named context from the kubeconfig.
func RESTConfig(contextName string) (*rest.Config, error) {
	if InCluster() && contextName == InClusterContextName {
		return rest.InClusterConfig()
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	return cfg.ClientConfig()
}

func kubeClient(contextName string) (*kubernetes.Clientset, error) {
	restCfg, err := RESTConfig(contextName)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restCfg)
}
