package http_logrus_test

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/logging/logrus"
	"github.com/sirupsen/logrus"
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
		http_logrus.Middleware(
			logrus.NewEntry(s.logrusBaseTestSuite.logger).WithField("http.handler.group", "my_service"),
			http_logrus.WithDecider(func(w httpwares.WrappedResponseWriter, r *http.Request) bool {
				return r.URL.Path != "/blah"
			}),
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

func (s *logrusMiddlewareTestSuite) TestPing_WithNoEndLogging() {
	req, _ := http.NewRequest("GET", "https://something.local/blah", nil)
	msgs := s.makeSuccessfulRequestWithAssertions(req, 1, "server")

	assert.Contains(s.T(), msgs[0], `"custom_tags.string": "something"`, "all lines must contain `custom_tags.string` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"custom_tags.int": 1337`, "all lines must contain `custom_tags.int` set by AddFields")
	assert.Contains(s.T(), msgs[0], `"level": "warning"`, "warningf handler myst be logged as this..")
	assert.Contains(s.T(), msgs[0], `"msg": "handler_log"`, "handler's message must contain user message")
}
