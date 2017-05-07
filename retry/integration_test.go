// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_retry_test

import (
	"testing"

	"net/http"

	"strings"

	"sync"
	"time"

	"io/ioutil"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/retry"
	"github.com/mwitkow/go-httpwares/testing"
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
			ClientTripperware: httpwares.TripperwareChain{
				http_retry.Tripperware(http_retry.WithMax(5)),
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

func (s *RetryTripperwareSuite) createRequest(ctx context.Context) *http.Request {
	content := strings.NewReader(expectedContent)
	req, _ := http.NewRequest("GET", "https://something.local/someurl", content)
	req = req.WithContext(ctx)
	return req
}

func (s *RetryTripperwareSuite) TestRetryPasses() {
	s.f.resetFailingConfiguration(3, noSleep)
	req := s.createRequest(s.SimpleCtx())
	resp, err := s.NewClient().Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
	require.EqualValues(s.T(), 3, s.f.requestCount(), "3 requests should be retried to meet the modulo")
}
