package rbac

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	"github.com/open-cluster-management/multicluster-metrics-server-proxy/pkg/util"
)

const (
	managedClusterAPIPath = "/apis/cluster.open-cluster-management.io/v1/managedclusters"
)

var ManagedClusterGVR schema.GroupVersionResource = schema.GroupVersionResource{
	Group:    "cluster.open-cluster-management.io",
	Version:  "v1",
	Resource: "managedclusters",
}

func sendHTTPRequest(url string, verb string, token string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr}
	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		klog.Fatalf("failed to new http request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return client.Do(req)
}

func GetUserClusterList(token string) []string {
	url := util.GetAPIHost() + managedClusterAPIPath
	resp, err := sendHTTPRequest(url, "GET", token)
	if err != nil {
		klog.Fatalf("failed to send http request: %v", err)
	}

	var clusters clusterv1.ManagedClusterList
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("failed to read response body: %v", err)
	}
	requestBody := ioutil.NopCloser(bytes.NewBuffer(body))
	err = json.NewDecoder(requestBody).Decode(&clusters)
	if err != nil && resp.StatusCode != http.StatusForbidden {
		klog.Errorf("failed to decode response json body: %v", err)
	}

	clusterNames := []string{}
	for _, cluster := range clusters.Items {
		clusterNames = append(clusterNames, cluster.GetName())
	}

	if len(clusterNames) == 0 {
		allClusters := getAllManagedClusterNames()
		for _, name := range allClusters {
			url := util.GetAPIHost() + managedClusterAPIPath + "/" + name
			resp, err := sendHTTPRequest(url, "GET", token)
			if err != nil {
				klog.Fatalf("failed to send http request: %v", err)
			}

			var cluster clusterv1.ManagedCluster
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Errorf("failed to read response body: %v", err)
			}
			requestBody := ioutil.NopCloser(bytes.NewBuffer(body))
			err = json.NewDecoder(requestBody).Decode(&cluster)
			if err != nil && resp.StatusCode != http.StatusForbidden {
				klog.Errorf("failed to decode response json body: %v, response body: %s", err, body)
				continue
			}
			clusterNames = append(clusterNames, cluster.GetName())
		}
	}
	return clusterNames
}

func getAllManagedClusterNames() []string {
	dynamicClient, err := util.NewDynamicClient(os.Getenv("KUBECONFIG"))
	if err != nil {
		klog.Fatal("failed to NewDynamicClient: %v", err)
	}

	clusters, err := dynamicClient.Resource(ManagedClusterGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to get all managedclusters: %v", err)
	}
	clusterNames := []string{}
	for _, cluster := range clusters.Items {
		clusterNames = append(clusterNames, cluster.GetName())
	}

	klog.Infof("all cluster names: %v", clusterNames)
	return clusterNames
}
