package http_logrus_test

import (
	"bytes"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"io/ioutil"

	"io"
	"mime/multipart"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/logging/logrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	nullLogger = &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.PanicLevel,
	}
)

func TestLogrusContentCaptureSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	alwaysOnDecider := func(req *http.Request) bool { return true }

	s := &logrusContentCaptureSuite{newLogrusBaseTestSuite(t)}
	s.logrusBaseTestSuite.logger.Level = logrus.DebugLevel // most of our log statements are on debug level.
	// In this suite we have all the Tripperware, but no Middleware.
	s.WaresTestSuite.ServerMiddleware = []httpwares.Middleware{
		http_logrus.Middleware(logrus.NewEntry(nullLogger).WithField("http.handler.group", "somegroup")),
		http_logrus.ContentCaptureMiddleware(logrus.NewEntry(s.logger), alwaysOnDecider),
	}
	s.WaresTestSuite.ClientTripperware = []httpwares.Tripperware{
		http_logrus.ContentCaptureTripperware(logrus.NewEntry(s.logger), alwaysOnDecider),
	}
	suite.Run(t, s)
}

type logrusContentCaptureSuite struct {
	*logrusBaseTestSuite
}

func (s *logrusContentCaptureSuite) getServerAndClientLogs(req *http.Request, expectedServer int, expectedClient int) (server []string, client []string) {
	c := s.NewClient()
	newReq := req.WithContext(s.SimpleCtx())
	_, err := c.Do(newReq)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, expectedClient+expectedServer, "this call should result in a different number of log statments")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"http.host": "`+req.URL.Host+`"`, "all lines must contain http.host")
		assert.Contains(s.T(), m, `"http.url.path": "`+req.URL.Path+`"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"level": "info"`, "body captures captures should be logged as info")
		if strings.Contains(m, `"span.kind": "server"`) {
			server = append(server, m)
		} else if strings.Contains(m, `"span.kind": "client"`) {
			client = append(client, m)
		} else {
			assert.Fail(s.T(), "message %v has no span kind", m)
		}
	}
	return server, client
}

func (s *logrusContentCaptureSuite) TestCapture_SimpleJSONBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/json", content)
	req.Header.Set("content-type", "application/JSON")
	serverMsgs, clientMsgs := s.getServerAndClientLogs(req, 2, 2)
	assert.Contains(s.T(), clientMsgs[0], `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), serverMsgs[0], `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), clientMsgs[1], `"http.response.body_json": {`, "response capture should log messages as structued json")
	assert.Contains(s.T(), serverMsgs[1], `"http.response.body_json": {`, "response capture should log messages as structued json")
}

func (s *logrusContentCaptureSuite) TestCapture_PlainTextBothWays() {
	content := new(bytes.Buffer)
	content.WriteString(`Lorem Ipsum, who cares?`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/plain", content)
	req.Header.Set("content-type", "text/plain")
	serverMsgs, clientMsgs := s.getServerAndClientLogs(req, 2, 2)
	assert.Contains(s.T(), clientMsgs[0], `"http.request.body_raw": "`, "request capture should log messages as strings")
	assert.Contains(s.T(), serverMsgs[0], `"http.request.body_raw": "`, "request capture should log messages as strings")
	assert.Contains(s.T(), clientMsgs[1], `"http.response.body_raw": "`, "response capture should log messages as strings")
	assert.Contains(s.T(), serverMsgs[1], `"http.response.body_raw": "`, "response capture should log messages as strings")
}

func (s *logrusContentCaptureSuite) TestCapture_StreamFileUp() {
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
	serverMsgs, clientMsgs := s.getServerAndClientLogs(req, 2, 2)
	assert.Contains(s.T(), clientMsgs[0], `request body capture skipped, missing GetBody method while Body set`, "request should log a helpful error")
	assert.Contains(s.T(), serverMsgs[0], `request body capture skipped, content length negative`, "request should log a helpful error")
	assert.Contains(s.T(), clientMsgs[1], `"http.response.body_json": {`, "response capture should log pingback response as JSON")
	assert.Contains(s.T(), serverMsgs[1], `"http.response.body_json": {`, "response capture should log pingback response as JSON")
}

func (s *logrusContentCaptureSuite) TestCapture_ChunkResponse() {
	content := new(bytes.Buffer)
	content.WriteString(`{"somekey": "some_value", "someint": 4}`)
	req, _ := http.NewRequest("POST", "https://fakeaddress.fakeaddress.com/capture/request/chunked", content)
	req.Header.Set("content-type", "application/JSON")
	serverMsgs, clientMsgs := s.getServerAndClientLogs(req, 2, 2)
	assert.Contains(s.T(), clientMsgs[0], `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), serverMsgs[0], `"http.request.body_json": {`, "request capture should log messages as structued json")
	assert.Contains(s.T(), clientMsgs[1], `"response body capture skipped, content length negative"`, "client side should log a helpful message")
	assert.Contains(s.T(), serverMsgs[1], `"response body capture skipped, transfer encoding is not identity"`, "server side should log a helpful message about skipped body")
}
