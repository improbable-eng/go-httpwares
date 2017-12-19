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
	s.WaresTestSuite.ClientTripperware = []httpwares.Tripperware{
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
	s.T().Log(m)
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
