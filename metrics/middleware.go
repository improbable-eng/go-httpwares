// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"time"

	"github.com/mwitkow/go-httpwares"
)

// Middleware returns a http.Handler middleware that exports request metrics.
func Middleware(reporter Reporter) httpwares.Middleware {
	return func(next http.Handler) http.Handler {
		if reporter == nil {
			return next
		}
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			tracker := reporter.Track(req)
			start := time.Now()
			tracker.RequestStarted()
			req.Body = &body{
				parent: req.Body,
				done: func(size int) {
					tracker.RequestRead(time.Since(start), size)
				},
			}
			wrapped := &writer{
				parent: resp,
				started: func(status int) {
					tracker.ResponseStarted(time.Since(start), status, resp.Header())
				},
			}
			next.ServeHTTP(wrapped, req)
			tracker.ResponseDone(time.Since(start), wrapped.size, wrapped.status)
		})
	}
}
