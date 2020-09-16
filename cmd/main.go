// Copyright (c) 2020 Red Hat, Inc.

package main

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog"

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

	flagset.Parse(os.Args[1:])
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

	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(cfg.listenAddress, nil); err != nil {
		klog.Fatalf("failed to ListenAndServe: %v", err)
	}
}

func ErrorHandle(rw http.ResponseWriter, req *http.Request, err error) {
	token := req.Header.Get("acm-access-token-cookie")
	if token == "" {
		rw.WriteHeader(http.StatusUnauthorized)
	}
}

func ModifyResponse(r *http.Response) error {
	token := r.Request.Header.Get("acm-access-token-cookie")
	if token == "" {
		return errors.New("found unauthorized user")
	}

	projectList := util.FetchUserProjectList(token)
	if len(projectList) == 0 || len(util.GetAllManagedClusterNames()) == 0 {
		r.Body = http.NoBody
		return errors.New("no project found")
	}

	return nil
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(target)
	if err != nil {
		klog.Fatalf("failed to parse url: %v", err)
	}

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = ModifyResponse
	proxy.ErrorHandler = ErrorHandle
	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	util.ModifyMetricsQueryParams(req)
	proxy.ServeHTTP(res, req)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	serveReverseProxy(os.Getenv("METRICS_SERVER"), res, req)
}
