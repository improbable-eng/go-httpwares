// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

import (
	"time"

	"net/http"

	"net"
	"os"
	"syscall"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

const (
	failedResolution  = "resolution"
	failedConnRefused = "refused"
	failedTimeout     = "timeout"
	failedUnknown     = "unknown"
)

// Tripperware is a client-side http ware for monitoring calls to http services.
//
// Calls are labeled by the http_ctxtags `TagForCallService` http_ctxtags.Tripperware. By default, these are inferred
// from hostnames.
//
// The following monitoring variables can be created if opted in using options:
//
//      http_client_requests_total
//      http_client_response_headers_duration_seconds
// Please note: errors in handled respo
//
// Please note that the instantiation of this Tripperware can panic if it has been previously instantiated with other
// options due to clashes in Prometheus metric names.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	requestHandledCounter := buildClientHandledCounter(o)
	requestErredCounter := buildClientErroredCounter(o)
	responseHeadersHistogram := buildClientResponseHeadersHistogram(o)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			serviceName := serviceNameFromTags(req)
			startTime := time.Now()
			resp, err := next.RoundTrip(req)
			if err != nil {
				failCode, failReason := httpErrorToLabelAndCode(err)
				requestHandledCounter.WithLabelValues(serviceName, sanitizeMethod(req.Method), sanitizeCode(failCode)).Inc()
				requestErredCounter.WithLabelValues(serviceName, sanitizeMethod(req.Method), failReason).Inc()
			} else {
				requestHandledCounter.WithLabelValues(serviceName, sanitizeMethod(req.Method), sanitizeCode(resp.StatusCode)).Inc()
				if responseHeadersHistogram != nil {
					responseHeadersHistogram.WithLabelValues(serviceName, sanitizeMethod(req.Method)).Observe(timeDiffToSeconds(startTime))
				}
			}
			return resp, err
		})
	}
}

func serviceNameFromTags(req *http.Request) string {
	serviceName := "unspecified"
	tags := http_ctxtags.ExtractOutbound(req).Values()
	if s, ok := tags[http_ctxtags.TagForCallService].(string); ok {
		serviceName = s
	}
	return serviceName
}

func buildClientHandledCounter(o *options) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: o.namespace,
			Subsystem: "client",
			Name:      "requests_total",
			Help:      "Total number of requests completed on the server.",
		}, []string{"service_name", "method", "code"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.CounterVec)
	}
	panic("failed registering handled_total error in http_prometheus: %v" + err.Error())
}

func buildClientErroredCounter(o *options) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: o.namespace,
			Subsystem: "client",
			Name:      "request_errors_total",
			Help:      "Total number of requests failed on the client side.",
		}, []string{"service_name", "method", "fail_reason"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.CounterVec)
	}
	panic("failed registering request_errors_total error in http_prometheus: %v" + err.Error())
}

func buildClientResponseHeadersHistogram(o *options) *prometheus.HistogramVec {
	if len(o.responseHeadersHistogramBuckets) == 0 {
		return nil
	}
	cv := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: o.namespace,
			Subsystem: "client",
			Name:      "response_headers_duration_seconds",
			Help:      "Latency (seconds) until HTTP response headers are received by the client.",
			Buckets:   o.responseHeadersHistogramBuckets,
		}, []string{"service_name", "method"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.HistogramVec)
	}
	panic("failed registering response_headers_duration_seconds error in http_prometheus: %v" + err.Error())
}

func httpErrorToLabelAndCode(err error) (int, string) {
	// For list of informal code mappings:
	// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
	if netErr, ok := err.(*net.OpError); ok {
		switch nestErr := netErr.Err.(type) {
		case *net.DNSError:
			return 599, failedResolution
		case *os.SyscallError:
			if nestErr.Err == syscall.ECONNREFUSED {
				return 599, failedConnRefused
			}
			return 599, failedUnknown
		}
		if netErr.Timeout() {
			return 598, failedTimeout
		}
	} else if err == context.Canceled || err == context.DeadlineExceeded {
		return 598, failedTimeout
	}
	return 599, failedUnknown
}
