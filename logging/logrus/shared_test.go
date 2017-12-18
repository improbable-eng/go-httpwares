package http_logrus_test

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"

	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/tags/logrus"
	"github.com/improbable-eng/go-httpwares/testing"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCodeImATeapot = http.StatusTeapot
)

func requestCaptureDeciderForTest(req *http.Request) bool {
	return strings.HasPrefix(req.URL.Path, "/capture/request/")
}

func responseCaptureDeciderForTest(req *http.Request, code int) bool {
	return strings.HasPrefix(req.URL.Path, "/capture/request/")
}

type loggingHandler struct {
	*testing.T
}

func (a *loggingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	assert.NotNil(a.T, ctx_logrus.Extract(req), "handlers must have access to the loggermust have ")
	http_ctxtags.ExtractInbound(req).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)
	ctx_logrus.Extract(req).Warningf("handler_log")
	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func handlerForTestingOfCaptures(t *testing.T) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/capture/request/chunked", handlerChunked())
	m.HandleFunc("/capture/request/plain", handlerPlainText())
	m.Handle("/", &loggingHandler{t})
	return m
}

func handlerPlainText() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("content-type", "text/html")
		resp.WriteHeader(200)
		resp.Write([]byte(`<body><head>Nothing</head></body>`))
	}
}

func handlerChunked() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("content-type", "text/plain")
		resp.Header().Set("Transfer-Encoding", "chunked")
		resp.WriteHeader(200)
		chunkedWriter := httputil.NewChunkedWriter(resp)
		for i := 0; i < 100; i++ {
			chunkedWriter.Write([]byte("value"))
			resp.(http.Flusher).Flush()
		}
		chunkedWriter.Close()
	}
}

type logrusBaseTestSuite struct {
	*httpwares_testing.WaresTestSuite
	buffer         *bytes.Buffer
	threadedBuffer *httpwares_testing.MutexReadWriter
	logger         *logrus.Logger
}

func newLogrusBaseTestSuite(t *testing.T) *logrusBaseTestSuite {
	b := &bytes.Buffer{}
	threadedBuffer := httpwares_testing.NewMutexReadWriter(b)
	logger := logrus.New()
	logger.Out = threadedBuffer
	logger.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	s := &logrusBaseTestSuite{
		logger:         logger,
		buffer:         b,
		threadedBuffer: threadedBuffer,
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: handlerForTestingOfCaptures(t),
		},
	}
	return s
}

func (s *logrusBaseTestSuite) SetupTest() {
	// Always delete all entries between tests.
	s.threadedBuffer.Mutex.Lock()
	s.buffer.Reset()
	s.threadedBuffer.Mutex.Unlock()
}

func (s *logrusBaseTestSuite) getOutputJSONs() []string {
	// So the final `handled` method may happen after the client completed the response. So wait here.
	time.Sleep(15 * time.Millisecond)
	ret := []string{}
	dec := json.NewDecoder(s.threadedBuffer)
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

func (s *logrusBaseTestSuite) makeSuccessfulRequestWithAssertions(req *http.Request, expectedLogMessages int, expectedKind string) []string {
	client := s.NewClient()
	newReq := req.WithContext(s.SimpleCtx())
	_, err := client.Do(newReq)
	require.NoError(s.T(), err, "call shouldn't fail")
	msgs := s.getOutputJSONs()
	require.Len(s.T(), msgs, expectedLogMessages, "this call should result in a different number of log statments")
	for _, m := range msgs {
		assert.Contains(s.T(), m, `"span.kind": "`+expectedKind+`"`, "all lines must contain indicator of being the right kind call")
		assert.Contains(s.T(), m, `"http.host": "`+req.URL.Host+`"`, "all lines must contain http.host from http_ctxtags")
		assert.Contains(s.T(), m, `"http.url.path": "`+req.URL.Path+`"`, "all lines must contain method name")
	}
	return msgs
}
