package util

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func GetKubeClient(kubeConfig string) kubernetes.Interface {
	var (
		err    error
		config *rest.Config
	)

	if kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Fatal("unable to build rest config based on provided path to kubeconfig file")
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatal("cannot find Service Account in pod to build in-cluster rest config")
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to instantiate Kubernetes client: %v", err)
	}
	return kubeClient
}
