package metric

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	"github.com/open-cluster-management/rbac-query-proxy/pkg/util"
)

const (
	managedClusterAPIPath = "/apis/cluster.open-cluster-management.io/v1/managedclusters"
	hubClusterName        = "hub_cluster"
)

var ManagedClusterGVR schema.GroupVersionResource = schema.GroupVersionResource{
	Group:    "cluster.open-cluster-management.io",
	Version:  "v1",
	Resource: "managedclusters",
}

func ModifyMetricsQueryParams(req *http.Request) {
	token := req.Header.Get("X-Forwarded-Access-Token")
	if token == "" {
		klog.Fatalf("failed to get token from http header")
	}

	managedClusterURL := util.GetAPIHost() + managedClusterAPIPath
	resp, err := sendHTTPRequest(managedClusterURL, "GET", token)
	if err != nil {
		klog.Fatalf("failed to send http request: %v", err)
	}

	// user is admin and have permission to access to /v1/managedclusters endpoint
	if resp.StatusCode == http.StatusOK {
		return
	}

	clusterList := GetUserClusterList(token)
	if len(clusterList) == 0 {
		klog.Fatalf("user have not permission to access cluster metrics")
	}

	klog.Infof("resp.StatusCode is: %v", resp.StatusCode)
	klog.Infof("len(clusterList) is: %v", len(clusterList))

	queryValues := req.URL.Query()
	if len(queryValues) == 0 {
		return
	}

	klog.Infof("URL is: %s", req.URL)
	klog.Infof("URL path is: %v", req.URL.Path)
	klog.Infof("URL RawQuery is: %v", req.URL.RawQuery)

	klog.Info("original URL is:")
	for k, v := range queryValues {
		fmt.Printf("%v = %v\n", k, v)
	}

	originalQuery := queryValues.Get("query")
	if len(originalQuery) == 0 {
		return
	}

	klog.Infof("user is %v", req.Header.Get("X-Forwarded-User"))
	modifiedQuery := updateQueryParams(originalQuery, clusterList)
	queryValues.Del("query")
	queryValues.Add("query", modifiedQuery)
	req.URL.RawQuery = queryValues.Encode()

	// just for testing
	queryValues = req.URL.Query()
	klog.Info("modified URL is:")
	for k, v := range queryValues {
		fmt.Printf("%v = %v\n", k, v)
	}
}

func updateQueryParams(originalQuery string, clusterList []string) string {
	klog.Infof("originalQuery: %s", originalQuery)
	klog.Infof("clusterList: %s", clusterList)

	if len(clusterList) == 0 {
		return originalQuery
	}
	// should get these metrics from multicluster-monitoring-operator
	metricNameList := []string{
		":node_memory_MemAvailable_bytes:sum",
		"cluster:capacity_cpu_cores:sum",
		"cluster:capacity_memory_bytes:sum",
		"cluster:container_cpu_usage:ratio",
		"cluster:container_spec_cpu_shares:ratio",
		"cluster:cpu_usage_cores:sum",
		"cluster:memory_usage:ratio",
		"cluster:memory_usage_bytes:sum",
		"cluster:usage:resources:sum",
		"cluster_infrastructure_provider",
		"cluster_version",
		"cluster_version_payload",
		"container_cpu_cfs_throttled_periods_total",
		"container_memory_cache",
		"container_memory_rss",
		"container_memory_swap",
		"container_memory_working_set_bytes",
		"container_network_receive_bytes_total",
		"container_network_receive_packets_dropped_total",
		"container_network_receive_packets_total",
		"container_network_transmit_bytes_total",
		"container_network_transmit_packets_dropped_total",
		"container_network_transmit_packets_total",
		"haproxy_backend_connections_total",
		"instance:node_cpu_utilisation:rate1m",
		"instance:node_load1_per_cpu:ratio",
		"instance:node_memory_utilisation:ratio",
		"instance:node_network_receive_bytes_excluding_lo:rate1m",
		"instance:node_network_receive_drop_excluding_lo:rate1m",
		"instance:node_network_transmit_bytes_excluding_lo:rate1m",
		"instance:node_network_transmit_drop_excluding_lo:rate1m",
		"instance:node_num_cpu:sum",
		"instance:node_vmstat_pgmajfault:rate1m",
		"instance_device:node_disk_io_time_seconds:rate1m",
		"instance_device:node_disk_io_time_weighted_seconds:rate1m",
		"kube_node_status_allocatable_cpu_cores",
		"kube_node_status_allocatable_memory_bytes",
		"kube_pod_container_resource_limits_cpu_cores",
		"kube_pod_container_resource_limits_memory_bytes",
		"kube_pod_container_resource_requests_cpu_cores",
		"kube_pod_container_resource_requests_memory_bytes",
		"kube_pod_info",
		"kube_resourcequota",
		"machine_cpu_cores",
		"machine_memory_bytes",
		"mixin_pod_workload",
		"node_cpu_seconds_total",
		"node_filesystem_avail_bytes",
		"node_filesystem_size_bytes",
		"node_memory_MemAvailable_bytes",
		"node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate",
		"node_namespace_pod_container:container_memory_cache",
		"node_namespace_pod_container:container_memory_rss",
		"node_namespace_pod_container:container_memory_swap",
		"node_namespace_pod_container:container_memory_working_set_bytes",
		"node_netstat_Tcp_OutSegs",
		"node_netstat_Tcp_RetransSegs",
		"node_netstat_TcpExt_TCPSynRetrans",
		"up",
	}

	modifiedQueryStr := originalQuery
	for _, metricName := range metricNameList {
		// match metrics_name or metrics_name{key=value}
		reg := `\b` + metricName + `\b\s*{*`
		result := regexp.MustCompile(reg).FindString(originalQuery)
		if len(result) == 0 {
			continue
		}

		if !strings.Contains(result, `{`) {
			queryParam := metricName + "{cluster=~\"" + strings.Join(clusterList, "|") + "\"}"
			modifiedQueryStr = strings.ReplaceAll(originalQuery, metricName, queryParam)
		} else {
			index := strings.Index(originalQuery, "{")
			queryParam := "{cluster=~\"" + strings.Join(clusterList, "|") + "\","
			modifiedQueryStr = originalQuery[:index] + queryParam + originalQuery[index+1:]
		}
	}

	return modifiedQueryStr
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
	clusterNames := []string{}
	allClusters := getAllManagedClusterNames()
	for _, name := range allClusters {
		url := util.GetAPIHost() + managedClusterAPIPath + "/" + name
		resp, err := sendHTTPRequest(url, "GET", token)
		if err != nil {
			klog.Errorf("failed to send http request: %v", err)
			continue
		}

		var cluster clusterv1.ManagedCluster
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Errorf("failed to read response body: %v", err)
			continue
		}

		requestBody := ioutil.NopCloser(bytes.NewBuffer(body))
		err = json.NewDecoder(requestBody).Decode(&cluster)
		if err != nil && resp.StatusCode != http.StatusForbidden {
			klog.Errorf("failed to decode response json body: %v, response body: %s", err, body)
			continue
		}

		clusterNames = append(clusterNames, cluster.GetName())
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
