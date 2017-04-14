package http_ctxtags_test

import (
	"net/http"
	"testing"

	"github.com/mwitkow/go-httpwares/tags"
	"github.com/mwitkow/go-httpwares/testing"
	"github.com/pressly/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type assertingHandler struct {
	*testing.T
}

func (a *assertingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	assert.True(a, http_ctxtags.ExtractInbound(req).Has("peer.address"), "ctxtags must have peer.address at least")
	assert.Equal(a, "someservice", http_ctxtags.ExtractInbound(req).Values()["http.service"], "ctxtags must have peer.address at least")

	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func TestTaggingSuite(t *testing.T) {
	chiRouter := chi.NewRouter()
	chiRouter.Use(
		http_ctxtags.Middleware(
			http_ctxtags.WithTagExtractor(http_ctxtags.ChiRouteTagExtractor),
			http_ctxtags.WithServiceName("someservice"),
		))
	chiRouter.Mount("/", &assertingHandler{t})
	s := &TaggingSuite{
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: chiRouter,
		},
	}
	suite.Run(t, s)
}

type TaggingSuite struct {
	*httpwares_testing.WaresTestSuite
}

func (s *TaggingSuite) SetupTest() {
}

func (s *TaggingSuite) TestTagsAreSet() {
	client := s.NewClient()
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
}
