// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_ctxtags

import "net/http"

var (
	DefaultServiceName = "unspecified"

	defaultOptions = &options{
		tagExtractors:      []RequestTagExtractorFunc{},
		defaultServiceName: DefaultServiceName,
	}
)

type options struct {
	tagExtractors      []RequestTagExtractorFunc
	defaultServiceName string
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

// WithTagExtractor adds another request tag extractor, allowing you to customize what tags get prepopulated from the request.
func WithTagExtractor(f RequestTagExtractorFunc) Option {
	return func(o *options) {
		o.tagExtractors = append(o.tagExtractors, f)
	}
}

// WithServiceName is an option that allows you to track requests to different URL under the same service name.
//
// For client side requests, you can track external, and internal service names by using WithServiceName("github").
// For server side you can track logical groups of http.Handlers into a single service.
func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.defaultServiceName = serviceName
	}
}
