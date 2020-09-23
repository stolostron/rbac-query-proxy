// Copyright (c) 2020 Red Hat, Inc.

package proxy

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestNewEmptyMatrixHTTPBody(t *testing.T) {
	body := newEmptyMatrixHTTPBody()
	bodyStr, _ := ioutil.ReadAll(body)

	gr, err := gzip.NewReader(bytes.NewBuffer([]byte(bodyStr)))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		log.Fatal(err)
	}

	var decompressedBuff bytes.Buffer
	gr, err = gzip.NewReader(bytes.NewBuffer([]byte(data)))
	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		t.Errorf("failed to ReadAll: %v", err)
	}

	decompressedBuff.Write(data)
	emptyMatrix := `{"status":"success","data":{"resultType":"matrix","result":[]}}`
	if decompressedBuff.String() != emptyMatrix {
		t.Errorf("(%v) is not the expected: (%v)", decompressedBuff.String(), emptyMatrix)
	}
}

func TestGzipWrite(t *testing.T) {
	originalStr := "test"
	var compressedBuff bytes.Buffer
	err := gzipWrite(&compressedBuff, []byte(originalStr))
	if err != nil {
		t.Errorf("failed to compressed: %v", err)
	}
	var decompressedBuff bytes.Buffer
	gr, err := gzip.NewReader(bytes.NewBuffer(compressedBuff.Bytes()))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		t.Errorf("failed to decompressed: %v", err)
	}
	decompressedBuff.Write(data)
	if decompressedBuff.String() != originalStr {
		t.Errorf("(%v) is not the expected: (%v)", originalStr, decompressedBuff.String())
	}
}

func TestProxyRequest(t *testing.T) {
	req := http.Request{}
	req.URL = &url.URL{}
	req.Header = http.Header(map[string][]string{})
	proxyRequest(&req)
	if req.Body != nil {
		t.Errorf("(%v) is not the expected nil", req.Body)
	}
	if req.Header.Get("Content-Type") != "" {
		t.Errorf("(%v) is not the expected: (\"\")", req.Header.Get("Content-Type"))
	}

	req.Method = http.MethodGet
	pathList := []string{
		"/api/v1/query",
		"/api/v1/query_range",
		"/api/v1/series",
	}

	for _, path := range pathList {
		req.URL.Path = path
		proxyRequest(&req)
		if req.Method != http.MethodPost {
			t.Errorf("(%v) is not the expected: (%v)", http.MethodPost, req.Method)
		}

		if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("(%v) is not the expected: (%v)", req.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
		}

		if req.Body == nil {
			t.Errorf("(%v) is not the expected non-nil", req.Body)
		}

		if req.URL.Scheme != "" {
			t.Errorf("(%v) is not the expected \"\"", req.URL.Scheme)
		}

		if req.URL.Host != "" {
			t.Errorf("(%v) is not the expected \"\"", req.URL.Host)
		}
	}
}
