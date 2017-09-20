// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

var (
	defaultOptions = &options{
		levelFunc:                 nil,
		levelForConnectivityError: logrus.WarnLevel,
		requestCaptureFunc:        func(r *http.Request) bool { return false },
		responseCaptureFunc:       func(r *http.Request, status int) bool { return false },
	}
)

type options struct {
	levelFunc                 CodeToLevel
	wrapLevelFromReqFunc      RequestToLevel
	levelForConnectivityError logrus.Level
	requestCaptureFunc        func(r *http.Request) bool
	responseCaptureFunc       func(r *http.Request, status int) bool
}

func evaluateTripperwareOpts(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultTripperwareCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateMiddlewareOpts(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultMiddlewareCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// CodeToLevel user functions define the mapping between HTTP status codes and logrus log levels.
type CodeToLevel func(httpStatusCode int) logrus.Level

// WithLevels customizes the function that maps HTTP client or server side status codes to log levels.
//
// By default `DefaultMiddlewareCodeToLevel` is used for server-side middleware, and `DefaultTripperwareCodeToLevel`
// is used for client-side tripperware.
//
// Note that this option can be overwritten by WithRequestLevels option if it is set as well.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// RequestToLevel user functions define the mapping between so far figured level and observed request details to logrus log levels.
type RequestToLevel func(level logrus.Level, req *http.Request) logrus.Level

// WithRequestLevels customizes the function that maps HTTP server side request to log levels.
//
// This is only supported by middleware.
//
// By default CodeToLevel option is used. If RequestToLevel function is set it will be invoked after CodeToLevel.
func WithRequestLevels(f RequestToLevel) Option {
	return func(o *options) {
		o.wrapLevelFromReqFunc = f
	}
}

// WithConnectivityErrorLevel customizes
func WithConnectivityErrorLevel(level logrus.Level) Option {
	return func(o *options) {
		o.levelForConnectivityError = level
	}
}

// WithRequestBodyCapture enables recording of request body pre-handling/pre-call.
//
// The body will be recorded as a separate log message. Body of `application/json` will be captured as
// http.request.body_json (in structured JSON form) and others will be captured as http.request.body_raw logrus field
// (raw base64-encoded value).
//
// For tripperware, only requests with Body of type `bytes.Buffer`, `strings.Reader`, `bytes.Buffer`, or with
// a specified `GetBody` function will be captured.
//
// For middleware, only requests with a set Content-Length will be captured, with no streaming or chunk encoding supported.
//
// This option creates a copy of the body per request, so please use with care.
func WithRequestBodyCapture(deciderFunc func(r *http.Request) bool) Option {
	return func(o *options) {
		o.requestCaptureFunc = deciderFunc
	}
}

// WithResponseBodyCapture enables recording of response body post-handling/post-call.
//
// The body will be recorded as a separate log message. Body of `application/json` will be captured as
// http.response.body_json (in structured JSON form) and others will be captured as http.response.body_raw logrus field
// (raw base64-encoded value).
//
// Only responses with Content-Length will be captured, with non-default Transfer-Encoding not being supported.
func WithResponseBodyCapture(deciderFunc func(r *http.Request, status int) bool) Option {
	return func(o *options) {
		o.responseCaptureFunc = deciderFunc
	}
}

// DefaultMiddlewareCodeToLevel is the default of a mapper between HTTP server-side status codes and logrus log levels.
func DefaultMiddlewareCodeToLevel(httpStatusCode int) logrus.Level {
	if httpStatusCode < 400 || httpStatusCode == http.StatusNotFound {
		return logrus.InfoLevel
	} else if httpStatusCode < 500 {
		return logrus.WarnLevel
	} else if httpStatusCode < 600 {
		return logrus.ErrorLevel
	} else {
		return logrus.ErrorLevel
	}
}

// DefaultTripperwareCodeToLevel is the default of a mapper between HTTP client-side status codes and logrus log levels.
func DefaultTripperwareCodeToLevel(httpStatusCode int) logrus.Level {
	if httpStatusCode < 400 {
		return logrus.DebugLevel
	} else if httpStatusCode < 500 {
		return logrus.InfoLevel
	} else {
		return logrus.WarnLevel
	}
}
