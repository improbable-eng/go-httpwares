// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_retry

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
	"context"
)

// Tripperware is client side HTTP ware that retries the requests.
//
// Be default this retries safe and idempotent requests 3 times with a linear delay of 100ms. This behaviour can be
// customized using With* parameter options.
//
// Requests that have `http_retry.Enable` set on them will always be retried.
func Tripperware(opts ...Option) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		o := evaluateOptions(opts)
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Short-circuit to avoid allocations.
			if !o.decider(req) && !isEnabled(req.Context()) {
				return next.RoundTrip(req)
			}
			if o.maxRetry == 0 {
				return next.RoundTrip(req)
			}

			var lastErr error
			for attempt := uint(0); attempt < o.maxRetry; attempt++ {
				if err := waitRetryBackoff(attempt, req.Context(), o); err != nil {
					return nil, err
				}

			}
			startTime := time.Now()
			resp, err := next.RoundTrip(req)
			fields := logrus.Fields{
				"system":        SystemField,
				"span.kind":     "client",
				"http.url.path": req.URL.Path,
				"http.time_ms":  timeDiffToMilliseconds(startTime),
			}
			for k, v := range http_ctxtags.ExtractOutbound(req).Values() {
				fields[k] = v
			}
			level := logrus.DebugLevel
			msg := "request completed"
			if err != nil {
				fields[logrus.ErrorKey] = err
				level = o.levelForConnectivityError
				msg = "request failed to execute, see err"
			} else {
				fields["http.proto_major"] = resp.ProtoMajor
				fields["http.response_bytes"] = resp.ContentLength
				fields["http.status"] = resp.StatusCode

				level = o.levelFunc(resp.StatusCode)
			}
			levelLogf(entry.WithFields(fields), level, msg)
			return resp, err
		})
	}
}

func waitRetryBackoff(attempt uint, parentCtx context.Context, opt *options) error {
	var waitTime time.Duration = 0
	if attempt > 0 {
		waitTime = opt.backoffFunc(attempt)
	}
	if waitTime > 0 {
		select {
		case <-parentCtx.Done():
			return parentCtx.Err()
		case <-time.Tick(waitTime):
		}
	}
	return nil
}
