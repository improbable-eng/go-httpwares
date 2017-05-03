package http_prometheus_test

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"fmt"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/monitoring/prometheus"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/mwitkow/go-httpwares/testing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	testHandlerGroup = "testingHandler"
	testHandlerName  = "testingHandler"
	testServiceName  = "testingExternalService"
)

func TestPrometheusSuite(t *testing.T) {
	s := &PrometheusSuite{
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: http_ctxtags.HandlerName(testHandlerName)(httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode)),
			ServerMiddleware: []httpwares.Middleware{
				http_ctxtags.Middleware(testHandlerGroup),
				http_prometheus.Middleware(
					http_prometheus.WithNamespace("http"),
					http_prometheus.WithResponseSizeHistogram(),
					http_prometheus.WithResponseHeadersLatencyHistogram(),
					http_prometheus.WithRequestCompletionLatencyHistogram(),
				),
			},
			ClientTripperware: httpwares.TripperwareChain{
				http_ctxtags.Tripperware(http_ctxtags.WithServiceName(testServiceName)),
			},
		},
	}
	suite.Run(t, s)
}

type PrometheusSuite struct {
	*httpwares_testing.WaresTestSuite
}

func (s *PrometheusSuite) SetupTest() {
}

func (s *PrometheusSuite) makeCall(t *testing.T, method string, expectedCode int) {
	client := s.NewClient() // client always dials localhost.
	req, _ := http.NewRequest(method, fmt.Sprintf("https://fakeaddress.fakeaddress.com/someurl?code=%d", expectedCode), nil)
	req = req.WithContext(s.SimpleCtx())
	_, err := client.Do(req)
	require.NoError(t, err, "call shouldn't fail")
}

func (s *PrometheusSuite) TestHandledCounterCountsValues() {
	for _, tcase := range []struct {
		method string
		code   int
	}{
		{method: "GET", code: 200},
		{method: "POST", code: 201},
		{method: "POST", code: 350},
		{method: "HEAD", code: 403},
	} {
		s.T().Run(fmt.Sprintf("%s_%d", tcase.method, tcase.code), func(t *testing.T) {
			codeStr := strconv.Itoa(tcase.code)
			lowerMethod := strings.ToLower(tcase.method)
			beforeRequestCounter := sumCountersForMetricAndLabels(t, "http_server_requests_total", testHandlerGroup, testHandlerName, codeStr, lowerMethod)
			s.makeCall(t, tcase.method, tcase.code)
			afterRequestCounter := sumCountersForMetricAndLabels(t, "http_server_requests_total", testHandlerGroup, testHandlerName, codeStr, lowerMethod)
			assert.Equal(t, beforeRequestCounter+1, afterRequestCounter, "request counter for this handler should increase")
		})
	}
}

func (s *PrometheusSuite) TestMiddlewareResponseHeaderDurations() {
	beforeBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_response_headers_duration_seconds_count", testHandlerGroup, testHandlerName, "get")
	s.makeCall(s.T(), "GET", 201)
	afterBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_response_headers_duration_seconds_count", testHandlerGroup, testHandlerName, "get")
	assert.Equal(s.T(), beforeBucketCount+1, afterBucketCount, "we should increment at least one bucket")
}

func (s *PrometheusSuite) TestMiddlewareRequestCompleteDuration() {
	beforeBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_request_duration_seconds_count", testHandlerGroup, testHandlerName, "head")
	s.makeCall(s.T(), "HEAD", 201)
	afterBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_request_duration_seconds_count", testHandlerGroup, testHandlerName, "head")
	assert.Equal(s.T(), beforeBucketCount+1, afterBucketCount, "we should increment at least one bucket")
}

func (s *PrometheusSuite) TestMiddlewareResponseSize() {
	beforeBucketSum := sumCountersForMetricAndLabels(s.T(), "http_server_response_size_bytes_sum", testHandlerGroup, testHandlerName, "get")
	beforeBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_response_size_bytes_count", testHandlerGroup, testHandlerName, "get")
	s.makeCall(s.T(), "GET", 201)
	afterBucketSum := sumCountersForMetricAndLabels(s.T(), "http_server_response_size_bytes_sum", testHandlerGroup, testHandlerName, "get")
	afterBucketCount := sumCountersForMetricAndLabels(s.T(), "http_server_response_size_bytes_count", testHandlerGroup, testHandlerName, "get")
	assert.Equal(s.T(), beforeBucketCount+1, afterBucketCount, "we should increment at least one bucket")
	assert.True(s.T(), beforeBucketSum < afterBucketSum, "our sum should have increased by non zero sum of bytes transferred.")
}

func fetchPrometheusLines(t *testing.T, metricName string, matchingLabelValues ...string) []string {
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "failed creating request for Prometheus handler")
	prometheus.Handler().ServeHTTP(resp, req)
	reader := bufio.NewReader(resp.Body)
	ret := []string{}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err, "error reading stuff")
		}
		if !strings.HasPrefix(line, metricName) {
			continue
		}
		matches := true
		for _, labelValue := range matchingLabelValues {
			if !strings.Contains(line, `"`+labelValue+`"`) {
				matches = false
			}
		}
		if matches {
			ret = append(ret, line)
		}

	}
	return ret
}

func sumCountersForMetricAndLabels(t *testing.T, metricName string, matchingLabelValues ...string) int {
	count := 0
	for _, line := range fetchPrometheusLines(t, metricName, matchingLabelValues...) {
		valueString := line[strings.LastIndex(line, " ")+1: len(line)-1]
		valueFloat, err := strconv.ParseFloat(valueString, 32)
		require.NoError(t, err, "failed parsing value for line: %v", line)
		count += int(valueFloat)
	}
	return count
}
