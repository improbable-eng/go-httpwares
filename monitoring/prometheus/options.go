// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus

import "github.com/prometheus/client_golang/prometheus"

var (
	// DefaultResponseSizeHistogram are the default (if enabled) buckets for response size histograms.
	DefaultResponseSizeHistogram = []float64{32.0, 256.0, 2048.0, 16384.0, 131072.0, 1048576}
	// DefaultLatencyHistogram defines the default (if enabled) buckets for latency histograms.
	DefaultLatencyHistogram = prometheus.DefBuckets
	disabledHistogram       = []float64{}

	defaultOptions = &options{
		namespace:                       "http",
		sizeHistogramBuckets:            disabledHistogram,
		responseHeadersHistogramBuckets: disabledHistogram,
		requestHistogramBuckets:         disabledHistogram,
	}
)

type options struct {
	namespace string
	registry  prometheus.Registerer

	sizeHistogramBuckets            []float64
	responseHeadersHistogramBuckets []float64
	requestHistogramBuckets         []float64
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.registry = prometheus.DefaultRegisterer
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// WithNamespace customizes the Prometheus namespace (first component before first underscore) of all the metrics.
func WithNamespace(prometheusNamespace string) Option {
	return func(o *options) {
		o.namespace = prometheusNamespace
	}
}

// WithResponseSizeHistogram enables the middleware to record the sizes of response messages in bytes.
//
// Optionally, you can provide your own histogram buckets for the measurements. If not provided DefaultResponseSizeHistogram is used.
func WithResponseSizeHistogram(bucketValues ...float64) Option {
	return func(o *options) {
		if len(bucketValues) == 0 {
			o.sizeHistogramBuckets = DefaultResponseSizeHistogram
		} else {
			o.sizeHistogramBuckets = bucketValues
		}
	}
}

// WithResponseHeadersLatencyHistogram enables the middleware to record the latency to headers response in seconds.
//
// Optionally, you can provide your own histogram buckets for the measurements. If not provided DefaultLatencyHistogram is used.
func WithResponseHeadersLatencyHistogram(bucketValues ...float64) Option {
	return func(o *options) {
		if len(bucketValues) == 0 {
			o.responseHeadersHistogramBuckets = DefaultResponseSizeHistogram
		} else {
			o.responseHeadersHistogramBuckets = bucketValues
		}
	}
}

// WithRequestCompletionLatencyHistogram enables the middleware to record the latency to completion of request.
//
// Optionally, you can provide your own histogram buckets for the measurements. If not provided DefaultLatencyHistogram is used.
func WithRequestCompletionLatencyHistogram(bucketValues ...float64) Option {
	return func(o *options) {
		if len(bucketValues) == 0 {
			o.requestHistogramBuckets = DefaultResponseSizeHistogram
		} else {
			o.requestHistogramBuckets = bucketValues
		}
	}
}
