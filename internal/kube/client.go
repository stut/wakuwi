package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func kubeClient(contextName string) (*kubernetes.Clientset, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName},
	)
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restCfg)
}
