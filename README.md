# multicluster-metrics-server-proxy

The kube-rbac-proxy is a small HTTP reverse proxy for muticluster metrics server.

## Usage

```
$ git clone git@github.com:open-cluster-management/multicluster-metrics-server-proxy.git
$ cd multicluster-metrics-server-proxy
$ go build && ./multicluster-metrics-server-proxy --metrics-server="http://localhost:9090/"
I0723 15:40:18.460797   21486 main.go:48] Proxy server will running on: :3002
I0723 15:40:18.460913   21486 main.go:49] Metrics server is: http://localhost:9090/
```