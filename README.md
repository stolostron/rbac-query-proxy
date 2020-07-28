# rbac-query-proxy

The rbac-query-proxy is a small HTTP reverse proxy.

## Usage

```
$ git clone git@github.com:open-cluster-management/rbac-query-proxy.git
$ cd rbac-query-proxy
$ go build && ./rbac-query-proxy --metrics-server="http://localhost:9090/"
I0723 15:40:18.460797   21486 main.go:48] Proxy server will running on: :3002
I0723 15:40:18.460913   21486 main.go:49] Metrics server is: http://localhost:9090/
```