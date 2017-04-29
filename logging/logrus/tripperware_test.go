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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	b := &bytes.Buffer{}
	log := logrus.New()
	log.Out = b
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	s := &LogrusTripperwareSuite{
		log:    logrus.NewEntry(log),
		buffer: b,
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: &loggingHandler{t},
			ClientTripperware: httpwares.TripperwareChain{
				http_ctxtags.Tripperware(),
				http_logrus.Tripperware(logrus.NewEntry(log), http_logrus.WithLevels(customTripperwareCodeToLevel)),
			},
		},
	}
	suite.Run(t, s)
}

type LogrusTripperwareSuite struct {
	*httpwares_testing.WaresTestSuite
	buffer *bytes.Buffer
	log    *logrus.Entry
}

func (s *LogrusTripperwareSuite) SetupTest() {
	s.buffer.Reset()
}

func (s *LogrusTripperwareSuite) getOutputJSONs() []string {
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

func (s *LogrusTripperwareSuite) TestSuccessfulCall() {
	client := s.NewClient() // client always dials localhost.
	req, _ := http.NewRequest("GET", "https://fakeaddress.fakeaddress.com/someurl", nil)
	req = req.WithContext(s.SimpleCtx())
	_, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, 1, "one log statements should be logged")
	m := msgs[0]
	s.T().Log(m)
	assert.Contains(s.T(), m, `"span.kind": "client"`, "all lines must contain indicator of being a client call")
	assert.Contains(s.T(), m, `"http.host": "fakeaddress.fakeaddress.com"`, "all lines must contain http.host from http_ctxtags")
	assert.Contains(s.T(), m, `"http.url.path": "/someurl"`, "all lines must contain method name")
	assert.Contains(s.T(), m, `"level": "debug"`, "warningf handler myst be logged as this..")
	assert.Contains(s.T(), m, `"msg": "request completed"`, "interceptor message must contain string")
	assert.Contains(s.T(), m, `"http.time_ms":`, "interceptor log statement should contain execution time")
}

func (s *LogrusTripperwareSuite) TestSuccessfulCall_WithRemap() {
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
		s.buffer.Reset()
		client := s.NewClient()
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://something.local/someurl?code=%d", tcase.code), nil)
		req = req.WithContext(s.SimpleCtx())
		_, err := client.Do(req)
		require.NoError(s.T(), err, "call shouldn't fail")
		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 1, "only one message is logged")
		m := msgs[0]
		assert.Contains(s.T(), m, `"http.host": "something.local"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"http.url.path": "/someurl"`, "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"http.status": %d`, tcase.code), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}
