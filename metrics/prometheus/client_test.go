// Copyright 2017 Mark Nevill. All Rights Reserved.
// See LICENSE for licensing terms.

package http_prometheus_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/metrics"
	"github.com/improbable-eng/go-httpwares/metrics/prometheus"
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const clientMetricName = "http_tripper_completed_requests_total"

func findMetric(t *testing.T, name string) *io_prometheus_client.MetricFamily {
	metrics, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	var found *io_prometheus_client.MetricFamily
	var names []string
	for _, m := range metrics {
		if m.Name == nil {
			continue
		}
		names = append(names, *m.Name)
		if *m.Name == name {
			found = m
		}
	}
	assert.Contains(t, names, name)
	require.NotNil(t, found)
	return found
}

func TestPrometheusClientMetricLables(t *testing.T) {
	client := httpwares.WrapClient(
		http.DefaultClient,
		http_ctxtags.Tripperware(http_ctxtags.WithServiceName("testing")),
		func(next http.RoundTripper) http.RoundTripper {
			return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				http_ctxtags.ExtractOutbound(req).Set(http_ctxtags.TagForHandlerName, "testhandler")
				return next.RoundTrip(req)
			})
		},
		http_metrics.Tripperware(http_prometheus.ClientMetrics()),
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)

	metricFamily := findMetric(t, clientMetricName)
	var expectedMetric io_prometheus_client.Metric
	err = jsonpb.Unmarshal(bytes.NewBufferString(`{
	"label": [
		{ "name": "handler", "value": "unknown.testhandler" },
		{ "name": "host", "value": "" },
		{ "name": "method", "value": "GET" },
		{ "name": "name", "value": "testing" },
		{ "name": "path", "value": "" },
		{ "name": "status", "value": "200" }
	],
	"counter": { "value": 1 }
	}`), &expectedMetric)
	require.NoError(t, err)
	require.Equal(t, []*io_prometheus_client.Metric{&expectedMetric}, metricFamily.Metric)
}
