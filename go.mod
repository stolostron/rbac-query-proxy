module github.com/stolostron/rbac-query-proxy

go 1.14

require (
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/prom-label-proxy v0.0.0-20200605071327-9371ee4a9422
	github.com/prometheus/prometheus v1.8.2-0.20200507164740-ecee9c8abfd1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog v1.0.0
	open-cluster-management.io/api v0.5.0
	sigs.k8s.io/controller-runtime v0.6.3
)
