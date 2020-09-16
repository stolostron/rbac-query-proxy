// Copyright (c) 2020 Red Hat, Inc.

package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	projectv1 "github.com/openshift/api/project/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	clusterclientset "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
)

const (
	projectsAPIPath       = "/apis/project.openshift.io/v1/projects"
	managedClusterAPIPath = "/apis/cluster.open-cluster-management.io/v1/managedclusters"
)

var allManagedClusterNames map[string]string

func GetAllManagedClusterNames() map[string]string {
	return allManagedClusterNames
}

func ModifyMetricsQueryParams(req *http.Request) {
	userName := req.Header.Get("user")
	klog.Infof("user is %v", userName)
	klog.Infof("URL is: %s", req.URL)
	klog.Infof("URL path is: %v", req.URL.Path)
	klog.Infof("URL RawQuery is: %v", req.URL.RawQuery)
	token := req.Header.Get("acm-access-token-cookie")
	if token == "" {
		klog.Errorf("failed to get token from http header")
	}

	projectList := FetchUserProjectList(token)
	if canAccessAllClusters(projectList) {
		klog.Infof("user <%v> have access to all clusters", userName)
		return
	}

	clusterList := fetchUserClusterList(token, projectList)
	klog.Infof("user <%v> have access to these clusters: %v", userName, clusterList)

	// TODO: use clusterList to modify query url

	return
}

func WatchManagedCluster() {
	allManagedClusterNames = map[string]string{}
	clusterClient, err := clusterclientset.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		klog.Fatalf("failed to new cluster clientset: %v", err)
	}

	watchlist := cache.NewListWatchFromClient(clusterClient.ClusterV1().RESTClient(), "managedclusters", v1.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&clusterv1.ManagedCluster{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				clusterName := obj.(*clusterv1.ManagedCluster).Name
				klog.Infof("added a managedcluster: %s \n", obj.(*clusterv1.ManagedCluster).Name)
				allManagedClusterNames[clusterName] = clusterName
			},

			DeleteFunc: func(obj interface{}) {
				clusterName := obj.(*clusterv1.ManagedCluster).Name
				klog.Infof("deleted a managedcluster: %s \n", obj.(*clusterv1.ManagedCluster).Name)
				delete(allManagedClusterNames, clusterName)
			},

			UpdateFunc: func(oldObj, newObj interface{}) {
				clusterName := newObj.(*clusterv1.ManagedCluster).Name
				klog.Infof("changed a managedcluster: %s \n", newObj.(*clusterv1.ManagedCluster).Name)
				allManagedClusterNames[clusterName] = clusterName
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second * 30)
		klog.Infof("found %v clusters", len(allManagedClusterNames))
	}
}

func sendHTTPRequest(url string, verb string, token string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr}
	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		klog.Errorf("failed to new http request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return client.Do(req)
}

func FetchUserProjectList(token string) []string {
	projectList := []string{}
	url := GetAPIHost() + projectsAPIPath
	resp, err := sendHTTPRequest(url, "GET", token)
	if err != nil {
		klog.Errorf("failed to send http request: %v", err)
		return projectList
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("failed to read response body: %v", err)
		return projectList
	}

	var projects projectv1.ProjectList
	requestBody := ioutil.NopCloser(bytes.NewBuffer(body))
	err = json.NewDecoder(requestBody).Decode(&projects)
	if err != nil {
		klog.Errorf("failed to decode response json body: %v, response body: %s", err, body)
		return projectList
	}

	for _, p := range projects.Items {
		projectList = append(projectList, p.Name)
	}

	return projectList
}

// Contains is used to check whether a list contains string s
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// canAccessAllClusters check user have permission to access all clusters
func canAccessAllClusters(projectList []string) bool {

	for name := range allManagedClusterNames {
		if !Contains(projectList, name) {
			return false
		}
	}

	return true
}

func fetchUserClusterList(token string, projectList []string) []string {
	clusterList := []string{}
	if len(projectList) == 0 {
		return clusterList
	}

	for _, projectName := range projectList {
		clusterName, ok := allManagedClusterNames[projectName]
		if ok {
			clusterList = append(clusterList, clusterName)
		}
	}

	return clusterList
}

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getKubeConfig(kubeConfig string) *rest.Config {
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
	return config
}

func GetAPIHost() string {
	config := getKubeConfig(os.Getenv("KUBECONFIG"))
	return config.Host
}

func NewDynamicClient(kubeConfig string) (dynamic.Interface, error) {
	config := getKubeConfig(kubeConfig)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}

func GetKubeClient(kubeConfig string) kubernetes.Interface {
	config := getKubeConfig(kubeConfig)
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to instantiate Kubernetes client: %v", err)
	}
	return kubeClient
}
