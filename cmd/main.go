// Copyright (c) 2020 Red Hat, Inc.

package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog"

	"github.com/open-cluster-management/rbac-query-proxy/pkg/proxy"
	"github.com/open-cluster-management/rbac-query-proxy/pkg/util"
)

const (
	defaultListenAddress = "0.0.0.0:3002"
)

type config struct {
	listenAddress      string
	metricServer       string
	kubeconfigLocation string
}

func main() {

	cfg := config{}

	klogFlags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(klogFlags)
	flagset := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	flagset.AddGoFlagSet(klogFlags)

	flagset.StringVar(&cfg.listenAddress, "listen-address", defaultListenAddress, "The address HTTP server should listen on.")
	flagset.StringVar(&cfg.metricServer, "metrics-server", "", "The address the metrics server should run on.")

	_ = flagset.Parse(os.Args[1:])
	if err := os.Setenv("METRICS_SERVER", cfg.metricServer); err != nil {
		klog.Fatalf("failed to Setenv: %v", err)
	}

	//Kubeconfig flag
	flagset.StringVar(&cfg.kubeconfigLocation, "kubeconfig", "", "Path to a kubeconfig file, specifying how to connect to the API server. If unset, in-cluster configuration will be used")

	klog.Infof("proxy server will running on: %s", cfg.listenAddress)
	klog.Infof("metrics server is: %s", cfg.metricServer)
	klog.Infof("kubeconfig is: %s", cfg.kubeconfigLocation)

	// watch all managed clusters
	go util.WatchManagedCluster()

	http.HandleFunc("/", proxy.HandleRequestAndRedirect)
	if err := http.ListenAndServe(cfg.listenAddress, nil); err != nil {
		klog.Fatalf("failed to ListenAndServe: %v", err)
	}
}
