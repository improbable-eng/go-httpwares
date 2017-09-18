// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/reporter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serverStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_handler_started_requests_total",
			Help: "Count of started requests.",
		},
		[]string{"name", "handler", "host", "path", "method"},
	)
	serverCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_handler_completed_requests_total",
			Help: "Count of completed requests.",
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	serverLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_handler_completed_latency_seconds",
			Help:    "Latency of completed requests.",
			Buckets: []float64{.01, .03, .1, .3, 1, 3, 10, 30, 100, 300},
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	serverRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_handler_request_size_bytes",
			Help:    "Size of received requests.",
			Buckets: prometheus.ExponentialBuckets(32, 32, 6),
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	serverResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_handler_response_size_bytes",
			Help:    "Size of sent responses.",
			Buckets: prometheus.ExponentialBuckets(32, 32, 6),
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)

	serverInit     sync.Once
	serverHistInit sync.Once
	serverSizeInit sync.Once
)

func serverMetrics(opts ...opt) http_reporter.Reporter {
	o := evalOpts(opts)
	serverInit.Do(func() {
		prometheus.MustRegister(serverStarted)
		prometheus.MustRegister(serverCompleted)
	})
	if o.latency {
		serverHistInit.Do(func() {
			prometheus.MustRegister(serverLatency)
		})
	}
	if o.sizes {
		serverSizeInit.Do(func() {
			prometheus.MustRegister(serverRequestSize)
			prometheus.MustRegister(serverResponseSize)
		})
	}
	return &serverReporter{opts: o}
}

// Middleware returns a http.Handler middleware that exports request metrics.
// It is using Reporter Middleware with Prometheus Server Metrics Reporter.
// If the tags middleware is used, this should be placed after tags to pick up metadata.
// This middleware assumes HTTP/1.x-style requests/response behaviour. It will not work with servers that use
// hijacking, pushing, or other similar features.
func Middleware(opts ...opt) httpwares.Middleware {
	return http_reporter.Middleware(serverMetrics(opts...))
}

type serverReporter struct {
	opts *options
}

func (r *serverReporter) Track(req *http.Request) http_reporter.Tracker {
	return &serverTracker{
		opts: r.opts,
		meta: reqMeta(req, r.opts, true),
	}
}

type serverTracker struct {
	opts *options
	*meta
}

func (t *serverTracker) RequestStarted() {
	serverStarted.WithLabelValues(t.name, t.handler, t.host, t.path, t.method).Inc()
}

func (t *serverTracker) RequestRead(duration time.Duration, size int) {
	if t.opts.sizes {
		serverRequestSize.WithLabelValues(t.name, t.handler, t.host, t.path, t.method).Observe(float64(size))
	}
}

func (t *serverTracker) ResponseStarted(duration time.Duration, code int, header http.Header) {
}

func (t *serverTracker) ResponseDone(duration time.Duration, code int, size int) {
	status := strconv.Itoa(code)
	serverCompleted.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, status).Inc()
	if t.opts.latency {
		serverLatency.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, status).Observe(duration.Seconds())
	}
	if t.opts.sizes {
		serverResponseSize.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, strconv.Itoa(code)).Observe(float64(size))
	}
}
