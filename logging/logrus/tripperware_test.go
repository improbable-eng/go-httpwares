// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"mime/multipart"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/logging/logrus"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func customTripperwareCodeToLevel(statusCode int) logrus.Level {
	if statusCode == testCodeImATeapot {
		// Make this a special case for tests, and an error.
		return logrus.ErrorLevel
	}
	level := http_logrus.DefaultTripperwareCodeToLevel(statusCode)
	return level
}

func TestLogrusTripperwareSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	s := &logrusTripperwareSuite{newLogrusBaseTestSuite(t)}
	s.logrusBaseTestSuite.logger.Level = logrus.DebugLevel // most of our log statements are on debug level.
	// In this suite we have all the Tripperware, but no Middleware.
	s.WaresTestSuite.ClientTripperware = httpwares.TripperwareChain{
		http_ctxtags.Tripperware(),
		http_logrus.Tripperware(
			logrus.NewEntry(s.logrusBaseTestSuite.logger),
			http_logrus.WithLevels(customTripperwareCodeToLevel),
			http_logrus.WithRequestBodyCapture(requestCaptureDeciderForTest),
			http_logrus.WithResponseBodyCapture(responseCaptureDeciderForTest),
		),
	}
	suite.Run(t, s)
}

type logrusTripperwareSuite struct {
	*logrusBaseTestSuite
}

func (s *logrusTripperwareSuite) TestSuccessfulCall() {
	req, _ := http.NewRequest("GET", "https://fakeaddress.fakeaddress.com/someurl", nil)
	msgs := s.makeSuccessfulRequestWithAssertions(req, 1, "client")
	m := msgs[0]
	assert.Contains(s.T(), m, `"level": "debug"`, "handlers by default log on debug")
	assert.Contains(s.T(), m, `"msg": "request completed"`, "interceptor message must contain string")
	assert.Contains(s.T(), m, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusTripperwareSuite) TestSuccessfulCall_WithRemap() {
	for _, tcase := range []struct {
		code  int
		level logrus.Level
		msg   string
	}{
		{
			code:  http.StatusInternalServerError,
			level: logrus.WarnLevel,
			msg:   "Internal (500) must remap to WarnLevel in DefaultTripperwareCodeLevels",
		},
		{
			code:  http.StatusNotFound,
			level: logrus.InfoLevel,
			msg:   "NotFound (404) must remap to InfoLevel in DefaultTripperwareCodeLevels",
		},
		{
			code:  http.StatusBadRequest,
			level: logrus.InfoLevel,
			msg:   "BadRequest (400) must remap to InfoLevel in DefaultTripperwareCodeLevels",
		},
		{
			code:  http.StatusTeapot,
			level: logrus.ErrorLevel,
			msg:   "ImATeapot is overwritten to ErrorLevel with customMiddlewareCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://something.local/someurl?code=%d", tcase.code), nil)
		msgs := s.makeSuccessfulRequestWithAssertions(req, 1, "client")
		m := msgs[0]
		assert.Contains(s.T(), m, `"http.host": "something.local"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"http.url.path": "/someurl"`, "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"http.status": %d`, tcase.code), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}

func (s *logrusTripperwareSuite) TestCapture_SimpleJSONBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/json", content)
	req.Header.Set("content-type", "application/JSON")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 3, "client")

	reqMsg, respMsg, finalMsg := msgs[0], msgs[1], msgs[2]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), respMsg, `"http.response.body_json": {`, "response capture should log messages as structued json")
	assert.Contains(s.T(), respMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusTripperwareSuite) TestCapture_PlainTextBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`Lorem Ipsum, who cares?`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/plain", content)
	req.Header.Set("content-type", "text/plain")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 3, "client")

	reqMsg, respMsg, finalMsg := msgs[0], msgs[1], msgs[2]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_raw": "`, "request capture should log messages as strings")
	assert.Contains(s.T(), respMsg, `"http.response.body_raw": "`, "response capture should log messages as strings")
	assert.Contains(s.T(), respMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusTripperwareSuite) TestCapture_StreamFileUp() {
	// Make a request that simulates a file upload.
	reader, writer := io.Pipe()
	multipartContent := multipart.NewWriter(writer)
	go func() {
		mimeWriter, _ := multipartContent.CreateFormFile("somefield", "filename.txt")
		for i := 0; i < 10; i++ {
			mimeWriter.Write([]byte("something\n"))
		}
		multipartContent.Close()
		writer.Close()
	}()

	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/json", reader)
	req.Header.Set("content-type", multipartContent.FormDataContentType())
	msgs := s.makeSuccessfulRequestWithAssertions(req, 4, "client")

	reqMsg, finalMsg := msgs[0], msgs[2]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `request body capture skipped, missing GetBody method while Body set`, "request should log a helpful error")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusTripperwareSuite) TestCapture_ChunkResponse() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/chunked", content)
	req.Header.Set("content-type", "application/JSON")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 3, "client")
	respMsg, finalMsg := msgs[1], msgs[2]
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), respMsg, `response body capture skipped, content length negative`, "response capture should log a helpful message")
	assert.Contains(s.T(), respMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusTripperwareSuite) assertCommonLogLinesForRequest(req *http.Request) {
}
