package metric

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/open-cluster-management/multicluster-metrics-server-proxy/pkg/rbac"
	"github.com/open-cluster-management/multicluster-metrics-server-proxy/pkg/token"
)

func ModifyMetricsQueryParams(req *http.Request) {

	log.Printf("request headers:")
	for k, v := range req.Header {
		fmt.Printf("%v = %v\n", k, v)
	}

	userToken := req.Header.Get("Token")
	if len(userToken) == 0 {
		return
	}

	username := token.ParseUserNameFromToken(req)
	clusterList := rbac.GetUserClusterList(username)

	queryValues, err := url.ParseQuery(req.URL.RawQuery)
	if len(queryValues) == 0 || err != nil {
		return
	}

	originalQuery := queryValues.Get("query")
	metricName, isValidMetric := containsMetricName(originalQuery)
	// TODO: add existed cluster params
	if isValidMetric {
		// add clustername to query params for filter metrics, for example:
		// cluster:capacity_cpu_cores:sum -> cluster:capacity_cpu_cores:sum{cluster=~"cluster1|cluster2"}
		modifiedQuery := metricName + "{cluster=~\"" + strings.Join(clusterList, "|") + "\"}"
		modifiedQueryStr := strings.ReplaceAll(originalQuery, metricName, modifiedQuery)
		queryValues.Del("query")
		queryValues.Add("query", modifiedQueryStr)
	}

	req.URL.RawQuery = queryValues.Encode()
}

func containsMetricName(queryStr string) (string, bool) {
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

	for _, metricName := range metricNameList {
		if strings.Contains(queryStr, metricName) {
			return metricName, true
		}
	}

	return "", false
}
