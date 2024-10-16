> [!IMPORTANT]  
> This repository is archived, as the code has moved to the [multicluster-observability-operator](https://github.com/stolostron/multicluster-observability-operator/tree/main/proxy) repo.

# rbac-query-proxy

The rbac-query-proxy is a small HTTP reverse proxy, that can perform RBAC authorization against the server. Helper service that acts a multicluster metrics RBAC proxy.

## Prerequisites

- You must install [Open Cluster Management Observabilty](https://github.com/stolostron/multicluster-observability-operator)

## How to build image

```
$ docker build -f Dockerfile.prow -t rbac-query-proxy:latest .
```

Now, you can use this image to replace the rbac-query-proxy component and verify your PRs.

Rebuild Image: Thu May 19 15:01:17 EDT 2022
