package brotli

import (
	"net/http"
	"strings"
)

// Request filter conditions
type RequestFilter interface {
	ShouldCompress(req *http.Request) bool
}

// interface verification
var (
	_ RequestFilter = &CommonRequestFilter{}
	_ RequestFilter = &RequestApiFilter{}
)

// CommonRequestFilter judge via common easy criteria like
// http method, accept-encoding header, etc.
type CommonRequestFilter struct{}

// NewCommonRequestFilter ...
func NewCommonRequestFilter() *CommonRequestFilter {
	return &CommonRequestFilter{}
}

// ShouldCompress implements RequestFilter interface
func (c *CommonRequestFilter) ShouldCompress(req *http.Request) bool {
	return req.Method != http.MethodHead &&
		req.Method != http.MethodOptions &&
		req.Header.Get("Upgrade") == "" &&
		strings.Contains(req.Header.Get("Accept-Encoding"), "br")
}

type RequestApiFilter struct {
	path []string
}

func NewRequestApiFilter(path []string) *RequestApiFilter {
	return &RequestApiFilter{path: path}
}

func (r *RequestApiFilter) ShouldCompress(req *http.Request) bool {
	if len(r.path) == 0 {
		return true
	}
	curPath := req.RequestURI
	for _, item := range r.path {
		if item == curPath {
			return true
		}
	}
	return false
}
