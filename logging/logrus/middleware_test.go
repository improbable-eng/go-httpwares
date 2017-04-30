// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_logrus_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"mime/multipart"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/logging/logrus"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/mwitkow/go-httpwares/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testCodeImATeapot = http.StatusTeapot
)

type loggingHandler struct {
	*testing.T
}

func (a *loggingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	assert.NotNil(a.T, http_logrus.Extract(req), "handlers must have access to the loggermust have ")
	http_ctxtags.ExtractInbound(req).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	http_logrus.Extract(req).Warningf("handler_log")
	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

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
	b := &bytes.Buffer{}
	log := logrus.New()
	log.Out = b
	log.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	s := &LogrusMiddlewareSuite{
		log:    logrus.NewEntry(log),
		buffer: b,
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: handlerForTestingOfCaptures(t),
			ServerMiddleware: []httpwares.Middleware{
				http_ctxtags.Middleware("my_service"),
				http_logrus.Middleware(
					logrus.NewEntry(log),
					http_logrus.WithLevels(customMiddlewareCodeToLevel),
					http_logrus.WithRequestBodyCapture(requestCaptureDeciderForTest),
					http_logrus.WithResponseBodyCapture(responseCaptureDeciderForTest),
				),
			},
		},
	}
	suite.Run(t, s)
}

type LogrusMiddlewareSuite struct {
	*httpwares_testing.WaresTestSuite
	buffer *bytes.Buffer
	log    *logrus.Entry
}

func (s *LogrusMiddlewareSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *LogrusMiddlewareSuite) TearDownTest() {
	s.buffer.Reset()
}

func (s *LogrusMiddlewareSuite) getOutputJSONs() []string {
	// So the final `handled` method may happen after the client completed the response. So wait here.
	time.Sleep(50 * time.Millisecond)
	ret := []string{}
	dec := json.NewDecoder(s.buffer)
	for {
		var val map[string]json.RawMessage
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from Logrus JSON: %v", err)
		}
		out, _ := json.MarshalIndent(val, "", "  ")
		ret = append(ret, string(out))
	}
	return ret
}

func (s *LogrusMiddlewareSuite) TestPing_WithCustomTags() {
	client := s.NewClient()
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	req = req.WithContext(s.SimpleCtx())
	_, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "server"`, "all lines must contain indicator of being a server call")
		assert.Contains(s.T(), m, `"http.host": "something.local"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"http.url.path": "/someurl"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"custom_tags.string": "something"`, "all lines must contain `custom_tags.string` set by AddFields")
		assert.Contains(s.T(), m, `"custom_tags.int": 1337`, "all lines must contain `custom_tags.int` set by AddFields")
	}
	assert.Contains(s.T(), msgs[0], `"level": "warning"`, "warningf handler myst be logged as this..")
	assert.Contains(s.T(), msgs[0], `"msg": "handler_log"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "handled"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "~200 status codes must be logged as info by default.")
	assert.Contains(s.T(), msgs[1], `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *LogrusMiddlewareSuite) TestPingError_WithCustomLevels() {
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
		s.buffer.Reset()
		client := s.NewClient()
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://something.local/someurl?code=%d", tcase.code), nil)
		req = req.WithContext(s.SimpleCtx())
		_, err := client.Do(req)
		require.NoError(s.T(), err, "call shouldn't fail")
		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 2, "both the handler and the interceptor print messages")
		m := msgs[1]
		assert.Contains(s.T(), m, `"http.host": "something.local"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"http.url.path": "/someurl"`, "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"http.status": %d`, tcase.code), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}

