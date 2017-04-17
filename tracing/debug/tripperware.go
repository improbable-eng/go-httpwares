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

// Tripperware returns a piece of client-side Tripperware that puts requests on a status page.
func Tripperware(opts ...Option) httpwares.Tripperware {
	o := evaluateOptions(opts)
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if o.filterOutFunc != nil && !o.filterOutFunc(req) {
				return next.RoundTrip(req)

			}
			tr := trace.New(operationNameFromUrl(req), req.URL.String())
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
			tr.Finish()
			return resp, err
		})
	}
}

func operationNameFromUrl(req *http.Request) string {
	if tags := http_ctxtags.ExtractOutbound(req); tags.Has(http_ctxtags.TagForCallService) {
		vals := tags.Values()
		method := "unknown"
		if val, ok := vals[http_ctxtags.TagForCallMethod].(string); ok {
			method = val
		}
		return fmt.Sprintf("http.Sent.%v.%s", vals[http_ctxtags.TagForCallService], method)
	}
	return fmt.Sprintf("http.Sent.%v.%v", req.URL.Host, req.URL.Path)
}
