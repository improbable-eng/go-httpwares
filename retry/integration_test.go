// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_retry_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/retry"
	"github.com/improbable-eng/go-httpwares/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

var (
	noSleep         = 0 * time.Second
	retryTimeout    = 50 * time.Millisecond
	failureCode     = http.StatusServiceUnavailable
	expectedContent = "SomeReallyLongString"
)

func requestDeciderForTesting(req *http.Request) bool {
	return req.Method == "GET" || req.Method == "PUT"
}

type failingHandler struct {
	*testing.T
	reqCounter uint
	reqModulo  uint
	reqSleep   time.Duration
	mu         sync.Mutex
}

func (f *failingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if f.maybeFailRequest() {
		resp.WriteHeader(failureCode)
		return
	}
	content, err := ioutil.ReadAll(req.Body)
	assert.NoError(f.T, err, "should have a body")
	assert.EqualValues(f.T, expectedContent, string(content), "the body should be sent")
	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func (f *failingHandler) resetFailingConfiguration(modulo uint, sleepTime time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqCounter = 0
	f.reqModulo = modulo
	f.reqSleep = sleepTime
}

func (f *failingHandler) requestCount() uint {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.reqCounter
}

func (f *failingHandler) maybeFailRequest() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqCounter += 1
	if (f.reqModulo > 0) && (f.reqCounter%f.reqModulo == 0) {
		return false
	}
	time.Sleep(f.reqSleep)
	return true
}

func TestRetryTripperwareSuite(t *testing.T) {
	f := &failingHandler{T: t}
	s := &RetryTripperwareSuite{
		f: f,
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: f,
			ClientTripperware: []httpwares.Tripperware{
				http_retry.Tripperware(
					http_retry.WithMax(5),
					http_retry.WithDecider(requestDeciderForTesting),
					http_retry.WithBackoff(http_retry.BackoffLinear(retryTimeout)),
				),
			},
		},
	}
	suite.Run(t, s)
}

type RetryTripperwareSuite struct {
	*httpwares_testing.WaresTestSuite
	f *failingHandler
}

func (s *RetryTripperwareSuite) SetupTest() {
}

func (s *RetryTripperwareSuite) createRequest(method string, ctx context.Context) *http.Request {
	content := strings.NewReader(expectedContent)
	req, _ := http.NewRequest(method, "https://something.local/someurl", content)
	req = req.WithContext(ctx)
	return req
}

func (s *RetryTripperwareSuite) TestRetryPassesOnEnabled() {
	s.f.resetFailingConfiguration(3, noSleep)
	req := s.createRequest("PUT", s.SimpleCtx())
	resp, err := s.NewClient().Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should succeed")
	require.EqualValues(s.T(), 3, s.f.requestCount(), "3 requests should be retried to meet the modulo")
}

func (s *RetryTripperwareSuite) TestRetryFailsOnMoreThanRetryCount() {
	s.f.resetFailingConfiguration(10, noSleep)
	req := s.createRequest("PUT", s.SimpleCtx())
	resp, err := s.NewClient().Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	assert.Equal(s.T(), failureCode, resp.StatusCode, "failure code should be propagated")
	require.EqualValues(s.T(), 5, s.f.requestCount(), "backend should see all the retry calls")
}

func (s *RetryTripperwareSuite) TestRetryFailsOnNonRetriable() {
	s.f.resetFailingConfiguration(2, noSleep)
	req := s.createRequest("POST", s.SimpleCtx()) // post is not retriable
	resp, err := s.NewClient().Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	assert.Equal(s.T(), failureCode, resp.StatusCode, "failure code should be propagated")
	require.EqualValues(s.T(), 1, s.f.requestCount(), "backend should see only one request")
}

func (s *RetryTripperwareSuite) TestRetryWorksWithAnEnable() {
	s.f.resetFailingConfiguration(3, noSleep)
	req := s.createRequest("POST", s.SimpleCtx()) // post is not retriable
	resp, err := s.NewClient().Do(http_retry.Enable(req))
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should succeed")
	require.EqualValues(s.T(), 3, s.f.requestCount(), "backend should see all the retry calls")
}

func (s *RetryTripperwareSuite) TestTimesoutTheContextAnyway() {
	s.f.resetFailingConfiguration(4, noSleep)
	ctx, _ := context.WithTimeout(s.SimpleCtx(), 2*retryTimeout) // should be enough for 2 calls
	req := s.createRequest("PUT", ctx)                           // PUT is retriable
	_, err := s.NewClient().Do(req)
	require.Error(s.T(), err, "call should fail with a context deadline exceeded")
	require.EqualValues(s.T(), 2, s.f.requestCount(), "backend should see two calls")
}
