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
		discarder:   DefaultResponseDiscarder,
		maxRetry:    3,
		backoffFunc: BackoffLinear(100 * time.Millisecond),
	}
)

type options struct {
	decider     RequestRetryDeciderFunc
	discarder   ResponseDiscarderFunc
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

// RequestRetryDeciderFunc decides whether the given function is idempotent and safe or to retry.
type RequestRetryDeciderFunc func(req *http.Request) bool

// ResponseDiscarderFunc decides when to discard a response and retry the request again (on true).
type ResponseDiscarderFunc func(resp *http.Response) bool

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

// WithDecider is a function that allows users to customize the logic that decides whether a request is retriable.
func WithDecider(f RequestRetryDeciderFunc) Option {
	return func(o *options) {
		o.decider = f
	}
}

// WithResponseDiscarder is a function that decides whether a given response should be discarded and another request attempted.
func WithResponseDiscarder(f RequestRetryDeciderFunc) Option {
	return func(o *options) {
		o.decider = f
	}
}

// DefaultRetriableDecider is the default implementation that retries only indempotent and safe requests (GET, OPTION, HEAD).
//
// It is fairly conservative and heeds the of http://restcookbook.com/HTTP%20Methods/idempotency.
func DefaultRetriableDecider(req *http.Request) bool {
	if req.Method == "GET" || req.Method == "OPTION" || req.Method == "HEAD" {
		return true
	}
	return false
}

// DefaultResponseDiscarder is the default implementation that discards responses in order to try again.
//
// It is fairly conservative and rejects (and thus retries) responses with 500, 503 and 504 status codes.
// See https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#5xx_Server_error
func DefaultResponseDiscarder(resp *http.Response) bool {
	return resp.StatusCode == 500 || resp.StatusCode == 503 || resp.StatusCode == 504
}
