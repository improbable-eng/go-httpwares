// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_retry

import (
	"net/http"

	"golang.org/x/net/context"
)

type ctxMarker struct{}

var (
	ctxEnableRetry = &ctxMarker{}
)

// Enable turns on the retry logic for a given request, regardless of what the retry decider says.
//
// Please make sure you do not pass around this request's context.
func Enable(req *http.Request) *http.Request {
	if isEnabled(req.Context()) {
		return req
	}
	return req.WithContext(EnableContext(req.Context()))
}

// Enable turns on the retry logic for a given request's context, regardless of what the retry decider says.
//
// Please make sure you do not pass around this request's context.
func EnableContext(ctx context.Context) context.Context {
	if isEnabled(ctx) {
		return ctx
	}
	return context.WithValue(ctx, ctxEnableRetry, true)
}

func isEnabled(ctx context.Context) bool {
	_, ok := ctx.Value(ctxEnableRetry).(bool)
	return ok
}
