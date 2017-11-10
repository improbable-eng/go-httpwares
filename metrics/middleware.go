// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"time"

	"github.com/improbable-eng/go-httpwares"
)

// Middleware returns a http.Handler middleware that exports request metrics.
// If the tags middleware is used, this should be placed after tags to pick up metadata.
// This middleware assumes HTTP/1.x-style requests/response behaviour. It will not work with servers that use
// hijacking, pushing, or other similar features.
func Middleware(reporter Reporter) httpwares.Middleware {
	return func(next http.Handler) http.Handler {
		if reporter == nil {
			return next
		}
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			tracker := reporter.Track(req)
			start := time.Now()
			tracker.RequestStarted()
			req.Body = wrapBody(req.Body, func(size int) {
				tracker.RequestRead(time.Since(start), size)
			})
			wrapped := wrapWriter(resp, func(status int) {
				tracker.ResponseStarted(time.Since(start), status, resp.Header())
			})
			next.ServeHTTP(wrapped, req)
			tracker.ResponseDone(time.Since(start), wrapped.Status(), wrapped.Size())
		})
	}
}
