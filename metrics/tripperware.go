// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_metrics

import (
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/improbable-eng/go-httpwares"
)

// Tripperware returns a new client-side ware that exports request metrics.
// If the tags tripperware is used, this should be placed after tags to pick up metadata.
func Tripperware(reporter Reporter) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		if reporter == nil {
			return next
		}
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			tracker := reporter.Track(req)
			start := time.Now()
			tracker.RequestStarted()

			// If present, wrap body to track number of bytes written
			reqSize := 0
			if req.Body != nil {
				req.Body = wrapBody(req.Body, func(size int) {
					reqSize = size
				})
			}

			// Use httptrace to get notified when writing request completed
			trace := httptrace.ContextClientTrace(req.Context())
			if trace == nil {
				trace = &httptrace.ClientTrace{}
			}
			prevWroteRequest := trace.WroteRequest
			trace.WroteRequest = func(info httptrace.WroteRequestInfo) {
				tracker.RequestRead(time.Since(start), reqSize)
				if prevWroteRequest != nil {
					prevWroteRequest(info)
				}
			}
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

			resp, err := next.RoundTrip(req)
			dur := time.Since(start)
			if err != nil {
				tracker.ResponseStarted(dur, 599, nil)
				tracker.ResponseDone(dur, 599, 0)
				return resp, err
			}
			tracker.ResponseStarted(dur, resp.StatusCode, resp.Header)
			resp.Body = wrapBody(resp.Body, func(size int) {
				tracker.ResponseDone(time.Since(start), resp.StatusCode, size)
			})
			return resp, err
		})
	}
}
