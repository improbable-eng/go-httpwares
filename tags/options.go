// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package http_ctxtags

import (
	"net"
	"net/http"
	"strings"
)

var (
	defaultOptions = &options{
		tagExtractors:           []RequestTagExtractorFunc{},
		serviceName:             "",
		serviceNameDetectorFunc: DefaultServiceNameDetector,
	}
)

const (
	DefaultServiceName = "unspecified"
)

type options struct {
	tagExtractors           []RequestTagExtractorFunc
	serviceName             string
	serviceNameDetectorFunc serviceNameDetectorFunc
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// RequestTagExtractorFunc is a signature of user-customizeable functions for extracting tags from requests.
type RequestTagExtractorFunc func(req *http.Request) map[string]interface{}

type serviceNameDetectorFunc func(req *http.Request) string

// WithTagExtractor adds another request tag extractor, allowing you to customize what tags get prepopulated from the request.
func WithTagExtractor(f RequestTagExtractorFunc) Option {
	return func(o *options) {
		o.tagExtractors = append(o.tagExtractors, f)
	}
}

// WithServiceName is an option for client-side wares that explicitly states the name of the service called.
//
// This option takes precedence over the WithServiceNameDetector values.
//
// For example WithServiceName("github").
func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

// WithServiceNameDetector allows you to customize the function for automatically detecting the service name from URLs.
//
// By default it uses the `DefaultServiceNameDetector`.
func WithServiceNameDetector(fn serviceNameDetectorFunc) Option {
	return func(o *options) {
		o.serviceNameDetectorFunc = fn
	}
}

// DefaultServiceNameDetector is the default detector of services from URLs.
func DefaultServiceNameDetector(req *http.Request) string {
	host := req.URL.Host
	if strings.Contains(host, ":") {
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		} else {
			return DefaultServiceName
		}
	}
	suffix := matchedWellKnown(host)
	if suffix == "" {
		return DefaultServiceName
	}
	parts := strings.Split(host[:len(host)-len(suffix)], ".")
	if len(parts) == 0 {
		return DefaultServiceName
	}
	return parts[len(parts)-1]
}

func matchedWellKnown(hostAddr string) string {
	for _, suffix := range []string{".com", ".net", ".org", ".io"} {
		if strings.HasSuffix(hostAddr, suffix) {
			return suffix
		}
	}
	return ""
}
