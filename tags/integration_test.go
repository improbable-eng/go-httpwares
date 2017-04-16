package http_ctxtags_test

import (
	"net/http"
	"testing"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
	"github.com/mwitkow/go-httpwares/testing"
	"github.com/pressly/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type assertingHandler struct {
	*testing.T
	serviceName string
	methodName  string
}

func (a *assertingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	assert.True(a, http_ctxtags.ExtractInbound(req).Has("peer.address"), "ctxtags must have peer.address at least")
	assert.Equal(a, a.serviceName, http_ctxtags.ExtractInbound(req).Values()["http.handler.service"], "ctxtags must have the service name")
	if a.methodName != "" {
		assert.Equal(a, a.methodName, http_ctxtags.ExtractInbound(req).Values()["http.handler.method"], "ctxtags must have the method name set")
	}

	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func TestTaggingSuite(t *testing.T) {
	chiRouter := chi.NewRouter()
	chiRouter.Use(
		http_ctxtags.Middleware(
			http_ctxtags.WithTagExtractor(http_ctxtags.ChiRouteTagExtractor),
			http_ctxtags.WithServiceName("someservice"),
		))
	chiRouter.Mount("/", &assertingHandler{T: t, serviceName: "someservice"})
	// This route will check whether the TagHandler passes the right metadata.
	chiRouter.Mount("/myservice/mymethod",
		http_ctxtags.TagHandler("MyService", "MyMethod", &assertingHandler{
			T: t, serviceName: "MyService", methodName: "MyMethod"}))

	s := &TaggingSuite{
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: chiRouter,
			ClientTripperware: httpwares.TripperwareChain{
				http_ctxtags.Tripperware(http_ctxtags.WithServiceName("someclientservice")),
			},
		},
	}
	suite.Run(t, s)
}

type TaggingSuite struct {
	*httpwares_testing.WaresTestSuite
}

func (s *TaggingSuite) SetupTest() {
}

func (s *TaggingSuite) TestDefaultTagsAreSet() {
	client := s.NewClient()
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
	requestTags := http_ctxtags.ExtractOutbound(resp.Request)
	assert.NotEmpty(s.T(), requestTags, "request leaving client has tags from client tripperware")
	assert.Equal(s.T(), "someclientservice", requestTags.Values()["http.call.service"], "request leaving client has tags from client tripperware")

}

func (s *TaggingSuite) TestCustomTagsAreBeingUsed() {
	client := s.NewClient()
	req, _ := http.NewRequest("GET", "https://something.local/myservice/mymethod", nil)
	req = http_ctxtags.TagRequest(req, "MyService_call", "MyMethod_call")
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
	requestTags := http_ctxtags.ExtractOutbound(resp.Request)
	assert.NotEmpty(s.T(), requestTags, "request leaving client has tags from client tripperware")
	assert.Equal(s.T(), "MyService_call", requestTags.Values()["http.call.service"], "request should have serviceName updated by TagRequest")
	assert.Equal(s.T(), "MyMethod_call", requestTags.Values()["http.call.method"], "request should have methodName updated by TagRequest")
}
