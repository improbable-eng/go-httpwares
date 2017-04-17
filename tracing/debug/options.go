// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_debug

import "net/http"

var (
	defaultOptions = &options{
		filterOutFunc:       nil,
		statusCodeErrorFunc: DefaultStatusCodeIsError,
	}
)

// FilterFunc allows users to provide a function that filters out certain methods from being traced.
//
// If it returns false, the given request will not be traced.
type FilterFunc func(req *http.Request) bool

// StatusCodeIsError allows the customization of which requests are considered errors in the tracing system.
type StatusCodeIsError func(statusCode int) bool

type options struct {
	filterOutFunc       FilterFunc
	statusCodeErrorFunc StatusCodeIsError
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

// WithFilterFunc customizes the function used for deciding whether a given call is traced or not.
func WithFilterFunc(f FilterFunc) Option {
	return func(o *options) {
		o.filterOutFunc = f
	}
}

// WithStatusCodeIsError customizes the function used for deciding whether a given call was an error
func WithStatusCodeIsError(f StatusCodeIsError) Option {
	return func(o *options) {
		o.statusCodeErrorFunc = f
	}
}

// DefaultStatusCodeIsError defines a function that says whether a given request is an error based on a code.
func DefaultStatusCodeIsError(statusCode int) bool {
	if statusCode < 500 {
		return false
	}
	return true
}
