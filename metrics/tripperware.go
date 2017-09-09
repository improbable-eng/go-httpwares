// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"time"

	"github.com/mwitkow/go-httpwares"
)

// Tripperware returns a new client-side ware that exports request metrics.
func Tripperware(reporter Reporter) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		if reporter == nil {
			return next
		}
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			tracker := reporter.Track(req)
			start := time.Now()
			tracker.RequestStarted()
			req.Body = &body{
				parent: req.Body,
				done: func(size int) {
					tracker.RequestRead(time.Since(start), size)
				},
			}

			resp, err := next.RoundTrip(req)
			dur := time.Since(start)
			if err != nil {
				tracker.ResponseStarted(dur, 599, nil)
				tracker.ResponseDone(dur, 599, 0)
				return resp, err
			}
			tracker.ResponseStarted(dur, resp.StatusCode, resp.Header)
			resp.Body = &body{
				parent: resp.Body,
				done: func(size int) {
					tracker.ResponseDone(time.Since(start), size, resp.StatusCode)
				},
			}
			return resp, err
		})
	}
}
