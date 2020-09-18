// Copyright (c) 2020 Red Hat, Inc.

package proxy

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
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
		r.Body = newEmptyMatrixHTTPBody()
		return errors.New("no project or cluster found")
	}

	return nil
}

func newEmptyMatrixHTTPBody() io.ReadCloser {
	var bodyBuff bytes.Buffer
	gz := gzip.NewWriter(&bodyBuff)
	if _, err := gz.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`)); err != nil {
		klog.Errorf("failed to write body: %v", err)
	}

	if err := gz.Close(); err != nil {
		klog.Errorf("failed to close gzip writer: %v", err)
	}

	var gzipBuff bytes.Buffer
	err := gzipWrite(&gzipBuff, bodyBuff.Bytes())
	if err != nil {
		klog.Errorf("failed to write with gizp: %v", err)
	}

	return ioutil.NopCloser(bytes.NewBufferString(gzipBuff.String()))
}

func gzipWrite(w io.Writer, data []byte) error {
	gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	defer gw.Close()
	gw.Write(data)
	return err
}

func proxyRequest(r *http.Request) {

	r.URL.Scheme = serverScheme
	r.URL.Host = serverHost

	if r.Method == http.MethodGet {
		if strings.HasSuffix(r.URL.Path, "/api/v1/query") ||
			strings.HasSuffix(r.URL.Path, "/api/v1/query_range") ||
			strings.HasSuffix(r.URL.Path, "/api/v1/series") {
			r.Method = http.MethodPost
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Body = ioutil.NopCloser(strings.NewReader(r.URL.RawQuery))
		}
	}
}
