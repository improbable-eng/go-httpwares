// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/improbable-eng/go-httpwares/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clientStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_tripper_started_requests_total",
			Help: "Count of started requests.",
		},
		[]string{"name", "handler", "host", "path", "method"},
	)
	clientCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_tripper_completed_requests_total",
			Help: "Count of completed requests.",
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	clientLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_tripper_completed_latency_seconds",
			Help:    "Latency of completed requests.",
			Buckets: []float64{.01, .03, .1, .3, 1, 3, 10, 30, 100, 300},
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	clientRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_tripper_request_size_bytes",
			Help:    "Size of sent requests.",
			Buckets: prometheus.ExponentialBuckets(32, 32, 6),
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)
	clientResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_tripper_response_size_bytes",
			Help:    "Size of received responses.",
			Buckets: prometheus.ExponentialBuckets(32, 32, 6),
		},
		[]string{"name", "handler", "host", "path", "method", "status"},
	)

	clientInit     sync.Once
	clientHistInit sync.Once
	clientSizeInit sync.Once
)

func ClientMetrics(opts ...opt) http_metrics.Reporter {
	o := evalOpts(opts)
	clientInit.Do(func() {
		prometheus.MustRegister(clientStarted)
		prometheus.MustRegister(clientCompleted)
	})
	if o.latency {
		clientHistInit.Do(func() {
			prometheus.MustRegister(clientLatency)
		})
	}
	if o.sizes {
		clientSizeInit.Do(func() {
			prometheus.MustRegister(clientRequestSize)
			prometheus.MustRegister(clientResponseSize)
		})
	}
	return &clientReporter{opts: o}
}

type clientReporter struct {
	opts *options
}

func (r *clientReporter) Track(req *http.Request) http_metrics.Tracker {
	return &clientTracker{
		opts: r.opts,
		meta: reqMeta(req, r.opts, false),
	}
}

type clientTracker struct {
	opts *options
	*meta
}

func (t *clientTracker) RequestStarted() {
	clientStarted.WithLabelValues(t.name, t.handler, t.host, t.path, t.method).Inc()
}

func (t *clientTracker) RequestRead(duration time.Duration, size int) {
	if t.opts.sizes {
		clientRequestSize.WithLabelValues(t.name, t.handler, t.host, t.path, t.method).Observe(float64(size))
	}
}

func (t *clientTracker) ResponseStarted(duration time.Duration, code int, header http.Header) {
	status := strconv.Itoa(code)
	clientCompleted.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, status).Inc()
	if t.opts.latency {
		clientLatency.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, status).Observe(duration.Seconds())
	}
}

func (t *clientTracker) ResponseDone(duration time.Duration, code int, size int) {
	if t.opts.sizes {
		clientResponseSize.WithLabelValues(t.name, t.handler, t.host, t.path, t.method, strconv.Itoa(code)).Observe(float64(size))
	}
}
