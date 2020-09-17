// Copyright (c) 2020 Red Hat, Inc.

package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"time"

	"k8s.io/klog"
)

const (
	caPath   = "/var/rbac_proxy/ca"
	certPath = "/var/rbac_proxy/certs"
)

func getTLSTransport() (*http.Transport, error) {

	caCertFile := path.Join(caPath, "./ca.crt")
	tlsKeyFile := path.Join(certPath, "./tls.key")
	tlsCrtFile := path.Join(certPath, "./tls.crt")

	// Load Server CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		klog.Error("failed to load server ca cert file")
		return nil, err
	}
	// Load client cert signed by Client CA
	cert, err := tls.LoadX509KeyPair(tlsCrtFile, tlsKeyFile)
	if err != nil {
		klog.Error("failed to load client cert/key")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		TLSClientConfig:     tlsConfig,
	}, nil
}
