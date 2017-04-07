// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_opentracing

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
)

var (
	defaultOptions = &options{
		filterOutFunc:       nil,
		statusCodeErrorFunc: DefaultStatusCodeIsError,
		tracer:              nil,
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
	tracer              opentracing.Tracer
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	if optCopy.tracer == nil {
		optCopy.tracer = opentracing.GlobalTracer()
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

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

func DefaultStatusCodeIsError(statusCode int) bool {
	if statusCode < 400 {
		return false
	} else if statusCode == 404 { // not found shouldn't really be an error, too common.
		return false
	}
	return true
}
