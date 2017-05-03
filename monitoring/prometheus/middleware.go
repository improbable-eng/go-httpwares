// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

import (
	"time"

	"net/http"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/prometheus/client_golang/prometheus"
)

// Middleware is a server-side http ware for monitoring handlers using Prometheus counters and histograms.
//
// Handlers are labeled by the http_ctxtags `TagForHandlerGroup` and `TagForHandlerName` applied using the http_ctxtags
// Middleware and HandlerName methods. These values are used as labels for all requests.
//
// The following monitoring variables can be created if opted in using options:
//
//      http_server_requests_total
//      http_server_response_size_bytes
//      http_server_response_headers_duration_seconds
//      http_server_request_duration_seconds
//
//
// Please note that the instantiation of this Middleware can panic if it has been previously instantiated with other
// options due to clashes in Prometheus metric names.
func Middleware(opts ...Option) httpwares.Middleware {
	return func(nextHandler http.Handler) http.Handler {
		o := evaluateOptions(opts)
		requestHandledCounter := buildServerHandledCounter(o)
		responseSizeHistogram := buildServerResponseSizeHistogram(o)
		responseHeadersHistogram := buildServerResponseHeadersHistogram(o)
		requestHistogram := buildServerRequestCompletionHistogram(o)
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			handlerGroup, handlerName := handlerInfoFromRequest(req)
			startTime := time.Now()
			wrappedResp := httpwares.WrapResponseWriter(resp)
			if responseHeadersHistogram != nil {
				wrappedResp.ObserveWriteHeader(func(writer httpwares.WrappedResponseWriter, code int) {
					responseHeadersHistogram.WithLabelValues(handlerGroup, handlerName, sanitizeMethod(req.Method)).Observe(timeDiffToSeconds(startTime))
				})
			}
			nextHandler.ServeHTTP(wrappedResp, req)

			requestHandledCounter.WithLabelValues(handlerGroup, handlerName, sanitizeMethod(req.Method), sanitizeCode(wrappedResp.StatusCode())).Inc()
			if requestHistogram != nil {
				requestHistogram.WithLabelValues(handlerGroup, handlerName, sanitizeMethod(req.Method)).Observe(timeDiffToSeconds(startTime))
			}
			if responseSizeHistogram != nil {
				responseSizeHistogram.WithLabelValues(handlerGroup, handlerName, sanitizeMethod(req.Method)).Observe(float64(wrappedResp.MessageLength()))
			}
		})
	}
}

func handlerInfoFromRequest(req *http.Request) (handlerGroup string, handlerName string) {
	handlerGroup = "unspecified"
	handlerName = "unspecified"
	tags := http_ctxtags.ExtractInbound(req).Values()
	if g, ok := tags[http_ctxtags.TagForHandlerGroup].(string); ok {
		handlerGroup = g
	}
	if n, ok := tags[http_ctxtags.TagForHandlerName].(string); ok {
		handlerName = n
	}
	return handlerGroup, handlerGroup
}

func buildServerHandledCounter(o *options) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: o.namespace,
			Subsystem: "server",
			Name:      "requests_total",
			Help:      "Total number of requests completed on the server.",
		}, []string{"handler_group", "handler_name", "method", "code"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.CounterVec)
	}
	panic("failed registering handled_total error in http_prometheus: %v" + err.Error())
}

func buildServerResponseSizeHistogram(o *options) *prometheus.HistogramVec {
	if len(o.sizeHistogramBuckets) == 0 {
		return nil
	}
	cv := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: o.namespace,
			Subsystem: "server",
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes (optional).",
			Buckets:   o.sizeHistogramBuckets,
		}, []string{"handler_group", "handler_name", "method"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.HistogramVec)
	}
	panic("failed registering response_size_bytes error in http_prometheus: %v" + err.Error())
}

func buildServerResponseHeadersHistogram(o *options) *prometheus.HistogramVec {
	if len(o.responseHeadersHistogramBuckets) == 0 {
		return nil
	}
	cv := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: o.namespace,
			Subsystem: "server",
			Name:      "response_headers_duration_seconds",
			Help:      "Latency (seconds) until HTTP response headers are returned (optional).",
			Buckets:   o.responseHeadersHistogramBuckets,
		}, []string{"handler_group", "handler_name", "method"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.HistogramVec)
	}
	panic("failed registering response_headers_duration_seconds error in http_prometheus: %v" + err.Error())
}

func buildServerRequestCompletionHistogram(o *options) *prometheus.HistogramVec {
	if len(o.requestHistogramBuckets) == 0 {
		return nil
	}
	cv := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: o.namespace,
			Subsystem: "server",
			Name:      "request_duration_seconds",
			Help:      "Latency (seconds) until HTTP request is fully handled by the server (optional).",
			Buckets:   o.requestHistogramBuckets,
		}, []string{"handler_group", "handler_name", "method"})
	err := o.registry.Register(cv)
	if err == nil {
		return cv
	} else if aeErr, ok := err.(*prometheus.AlreadyRegisteredError); ok {
		return aeErr.ExistingCollector.(*prometheus.HistogramVec)
	}
	panic("failed registering request_duration_seconds error in http_prometheus: %v" + err.Error())
}

func timeDiffToSeconds(start time.Time) float64 {
	d := time.Now().Sub(start).Seconds()
	if d < 0.0 {
		return 0.0
	}
	return d
}
