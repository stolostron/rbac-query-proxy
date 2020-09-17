// Copyright (c) 2020 Red Hat, Inc.

package proxy

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/open-cluster-management/rbac-query-proxy/pkg/util"
	"k8s.io/klog"
)

var (
	serverScheme = ""
	serverHost   = ""
)

// HandleRequestAndRedirect is used to init proxy handler
func HandleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	serverURL, err := url.Parse(os.Getenv("METRICS_SERVER"))
	if err != nil {
		klog.Errorf("failed to parse url: %v", err)
	}
	serverHost = serverURL.Host
	serverScheme = serverURL.Scheme

	// create the reverse proxy
	tlsTransport, err := getTLSTransport()
	if err != nil {
		klog.Fatalf("failed to create tls transport: %v", err)
	}

	proxy := httputil.ReverseProxy{
		Director:  proxyRequest,
		Transport: tlsTransport,
	}

	//proxy.ModifyResponse = modifyResponse
	//proxy.ErrorHandler = errorHandle
	//req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	//req.Host = reqUrl.Host

	//util.ModifyMetricsQueryParams(req)
	proxy.ServeHTTP(res, req)
}

func errorHandle(rw http.ResponseWriter, req *http.Request, err error) {
	token := req.Header.Get("acm-access-token-cookie")
	if token == "" {
		rw.WriteHeader(http.StatusUnauthorized)
	}
}

func modifyResponse(r *http.Response) error {
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

func proxyRequest(r *http.Request) {

	r.URL.Scheme = serverScheme
	r.URL.Host = serverHost

	if r.Method == http.MethodGet {
		r.Method = http.MethodPost
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Body = ioutil.NopCloser(strings.NewReader(r.URL.Query().Get("query")))
	}
}
