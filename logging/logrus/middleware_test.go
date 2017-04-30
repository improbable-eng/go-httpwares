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

func customMiddlewareCodeToLevel(statusCode int) logrus.Level {
	if statusCode == testCodeImATeapot {
		// Make this a special case for tests, and an error.
		return logrus.ErrorLevel
	}
	level := http_logrus.DefaultMiddlewareCodeToLevel(statusCode)
	return level
}

func TestLogrusMiddlewareSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	s := &logrusMiddlewareTestSuite{newLogrusBaseTestSuite(t)}
	// In this suite we have all the Middleware, but no Tripperware.
	s.WaresTestSuite.ServerMiddleware = []httpwares.Middleware{
		http_ctxtags.Middleware("my_service"),
		http_logrus.Middleware(
			logrus.NewEntry(s.logrusBaseTestSuite.logger),
			http_logrus.WithLevels(customMiddlewareCodeToLevel),
			http_logrus.WithRequestBodyCapture(requestCaptureDeciderForTest),
			http_logrus.WithResponseBodyCapture(responseCaptureDeciderForTest),
		),
	}
	suite.Run(t, s)
}

type logrusMiddlewareTestSuite struct {
	*logrusBaseTestSuite
}

func (s *logrusMiddlewareTestSuite) TestPing_WithCustomTags() {
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	msgs := s.makeSuccessfulRequestWithAssertions(req, 2, "server")

	// Assert custom tags exist
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"custom_tags.string": "something"`, "all lines must contain `custom_tags.string` set by AddFields")
		assert.Contains(s.T(), m, `"custom_tags.int": 1337`, "all lines must contain `custom_tags.int` set by AddFields")
	}
	assert.Contains(s.T(), msgs[0], `"level": "warning"`, "warningf handler myst be logged as this..")
	assert.Contains(s.T(), msgs[0], `"msg": "handler_log"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "handled"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "~200 status codes must be logged as info by default.")
	assert.Contains(s.T(), msgs[1], `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusMiddlewareTestSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  int
		level logrus.Level
		msg   string
	}{
		{
			code:  http.StatusInternalServerError,
			level: logrus.ErrorLevel,
			msg:   "Internal (500) must remap to ErrorLevel in DefaultMiddlewareCodeToLevel",
		},
		{
			code:  http.StatusNotFound,
			level: logrus.InfoLevel,
			msg:   "NotFound (404) must remap to InfoLevel in DefaultMiddlewareCodeToLevel",
		},
		{
			code:  http.StatusBadRequest,
			level: logrus.WarnLevel,
			msg:   "BadRequest (400) must remap to WarnLevel in DefaultMiddlewareCodeToLevel",
		},
		{
			code:  http.StatusTeapot,
			level: logrus.ErrorLevel,
			msg:   "ImATeapot is overwritten to ErrorLevel with customMiddlewareCodeToLevel override, which probably didn't work",
		},
	} {
		s.SetupTest()
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://something.local/someurl?code=%d", tcase.code), nil)
		msgs := s.makeSuccessfulRequestWithAssertions(req, 2, "server")
		m := msgs[1]
		assert.Contains(s.T(), m, fmt.Sprintf(`"http.status": %d`, tcase.code), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}

func (s *logrusMiddlewareTestSuite) TestCapture_SimpleJSONBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/json", content)
	req.Header.Set("content-type", "Application/JSON")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 4, "server")

	reqMsg, respMsg, finalMsg := msgs[0], msgs[2], msgs[3]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), respMsg, `"http.response.body_json": {`, "response capture should log messages as structued json")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusMiddlewareTestSuite) TestCapture_PlainTextBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`Lorem Ipsum, who cares?`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/plain", content)
	req.Header.Set("content-type", "text/plain")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 3, "server")

	reqMsg, respMsg, finalMsg := msgs[0], msgs[1], msgs[2]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_raw": "`, "request capture should log messages as strings")
	assert.Contains(s.T(), respMsg, `"http.response.body_raw": "`, "response capture should log messages as strings")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusMiddlewareTestSuite) TestCapture_ChunkResponse() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/chunked", content)
	req.Header.Set("content-type", "application/json")
	msgs := s.makeSuccessfulRequestWithAssertions(req, 3, "server")

	respMsg, finalMsg := msgs[1], msgs[2]
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), respMsg, `response body capture skipped, transfer encoding is not identity`, "response capture should log a helpful message")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *logrusMiddlewareTestSuite) TestCapture_StreamFileUp() {
	// Simulate an async upload of a file.
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
	msgs := s.makeSuccessfulRequestWithAssertions(req, 4, "server")

	reqMsg, finalMsg := msgs[0], msgs[3]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `request body capture skipped, content length negative`, "request should log a helpful error")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}
