package rclone

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func getK8sClient() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}
	config, err := loadKubeConfig()
	if err != nil {
		return nil, err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func loadKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	if err != rest.ErrNotInCluster {
		return nil, err
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetCSIDriverNamespace returns the namespace where the CSI driver is running
func GetCSIDriverNamespace() (string, error) {
	if ns := os.Getenv("CSI_NAMESPACE"); ns != "" {
		return ns, nil
	}

	return "default", nil
}
