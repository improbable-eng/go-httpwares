// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_opentracing_test

import (
	"testing"

	"fmt"
	"net/http"

	"strings"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/improbable-eng/go-httpwares/testing"
	"github.com/improbable-eng/go-httpwares/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

var (
	fakeInboundTraceId = 1337
	fakeInboundSpanId  = 999
)

type assertingHandler struct {
	*testing.T
}

func (a *assertingHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	assert.NotNil(a.T, opentracing.SpanFromContext(req.Context()), "handlers must have the spancontext in their context, otherwise propagation will fail")
	tags := http_ctxtags.ExtractInbound(req)
	assert.True(a.T, tags.Has("trace.traceid"), "handlers should see traceid in tags")
	assert.True(a.T, tags.Has("trace.spanid"), "handlers should see traceid in tags")
	httpwares_testing.PingBackHandler(httpwares_testing.DefaultPingBackStatusCode).ServeHTTP(resp, req)
}

func TestTaggingSuite(t *testing.T) {
	mockTracer := mocktracer.New()
	s := &OpentracingSuite{
		WaresTestSuite: &httpwares_testing.WaresTestSuite{
			Handler: http_ctxtags.HandlerName("assert_method")(&assertingHandler{t}),
			ServerMiddleware: []httpwares.Middleware{
				http_ctxtags.Middleware("assert_service"),
				http_opentracing.Middleware(http_opentracing.WithTracer(mockTracer)),
			},
			ClientTripperware: []httpwares.Tripperware{
				http_ctxtags.Tripperware(http_ctxtags.WithServiceName("assert_service")),
				http_opentracing.Tripperware(http_opentracing.WithTracer(mockTracer)),
			},
		},
		mockTracer: mockTracer,
	}
	suite.Run(t, s)
}

type OpentracingSuite struct {
	*httpwares_testing.WaresTestSuite
	mockTracer *mocktracer.MockTracer
}

func (s *OpentracingSuite) SetupTest() {
	s.mockTracer.Reset()
}

func (s *OpentracingSuite) createContextFromFakeHttpRequestParent(ctx context.Context) context.Context {
	hdr := http.Header{}
	hdr.Set("mockpfx-ids-traceid", fmt.Sprint(fakeInboundTraceId))
	hdr.Set("mockpfx-ids-spanid", fmt.Sprint(fakeInboundSpanId))
	hdr.Set("mockpfx-ids-sampled", fmt.Sprint(true))
	parentSpanContext, err := s.mockTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(hdr))
	require.NoError(s.T(), err, "parsing a fake HTTP request headers shouldn't fail, ever")
	fakeSpan := s.mockTracer.StartSpan(
		"/fake/parent/http/request",
		// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		opentracing.ChildOf(parentSpanContext),
	)
	fakeSpan.Finish()
	return opentracing.ContextWithSpan(ctx, fakeSpan)
}

func (s *OpentracingSuite) assertTracesCreated(groupOrServiceName string) (clientSpan *mocktracer.MockSpan, serverSpan *mocktracer.MockSpan) {
	spans := s.mockTracer.FinishedSpans()
	for _, span := range spans {
		s.T().Logf("span: %v, tags: %v", span, span.Tags())
	}
	require.Len(s.T(), spans, 3, "should record 3 spans: one fake inbound, one client, one server")
	traceIdAssert := fmt.Sprintf("traceId=%d", fakeInboundTraceId)
	for _, span := range spans {
		assert.Contains(s.T(), span.String(), traceIdAssert, "not part of the fake parent trace: %v", span)
		if strings.HasPrefix(span.OperationName, groupOrServiceName) {
			kind := fmt.Sprintf("%v", span.Tag("span.kind"))
			if kind == "client" {
				clientSpan = span
			} else if kind == "server" {
				serverSpan = span
			}
			assert.EqualValues(s.T(), span.Tag("component"), "http", "span must be tagged with http component")
		}
	}
	require.NotNil(s.T(), clientSpan, "client span must be there")
	require.NotNil(s.T(), serverSpan, "server span must be there")
	return clientSpan, serverSpan
}

func (s *OpentracingSuite) TestPropagatesTraces() {
	client := s.NewClient()
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx())
	req, _ := http.NewRequest("GET", "https://something.local/someurl", nil)
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), httpwares_testing.DefaultPingBackStatusCode, resp.StatusCode, "response should have the same type")
	clientSpan, serverSpan := s.assertTracesCreated("assert_service")
	assert.Equal(s.T(), "GET", clientSpan.Tag("http.method"), "client span needs the correct method marking")
	assert.Equal(s.T(), "GET", serverSpan.Tag("http.method"), "server span needs the correct method marking")
	assert.EqualValues(s.T(), httpwares_testing.DefaultPingBackStatusCode, clientSpan.Tag("http.status_code"), "client span needs the correct status code marking")
	assert.EqualValues(s.T(), httpwares_testing.DefaultPingBackStatusCode, serverSpan.Tag("http.status_code"), "server span needs the correct status code marking")
}

func (s *OpentracingSuite) TestPropagatesErrors() {
	client := s.NewClient()
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx())
	req, _ := http.NewRequest("POST", "https://something.local/someurl?code=501", nil)
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "call shouldn't fail")
	require.Equal(s.T(), 501, resp.StatusCode, "response should have the same type")
	clientSpan, serverSpan := s.assertTracesCreated("assert_service")
	assert.Equal(s.T(), "POST", clientSpan.Tag("http.method"), "client span needs the correct method marking")
	assert.Equal(s.T(), "POST", serverSpan.Tag("http.method"), "server span needs the correct method marking")
	assert.EqualValues(s.T(), 501, clientSpan.Tag("http.status_code"), "client span needs the correct status code marking")
	assert.EqualValues(s.T(), 501, serverSpan.Tag("http.status_code"), "server span needs the correct status code marking")
}

func (s *OpentracingSuite) TestTripperwareHandlesErrors() {
	client := httpwares.WrapClient(http.DefaultClient, s.WaresTestSuite.ClientTripperware...)
	ctx := s.createContextFromFakeHttpRequestParent(s.SimpleCtx())
	req, _ := http.NewRequest("POST", "https://whatever.doesntexist/someurl?code=501", nil)
	req = req.WithContext(ctx)
	_, err := client.Do(req)
	require.Error(s.T(), err, "call should fail with resolution error")
	assert.Len(s.T(), s.mockTracer.FinishedSpans(), 2, "we should record two spans: fake one and client one")
}
