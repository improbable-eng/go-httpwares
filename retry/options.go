// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_retry

import (
	"net/http"
	"time"
)

var (
	defaultOptions = &options{
		decider:     DefaultRetriableDecider,
		maxRetry:    3,
		backoffFunc: BackoffLinear(100 * time.Millisecond),
	}
)

type options struct {
	decider     RetriableDeciderFunc
	maxRetry    uint
	backoffFunc BackoffFunc
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

// RetriableDeciderFunc decides whether the given function is idempotent and safe or to retry.
type RetriableDeciderFunc func(req *http.Request) bool

// BackoffFunc denotes a family of functions that controll the backoff duration between call retries.
//
// They are called with an identifier of the attempt, and should return a time the system client should
// hold off for. If the time returned is longer than the `context.Context.Deadline` of the request
// the deadline of the request takes precedence and the wait will be interrupted before proceeding
// with the next iteration.
type BackoffFunc func(attempt uint) time.Duration

// WithMax sets the maximum number of retries on this call, or this interceptor.
func WithMax(maxRetries uint) Option {
	return func(o *options) {
		o.maxRetry = maxRetries
	}
}

// WithBackoff sets the `BackoffFunc `used to control time between retries.
func WithBackoff(bf BackoffFunc) Option {
	return func(o *options) {
		o.backoffFunc = bf
	}
}

// WithLevels customizes the function that maps HTTP client or server side status codes to log levels.
//
// By default `DefaultMiddlewareCodeToLevel` is used for server-side middleware, and `DefaultTripperwareCodeToLevel`
// is used for client-side tripperware.
func WithDecider(f RetriableDeciderFunc) Option {
	return func(o *options) {
		o.decider = f
	}
}

// It is fairly conservative and heeds the of http://restcookbook.com/HTTP%20Methods/idempotency.
func DefaultRetriableDecider(req *http.Request) bool {
	if req.Method == "GET" || req.Method == "OPTION" || req.Method == "HEAD" {
		return true
	}
	return false
}
