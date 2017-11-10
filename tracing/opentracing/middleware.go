// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_opentracing

import (
	"log"
	"net/http"

	"fmt"

	"github.com/improbable-eng/go-httpwares"
	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	httpTag = opentracing.Tag{string(ext.Component), "http"}
)

// Middleware returns a http.Handler middleware values for request tags.
func Middleware(opts ...Option) httpwares.Middleware {
	o := evaluateOptions(opts)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if o.filterOutFunc != nil && !o.filterOutFunc(req) {
				next.ServeHTTP(resp, req)
				return
			}
			tags := http_ctxtags.ExtractInbound(req)
			newReq, serverSpan := newServerSpanFromInbound(req, o.tracer)
			hackyInjectOpentracingIdsToTags(serverSpan, tags)
			newResp := httpwares.WrapResponseWriter(resp)
			next.ServeHTTP(newResp, newReq)

			// The other middleware could have changed the tags, so only update the tags here.
			for k, v := range tags.Values() {
				serverSpan.SetTag(k, v)
			}
			serverSpan.SetOperationName(operationNameFromReqHandler(req))
			ext.HTTPStatusCode.Set(serverSpan, uint16(newResp.StatusCode()))
			if o.statusCodeErrorFunc(newResp.StatusCode()) {
				ext.Error.Set(serverSpan, true)
			}
			serverSpan.Finish()
		})
	}
}

func newServerSpanFromInbound(req *http.Request, tracer opentracing.Tracer) (*http.Request, opentracing.Span) {
	parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		logServerErr(req, "http_opentracing: failed parsing trace information: %v", err)
	}

	serverSpan := tracer.StartSpan(
		"placeholder_name",
		// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
		ext.RPCServerOption(parentSpanContext),
		httpTag,
	)

	ext.HTTPMethod.Set(serverSpan, req.Method)
	ext.HTTPUrl.Set(serverSpan, req.URL.String())
	newReq := req.WithContext(opentracing.ContextWithSpan(req.Context(), serverSpan))
	return newReq, serverSpan
}

func operationNameFromReqHandler(req *http.Request) string {
	if tags := http_ctxtags.ExtractInbound(req); tags.Has(http_ctxtags.TagForHandlerGroup) {
		vals := tags.Values()
		method := "unknown"
		if val, ok := vals[http_ctxtags.TagForHandlerName].(string); ok {
			method = val
		}
		return fmt.Sprintf("%v:%s", vals[http_ctxtags.TagForHandlerGroup], method)
	}
	if req.URL.Host != "" {
		return fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)
	} else {
		return fmt.Sprintf("%s%s", req.Host, req.URL.Path)
	}
}

func logServerErr(req *http.Request, fmt string, args ...interface{}) {
	if v, ok := req.Context().Value(http.ServerContextKey).(*http.Server); ok {
		if logger := v.ErrorLog; logger != nil {
			logger.Printf(fmt, args)
		} else {
			log.Printf(fmt, args)
		}
	} else {
		log.Printf(fmt, args)
	}
}
