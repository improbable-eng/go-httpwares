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

	"github.com/Sirupsen/logrus"
	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/logging/logrus"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/mwitkow/go-httpwares/testing"
	"github.com/opentracing/opentracing-go/mocktracer"
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
	httpwares_ctxtags.Extract(req).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	http_logrus.Extract(req).Warningf("handler_log")
	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

type OpentracingSuite struct {
	*httpwares_testing.WaresTestSuite
	mockTracer *mocktracer.MockTracer
}

func (s *OpentracingSuite) SetupTest() {
	s.mockTracer.Reset()
}

func customCodeToLevel(statusCode int) logrus.Level {
	if statusCode == testCodeImATeapot {
		// Make this a special case for tests, and an error.
		return logrus.ErrorLevel
	}
	level := http_logrus.DefaultCodeToLevel(statusCode)
	return level
}

func TestLogrusLoggingSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	b := &bytes.Buffer{}
	log := logrus.New()
	log.Out = b
	log.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	s := &LogrusLoggingSuite{
		log:    logrus.NewEntry(log),
		buffer: b,
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: &loggingHandler{t},
			ServerMiddleware: []httpwares.Middleware{
				httpwares_ctxtags.Middleware(),
				http_logrus.Middleware(logrus.NewEntry(log), http_logrus.WithLevels(customCodeToLevel)),
			},
		},
	}
	suite.Run(t, s)
}

type LogrusLoggingSuite struct {
	*httpwares_testing.WaresTestSuite
	buffer *bytes.Buffer
	log    *logrus.Entry
}

func (s *LogrusLoggingSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *LogrusLoggingSuite) getOutputJSONs() []string {
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

func (s *LogrusLoggingSuite) TestPing_WithCustomTags() {
	client := s.NewClient()
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	req = req.WithContext(s.SimpleCtx())
	_, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
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

func (s *LogrusLoggingSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  int
		level logrus.Level
		msg   string
	}{
		{
			code:  http.StatusInternalServerError,
			level: logrus.ErrorLevel,
			msg:   "Internal (500) must remap to ErrorLevel in DefaultCodeToLevel",
		},
		{
			code:  http.StatusNotFound,
			level: logrus.InfoLevel,
			msg:   "NotFound (404) must remap to InfoLevel in DefaultCodeToLevel",
		},
		{
			code:  http.StatusBadRequest,
			level: logrus.WarnLevel,
			msg:   "BadRequest (400) must remap to WarnLevel in DefaultCodeToLevel",
		},
		{
			code:  http.StatusTeapot,
			level: logrus.ErrorLevel,
			msg:   "ImATeapot is overwritten to ErrorLevel with customCodeToLevel override, which probably didn't work",
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