func (s *LogrusMiddlewareSuite) TestCapture_SimpleJSONBothWays() {
	client := s.NewClient() // client always dials localhost.
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/json", content)
	req = req.WithContext(s.SimpleCtx())
	req.Header.Set("content-type", "application/JSON")
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	pingBack, err := httpwares_testing.DecodePingBack(resp)
	require.NoError(s.T(), err, "decoding pingback response must not fail, otherwise we change the behaviour of the client")
	assert.Equal(s.T(), "application/JSON", pingBack.Headers["Content-Type"], "the content must be preserved")

	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 4, "three log statements should be logged: captured req, handler, captured resp, and final one")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "server"`, "all lines must contain indicator of being a server call")
		assert.Contains(s.T(), m, `"http.host": "fakeaddress.fakeaddress.com"`, "all lines must contain http.host from http_ctxtags")
		assert.Contains(s.T(), m, `"http.url.path": "/capture/request/json"`, "all lines must contain method name")
	}
	reqMsg, respMsg, finalMsg := msgs[0], msgs[2], msgs[3]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), respMsg, `"http.response.body_json": {`, "response capture should log messages as structued json")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *LogrusMiddlewareSuite) TestCapture_PlainTextBothWays() {
	client := s.NewClient() // client always dials localhost.
	content := new(bytes.Buffer)
	content.WriteString(`Lorem Ipsum, who cares?`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/plain", content)
	req = req.WithContext(s.SimpleCtx())
	req.Header.Set("content-type", "text/plain")
	_, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 3, "three log statements should be logged: captured req, captured resp, and final one")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "server"`, "all lines must contain indicator of being a server call")
		assert.Contains(s.T(), m, `"http.host": "fakeaddress.fakeaddress.com"`, "all lines must contain http.host from http_ctxtags")
		assert.Contains(s.T(), m, `"http.url.path": "/capture/request/plain"`, "all lines must contain method name")
	}
	reqMsg, respMsg, finalMsg := msgs[0], msgs[1], msgs[2]

	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `"http.request.body_raw": "`, "request capture should log messages as strings")
	assert.Contains(s.T(), respMsg, `"http.response.body_raw": "`, "response capture should log messages as strings")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *LogrusMiddlewareSuite) TestCapture_ChunkResponse() {
	client := s.NewClient() // client always dials localhost.
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/chunked", content)
	req = req.WithContext(s.SimpleCtx())
	req.Header.Set("content-type", "application/JSON")
	_, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 3, "three log statements should be logged: captured req, captured resp, and final one")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "server"`, "all lines must contain indicator of being a server call")
		assert.Contains(s.T(), m, `"http.host": "fakeaddress.fakeaddress.com"`, "all lines must contain http.host from http_ctxtags")
		assert.Contains(s.T(), m, `"http.url.path": "/capture/request/chunked"`, "all lines must contain method name")
	}
	respMsg, finalMsg := msgs[1], msgs[2]
	assert.Contains(s.T(), respMsg, `"level": "info"`, "response captures should be logged as info")
	assert.Contains(s.T(), respMsg, `response body capture skipped, transfer encoding is not identity`, "response capture should log a helpful message")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *LogrusMiddlewareSuite) TestCapture_StreamFileUp() {
	client := s.NewClient() // client always dials localhost.
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
	req = req.WithContext(s.SimpleCtx())
	resp, err := client.Do(req)
	pingBack, err := httpwares_testing.DecodePingBack(resp)
	require.NoError(s.T(), err, "decoding pingback response must not fail, otherwise we change the behaviour of the client")
	assert.Contains(s.T(), pingBack.Headers["Content-Type"], "multipart/form-data", "the content must be preserved")

	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 4, "three log statements should be logged: captured req, captured resp, handler, and final one")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "server"`, "all lines must contain indicator of being a server call")
		assert.Contains(s.T(), m, `"http.host": "fakeaddress.fakeaddress.com"`, "all lines must contain http.host from http_ctxtags")
		assert.Contains(s.T(), m, `"http.url.path": "/capture/request/json"`, "all lines must contain method name")
	}
	reqMsg, finalMsg := msgs[0], msgs[3]
	assert.Contains(s.T(), reqMsg, `"level": "info"`, "request captures should be logged as info")
	assert.Contains(s.T(), reqMsg, `request body capture skipped, content length negative`, "request should log a helpful error")
	assert.Contains(s.T(), finalMsg, `"http.time_ms":`, "interceptor log statement should contain execution time")
}
