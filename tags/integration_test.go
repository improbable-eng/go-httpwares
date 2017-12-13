package http_ctxtags_test

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi"
	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/tags/chi"
	"github.com/improbable-eng/go-httpwares/testing"
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
	assert.Equal(a, a.serviceName, http_ctxtags.ExtractInbound(req).Values()["http.handler.group"], "ctxtags must have the service name")
	if a.methodName != "" {
		assert.Equal(a, a.methodName, http_ctxtags.ExtractInbound(req).Values()["http.handler.name"], "ctxtags must have the method name set")
	}

	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func TestTaggingSuite(t *testing.T) {
	chiRouter := chi.NewRouter()
	chiRouter.Use(
		http_ctxtags.Middleware(
			"someservice",
			http_ctxtags.WithTagExtractor(http_chitags.ChiRouteTagExtractor),
		))
	chiRouter.Mount("/", &assertingHandler{T: t, serviceName: "someservice"})
	// This route will check whether the HandlerName passes the right metadata.
	chiMyService := chi.NewRouter()
	chiMyService.Use(append(chiRouter.Middlewares(), http_ctxtags.Middleware("MyService"))...)
	chiMyService.Handle("/myservice/mymethod",
		http_ctxtags.HandlerName("MyMethod")(&assertingHandler{
			T: t, serviceName: "MyService", methodName: "MyMethod"}))

	s := &TaggingSuite{
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: chiRouter,
			ClientTripperware: []httpwares.Tripperware{
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
	client := httpwares.WrapClient(s.NewClient(), http_ctxtags.Tripperware(http_ctxtags.WithServiceName("MyServiceCapitalised")))
	req, _ := http.NewRequest("GET", "https://something.local/myservice/mymethod", nil)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
	requestTags := http_ctxtags.ExtractOutbound(resp.Request)
	assert.NotEmpty(s.T(), requestTags, "request leaving client has tags from client tripperware")
	assert.Equal(s.T(), "MyServiceCapitalised", requestTags.Values()["http.call.service"], "request should have serviceName updated by TagRequest")
}
