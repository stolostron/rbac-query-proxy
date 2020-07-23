# Build the multicluster-metrics-server-proxy binary
FROM golang:1.13.13 as builder

# Copy in the go src
WORKDIR /go/src/github.com/open-cluster-management/multicluster-metrics-server-proxy

COPY pkg/    pkg/
COPY main.go ./
COPY go.mod ./

RUN export GO111MODULE=on && go mod tidy

RUN export GO111MODULE=on \
    && CGO_ENABLED=0 GOOS=linux go build -a -o multicluster-metrics-server-proxy main.go \
    && strip multicluster-metrics-server-proxy


FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /
COPY --from=builder /go/src/github.com/open-cluster-management/multicluster-metrics-server-proxy/multicluster-metrics-server-proxy .

EXPOSE 3002

ENTRYPOINT ["/multicluster-metrics-server-proxy"]
