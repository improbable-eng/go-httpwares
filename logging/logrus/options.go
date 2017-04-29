// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

var (
	defaultOptions = &options{
		levelFunc:                 nil,
		levelForConnectivityError: logrus.WarnLevel,
	}
)

type options struct {
	levelFunc                 CodeToLevel
	levelForConnectivityError logrus.Level
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateTripperwareOpts(opts []Option) *options {
	o := evaluateOptions(opts)
	if o.levelFunc == nil {
		o.levelFunc = DefaultTripperwareCodeToLevel
	}
	return o
}

func evaluateMiddlewareOpts(opts []Option) *options {
	o := evaluateOptions(opts)
	if o.levelFunc == nil {
		o.levelFunc = DefaultMiddlewareCodeToLevel
	}
	return o
}

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(httpStatusCode int) logrus.Level

// WithLevels customizes the function that maps HTTP client or server side status codes to log levels.
//
// By default `DefaultMiddlewareCodeToLevel` is used for server-side middleware, and `DefaultTripperwareCodeToLevel`
// is used for client-side tripperware.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithConnectivityErrorLevel customizes
func WithConnectivityErrorLevel(level logrus.Level) Option {
	return func(o *options) {
		o.levelForConnectivityError = level
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
