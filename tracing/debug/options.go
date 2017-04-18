// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_debug

import "net/http"

var (
	defaultOptions = &options{
		filterFunc:          nil,
		statusCodeErrorFunc: DefaultIsStatusCodeAnError,
	}
)

// FilterFunc allows users to provide a function that filters out certain methods from being traced.
//
// If it returns false, the given request will not be traced.
type FilterFunc func(req *http.Request) bool

// IsStatusCodeAnErrorFunc allows the customization of which requests are considered errors in the tracing system.
type IsStatusCodeAnErrorFunc func(statusCode int) bool

type options struct {
	filterFunc          FilterFunc
	statusCodeErrorFunc IsStatusCodeAnErrorFunc
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
		o.filterFunc = f
	}
}

// WithIsStatusCodeAnError customizes the function used for deciding whether a given call was an error
func WithIsStatusCodeAnError(f IsStatusCodeAnErrorFunc) Option {
	return func(o *options) {
		o.statusCodeErrorFunc = f
	}
}

// DefaultIsStatusCodeAnError defines a function that says whether a given request is an error based on a code.
func DefaultIsStatusCodeAnError(statusCode int) bool {
	return statusCode >= 500
}
