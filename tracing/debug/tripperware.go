// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package http_debug

import (
	"fmt"
	"net/http"

	"github.com/mwitkow/go-httpwares"
	"github.com/mwitkow/go-httpwares/tags"
	"golang.org/x/net/trace"
)

// Tripperware returns a piece of client-side Tripperware that puts requests on the `/debug/requests` page.
//
// The data logged will be: request headers, request ctxtags, response headers and response length.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if o.filterFunc != nil && !o.filterFunc(req) {
				return next.RoundTrip(req)

			}
			tr := trace.New(operationNameFromUrl(req), req.URL.String())
			defer tr.Finish()
			tr.LazyPrintf("%v %v HTTP/%d.%d", req.Method, req.URL, req.ProtoMajor, req.ProtoMinor)
			tr.LazyPrintf("Host: %v", hostFromReq(req))
			for k, _ := range req.Header {
				tr.LazyPrintf("%v: %v", k, req.Header.Get(k))
			}
			tr.LazyPrintf("invoking next chain")
			resp, err := next.RoundTrip(req)
			tr.LazyPrintf("tags: ")
			for k, v := range http_ctxtags.ExtractInbound(req).Values() {
				tr.LazyPrintf("%v: %v", k, v)
			}
			if err != nil {
				tr.LazyPrintf("Error on response: %v", err)
				tr.SetError()
			} else {
				tr.LazyPrintf("Response: %d, length: %d", resp.Status, resp.ContentLength)
				for k, _ := range resp.Header {
					tr.LazyPrintf("%v: %v", k, resp.Header.Get(k))
				}
				if o.statusCodeErrorFunc(resp.StatusCode) {
					tr.SetError()
				}
			}
			return resp, err
		})
	}
}

func operationNameFromUrl(req *http.Request) string {
	if tags := http_ctxtags.ExtractOutbound(req); tags.Has(http_ctxtags.TagForCallService) {
		vals := tags.Values()
		return fmt.Sprintf("%v.%s", vals[http_ctxtags.TagForCallService], req.Method)
	}
	return fmt.Sprintf("%s%s", req.URL.Host, req.URL.Path)
}
