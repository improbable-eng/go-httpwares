// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
)

// Tripperware is a server-side http ware for logging using logrus.
//
// This tripperware *does not* propagate a context-based logger, but act as a logger of requests.
// This includes logging of errors.
func Tripperware(entry *logrus.Entry, opts ...Option) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		o := evaluateTripperwareOpts(opts)
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
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
